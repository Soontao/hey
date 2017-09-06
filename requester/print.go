// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requester

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

const (
	barChar = "∎"
)

type Report struct {
	AvgTotal float64
	Fastest  float64
	Slowest  float64
	Average  float64
	RPS      float64

	AvgConn   float64
	AvgDNS    float64
	AvgReq    float64
	AvgRes    float64
	AvgDelay  float64
	ConnLats  []float64
	DNSLats   []float64
	ReqLats   []float64
	ResLats   []float64
	DelayLats []float64

	Results chan *result
	Total   time.Duration

	errorDist      map[string]int
	statusCodeDist map[int]int
	lats           []float64
	sizeTotal      int64

	output string

	w io.Writer
}

func newReport(w io.Writer, size int, results chan *result, output string, total time.Duration) *Report {
	return &Report{
		output:         output,
		Results:        results,
		Total:          total,
		statusCodeDist: make(map[int]int),
		errorDist:      make(map[string]int),
		w:              w,
	}
}

func (r *Report) finalize() {
	for res := range r.Results {
		if res.err != nil {
			r.errorDist[res.err.Error()]++
		} else {
			r.lats = append(r.lats, res.duration.Seconds())
			r.AvgTotal += res.duration.Seconds()
			r.AvgConn += res.connDuration.Seconds()
			r.AvgDelay += res.delayDuration.Seconds()
			r.AvgDNS += res.dnsDuration.Seconds()
			r.AvgReq += res.reqDuration.Seconds()
			r.AvgRes += res.resDuration.Seconds()
			r.ConnLats = append(r.ConnLats, res.connDuration.Seconds())
			r.DNSLats = append(r.DNSLats, res.dnsDuration.Seconds())
			r.ReqLats = append(r.ReqLats, res.reqDuration.Seconds())
			r.DelayLats = append(r.DelayLats, res.delayDuration.Seconds())
			r.ResLats = append(r.ResLats, res.resDuration.Seconds())
			r.statusCodeDist[res.statusCode]++
			if res.contentLength > 0 {
				r.sizeTotal += res.contentLength
			}
		}
	}
	r.RPS = float64(len(r.lats)) / r.Total.Seconds()
	r.Average = r.AvgTotal / float64(len(r.lats))
	r.AvgConn = r.AvgConn / float64(len(r.lats))
	r.AvgDelay = r.AvgDelay / float64(len(r.lats))
	r.AvgDNS = r.AvgDNS / float64(len(r.lats))
	r.AvgReq = r.AvgReq / float64(len(r.lats))
	r.AvgRes = r.AvgRes / float64(len(r.lats))
	//r.print()
}

func (r *Report) printCSV() {
	r.printf("response-time,DNS+dialup,DNS,Request-write,Response-delay,Response-read\n")
	for i, val := range r.lats {
		r.printf("%4.4f,%4.4f,%4.4f,%4.4f,%4.4f,%4.4f\n",
			val, r.ConnLats[i], r.DNSLats[i], r.ReqLats[i], r.DelayLats[i], r.ResLats[i])
	}
}

func (r *Report) print() {
	if r.output == "csv" {
		r.printCSV()
		return
	}

	if len(r.lats) > 0 {
		sort.Float64s(r.lats)
		r.Fastest = r.lats[0]
		r.Slowest = r.lats[len(r.lats)-1]
		r.printf("Summary:\n")
		r.printf("  Total:\t%4.4f secs\n", r.Total.Seconds())
		r.printf("  Slowest:\t%4.4f secs\n", r.Slowest)
		r.printf("  Fastest:\t%4.4f secs\n", r.Fastest)
		r.printf("  Average:\t%4.4f secs\n", r.Average)
		r.printf("  Requests/sec:\t%4.4f\n", r.RPS)
		if r.sizeTotal > 0 {
			r.printf("  Total data:\t%d bytes\n", r.sizeTotal)
			r.printf("  Size/request:\t%d bytes\n", r.sizeTotal/int64(len(r.lats)))
		}
		r.printHistogram()
		r.printLatencies()
		r.printf("\nDetails (Average, Fastest, Slowest):")
		r.printSection("DNS+dialup", r.AvgConn, r.ConnLats)
		r.printSection("DNS-lookup", r.AvgDNS, r.DNSLats)
		r.printSection("req write", r.AvgReq, r.ReqLats)
		r.printSection("resp wait", r.AvgDelay, r.DelayLats)
		r.printSection("resp read", r.AvgRes, r.ResLats)
		r.printStatusCodes()
	}
	if len(r.errorDist) > 0 {
		r.printErrors()
	}
	r.printf("\n")
}

// printSection prints details for http-trace fields
func (r *Report) printSection(tag string, avg float64, lats []float64) {
	sort.Float64s(lats)
	fastest, slowest := lats[0], lats[len(lats)-1]
	r.printf("\n  %s:\t", tag)
	r.printf(" %4.4f secs, %4.4f secs, %4.4f secs", avg, fastest, slowest)
}

// printLatencies prints percentile latencies.
func (r *Report) printLatencies() {
	pctls := []int{10, 25, 50, 75, 90, 95, 99}
	data := make([]float64, len(pctls))
	j := 0
	for i := 0; i < len(r.lats) && j < len(pctls); i++ {
		current := i * 100 / len(r.lats)
		if current >= pctls[j] {
			data[j] = r.lats[i]
			j++
		}
	}
	r.printf("\nLatency distribution:\n")
	for i := 0; i < len(pctls); i++ {
		if data[i] > 0 {
			r.printf("  %v%% in %4.4f secs\n", pctls[i], data[i])
		}
	}
}

func (r *Report) printHistogram() {
	bc := 10
	buckets := make([]float64, bc+1)
	counts := make([]int, bc+1)
	bs := (r.Slowest - r.Fastest) / float64(bc)
	for i := 0; i < bc; i++ {
		buckets[i] = r.Fastest + bs*float64(i)
	}
	buckets[bc] = r.Slowest
	var bi int
	var max int
	for i := 0; i < len(r.lats); {
		if r.lats[i] <= buckets[bi] {
			i++
			counts[bi]++
			if max < counts[bi] {
				max = counts[bi]
			}
		} else if bi < len(buckets)-1 {
			bi++
		}
	}
	r.printf("\nResponse time histogram:\n")
	for i := 0; i < len(buckets); i++ {
		// Normalize bar lengths.
		var barLen int
		if max > 0 {
			barLen = (counts[i]*40 + max/2) / max
		}
		r.printf("  %4.3f [%v]\t|%v\n", buckets[i], counts[i], strings.Repeat(barChar, barLen))
	}
}

// printStatusCodes prints status code distribution.
func (r *Report) printStatusCodes() {
	r.printf("\n\nStatus code distribution:\n")
	for code, num := range r.statusCodeDist {
		r.printf("  [%d]\t%d responses\n", code, num)
	}
}

func (r *Report) printErrors() {
	r.printf("\nError distribution:\n")
	for err, num := range r.errorDist {
		r.printf("  [%d]\t%s\n", num, err)
	}
}

func (r *Report) printf(s string, v ...interface{}) {
	fmt.Fprintf(r.w, s, v...)
}

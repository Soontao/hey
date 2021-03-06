# Hey Fork

[![Build Status](https://ci.fornever.org/job/hey/badge/icon)](https://ci.fornever.org/job/hey/)

a project like the Apache Bench, but written by golang

> I rewrite the logic, now it will run series tests, concurrent config will be disabled

PLEASE NOTE THAT: 

requests number and concurrently will be controlled by program, you could set the max concurrency by `-c`

## Installation

```bash
go get -u github.com/Soontao/hey
```

or [download binary](https://download.fornever.org/hey/latest/)

## Usage

just set the protocol and port (http)

```bash
hey localhost:8080
```

out:

```bash
concurrency:    50, RPS:  8058, req time avg:    5.30ms
concurrency:   100, RPS: 10862, req time avg:    8.28ms
concurrency:   200, RPS: 10687, req time avg:   15.70ms
concurrency:   500, RPS: 10344, req time avg:   44.89ms
concurrency:  1000, RPS: 10420, req time avg:   86.42ms
```

set max concurrency with `-c` param

hey runs provided number of requests in the provided concurrency level and prints stats.

It also supports HTTP2 endpoints.

```bash
Usage: hey [options...] <url>

Options:
  -n  Number of requests to run. Default is 200.
  -c  Number of requests to run concurrently. Total number of requests cannot
      be smaller than the concurrency level. Default is 50.
  -q  Rate limit, in seconds (QPS).
  -o  Output type. If none provided, a summary is printed.
      "csv" is the only supported alternative. Dumps the response
      metrics in comma-separated values format.

  -m  HTTP method, one of GET, POST, PUT, DELETE, HEAD, OPTIONS.
  -H  Custom HTTP header. You can specify as many as needed by repeating the flag.
      For example, -H "Accept: text/html" -H "Content-Type: application/xml" .
  -t  Timeout for each request in seconds. Default is 20, use 0 for infinite.
  -A  HTTP Accept header.
  -d  HTTP request body.
  -D  HTTP request body from file. For example, /home/user/file.txt or ./file.txt.
  -T  Content-type, defaults to "text/html".
  -a  Basic authentication, username:password.
  -x  HTTP Proxy address as host:port.
  -h2 Enable HTTP/2.

  -host	HTTP Host header.

  -disable-compression  Disable compression.
  -disable-keepalive    Disable keep-alive, prevents re-use of TCP
                        connections between different HTTP requests.
  -cpus                 Number of used cpu cores.
                        (default for current machine is 8 cores)
  -more                 Provides information on DNS lookup, dialup, request and
                        response timings.
```

Note: Requires go 1.7 or greater.

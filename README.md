# gopherlyzer

## Beta Release

It's the first beta implementation of our analysis described in [Static Trace-Based Deadlock Analysis for
Synchronous Mini-Go](http://www.home.hs-karlsruhe.de/~suma0002/publications/TraceBasedDeadlockAnalysisMiniGo.pdf).

The current prototype only covers a subset of the Go language. See our examples.
We plan to add support for full recursion, anonymous functions, labels, close for channels etc. in the near future.

## Description

We consider the problem of static deadlock detection for
programs in the Go programming language which make use of synchronous
channel communications. In our analysis, regular expressions extended
with a fork operator capture the communication behavior of a program.
Starting from a simple criterion that characterizes traces of deadlock-free
programs, we develop automata-based methods to check for deadlock-
freedom. The approach is implemented and evaluated with a series of
examples.

## How to use

gopherlyzer is a Go module and needs no external tools installed. It requires Go 1.26 or newer.

Running gopherlyzer:

go run . -p benchmarks/philo/main.go

Or build it first:

go build -o gopherlyzer .
./gopherlyzer -p benchmarks/philo/main.go

## Channel analysis backend

Resolving a channel operation to the channel it uses (needed to pair sends with
receives) was originally done by shelling out to the `peers` query of
`golang.org/x/tools/cmd/guru`. That query was removed upstream (see
[Go issue #59676](https://github.com/golang/go/issues/59676)), and the last
release that still implemented it no longer builds with a current Go
toolchain.

On 2026-10-25 guru was replaced by an in-process resolver in `oracleAPI`, built
on `go/packages` and `go/types`. For a send, receive or `close` it reports the
`make(chan ...)` site the channel originates from, following variables,
parameters through their call sites and struct fields within the analysed
package. Channels that arrive from another package, through a function value or
an interface are not tracked and simply yield no allocation site.

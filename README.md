# gopherlyzer

## Beta Release

It's the first beta implementation of our analysis described in [Static Trace-Based Deadlock Analysis for
Synchronous Mini-Go](http://www.home.hs-karlsruhe.de/~suma0002/publications/TraceBasedDeadlockAnalysisMiniGo.pdf).

* The current prototyp can't create a correct forkable expression for everything that can be written in Go. The parsing component needs to be enhanced so 
it can handle more code constructs, for example anonymous functions or breaks that use labels.

* The results from the prototype can be wrong in some cases because the described system in our publication covers 
more than what is currently implemented and there are still some implementation errors. 

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

Configure the Go environment:

export GOROOT=/usr/local/go

export GOPATH=`*path_to_workspace*`/bin

export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

Install Go Oracle:

go get golang.org/x/tools/cmd/oracle

Running gopherlyzer:

go run main.go -p benchmarks/philo/main.go 

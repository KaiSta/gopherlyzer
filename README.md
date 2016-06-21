# gopherlyzer

## Beta Release

It's the first beta implementation of our analysis described in *Static Trace-Based Deadlock Analysis for
Synchronous Mini-Go*. There are still bugs we intend to fix.

## Description

We consider the problem of static deadlock detection for
programs in the Go programming language which make use of synchronous
channel communications. In our analysis, regular expressions extended
with a fork operator capture the communication behavior of a program.
Starting from a simple criterion that characterizes traces of deadlock-free
programs, we develop automata-based methods to check for deadlock-
freedom. The approach is implemented and evaluated with a series of
examples.
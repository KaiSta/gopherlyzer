package main

import (
	"flag"
	"fmt"

	"./simplifier"
	//"./types"
	"time"

	"./FEBuilder"
	"./validation"
)

func main() {

	path := flag.String("p", "", "path to code file")
	flag.Parse()

	start := time.Now()
	s := simplify.New()
	parseGraph := s.Parse(*path)
	fmt.Println(parseGraph["main"].AllOps)
	//fmt.Println("Parse time:", time.Since(start))
	//start = time.Now()
	_, pmap, threads, cmap := febuilder.BuildExpression(parseGraph)
	fmt.Println("CloserMap", cmap)
	// fmt.Println("Expression Build time:", time.Since(start))
	fmt.Println("Threads:", threads)
	//  fmt.Println(r)
	fmt.Println("PartnerMap", pmap)
	// start = time.Now()
	validation.Run2(threads, pmap, cmap)
	// fmt.Println("Validation time:", time.Since(start))

	fmt.Println("Complete:", time.Since(start))
}

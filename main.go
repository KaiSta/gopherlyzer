package main

import (
    "./simplifier"
    "fmt"
    "flag"
    //"./types"
    "./validation"
    "./FEBuilder"
    "time"
)

func main() {
    
    path := flag.String("p", "", "path to code file")
    flag.Parse()
    
    start := time.Now()
    s := simplify.New()
    parseGraph := s.Parse(*path)
    //fmt.Println("Parse time:", time.Since(start))
    //start = time.Now()
    _, pmap, threads := febuilder.BuildExpression(parseGraph)
   // fmt.Println("Expression Build time:", time.Since(start))
    fmt.Println("Threads:",threads)
  //  fmt.Println(r)
   fmt.Println(pmap)
   // start = time.Now()
    validation.Run2(threads, pmap)
   // fmt.Println("Validation time:", time.Since(start))
    
    fmt.Println("Complete:", time.Since(start))  
}
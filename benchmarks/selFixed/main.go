package main

import (
    "fmt"
)

func provide1(x chan bool) {
    x <- true
}
func provide2(y chan bool) {
    y <- false
}

func main() {
    x := make(chan bool)
    y := make(chan bool)
    go provide1(x)
    go provide2(y)
    
    
    select {
        case z := <-x:
        fmt.Println(z)
        case z := <-y:
        fmt.Println(z)
    }
}
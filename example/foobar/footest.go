package main

func foo(x, y chan int) {
    <-x
    y <- 42
}

func main() {
    x := make(chan int)
    y := make(chan int)
    
    go foo(x,y)
    
    if true {
        x <- 42
    } else {
        <-y
    }
    
    for {
        x <- 42
    }
}
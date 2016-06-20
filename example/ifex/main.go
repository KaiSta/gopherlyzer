package main

func A(x,y chan int) {
    for {
        x <- 42
        <-y
    }
}

func main() {
    x := make(chan int)
    y := make(chan int)
    
    go A(x,y)
    
    for {
        <-x
        if true {
            y <- 42
        }
    }
}
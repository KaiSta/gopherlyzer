package main

func A(x,y chan int) {
    for {
        x <- 4
        <-y
    }
}

func B(x,y chan int) {
    for {
        k := <-x
        if k > 5 {
            y <- 4
        }
    }
}

func main() {
    a := make(chan int)
    b := make(chan int)
    
    go A(a,b)
    
    B(a,b)
}
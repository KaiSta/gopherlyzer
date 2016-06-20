package main

func A(x, z chan int) {
    x <- 4
    <-z
    x <- 5
}

func B(x, y chan int) {
    <-x
    <-y
}

func C(x,y,z chan int) {
    <-x
    C1(z)
    y <-3
}

func C1(z chan int) {
    z <- 4
}

func main() {
    x := make(chan int)
    y := make(chan int)
    z := make(chan int)
    
    go A(x,z)
    go B(x,y)
    C(x,y,z)
}

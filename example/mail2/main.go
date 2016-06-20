package main


func sel(x, y chan bool) {
    z := make(chan bool)
    go func() { z<- (<-x) }()
    go func() { z<- (<-y) }()
    <-z
}


func main() {
    x := make(chan bool)
    y := make(chan bool)
    go func() { x <- true }()
    go func() { y <- false }()
    sel(x,y)
    sel(x,y)
}
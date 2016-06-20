package main

func provide1(x chan bool) {
    x <- true
}
func provide2(y chan bool) {
    y <- false
}

func collect1(in, out chan bool) {
    out <- <-in
}
func collect2(in, out chan bool) {
    out <- <-in
}

func main() {
    x := make(chan bool)
    y := make(chan bool)
    go provide1(x)
    go provide2(y)
    
    
    z1 := make(chan bool)
    go collect1(x,z1)
    go collect2(y,z1)
    <-z1
    
    z2 := make(chan bool)
    go collect1(x,z2)
    go collect2(y,z2)
    <-z2 
}
package main

func a(x,y chan int) {
  for {
    x <- 42
    <-y
  }
}

func main() {
  x := make(chan int)
  y := make(chan int)
  go a(x,y)
  <-x
  <-x
  <-x
  y <- 42
}

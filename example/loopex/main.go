package main

func loop(x,y chan int) {
  for i := 0; i < 3; i++ {
    x <- 42
  }
  <-y
}

func main() {
  x := make(chan int)
  y := make(chan int)
  go loop(x,y)

  <-x
  <-x
  <-x
  y <- 42
}

package main

func a(x,y chan int) {
  y <- 42
  <-x
}

func b(x,y  chan int) {
  <-y
  x <- 42
}

func c(x chan int) {
  x <- 42
}

func d(x chan int) {
  <-x
}

func main() {
  x := make(chan int)
  y := make(chan int)

  go a(x,y)
  go b(x,y)
  go c(x)
  go d(x)

  x <- 42
  <-x
}

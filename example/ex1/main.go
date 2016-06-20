package main

func foo(x chan int) {
  x <- 42
  <-x
}

func main() {
  x := make(chan int)
  go foo(x)

  x <- 42
  <-x
}

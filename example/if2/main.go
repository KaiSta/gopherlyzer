package main


func foo(x chan int) {
  x <- 42
}

func main() {
  x := make(chan int)

  go foo(x)
  
  if 5 < 0 {
      <-x
  }
}

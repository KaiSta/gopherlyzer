package main

func main() {
  x := make(chan int)

  for {
    if true {
     x <- 42
    } else {
     <-x
    }

  }
}

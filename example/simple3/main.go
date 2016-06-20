package main

func bar(y chan int) {
  y <- 42
}

func foo() {
 y:= make(chan int)  
 go bar(y)
 <-y 
}

func main() {
  x := make(chan int)
  go foo()
  x <- 42
  <-x
}

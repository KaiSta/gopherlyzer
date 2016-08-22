package main

func foo(x chan int) {
	<-x
}

func main() {
	x := make(chan int)
	go foo(x)
	x <- 42
	close(x)
}

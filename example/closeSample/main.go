package main

func foo(x chan int) {
	<-x
	<-x
}

func main() {
	x := make(chan int)
	go foo(x)
	x <- 42
	if 5 > 8 {
		close(x)
	}

	//x <- 42
}

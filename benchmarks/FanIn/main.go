package main



func work1(out chan int) {
	for {
		out <- 42
	}
}
func work2(out chan int) {
	for {
		out <- 42
	}
}

func fanin(input1, input2, out chan int) {
	for {
		select {
			case s := <-input1:
				out <- s
			case s := <-input2:
				out <- s
		}
	}
}

func main() {
	input1 := make(chan int)
	input2 := make(chan int)
	out := make(chan int)
	go work1(input1)
	go work2(input2)
	go fanin(input1, input2, out)
	for {
		<-out
	}
}
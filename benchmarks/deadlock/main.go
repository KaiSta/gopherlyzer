package main

func Send(ch chan int) {
	ch <- 42
}

func Recv1(ch chan int, done chan int) {
	val := <-ch
        done <- val
}

func Recv2(ch chan int, done chan int) {
	val := <-ch
        done <- val
}

func Work() {
	for {
		
	}
}

func main() {
   	ch := make(chan int)
   	done := make(chan int)

	go Send(ch)
	go Recv1(ch, done)
	go Recv2(ch, done)
	go Work()

	<-done
	<-done
}

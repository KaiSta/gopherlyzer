package main

func generate(ch chan int) {
	for i := 2;;i++ {
		ch <- i
	}
}

func filter(in chan int, out chan int, prime int) {
	for {
		i := <-in
		if i % prime != 0 {
			out <- i
		}
	}
}

func main() {
	ch := make(chan int)
	
	go generate(ch)
	
	//------ 1 ------
	prime := <-ch
	//fmt.Println(prime)
	ch1 := make(chan int)
	go filter(ch, ch1, prime)
	
}

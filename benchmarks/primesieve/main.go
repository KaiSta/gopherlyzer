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
func filter2(in chan int, out chan int, prime int) {
	for {
		i := <-in
		if i % prime != 0 {
			out <- i
		}
	}
}
func filter3(in chan int, out chan int, prime int) {
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
	
	//------ 2 ------
	prime = <-ch1
	//fmt.Println(prime)
	ch2 := make(chan int)
	go filter2(ch1, ch2, prime)
	
	//------ 3 ------
	prime = <-ch2
	//fmt.Println(prime)
	ch3 := make(chan int)
	go filter3(ch2, ch3, prime)
	
}
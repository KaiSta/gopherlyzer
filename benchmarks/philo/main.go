package main

func phil(forks chan int) {
	for {
		<-forks
		<-forks
		
		forks <- 1
		forks <- 1
	}
}

func give1(forks chan int) {
	forks <- 1
}
func give2(forks chan int) {
	forks <- 1
}

func main() {
	forks := make(chan int)
	
	go give1(forks)
	go give2(forks)
	go phil(forks)
	
	for {
		<-forks
		<-forks
		
		forks <- 1
		forks <- 1
	}	
}
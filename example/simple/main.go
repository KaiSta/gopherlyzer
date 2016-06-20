package main


func first(x chan int) {
    x <- 43
}

func main() {
	x := make(chan int)    
    go first(x)
	
    if 5 > 0 {
        <-x
    } else {
        x <- 42
    }
}

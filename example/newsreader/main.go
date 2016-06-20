package main

func news1(ch chan string) {
   ch <- "1"
}

func news2(ch chan string) {
   ch <- "2"
}

func collecter(a chan string, b chan string) {
    a <- <-b
}

func collecter2(a  chan string, b chan string) {
    a <- <-b
}

func collect(n1,n2 chan string) {
   ch := make(chan string)
   
   go collecter(ch, n1)
   go collecter2(ch, n2)

   <-ch
   //fmt.Println(x)
}

func main() {
  n1 := make(chan string)
  n2 := make(chan string)

  go news1(n1)
  go news2(n2)

  collect(n1,n2)
//  collect(n1,n2)
}

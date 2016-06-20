package main

func foo(x,y chan int) { 
   select {
     case <-x:
     case <-y:
   }
}

func main() {
 x := make(chan int)
 y := make(chan int)

 go foo(x,y)

 x <- 42

}

package main

var a chan int

func foobar() {
  bar()
  select {
    case <-a:
      
    case a <- 42:
  }
}

func main() {
    a = make(chan int)
    go foobar()
    foo()
// 
     b := make(chan int)
//     z := <-b
//     z++
     <-a
//     a <- <-b
//     go baz(a)
//     baz(b)
    if 3 > 5 {
        <-b
        a <- 42
    } else {
        <-a
    }
}


func bar() {
  x := <-a
  x++
  foo()
}

func foo() {
  b := make(chan int)
  <-b
  go foo()
}

func baz(x chan int) {
  <-x
}

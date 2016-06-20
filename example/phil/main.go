package main



func philo1(forks chan int) {
    for {
        <-forks
        <-forks
        forks <- 1
        forks <- 1
    }
}
// 
// func philo2(forks chan int) {
//   //  for {
//         <-forks
//         <-forks
//         forks <- 1
//         forks <- 1
//   //  }
// }

func fork1(forks chan int) {
    forks <- 1
}
func fork2(forks chan int) {
    forks <- 1
}
// func fork3(forks chan int) {
//     forks <- 1
// }

func main() {
    forks := make(chan int)
    
    go fork1(forks)
    go fork2(forks)
  //  go fork3(forks)
    
    go philo1(forks)
    //go philo2(forks)
    
    for {
        <-forks
        <-forks
        forks <- 1
        forks <- 1
   }
}
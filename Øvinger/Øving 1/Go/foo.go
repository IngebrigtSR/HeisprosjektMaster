// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    "time"
)

var i = 0

func incrementing() {
    //TODO: increment i 1000000 times
    for j := 1; j < 1000000 ; j++ {
        i += 1;
    }
}

func decrementing() {
    //TODO: decrement i 1000000 times
    for j := 0 ; j < 1000000; j++ {
        i -= 1;
    }
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())    // I guess this is a hint to what GOMAXPROCS does...
	                                    // Try doing the exercise both with and without it!

    // TODO: Spawn both functions as goroutines

    go incrementing()

    go decrementing()
	
    // We have no way to wait for the completion of a goroutine (without additional syncronization of some sort)
    // We'll come back to using channels in Exercise 2. For now: Sleep.


    time.Sleep(1000*time.Millisecond)
    Println("The magic number is:", i)
}


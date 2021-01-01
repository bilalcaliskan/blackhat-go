package main

import (
	"fmt"
	"net"
)

func main() {
	/*
		The previous scanner scanned multiple ports in a single go (pun intended). But your goal now is to scan multiple
		ports concurrently, which will make your port scanner faster. To do this, you’ll harness the power of goroutines.
		Go will let you create as many goroutines as your system can handle, bound only by available memory.
		The most naive way to create a port scanner that runs concurrently is to wrap the call to Dial(network, address string)
		in a goroutine.
	*/
	for i := 1; i <= 1024; i++ {
		go func(j int) {
			address := fmt.Sprintf("scanme.nmap.org:%d", j)
			conn, err := net.Dial("tcp", address)
			if err != nil {
				return
			}
			conn.Close()
			fmt.Printf("%d open\n", j)
		}(i)
	}
}

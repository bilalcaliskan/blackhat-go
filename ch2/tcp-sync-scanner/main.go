package main

import (
	"fmt"
	"sync"
)

func worker(host string, proto string, ports chan int, wg *sync.WaitGroup) {
	for p := range ports {
		fmt.Printf("%s:%d\n", host, p)
		wg.Done()
	}
}

func multiPortScanConcurrentlyUsingWorkerPool(host string, proto string, wg *sync.WaitGroup) {
	ports := make(chan int, 100)
	for i := 0; i < cap(ports); i++ {
		go worker(host, proto, ports, wg)
	}
	for i := 1; i <= 1024; i++ {
		wg.Add(1)
		ports <- i
	}
	wg.Wait()
	close(ports)
}

func main() {
	/*
		To avoid inconsistencies, you’ll use a pool of goroutines to manage the concurrent work being performed. Using a for
		loop, you’ll create a certain number of worker goroutines as a resource pool. Then, in your main() “thread,” you’ll
		use a channel to provide work.
		To start, create a new program that has 100 workers, consumes a channel of int, and prints them to the screen.
		The worker(int, *sync.WaitGroup) function takes two arguments: a channel of type int and a pointer to a WaitGroup.
		The channel will be used to receive work, and the WaitGroup will be used to track when a single work item has been
		completed.
	*/
	var wg sync.WaitGroup
	proto := "tcp"
	host := "api.thevpnbeast.com"
	multiPortScanConcurrentlyUsingWorkerPool(host, proto, &wg)
	/*
		We have created a channel in function above like that;
			ports := make(chan int, 100)
		This allows the channel to be buffered, which means you can send it an item without waiting for a receiver to read
		the item. Buffered channels are ideal for maintaining and tracking work for multiple producers and consumers. You’ve
		capped the channel at 100, meaning it can hold 100 items before the sender will block. This is a slight performance
		increase, as it will allow all the workers to start immediately.
		Next, you use a for loop to start the desired number of workers—in this case, 100. In the worker(int, *sync.WaitGroup)
		function, you use range to continuously receive from the ports channel, looping until the channel is closed.
		After all the work has been completed, you close the channel. Waitgroups stops blocking the execution so compiler can
		reach the closing the channel line.
		You might notice something interesting here: the numbers are printed in no particular order. Welcome to the wonderful
		world of parallelism.
	*/
}
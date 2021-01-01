package main

import (
	"fmt"
	"net"
	"sort"
)

func workerMultiChannel(host, proto string, ports, results chan int) {
	for p := range ports {
		address := fmt.Sprintf("%s:%d", host, p)
		conn, err := net.Dial(proto, address)
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}
}

func multiPortScanConcurrentlyUsingMultipleChannels(host, proto string) {
	ports := make(chan int, 100)
	results := make(chan int)
	var openPorts []int

	for i := 0; i < cap(ports); i++ {
		go workerMultiChannel(host, proto, ports, results)
	}

	go func() {
		for i := 1; i <= 1024; i++ {
			ports <- i
		}
	}()

	for i := 0; i < 1024; i++ {
		port := <-results
		if port != 0 {
			openPorts = append(openPorts, port)
		}
	}

	close(ports)
	close(results)
	sort.Ints(openPorts)
	for _, port := range openPorts {
		fmt.Printf("%d open\n", port)
	}
}

func main() {
	// Multichannel Communication
	/*
		To complete the port scanner, you could plug in your code from earlier in the section, and it would work just fine.
		However, the printed ports would be unsorted, because the scanner wouldn’t check them in order.
		To solve this problem, you need to use a separate thread to pass the result of the port scan back to your main thread
		to order the ports before printing. Another benefit of this modification is that you can remove the dependency of a
		WaitGroup entirely, as you’ll have another method of tracking completion. For example, if you scan 1024 ports, you’re
		sending on the worker channel 1024 times, and you’ll need to send the result of that work back to the main thread 1024
		times. Because the number of work units sent and the number of results received are the same, your program can know
		when to close the channels and subsequently shut down the workers.
	*/
	proto := "tcp"
	host := "api.thevpnbeast.com"
	multiPortScanConcurrentlyUsingMultipleChannels(host, proto)
	/*
		If the port is closed, you’ll send a zero, and if its open, you will send the port to the results channel inside
		function workerMultiChannel. Also, you create a separate channel to communicate the results from the worker to the
		main thread. You then use a slice to store the results so you can sort them later. Next, you need to send to the
		workers in a separate goroutine because the result-gathering loop needs to start before more than 100 items of work
		can continue.

		The result-gathering loop receives on the results channel 1024 times. If the port doesn’t equal 0, it’s appended to
		the slice. After closing the channels, you’ll use sort to sort the slice of open ports. All that’s left is to loop
		over the slice and print the open ports to screen.

		There you have it: a highly efficient port scanner. Take some time to play around with the code—specifically, the number
		of workers. The higher the count, the faster your program should execute. But if you add too many workers, your results
		could become unreliable. When you’re writing tools for others to use, you’ll want to use a healthy default value that
		caters to reliability over speed. However, you should also allow users to provide the number of workers as an option.
	*/
}

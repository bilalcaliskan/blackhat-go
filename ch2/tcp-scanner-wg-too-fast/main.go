package main

import (
	"fmt"
	"net"
	"sync"
)

// The “Too Fast” Scanner Version With Wait Group
func multiPortScanConcurrently(host string, proto string, wg *sync.WaitGroup) {
	for i := 1; i <= 1024; i++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			address := fmt.Sprintf("%s:%d", host, j)
			conn, err := net.Dial(proto, address)
			if err != nil {
				// port is closed or filtered
				return
			}
			conn.Close()
			fmt.Printf("%d open on host %s\n", j, host)
		}(i)
	}
}

func main() {
	/*
		If we dont define a wait group, the code you just ran launches a single goroutine per connection, but the main goroutine
		who calls the method does not wait for child goroutines to complete. Therefore, the code completes and exits as soon
		as the for loop finishes its iterations, which may be faster than the network exchange of packets between your code
		and the target ports. You may not get accurate results for ports whose packets were still in-flight.
		There are a few ways to fix this. One is to use WaitGroup from the sync package, which is a thread-safe way to control
		concurrency. WaitGroup is a struct type and can be created like so:
			var wg sync.WaitGroup
		After created that, we will pass it to the function
		Once you’ve created WaitGroup, you can call a few methods on the struct. The first is Add(int), which increases an
		internal counter by the number provided. Next, Done() decrements the counter by one. Finally, Wait() blocks the
		execution of the goroutine in which it’s called, and will not allow further execution until the internal counter
		reaches zero. You can combine these calls to ensure that the main goroutine waits for all connections to finish.
	*/
	var wg sync.WaitGroup
	proto := "tcp"
	host := "api.thevpnbeast.com"
	multiPortScanConcurrently(host, proto, &wg)
	wg.Wait()
	/*
		Above function remains largely identical to our initial version(tcp-scanner-too-fast). However, you’ve added code
		that explicitly tracks the remaining work. In this version of the program, you create sync.WaitGroup which acts
		as a synchronized counter.
		You increment this counter via wg.Add(1) each time you create a goroutine to scan a port and a deferred call to wg.Done()
		decrements the counter whenever one unit of work has been performed. Your main() function calls wg.Wait(), which blocks
		until all the work has been done and your counter has returned to zero.

		This version of the program is better, but still incorrect. If you run this multiple times against multiple hosts, you
		might see inconsistent results. Scanning an excessive number of hosts or ports simultaneously may cause network or
		system limitations to skew your results. Go ahead and change 1024 to 65535, and the destination server to your
		localhost 127.0.0.1 in your code. If you want, you can use Wireshark or tcpdump to see how fast those connections
		are opened.
	*/
}

package main

import (
	"io"
	"log"
	"net"
)

func handle(src net.Conn) {
	dst, err := net.Dial("tcp", "joescatcam.website:80")
	if err != nil {
		log.Fatalln("Unable to connect to our reachable host")
	}
	defer dst.Close()

	// Run in goroutine to prevent io.Copy from blocking
	go func() {
		// Copy our source's output to the destination
		if _, err := io.Copy(dst, src); err != nil {
			log.Fatalln(err)
		}
	}()

	// Copy our destination's output back to our source
	if _, err := io.Copy(src, dst); err != nil {
		log.Fatalln(err)
	}
}

func runJoesProxyCom() {
	// Listen on local port 80
	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatalln("Unable to bind to port")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("Unable to accept connection")
		}
		go handle(conn)
	}
}

func main() {
	// Proxying a TCP Client
	/*
		Now that you have a solid foundation, you can take what you’ve learned up to this point and create a simple port
		forwarder to proxy a connection through an intermediary service or host. As mentioned earlier in this chapter, this
		is useful for trying to circumvent restrictive egress controls or to leverage a system to bypass network segmentation.
		Before laying out the code, consider this imaginary but realistic problem: Joe is an underperforming employee who
		works for ACME Inc. as a business analyst making a handsome salary based on slight exaggerations he included on his
		resume. (Did he really go to an Ivy League school? Joe, that’s not very ethical.) Joe’s lack of motivation is
		matched only by his love for cats—so much so that Joe installed cat cameras at home and hosted a site,
		joescatcam.website, through which he could remotely monitor the dander-filled fluff bags. One problem, though: ACME is
		onto Joe. They don’t like that he’s streaming his cat cam 24/7 in 4K ultra high-def, using valuable ACME network
		bandwidth. ACME has even blocked its employees from visiting Joe’s cat cam website.
		Joe has an idea. “What if I set up a port-forwarder on an internet-based system I control,” Joe says, “and force the
		redirection of all traffic from that host to joescatcam.website?” Joe checks at work the following day and confirms
		he can access his personal website, hosted at the joesproxy.com domain. Joe skips his afternoon meetings, heads to a
		coffee shop, and quickly codes a solution to his problem. He’ll forward all traffic received at http://joesproxy.com
		to http://joescatcam.website.
		Here’s Joe’s code, which he runs on the joesproxy.com server:
	*/
	runJoesProxyCom()
	/*
		Start by examining Joe’s handle(net.Conn) function. Joe connects to joescatcam.website (recall that this unreachable
		host isn’t directly accessible from Joe’s corporate workstation). Joe then uses Copy(Writer, Reader) two separate
		times. The first instance ensures that data from the inbound connection is copied to the joescatcam.website connection.
		The second instance ensures that data read from joescatcam.website is written back to the connecting client’s
		connection. Because Copy(Writer, Reader) is a blocking function, and will continue to block execution until the
		network connection is closed, Joe wisely wraps his first call to Copy(Writer, Reader) in a new goroutine. This
		ensures that execution within the handle(net.Conn) function continues, and the second Copy(Writer, Reader) call can
		be made.
		Joe’s proxy listens on port 80 and relays any traffic received from a connection to and from port 80 on
		joescatcam.website. Joe, that crazy and wasteful man, confirms that he can connect to joescatcam.website via
		joesproxy.com by connecting with curl.
		At this point, Joe has done it. He’s living the dream, wasting ACME-sponsored time and network bandwidth while he
		watches his cats. Today, there will be cats!
	*/
}
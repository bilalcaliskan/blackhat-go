package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func handle(src net.Conn, targetProto, connectionStr string) {
	dst, err := net.Dial(targetProto, connectionStr)
	if err != nil {
		log.Fatalf("Unable to connect to remote host %s!\n", connectionStr)
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

func main() {
	proxyProto := flag.String("proxyProto", "tcp", "Please provide a proxy server protocol")
	proxyPort := flag.Int("proxyPort", 3000, "Please provide a port to run proxy server on")
	targetProto := flag.String("targetProto", "tcp", "Please provide a target protocol to proxy")
	targetDns := flag.String("targetDns", "mail.google.com", "Please provide a target DNS to proxy")
	targetPort := flag.Int("targetPort", 443, "Please provide a target port to proxy")
	flag.Parse()

	connectionStr := fmt.Sprintf("%s:%d", *targetDns, *targetPort)
	listener, err := net.Listen(*proxyProto, fmt.Sprintf(":%d", *proxyPort))
	if err != nil {
		log.Fatalf("Unable to bind to port %d!\n", *proxyPort)
	}
	log.Printf("Server is listening on port %d!\n", *proxyPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln("Unable to accept connection!")
		}

		go handle(conn, *targetProto, connectionStr)
	}
}

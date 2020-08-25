package main

import (
	"bufio"
	"io"
	"log"
	"net"
)

// echo is a handler function that simply echoes received data.
func echo(conn net.Conn) {
	defer conn.Close()

	// Create a buffer to store received data
	b := make([]byte, 512)
	for {
		// Receive data via conn.Read into a buffer
		size, err := conn.Read(b[0:])
		if err == io.EOF {
			log.Println("Client disconnected")
			break
		}
		if err != nil {
			log.Println("Unexpected error")
			break
		}
		log.Printf("Received %d bytes: %s\n", size, string(b))

		// Send data via conn.Write
		if _, err := conn.Write(b[0:size]); err != nil {
			log.Fatalln("Unable to write data")
		}
	}
}

func improvedEcho(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	s, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalln("Unable to read data")
	}
	log.Printf("Read %d bytes: %s\n", len(s), s)

	log.Println("Writing data")
	writer := bufio.NewWriter(conn)
	if _, err := writer.WriteString(s); err != nil {
		log.Fatalln("Unable to write data")
	}
	writer.Flush()
}

func runEchoServer() {
	// Bind to TCP port 20080 on all interfaces.
	listener, err := net.Listen("tcp", ":20080")
	if err != nil {
		log.Fatalln("Unable to bind to port")
	}
	log.Println("Listening on 0.0.0.0:20080")
	for  {
		// Wait for connection. Create net.Conn on connection established. It blocks execution as it awaits client connections
		conn, err := listener.Accept()
		log.Println("Received connection")
		if err != nil {
			log.Fatalln("Unable to accept connection")
		}
		// Handle the connection. Using goroutine for concurrency
		go echo(conn)
	}
}

func runImprovedEchoServer() {
	// Bind to TCP port 20080 on all interfaces.
	listener, err := net.Listen("tcp", ":20081")
	if err != nil {
		log.Fatalln("Unable to bind to port")
	}
	log.Println("Listening on 0.0.0.0:20081")
	for  {
		// Wait for connection. Create net.Conn on connection established. It blocks execution as it awaits client connections
		conn, err := listener.Accept()
		log.Println("Received connection")
		if err != nil {
			log.Fatalln("Unable to accept connection")
		}
		// Handle the connection. Using goroutine for concurrency
		go improvedEcho(conn)
	}
}

func main() {
	/*
		As is customary for most languages, you’ll start by building an echo server to learn how to read and write data to
		and from a socket. To do this, you’ll use net.Conn, Go’s stream-oriented network connection, which we introduced when
		you built a port scanner. Based on Go’s documentation for the data type, Conn implements the Read([]byte) and
		Write([]byte) functions as defined for the Reader and Writer interfaces. Therefore, Conn is both a Reader and a
		Writer (yes, this is possible). This makes sense logically, as TCP connections are bidirectional and can be used to
		send (write) or receive (read) data.
		After creating an instance of Conn, you’ll be able to send and receive data over a TCP socket. However, a TCP server
		can’t simply manufacture a connection; a client must establish a connection. In Go, you can use net.Listen(network,
		address string) to first open a TCP listener on a specific port. Once a client connects, the Accept() method creates
		and returns a Conn object that you can use for receiving and sending data.
	*/
	runEchoServer()
	/*
		echo(net.Conn), which accepts a Conn instance as a parameter. It behaves as a connection handler to perform all
		necessary I/O. The function loops indefinitely, using a buffer to read and write data from and to the connection.
		The data is read into a variable named b and subsequently written back on the connection.
		Now you need to set up a listener that will call your handler. As mentioned previously, a server can’t manufacture
		a connection but must instead listen for a client to connect. Therefore, a listener, defined as tcp bound to port
		20080, is started on all interfaces by using the net.Listen(network, address string) function.
		Next, an infinite loop ensures that the server will continue to listen for connections even after one has been
		received. Within this loop, you call listener.Accept(), a function that blocks execution as it awaits client
		connections. When a client connects, this function returns a Conn instance. Recall from earlier discussions in this
		section that Conn is both a Reader and a Writer (it implements the Read([]byte) and Write([]byte) interface methods).
		The Conn instance is then passed to the echo(net.Conn) handler function. This call is prefaced with the go keyword,
		making it a concurrent call so that other connections don’t block while waiting for the handler function to complete.
		This is likely overkill for such a simple server, but we’ve included it again to demonstrate the simplicity of Go’s
		concurrency pattern, in case it wasn’t already clear. At this point, you have two lightweight threads running
		concurrently:
			- The main thread loops back and blocks on listener.Accept() while it awaits another connection.
			- The handler goroutine, whose execution has been transferred to the echo(net.Conn) function, proceeds to run,
			processing the data.
	*/


	// Improving the Code by Creating a Buffered Listener
	/*
		Above function runEchoServer() works perfectly fine but relies on fairly low-level function calls, buffer tracking,
		and iterative reads/writes. This is a somewhat tedious, error-prone process. Fortunately, Go contains other packages
		that can simplify this process and reduce the complexity of the code. Specifically, the bufio package wraps Reader
		and Writer to create a buffered I/O mechanism. The updated version of echo(net.Conn) function is detailed here, and
		an explanation of the changes follows.
	*/
	runImprovedEchoServer()
	/*
		No longer are you directly calling the Read([]byte) and Write([]byte) functions on the Conn instance; instead, you’re
		initializing a new buffered Reader and Writer via NewReader(io.Reader) and NewWriter(io.Writer). These calls both
		take, as a parameter, an existing Reader and Writer (remember, the Conn type implements the necessary functions to
		be considered both a Reader and a Writer).
		Both buffered instances contain complementary functions for reading and writing string data. ReadString(byte) takes
		a delimiter character used to denote how far to read, whereas WriteString(byte) writes the string to the socket.
		When writing data, you need to explicitly call writer.Flush() to flush write all the data to the underlying writer
		(in this case, a Conn instance).
		Although the previous example simplifies the process by using buffered I/O, you can reframe it to use the
		Copy(Writer, Reader) convenience function. Recall that this function takes as input a destination Writer and a
		source Reader, simply copying from source to destination.
		In this example, you’ll pass the conn variable as both the source and destination because you’ll be echoing the
		contents back on the established connection:
			func echo(conn net.Conn) {
		    	defer conn.Close()
		    	// Copy data from io.Reader to io.Writer via io.Copy().
				if _, err := io.Copy(conn, conn); err != nil {
					log.Fatalln("Unable to read/write data")
				}
			}
	*/
}
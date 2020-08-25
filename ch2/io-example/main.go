package main

import (
	"fmt"
	"log"
	"os"
)

// FooReader defines an io.Reader to read from stdin.
type FooReader struct{}

// Read reads data from stdin.
func (fooReader *FooReader) Read(b []byte) (int, error) {
	fmt.Print("in > ")
	return os.Stdin.Read(b)
}

// FooWriter defines an io.Writer to write to Stdout.
type FooWriter struct{}

// Write writes data to Stdout.
func (fooWriter *FooWriter) Write(b []byte) (int, error) {
	fmt.Print("out> ")
	return os.Stdout.Write(b)
}

func readAndWrite(writer *FooWriter, reader *FooReader) {
	// Create buffer to hold input/output
	input := make([]byte, 4096)
	// Use reader to read input
	s, err := reader.Read(input)
	if err != nil {
		log.Fatalln("Unable to read data")
	}
	fmt.Printf("Read %d bytes from stdin\n", s)
	// Use writer to write output
	s, err = writer.Write(input)
	if err != nil {
		log.Fatalln("Unable to write data")
	}
	fmt.Printf("Wrote %d bytes to stdout\n", s)
}

func main() {
	/*
		To create the examples in this section, you need to use two significant types that are crucial to essentially all
		input/output (I/O) tasks, whether you’re using TCP, HTTP, a filesystem, or any other means: io.Reader and io.Writer.
		Part of Go’s built-in io package, these types act as the cornerstone to any data transmission, local or networked.
		These types are defined in Go’s documentation as follows:
			type Reader interface {
		    	Read(p []byte) (n int, err error)
			}
			type Writer interface {
				Write(p []byte) (n int, err error)
			}
		Both types are defined as interfaces, meaning they can’t be directly instantiated. Each type contains the definition
		of a single exported function: Read or Write. As explained in Chapter 1, you can think of these functions as abstract
		methods that must be implemented on a type for it to be considered a Reader or Writer. For example, the following
		contrived type fulfills this contract and can be used anywhere a Reader is accepted:
			type FooReader struct {}
			func (fooReader *FooReader) Read(p []byte) (int, error) {
				// Read some data from somewhere, anywhere.
				return len(dataReadFromSomewhere), nil
			}
		Let’s take this knowledge and create something semi-usable: a custom Reader and Writer that wraps stdin and stdout.
		The code for this is a little contrived since Go’s os.Stdin and os.Stdout types already act as Reader and Writer,
		but then you wouldn’t learn anything if you didn’t reinvent the wheel every now and again, would you?
	*/
	// Instantiate reader and writer.
	var (
		reader FooReader
		writer FooWriter
	)
	readAndWrite(&writer, &reader)
	/*
		In the above readAndWrite() function, the data itself is copied into the byte slice passed to the function. This is
		consistent with the Reader interface prototype definition provided earlier in this section.
		Copying data from a Reader to a Writer is a fairly common pattern—so much so that Go’s io package contains a Copy()
		function that can be used to simplify the main() function. The function prototype is as follows:
			func Copy(dst io.Writer, src io.Reader) (written int64, error)
		This convenience function allows you to achieve the same programmatic behavior as before with below function!
	*/
}

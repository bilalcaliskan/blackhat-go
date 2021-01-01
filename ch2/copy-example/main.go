package main

import (
	"fmt"
	"io"
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

func readAndWriteWithCopying(writer *FooWriter, reader *FooReader) {
	if _, err := io.Copy(writer, reader); err != nil {
		log.Fatalln("Unable to read/write data")
	}
}

func main() {
	// Instantiate reader and writer.
	var (
		reader FooReader
		writer FooWriter
	)

	readAndWriteWithCopying(&writer, &reader)
	/*
		Notice that the explicit calls to reader.Read([]byte) and writer.Write([]byte) have been replaced with a single call
		to io.Copy(writer, reader) inside function readAndWriteWithCopying(). Under the covers, io.Copy(writer, reader) calls
		the Read([]byte) function on the provided reader, triggering the FooReader to read from stdin. Subsequently,
		io.Copy(writer, reader) calls the Write([]byte) function on the provided writer, resulting in a call to your
		FooWriter, which writes the data to stdout. Essentially, io.Copy(writer, reader) handles the sequential read-then-write
		process without all the petty details.
	*/
}

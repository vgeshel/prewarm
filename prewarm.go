package main

import (
	"os"
	"fmt"
	"flag"
)

func DoRead(file *os.File, offset int64, buf []byte, done chan int64) {
	_, err := file.ReadAt(buf, offset)

	if err != nil {
		panic(fmt.Sprintf("error reading at %d: %s", offset, err.Error())) 
	}

	done <- offset;
}

func main() {
	name := flag.String("file", "", "file to prewarm")
	chunk := flag.Int64("chunk", 1024 * 1024, "touch each part of this size")
	toread := flag.Int64("bufsize", 1024, "read this many bytes from each chunk")

	flag.Parse()
	
	fmt.Printf("name: %s\n", *name);
	
	if *name == "" {
		fmt.Printf("-file is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	info, err := os.Stat(*name)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s: %d\n", *name, info.Size())

	file, err := os.Open(*name)

	if err != nil {
		panic(err)
	}

	done := make(chan int64)
	buf := make([]byte, *toread)
	cnt := 0

	var pos int64
	for pos = 0; pos < info.Size(); pos += *chunk {
		go DoRead(file, pos, buf, done)
		cnt ++
	}

	for cnt > 0 {
		off := <-done;
		cnt --;

		fmt.Printf("done at %d\n", off)
	}
}

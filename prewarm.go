package main

import (
	"os"
	"fmt"
	"flag"
	"io/ioutil"
	s "strings"
	c "strconv"
)

func DoRead(file *os.File, offset int64, buf *[]byte, done chan int64) {
	_, err := file.ReadAt(*buf, offset)

	if err != nil {
		panic(fmt.Sprintf("error reading at %d: %s", offset, err.Error())) 
	}

	done <- offset;
}

func main() {
	name := flag.String("file", "", "file to prewarm")
	chunk := flag.Int64("chunk", 1024 * 1024, "touch each part of this size")
	toread := flag.Int64("bufsize", 1024, "read this many bytes from each chunk")
	use_sys := flag.Bool("sys", false, "use /sys/class/block/DEV/size to determine size instead of stat()")

	flag.Parse()
	
	fmt.Printf("name: %s\n", *name);
	
	if *name == "" {
		fmt.Printf("-file is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var size int64 = 0
	if *use_sys {
		parts := s.Split(*name, "/")
		basename := parts[len(parts) - 1]
		data, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/block/%s/size", basename))

		if err != nil {
			panic(fmt.Sprintf("could not read size file for %s: %s", name, err))
		}

		str := s.TrimSpace(string(data))

		blocks, err := c.ParseInt(str, 10, 64)

		if err != nil {
			panic(err)
		}

		size = blocks * 512
	} else {
		info, err := os.Stat(*name)
		
		if err != nil {
			panic(err)
		}

		size = info.Size()
	}

	if size == 0 {
		panic(fmt.Sprintf("unable to determine size of %s", name))
	}

	fmt.Printf("%s: %d\n", *name, size)

	file, err := os.Open(*name)

	if err != nil {
		panic(err)
	}

	done := make(chan int64)
	buf := make([]byte, *toread)
	cnt := 0

	var pos int64
	for pos = 0; pos < size; pos += *chunk {
		go DoRead(file, pos, &buf, done)
		cnt ++
	}

	for cnt > 0 {
		off := <-done;
		cnt --;

		fmt.Printf("done at %d\n", off)
	}
}

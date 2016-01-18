package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hpcloud/tail"
)

var threshold int
var filename string

func parse() {
	flag.IntVar(&threshold, "t", 10, "Requests/second upper limit")
	flag.StringVar(&filename, "f", "", "File to be processed")

	// parse flags and check values for their correctness
	flag.Parse()

	if threshold < 1 {
		log.Fatal("Threshold cannot be negative")
	}

	if filename == "" {
		log.Fatal("Access log file must be specified. Exiting")
	}
}

func openAndReadFile(filename string) *tail.Tail {
	t, err := tail.TailFile(filename, tail.Config{Follow: true})
	if err != nil {
		panic(err)
	}

	return t
}

func main() {
	parse()
	t := openAndReadFile(filename)
	queue := make(chan *Log)
	go process(t, queue)
	for x := range queue {
		fmt.Println(x)
	}

}

package main

import (
	"flag"
	_ "fmt"
	"github.com/hpcloud/tail"
	"log"
)

var treshold int
var timeWindow int
var filename string

func parse() {
	flag.IntVar(&treshold, "t", 10, "Requests/minute upper limit")
	flag.IntVar(&timeWindow, "w", 2, "Time window span (in minutes)")
	flag.StringVar(&filename, "f", "", "File to be processed")

	// parse flags and check values for their correctness
	flag.Parse()

	if treshold < 1 {
		log.Fatal("Threshold cannot be negative")
	}

	if timeWindow < 1 {
		log.Fatal("TimeWindow must be positive number")
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
	storage := NewCache()
	go Process(t, queue)
	go Store(storage, queue)
	Dispatcher(storage, treshold*timeWindow, timeWindow)
}

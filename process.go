package main

import (
	"errors"
	"log"
	"net"
	"regexp"
	"strconv"

	_ "fmt"

	"github.com/hpcloud/tail"
)

const (
	LogLength = 7
	// 8.8.8.8 - - [18/01/2016:09:23:07 +01.000] "POST /pages" 500 1280
	LogRegexp = `(.*?) (.*?) (.*?) \[(.*?)\] \"(.*?)\" (\d+) (\d+)`
)

type Log struct {
	Address net.IP
	Rfc     string
	User    string
	Time    string
	Method  string
	Status  int
	Bytes   int
}

func NewLog(fields []string) (*Log, error) {

	if len(fields) != LogLength {
		return nil, errors.New("Malformed log line")
	}
	ip := net.ParseIP(fields[0])
	if ip == nil {
		return nil, errors.New("Invalid IP address")
	}

	status, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil, err
	}

	bytes, err := strconv.Atoi(fields[6])
	if err != nil {
		return nil, err
	}

	entry := &Log{
		ip,
		fields[1],
		fields[2],
		fields[3],
		fields[4],
		status,
		bytes,
	}
	return entry, nil

}

func process(t *tail.Tail, queue chan *Log) {

	re, _ := regexp.Compile(LogRegexp)

	for line := range t.Lines {
		//fmt.Println(line.Text)

		fields := re.FindStringSubmatch(line.Text)[1:]

		entry, err := NewLog(fields)
		if err != nil {
			log.Printf("Error while parsing log, skipping.\n%s\n", err)
			continue
		}

		// send Log to the channel
		queue <- entry

	}
}

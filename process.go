package main

import (
	"errors"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "fmt"

	"github.com/hpcloud/tail"
)

const (
	LogLength = 7
	// 8.8.8.8 - - [18/01/2016:09:23:07 +01.000] "POST /pages" 500 1280
	LogRegexp        = `(.*?) (.*?) (.*?) \[(.*?)\] \"(.*?)\" (\d+) (\d+)`
	DateTimeTemplate = "02/01/2006:15:04:05 -0700"
)

type Log struct {
	Address       net.IP
	Rfc           string
	User          string
	Time          time.Time
	Method        string
	Section       string
	URLParameters string
	Status        int
	Bytes         int
}

func NewLog(fields []string) (*Log, error) {

	if len(fields) != LogLength {
		return nil, errors.New("Malformed log line")
	}
	ip := net.ParseIP(fields[0])
	if ip == nil {
		return nil, errors.New("Invalid IP address")
	}

	// split GET "/static/pages?something=2" into
	// ["GET", "static", "pages?something=2"]
	s := strings.SplitN(fields[4], " ", 2)

	method, url := s[0], s[1]
	url = strings.Trim(url, "/ ")

	s = strings.SplitN(url, "/", 2)

	section, urlparameters := s[0], ""

	if len(s) > 1 {
		urlparameters = s[1]
	}

	//fmt.Printf("url=%s, %s    %s\n", url, section, urlparameters)

	status, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil, err
	}

	bytes, err := strconv.Atoi(fields[6])
	if err != nil {
		return nil, err
	}

	datetime, err := time.Parse(DateTimeTemplate, fields[3])
	if err != nil {
		return nil, err
	}

	entry := &Log{
		ip,
		fields[1],
		fields[2],
		datetime,
		method,
		section,
		urlparameters,
		status,
		bytes,
	}
	return entry, nil

}

func Process(t *tail.Tail, queue chan *Log) {

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

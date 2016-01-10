package main

import (
	"net"
	"time"
)

type Log struct {
	Address net.IP
	Rfc     string
	User    string
	Method  string
	Request string
	Proto   string
	Status  string
	Bytes   uint64
}

func process(t *tail.Tail) {
	for _, line := range t.Lines {

	}
}

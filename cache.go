package main

import (
	"fmt"
	"sync"
)

type Entry struct {
	Id           string
	TotalHits    uint64
	StatusCodes  map[int]uint64
	MethodsStats map[string]uint64
}

func NewEntry() *Entry {
	return &Entry{
		"", 0,
		make(map[int]uint64),
		make(map[string]uint64),
	}
}

func (e *Entry) Stringer() string {
	return fmt.Sprintf("Id: >>%s<<\n\tStatus Stats: %v\n\tMethodStats %v\n",
		e.Id, e.StatusCodes, e.MethodsStats)
}

type Cache struct {
	m      sync.Mutex
	Memory map[string]*Entry
}

func NewCache() *Cache {
	c := new(Cache)
	c.Memory = make(map[string]*Entry)
	return c
}

func store(storage *Cache, queue chan *Log) {

	for l := range queue {

		storage.m.Lock()

		var e *Entry = nil
		if _, ok := storage.Memory[l.Section]; ok == false {
			fmt.Println("CACHE MISS")
			e = NewEntry()
			e.Id = l.Section
			storage.Memory[l.Section] = e
		} else {
			fmt.Println("CACHE HIT")
			e = storage.Memory[l.Section]
		}

		e.StatusCodes[l.Status]++
		e.MethodsStats[l.Method]++
		e.TotalHits++

		fmt.Println(e.Stringer())

		fmt.Println(storage)

		storage.m.Unlock()
	}
}

package main

import (
	"fmt"
	"sync"
)

const HeapLenght = 10

type Entry struct {
	Id           string
	TotalHits    uint64
	StatusCodes  map[int]uint64
	MethodsStats map[string]uint64

	// used by heap.Interface
	index int
}

func NewEntry() *Entry {
	return &Entry{
		"", 0,
		make(map[int]uint64),
		make(map[string]uint64),
		0,
	}
}

func (e *Entry) Stringer() string {
	return fmt.Sprintf("Id: >>%s<<\n\tStatus Stats: %v\n\tMethodStats %v\n",
		e.Id, e.StatusCodes, e.MethodsStats)
}

type Cache struct {
	m      sync.Mutex
	Memory map[string]*Entry
	Pq     PriorityQueue
}

func NewCache() *Cache {
	c := new(Cache)
	c.Memory = make(map[string]*Entry)
	// initialize priority queue
	c.Pq = make(PriorityQueue, 0, HeapLenght)
	return c
}

func Store(storage *Cache, queue chan *Log) {

	for l := range queue {

		storage.m.Lock()

		var e *Entry = nil
		var isNew bool
		if _, ok := storage.Memory[l.Section]; ok == false {

			e = NewEntry()
			e.Id = l.Section
			storage.Memory[l.Section] = e
			isNew = true

		} else {
			isNew = false
			e = storage.Memory[l.Section]
		}

		// update counters

		e.StatusCodes[l.Status]++
		e.MethodsStats[l.Method]++
		e.TotalHits++

		// update priority queue

		if isNew && storage.Pq.Len() < HeapLenght {
			storage.Pq = append(storage.Pq, e)
		} else if isNew {
			heap.Push(&storage.Pq, e)
			storage.Pq.update(e)
			for storage.Pq.Len() > HeapLenght {
				storage.Pq.Pop()
			}
		}

		fmt.Println(e.Stringer())
		storage.m.Unlock()
	}
}

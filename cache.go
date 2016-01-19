package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	HeapLenght            = 100
	StatsLimit            = 10
	StandardStatsInterval = 10
)

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
		-1,
	}
}

func (e *Entry) Stringer() string {
	return fmt.Sprintf("Section:/%s,\nHits: %d\n\tStatus Stats: %v\n\tMethodStats %v\n",
		e.Id, e.TotalHits, e.StatusCodes, e.MethodsStats)
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

		if isNew {
			storage.Pq.Push(e)
		} else {
			storage.Pq.update(e)
		}

		//TODO(marek): trim or not to trim the pq?
		/*
			if storage.Pq.Len() > HeapLenght {
				for i := HeapLenght - 1; i < storage.Pq.Len(); i++ {
					storage.Pq[i].index = -1
				}

				//trim priority queue
				storage.Pq = storage.Pq[:HeapLenght]
			}
		*/
		storage.m.Unlock()
	}
}

func StandardAlert(storage *Cache) {
	c := time.Tick(StandardStatsInterval * time.Second)

	var totalHits uint64 = 0
	statusCodes := make(map[int]uint64)

	for now := range c {

		totalHits = 0
		for k, _ := range statusCodes {
			statusCodes[k] = 0
		}

		storage.m.Lock()
		for _, val := range storage.Pq {
			totalHits += val.TotalHits
			for status, cnt := range val.StatusCodes {
				statusCodes[status/100] += cnt
			}

		}

		l := math.Min(float64(HeapLenght), float64(storage.Pq.Len()))
		limit := int(math.Min(StatsLimit, l))

		tophits := storage.Pq[:limit]
		storage.m.Unlock()

		fmt.Printf("======== %v ======\n", now)

		fmt.Printf("Total Hits: %d, Hit statistics:\n\t2xx: %d, 3xx: %d, 4xx: %d, 5xx: %d\n",
			totalHits, statusCodes[2], statusCodes[3], statusCodes[4], statusCodes[5])
		fmt.Printf("Sections with most hits so far (%d):\n", limit)
		for _, h := range tophits {
			fmt.Println(h.Stringer())
		}
		fmt.Printf("=======================================================\n")

	}

}

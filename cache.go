package main

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/oleiade/lane"
)

const (
	HeapLenght            = 100
	StatsLimit            = 10
	StandardStatsInterval = 10
)

type StatusCodeMap map[int]uint64

func (scm StatusCodeMap) String() string {
	s := ""
	for status, cnt := range scm {
		s += fmt.Sprintf("%d: %d, ", status, cnt)
	}
	return strings.Trim(s, ",")
}

type MethodStatsMap map[string]uint64

func (msm MethodStatsMap) String() string {
	s := ""
	for method, cnt := range msm {
		s += fmt.Sprintf("%s: %d, ", method, cnt)
	}
	return strings.Trim(s, ",")
}

type Entry struct {
	Id           string
	TotalHits    uint64
	StatusCodes  StatusCodeMap
	MethodsStats MethodStatsMap
	Logs         *lane.Deque
	HighRate     bool
	// used by heap.Interface
	index int
}

func NewEntry() *Entry {
	return &Entry{
		"", 0,
		make(StatusCodeMap),
		make(MethodStatsMap),
		lane.NewDeque(),
		false,
		-1,
	}
}

func (e *Entry) String() string {
	return fmt.Sprintf("Section:/%s,\nHits: %d\n\tStatus Stats: %s\n\tMethodStats %s\n",
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
		e.Logs.Append(l)

		storage.m.Unlock()
	}
}

func StandardAlert(storage *Cache,
	sink chan StatsEvent) {

	timer := time.Tick(StandardStatsInterval * time.Second)

	var totalHits uint64 = 0
	statusCodes := make(map[int]uint64)

	for now := range timer {

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
		tophits := make([]*Entry, limit)
		copy(tophits, storage.Pq[:limit])

		sink <- StatsEvent{now, totalHits, statusCodes, tophits}
		storage.m.Unlock()
	}

}

func AnalyzeAndTrimLogs(e *Entry, now time.Time,
	alert, calm chan AlertEvent,
	treshold, timeWindow int) {

	// setup TimeWindow to (Now - timeWindow minutes)
	TimeWindow := now.Add((time.Minute * time.Duration(timeWindow)) * (-1))

	for !e.Logs.Empty() && e.Logs.First().(*Log).Time.Before(TimeWindow) {
		e.Logs.Shift()
	}

	if e.Logs.Size() > treshold {
		e.HighRate = true
		alert <- AlertEvent{e, now}
	} else if e.HighRate {
		e.HighRate = false
		calm <- AlertEvent{e, now}
	}
}

func Dispatcher(storage *Cache, treshold, timeWindow int) {

	timer := time.Tick(time.Minute * time.Duration(timeWindow))

	alerts := make(chan AlertEvent, 10)
	calm := make(chan AlertEvent, 10)
	stats := make(chan StatsEvent, 10)

	defer close(alerts)
	defer close(calm)
	defer close(stats)

	go StandardAlert(storage, stats)
	go Output(alerts, calm, stats)

	for {
		//block until timer 'ticks'
		<-timer
		storage.m.Lock()
		now := time.Now()
		for _, e := range storage.Memory {
			go AnalyzeAndTrimLogs(e, now, alerts, calm, treshold, timeWindow)
		}
		storage.m.Unlock()
	}
}

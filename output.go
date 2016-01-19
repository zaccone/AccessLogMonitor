package main

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

type AlertEvent struct {
	Event *Entry
	Now   time.Time
}

type StatsEvent struct {
	Now         time.Time
	Hits        uint64
	StatusCodes map[int]uint64
	TopHits     []*Entry
}

func Output(alerts chan AlertEvent,
	calm chan AlertEvent,
	stats chan StatsEvent) {

	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)

	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)

	for {
		select {
		case alert := <-alerts:
			boldRed.Printf("High traffic generated on section /%s, hits %d, triggered at %v\n",
				alert.Event.Id, alert.Event.Logs.Size(), alert.Now)
		case alert := <-calm:
			boldGreen.Printf("%v: Traffic for section /%s is back to normal (%d hits in last time window)\n",
				alert.Now, alert.Event.Id, alert.Event.Logs.Size())
		case stat := <-stats:
			fmt.Printf("======== %v ======\n", stat.Now)
			fmt.Printf("Total Hits: %d, Hit statistics:\n\t2xx: %d, 3xx: %d, 4xx: %d, 5xx: %d\n",
				stat.Hits, stat.StatusCodes[2], stat.StatusCodes[3],
				stat.StatusCodes[4], stat.StatusCodes[5])
			fmt.Printf("Sections with most hits so far (%d):\n", len(stat.TopHits))
			for _, th := range stat.TopHits {
				fmt.Println(th)
			}
			fmt.Printf("\n")
		}
	}

}

package routers

import (
	"fmt"
)

type RouterId uint

type Template [][]RouterId

type Envelope struct {
	Dest    RouterId
	Hops    uint
	Message interface{}
}

func hasLink(routers []RouterId, id RouterId) bool {
	for _, node := range routers {
		if id == node {
			return true
		}
	}
	return false
}

func MakeRouters(t Template, logLevel string, printCons bool) (in []chan<- interface{}, out <-chan Envelope) {
	channels := make([]chan interface{}, len(t))
	framework := make(chan Envelope)

	in = make([]chan<- interface{}, len(t))
	for i := range channels {
		channels[i] = make(chan interface{})
		in[i] = channels[i]
	}
	out = framework

	if printCons {
		fmt.Println("+-------------------+")
		fmt.Println("|  Connection Type  |")
		fmt.Println("| 0 = No connection |")
		fmt.Println("| 1 = Connection    |")
		fmt.Println("| * = Self          |")
		fmt.Println("+-------------------+")
		fmt.Println()
		indexes := "  |"
		line := "--|"
		for i := 0; i < len(t); i++ {
			if i < 10 {
				indexes += fmt.Sprintf(" %v ", i)
			} else {
				indexes += fmt.Sprintf(" %v", i)
			}
			line += "---"
		}
		fmt.Println(indexes)
		fmt.Println(line)
	}
	for routerId, neighbourIds := range t {
		if printCons {
			line := ""
			if routerId < 10 {
				line += fmt.Sprintf("%v |", routerId)
			} else {
				line += fmt.Sprintf("%v|", routerId)
			}
			for i := 0; i < len(t); i++ {
				if i == routerId {
					line += " * "
					continue
				}
				if hasLink(neighbourIds, RouterId(i)) {
					line += " 1 "
				} else {
					line += " 0 "
				}
			}
			fmt.Println(line)
		}
		neighbours := make([]chan<- interface{}, len(neighbourIds))
		for i, id := range neighbourIds {
			neighbours[i] = channels[id]
		}

		go Router(RouterId(routerId), channels[routerId], neighbours, framework, logLevel)
	}

	return
}

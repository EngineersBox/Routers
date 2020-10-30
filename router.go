package routers

import (
	"fmt"
	"log"
	"math/rand"
)

// #### CONSTANTS ####

const hmsTime string = "03:04:05.999"

// #### ROUTERS ####

// Routers ... An array of RouterId
type Routers []RouterId

func (r Routers) contains(id RouterId) bool {
	for _, b := range r {
		if b == id {
			return true
		}
	}
	return false
}

// #### MESSAGES ####

// TopologyUpdate ... In order record of routers visited when initialising
type TopologyUpdate struct {
	ID   UUID
	IP   IPv4
	Path Routers
}

// NeighbourUpdate ... Pass the ID of the self to neighbours
type NeighbourUpdate struct {
	ID     RouterId
	ChanID <-chan interface{}
}

// NeighbourMap ... Mapping of RouterId to local channel index
type NeighbourMap map[RouterId]int

func (n NeighbourMap) getAllNotIn(values Routers) []int {
	returnValue := make([]int, 0)
	for r, i := range n {
		if !values.contains(r) {
			returnValue = append(returnValue, i)
		}
	}
	return returnValue
}

// #### MESSAGE PROCESSORS ####

//--- NeighbourUpdate ----
func processNeighbourUpdate(logLevel string, msg NeighbourUpdate, neighbours []chan<- interface{}, networkAddress IPv4, NMap NeighbourMap) {
	for i, channel := range neighbours {
		if fmt.Sprint(channel) == fmt.Sprint(msg.ChanID) {
			if logLevel == "verbose" {
				log.Printf("[%v] Updating channel mapping [%v] <~> [%v]",
					networkAddress.toString(false),
					msg.ID,
					i)
			}
			NMap[msg.ID] = i
			break
		}
	}
}

// ---- Envelope ----

// forwardEnvelope ... Calculate the shortest path to the destination and forward the message to the next router in the path
func forwardEnvelope(logLevel string, msg Envelope, RoutingTable DVRTable, self RouterId, networkAddress IPv4, neighbours []chan<- interface{}, NMap NeighbourMap, raw interface{}, RouterIPAddress IPv4) {
	msg.Hops++
	newPath := make(Path, 0)
	// Calculate the shortest path from the current node to the destination
	next := AsyncShortestPath(RoutingTable, self, msg.Dest, newPath)
	if len(next) > 0 {
		if logLevel != "none" {
			log.Printf("[%v] Found shortest path %v",
				networkAddress.toString(false),
				next)
		}
		if logLevel != "none" {
			log.Printf("| >> [%s] ~ [%v] {Envelope: %v} Forwarding to neighbours..",
				networkAddress.toString(false),
				neighbours[NMap[next[1]]],
				&raw)
		}
		// Send that to the next router in the path (index 0 is self)
		neighbours[NMap[next[1]]] <- msg
		return
	}
	// If there was no path (network not mapped deep enough)
	// then send to a random neighbour and send a new network mapping message
	nextHop := rand.Intn(len(neighbours))
	if logLevel != "none" {
		log.Printf("[%v] Shortest path not found, routing to random neighbour: %v",
			networkAddress.toString(false),
			nextHop)
		log.Printf("| >> [%s] ~ [%v] {Envelope: %v} Forwarding to neighbours..",
			networkAddress.toString(false),
			neighbours[nextHop],
			&raw)
	}
	// Send that to the next router in the path (index 0 is self)
	go func() {
		neighbours[nextHop] <- msg
	}()

	// Send new network mapping message
	_, nextHost := RouterIPAddress.firstHostID()
	CurrPath := make(Routers, 0)
	CurrPath = append(CurrPath, self)
	uuid, _ := uuid4()
	sendTopologyUpdate(logLevel, networkAddress, uuid, nextHost, CurrPath, neighbours[rand.Intn(len(neighbours))])
}

func processEnvelope(logLevel string, msg Envelope, self RouterId, framework chan<- Envelope, networkAddress IPv4, incoming <-chan interface{}, raw interface{}, RoutingTable DVRTable, neighbours []chan<- interface{}, NMap NeighbourMap, RouterIPAddress IPv4) {
	if msg.Dest == self {
		if logLevel != "none" {
			log.Printf("| << [%v] ~ [%v] {Envelope: %v} --TERMINATED-- HOPS: %v",
				networkAddress.toString(false),
				incoming,
				&raw,
				msg.Hops)
		}
		framework <- msg
	} else {
		forwardEnvelope(logLevel, msg, RoutingTable, self, networkAddress, neighbours, NMap, raw, RouterIPAddress)
	}
}

// ---- TopologyUpdate ----

// updateNeighboursSlidingWindow ... Create the link between routers in the routing DVRTable
func updateNeighboursSlidingWindow(logLevel string, msg TopologyUpdate, networkAddress IPv4, RoutingTable DVRTable) {
	for i := 0; i < len(msg.Path)-1; i++ {
		if logLevel == "verbose" {
			log.Printf("[%v] Updating DVRTable with router link [%v] <=> [%v]",
				networkAddress.toString(false),
				msg.Path[i],
				msg.Path[i+1])
		}
		// Dual pairings, ensure as a symmetric matrix (A^T = A)
		RoutingTable.put(msg.Path[i], msg.Path[i+1], 1)
		RoutingTable.put(msg.Path[i+1], msg.Path[i], 1)
	}
}

// forwardPathMsg ... Push the message to all neighbours to mirror the path through the network
func forwardPathMsg(logLevel string, msg TopologyUpdate, self RouterId, networkAddress IPv4, neighbours []chan<- interface{}, RouterIPAddress IPv4, NMap NeighbourMap) {
	msg.Path = append(msg.Path, self)
	validToSend := NMap.getAllNotIn(msg.Path)
	if len(validToSend) == 0 {
		if logLevel != "none" {
			log.Printf("[%s] Topology update invalidated [%v]",
				networkAddress.toString(false),
				msg.ID)
		}
		return
	}
	if logLevel == "verbose" {
		log.Printf("[%s] Forwarding topology update to neighbours...",
			networkAddress.toString(false))
	}
	for _, i := range validToSend {
		if logLevel == "verbose" {
			log.Printf("| >> [%s] ~ [%v] {TopologyUpdate: %v}",
				RouterIPAddress.toString(false),
				neighbours[i],
				msg.ID)
		}
		// Forward the message to all neighbours
		go func(ns *[]chan<- interface{}, idx int, ms TopologyUpdate) {
			(*ns)[idx] <- ms
		}(&neighbours, i, msg)
	}
}

func processPathMsg(logLevel string, self RouterId, networkAddress IPv4, RouterIPAddress IPv4, msg TopologyUpdate, RoutingTable DVRTable, neighbours []chan<- interface{}, NMap NeighbourMap) {
	if logLevel == "verbose" {
		log.Printf("[%v] Processing topology update [%v] <- {%v}",
			networkAddress.toString(false),
			msg.ID,
			msg.IP.toString(false))
	}
	if len(msg.Path) < 2 {
		// Single router ID in the TopologyUpdate, update the pairing with self ID
		RoutingTable.put(self, msg.Path[0], 1)
		RoutingTable.put(msg.Path[0], self, 1)
	} else {
		// Multiple router IDs in the TopologyUpdate, update as sliding window pairings
		updateNeighboursSlidingWindow(logLevel, msg, networkAddress, RoutingTable)
	}
	if !msg.Path.contains(self) {
		// If this is the first time the message has visited here, re-send to neighbours
		forwardPathMsg(logLevel, msg, self, networkAddress, neighbours, RouterIPAddress, NMap)
	} else {
		if logLevel == "verbose" {
			log.Printf("[%s] Topology update invalidated by cyclic traversal: [%v] %v",
				networkAddress.toString(false),
				msg.ID,
				msg.Path)
		}
	}
}

// #### ROUTER IMPLEMENTATION ####

func sendTopologyUpdate(logLevel string, networkAddress IPv4, newID UUID, nextHost IPv4, CurrPath Routers, neighbour chan<- interface{}) {
	if logLevel != "none" {
		log.Printf("[%v] Sending local topology update... [%v] -> {%v}",
			networkAddress.toString(false),
			newID,
			nextHost.toString(false))
	}
	// Update the neighbours with pathing and address info
	go func(ns *chan<- interface{}) {
		(*ns) <- TopologyUpdate{
			ID:   newID,
			IP:   networkAddress,
			Path: CurrPath,
		}
	}(&neighbour)
}

// mapNetwork ... Start network mapping of connections via TopologyUpdate and update mapping of neighbour channels with NeighbourUpdate
func mapNetwork(logLevel string, neighbours []chan<- interface{}, self RouterId, nextHost IPv4, networkAddress IPv4, incoming <-chan interface{}) {
	for i, n := range neighbours {
		CurrPath := make(Routers, 0)
		CurrPath = append(CurrPath, self)
		newID, _ := uuid4()

		// Update fouth quadrant of address with subnet reference
		nextHost.Quad4 += uint8(i)

		// Update the neighbours with the channel ID mapping
		if logLevel != "none" {
			log.Printf("[%v] Updating neighbours with router ID... {%v} -> {%v}",
				networkAddress.toString(false),
				incoming,
				n)
		}
		go func(selfID RouterId, ns chan<- interface{}, next <-chan interface{}) {
			ns <- NeighbourUpdate{
				selfID,
				next,
			}
		}(self, n, incoming)

		sendTopologyUpdate(logLevel, networkAddress, newID, nextHost, CurrPath, n)
	}
}

// Router ... Implementation of modified DVR based network router and/or switch.
//
// -- Features --
// - IPv4 subranging for neighbouring nodes
// - Dyamic shortest path
// - Support for dropouts with periodic updates
func Router(self RouterId, incoming <-chan interface{}, neighbours []chan<- interface{}, framework chan<- Envelope, logLevel string) {
	// Assign a new local network IP with subnet range poer of 2 encapsulating all neighbours
	RouterIPAddress := randomIPv4FromSubetSize(len(neighbours) + 1)
	_, networkAddress := RouterIPAddress.networkID()
	RoutingTable := make(DVRTable, 0)
	NMap := make(NeighbourMap, len(neighbours))

	if logLevel == "verbose" {
		log.Printf("[HOST: %v] -> Assigning CIDR block %v {Addresses: %v}",
			networkAddress.toString(false),
			RouterIPAddress.toString(true),
			len(neighbours)+1)
	}

	_, nextHost := RouterIPAddress.firstHostID()

	mapNetwork(logLevel, neighbours, self, nextHost, networkAddress, incoming)

	if logLevel == "verbose" {
		log.Printf("[%v] Sent local topology update to %v neighbours",
			networkAddress.toString(false),
			len(neighbours))
	}

	for {
		select {
		case raw := <-incoming:
			switch msg := raw.(type) {
			case Envelope:
				processEnvelope(logLevel, msg, self, framework, networkAddress, incoming, raw, RoutingTable, neighbours, NMap, RouterIPAddress)
			case NeighbourUpdate:
				processNeighbourUpdate(logLevel, msg, neighbours, networkAddress, NMap)
			case TopologyUpdate:
				processPathMsg(logLevel, self, networkAddress, RouterIPAddress, msg, RoutingTable, neighbours, NMap)
			default:
				log.Printf("[%v] received unexpected message %g\n", self, msg)
			}
		}
	}
}

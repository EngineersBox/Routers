# Routers

---

## Overview

Implementing a network of routers with arbitrary topology requires dynamic identification ofnetwork structure and neighbour mappings. This implementation is based around DistanceVector Routing (DVR) with local only tables. Each of the routers implements a IPv4 dynamicaddress for managing subnet changes locally, and also re-pathing for dynamic topologychanges.

## Message Types

In order to communicate network structure, self identification and actual message passing, thereare three types of messages implemented. 

Firstly, is the Topology Update, which is designed tostore the path the message took whilst travelling around the network. At each node, the pathingis updated in the DVR table and forwarded to neighbours. It is invalidated when it reaches acycle, such that the current node has been previously visited and is in the path list.

Second type of message is the Neighbour Update. Here, the routers send this message to theirneighbours to identify which channel maps to the connection between them, and the ID thatshould correspond. Without circulating this message, routers cannot identify which of their neighbours to send a given message to when finding the shortest path.

The last type of message is the Envelope, provided alongside the framework. No modificationsare made to this message and it is purely for sending through the network until reaching itsstored destination

## Network Mapping

In order to route messages, knowledge of the network topology is required. Routers utilise theTopologyUpdate​ message to spread through each path starting from themselves in order tomap the network connections.

In order to reduce cyclic redundancy, at each branching traversal, the message is passed onlyto routers who haven’t yet received the message. It is important to note that since every router is passing these messages this reduces the stress to the network but also doesn’t lose theguarantee of each router having knowledge of the network (assuming all routers have at leastone connection.

Below is an example network with the topology update pathing tree for each message split.Note the lack of looping in the tree, guaranteeing acyclic pathing until invalidation, at which pointthere is no alternative.

![images/Topology%20Update.svg](images/Topology%20Update.svg)

## Identifying Neighbours

In order to send messages to routers by ID, the router needs to map the channel addresses torouter IDs. This is the job of the ​NeighbourUpdate​ message, which contains the incomingchannel address and the routers ID. When another router receives this message it checks its listof channels and finds the matching one, which tells it the mapping of router ID to channel, thuscompleting the indexing.

In this implementation of routing, the channel IDs act kind of like MAC addresses (of differentformat), as they identify physical existence of network destinations. However, the usage ofthese channels is not like that of MAC addresses in typical routing implementations.

## Pathfinding

It’s all well and good to know what the network looks like, but without being able to traverse it, itbecomes redundant. Here Dijkstra’s shortest path algorithm is used to path through the mappednetwork for a given destination. Note the efficiency of this algorithm drops with larger quantities of routers, however for most networks it is sufficient.

## CIDR Block Addressing

Each router is assigned a random dynamic IPv4 address at startup, and a CIDR prefix based onthe amount of neighbours it has. Using classless subnets allows for immediate identification of neighbouring nodes and also relative addressing changes based on topology changes.Using the CIDR prefix, routing messages within a given subnet becomes a matter of deterministic connectivity, and also provides instantaneous invalidation of the current subnetprefix. Given any changes, a recalculation can be done in one of three ways:

* Static addressing
* Dynamic addressing
* DNS configuration

This implementation uses dynamic addressing as it was the happy medium of the three,however not complete as you will note with the inconsistent subnet overlap. Implementing DNSservers would have been optimal, but time constraints permitted otherwise.

![images/CIDR%20subnets.svg](images/CIDR%20subnets.svg)

### Static Addressing

If the host address is required to be static, then a re-assignment can be made with a subsequent re-calculation of CIDR prefixes and updated to the supernet. However, this can be costly and should be reserved for full topological supernet changes.

### Dynamic Addressing

If the host address is dynamic then the subnet can be reconfigured and an update to the CIDR prefix can be made to reflect these changes. This allows the global supernet addressing toremain intact whilst having localised changes propagate as and when needed. This does notaffect the ability to send and receive messages.

### DNS configuration

In the last situation, a DNS server sits as an intermediary between the supernet and subnet. Here we can make localised addressing changes and update the DNS records on the fly without losing actively queued messages or currently processing messages. A downside to this is that it does have a small dropout time when records are switched over, which can create isolated messages in a dead channel. However, this is extremely rare with DNS updates being microseconds at worst in this case.

### Router Processing Sequence

![images/Routers%20sequence.svg](images/Routers%20sequence.svg)
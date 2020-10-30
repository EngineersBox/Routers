package routers

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ---- CLASSLESS IPv4 ----

// IPv4 ... Implementation of IPv4 with classless subnets using CIDR blocks
type IPv4 struct {
	Quad1  uint8
	Quad2  uint8
	Quad3  uint8
	Quad4  uint8
	Prefix uint
}

// randomIPv4 ... Generate a random IPv4 with CIDR prefix of 0
func randomIPv4() IPv4 {
	return randomIPv4WithPrefix(0)
}

// randomIPv4WithPrefix ... Generate a random IPv4 address with a given CIDR prefix
func randomIPv4WithPrefix(prefix uint) IPv4 {
	rand.Seed(time.Now().UnixNano())
	return IPv4{
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		prefix,
	}
}

var ln2 = math.Log(2)

// ipCountToPrefix ... Calculate the minimum CIDR prefix to house a subnet of size {ipCount}
func ipCountToPrefix(ipCount int) uint {
	return uint(0 - ((math.Log(float64(ipCount)) - (32 * ln2)) / ln2))
}

// randomIPv4FromSubnetSize ... Given a subnet size, generate a random IPv4 address
func randomIPv4FromSubetSize(subnetSize int) IPv4 {
	return randomIPv4WithPrefix(ipCountToPrefix(subnetSize))
}

// toString ... Convert the IPv4 address to a string, conditionally showing CIDR prefix
func (ip IPv4) toString(printPrefix bool) string {
	if printPrefix {
		return fmt.Sprintf("%v.%v.%v.%v/%v", ip.Quad1, ip.Quad2, ip.Quad3, ip.Quad4, ip.Prefix)
	}
	return fmt.Sprintf("%v.%v.%v.%v", ip.Quad1, ip.Quad2, ip.Quad3, ip.Quad4)
}

// addressCountForSubnet ... Calculate the amount of addresses in the provided subnet
func addressCountForSubnet(subnet uint) float64 {
	return math.Pow(2, float64(32-subnet))
}

// networkID ... calculate the network ID based on the fourth quadrant and CIDR prefix
// Formula: floor(Host Address/Subnet Number of Hosts) * Subnet Number of Hosts
func (ip IPv4) networkID() (a uint, b IPv4) {
	if ip.Prefix == 0 {
		return uint(ip.Quad4), ip
	}
	addrCount := addressCountForSubnet(ip.Prefix)
	networkMultiplier := float64(ip.Quad4) / addrCount
	nextQuad := math.Floor(networkMultiplier) * addrCount
	return uint(nextQuad), IPv4{ip.Quad1, ip.Quad2, ip.Quad3, uint8(nextQuad), 32}
}

// broadcastID ... Calculate the broadcast ID (to all subnet nodes) based on network ID and CIDR prefix
// Formula: Host ID + (Subnet Number of Hosts-1)
func (ip IPv4) broadcastID() (uint, IPv4) {
	if ip.Prefix == 0 {
		return uint(ip.Quad4), ip
	}
	addrCount := uint(addressCountForSubnet(ip.Prefix))
	nwID, _ := ip.networkID()
	broadID := nwID + (addrCount - 1)
	return broadID, IPv4{ip.Quad1, ip.Quad2, ip.Quad3, uint8(broadID), 32}
}

// firstHostID ... Calculate the first subnet host address based on the network ID
// Formula: Network ID + 1
func (ip IPv4) firstHostID() (uint, IPv4) {
	if ip.Prefix == 0 {
		return uint(ip.Quad4), ip
	}
	nwID, _ := ip.networkID()
	return nwID + 1, IPv4{ip.Quad1, ip.Quad2, ip.Quad3, uint8(nwID + 1), 32}
}

// lastHostID ... Calculate the first subnet host address based on the broadcast ID
// Formula: Broadcast ID - 1
func (ip IPv4) lastHostID() (uint, IPv4) {
	if ip.Prefix == 0 {
		return uint(ip.Quad4), ip
	}
	broadID, _ := ip.broadcastID()
	return broadID - 1, IPv4{ip.Quad1, ip.Quad2, ip.Quad3, uint8(broadID - 1), 32}
}

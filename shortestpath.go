package routers

import "sync"

// Path ... Ordered record of pathing traversals
type Path []RouterId

// hasVisited ... Check if a given router has already been visited
func (p Path) hasVisited(id RouterId) bool {
	for _, node := range p {
		if id == node {
			return true
		}
	}
	return false
}

// AsyncShortestPath ... Finds the shortest path through a network with Dijkstra's algorithm
func AsyncShortestPath(table DVRTable, start RouterId, end RouterId, path Path) []RouterId {
	if table.getRow(start) == nil || table.getRow(end) == nil {
		return path
	}
	path = append(path, start)
	if start == end {
		// End of the line!
		return path
	}
	shortest := make([]RouterId, 0)
	// New WaitGroup to prevent exit until all goroutines terminate
	var wg sync.WaitGroup
	for idx, con := range table.getRow(start) {
		if path.hasVisited(idx) {
			continue
		}
		if con.(int) == 0 {
			continue
		}
		wg.Add(1)
		// Explore the neighbouring connections
		go func(g DVRTable, s RouterId, e RouterId, p Path, sp *[]RouterId, wg *sync.WaitGroup) {
			defer wg.Done()
			newPath := AsyncShortestPath(g, s, e, p)
			if len(newPath) <= 0 {
				// No path, exit
				return
			}
			// If there is a new path, check if its better than the current
			currentPath := len(*sp)
			if currentPath == 0 || (len(newPath) < currentPath) {
				// Update the current best path with the new best path
				(*sp) = newPath
			}
		}(table, idx, end, path, &shortest, &wg)
	}
	wg.Wait()
	return shortest
}

package game

import (
	"container/heap"
	"math"
)

// An ItemPath represents an A* node
type ItemPath struct {
	pos  Position
	priority float64
	index    int
}

type PriorityQueue []*ItemPath

func (pq PriorityQueue) Len() int           { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*ItemPath)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// FindPath executes an A* pathfinding algorithm on the map grid.
func FindPath(gm *GameMap, start, goal Position) []Position {
	if start == goal {
		return []Position{start}
	}

	// Check bounds
	if goal.X < 0 || goal.X >= gm.Width || goal.Y < 0 || goal.Y >= gm.Height {
		return nil
	}

	// Goal must not be impassable (like mountains or walls, unless we are constructing/mining them, in which case we target adjacent cells)
	if !IsWalkable(gm, goal) {
		// Try finding adjacent walkable cell
		adj := GetAdjacent(gm, goal)
		bestDist := math.MaxFloat64
		bestAdj := start
		found := false
		for _, a := range adj {
			if IsWalkable(gm, a) {
				dist := heuristic(a, goal)
				if dist < bestDist {
					bestDist = dist
					bestAdj = a
					found = true
				}
			}
		}
		if found {
			goal = bestAdj
		} else {
			return nil
		}
	}

	frontier := &PriorityQueue{}
	heap.Init(frontier)
	heap.Push(frontier, &ItemPath{pos: start, priority: 0})

	cameFrom := make(map[Position]Position)
	costSoFar := make(map[Position]float64)

	cameFrom[start] = start
	costSoFar[start] = 0

	for frontier.Len() > 0 {
		current := heap.Pop(frontier).(*ItemPath).pos

		if current == goal {
			break
		}

		for _, next := range GetAdjacent(gm, current) {
			if !IsWalkable(gm, next) && next != goal {
				continue
			}

			// Movement cost can vary by terrain
			newCost := costSoFar[current] + GetMovementCost(gm, next)
			if priorCost, exists := costSoFar[next]; !exists || newCost < priorCost {
				costSoFar[next] = newCost
				priority := newCost + heuristic(next, goal)
				heap.Push(frontier, &ItemPath{pos: next, priority: priority})
				cameFrom[next] = current
			}
		}
	}

	// Reconstruct path
	if _, exists := cameFrom[goal]; !exists {
		return nil
	}

	path := make([]Position, 0)
	curr := goal
	for curr != start {
		path = append([]Position{curr}, path...)
		curr = cameFrom[curr]
	}
	return path
}

func heuristic(a, b Position) float64 {
	return math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y))
}

// GetAdjacent returns neighboring orthogonal and diagonal positions.
func GetAdjacent(gm *GameMap, pos Position) []Position {
	adj := make([]Position, 0, 8)
	dirs := []Position{
		{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0},
		{X: -1, Y: -1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: 1, Y: 1},
	}
	for _, d := range dirs {
		nx, ny := pos.X+d.X, pos.Y+d.Y
		if nx >= 0 && nx < gm.Width && ny >= 0 && ny < gm.Height {
			adj = append(adj, Position{X: nx, Y: ny})
		}
	}
	return adj
}

// IsWalkable returns true if a tile is passable
func IsWalkable(gm *GameMap, pos Position) bool {
	if pos.X < 0 || pos.X >= gm.Width || pos.Y < 0 || pos.Y >= gm.Height {
		return false
	}
	tile := gm.Grid[pos.Y][pos.X]
	if tile.Type == TileMountain || tile.Type == TileWater {
		return false
	}
	if b, ok := gm.Buildings[pos]; ok {
		if b.Type == BuildingWall && !b.IsBlueprint {
			return false
		}
	}
	return true
}

// GetMovementCost returns the pathfinding cost multiplier for the terrain/structures
func GetMovementCost(gm *GameMap, pos Position) float64 {
	tile := gm.Grid[pos.Y][pos.X]
	cost := 1.0
	if tile.Type == TileStonySoil {
		cost = 1.3
	} else if tile.Type == TileSand {
		cost = 1.5
	}
	if b, ok := gm.Buildings[pos]; ok {
		if b.Type == BuildingSandbag {
			cost = 2.5
		} else if b.Type == BuildingDoor {
			cost = 1.1
		}
	}
	return cost
}

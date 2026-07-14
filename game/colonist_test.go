package game

import "testing"

func TestPathfinding(t *testing.T) {
	gm := GenerateMap(10, 10)
	// Clear path from 1,1 to 8,8
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gm.Grid[y][x].Type = TileSoil
		}
	}
	start := Position{X: 1, Y: 1}
	goal := Position{X: 8, Y: 8}
	path := FindPath(gm, start, goal)
	if len(path) == 0 {
		t.Error("Expected to find valid path across clear terrain")
	}

	// Block paths completely
	for y := 0; y < 10; y++ {
		gm.Grid[y][5].Type = TileMountain
	}
	blockedPath := FindPath(gm, start, goal)
	if len(blockedPath) > 0 {
		// Verify if it finds adjacent or completely fails
		for _, step := range blockedPath {
			if step == goal {
				t.Error("Should not reach goal through completely blocked mountain wall")
			}
		}
	}
}

func TestColonistNeeds(t *testing.T) {
	gm := GenerateMap(10, 10)
	colonist := gm.Colonists[0]
	colonist.Hunger = 50.0
	colonist.Rest = 50.0
	UpdateColonistNeeds(gm, colonist)
	if colonist.Hunger >= 50.0 {
		t.Errorf("Expected hunger to decay from 50, got %f", colonist.Hunger)
	}
}

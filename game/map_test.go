package game

import "testing"

func TestGenerateMap(t *testing.T) {
	gm := GenerateMap(50, 30)
	if gm.Width != 50 || gm.Height != 30 {
		t.Errorf("Expected 50x30 map, got %dx%d", gm.Width, gm.Height)
	}
	if len(gm.Colonists) != 3 {
		t.Errorf("Expected 3 starting colonists, got %d", len(gm.Colonists))
	}
	if len(gm.Items) == 0 {
		t.Error("Expected starting items to be spawned")
	}
}

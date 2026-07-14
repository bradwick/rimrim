package game

import "testing"

func TestDispatchAndExecuteJobs(t *testing.T) {
	gm := GenerateMap(20, 20)
	c := gm.Colonists[0]

	// Trigger emergency Hunger Eat Job
	c.Hunger = 5.0
	// Make sure starting Simple Meal is on the ground
	mealPos := Position{X: 10, Y: 10}
	spawnItem(gm, ItemMealSimple, mealPos, 5)

	DispatchJobs(gm)

	if c.CurrentJob == nil || c.CurrentJob.Type != JobEat {
		t.Errorf("Expected colonist to select Eat job under starving conditions, got %+v", c.CurrentJob)
	}

	// Trigger blueprint construction job
	c.Hunger = 100.0 // reset hunger
	c.CurrentJob = nil
	bpPos := Position{X: 5, Y: 5}
	gm.Buildings[bpPos] = &Building{
		ID:           1,
		Type:         BuildingWall,
		Pos:          bpPos,
		IsBlueprint:  true,
		WorkRequired: 10,
		WorkLeft:     10,
	}

	DispatchJobs(gm)
	if c.CurrentJob == nil || c.CurrentJob.Type != JobConstruct {
		t.Errorf("Expected colonist to select Construct job, got %+v", c.CurrentJob)
	}

	// Move directly to target of JobConstruct
	c.Pos = bpPos
	c.CurrentJob.PathIdx = len(c.CurrentJob.Path) // fake arrival

	ExecuteJobs(gm)
	if gm.Buildings[bpPos].WorkLeft != 9 {
		t.Errorf("Expected blueprint WorkLeft to decrement from 10 to 9, got %d", gm.Buildings[bpPos].WorkLeft)
	}
}

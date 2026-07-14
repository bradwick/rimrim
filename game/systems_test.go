package game

import "testing"

func TestPowerSystem(t *testing.T) {
	gm := GenerateMap(10, 10)

	// Build a generator
	genPos := Position{X: 1, Y: 1}
	gm.Buildings[genPos] = &Building{
		Type:         BuildingGenerator,
		Pos:          genPos,
		IsBlueprint:  false,
		PowerOn:      true,
		PowerOutput:  1000,
		Fuel:         50.0,
	}

	// Build a cooler consumer
	coolerPos := Position{X: 2, Y: 2}
	gm.Buildings[coolerPos] = &Building{
		Type:         BuildingCooler,
		Pos:          coolerPos,
		IsBlueprint:  false,
		PowerOn:      true,
		PowerOutput:  -200,
	}

	UpdatePower(gm)
	if !gm.Buildings[coolerPos].PowerOn {
		t.Error("Expected consumer to have power from generator")
	}

	// Remove generator fuel
	gm.Buildings[genPos].Fuel = 0
	UpdatePower(gm)
	if gm.Buildings[coolerPos].PowerOn {
		t.Error("Expected consumer to shut down without power fuel")
	}
}

func TestTemperatureAndSpoilage(t *testing.T) {
	gm := GenerateMap(10, 10)

	// Warm ambient temperature
	gm.OutdoorTemp = 25.0
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			gm.Grid[y][x].Temperature = 25.0
			gm.Grid[y][x].Roofed = true
		}
	}

	itemPos := Position{X: 3, Y: 3}
	gm.Items[itemPos] = &Item{
		Type:     ItemMealSimple,
		Qty:      10,
		SpoilsIn: 1.0, // 1 hour left
	}

	// Run multiple updates (60 iterations = 1 game hour equivalent)
	for i := 0; i < 60; i++ {
		UpdateTemperature(gm)
	}

	if _, ok := gm.Items[itemPos]; ok {
		t.Error("Expected food item to have rotted and deleted after warm heat simulation")
	}
}

package game

import (
	"math"
)

// UpdatePower recalculates generators, storage batteries, and consumer load.
func UpdatePower(gm *GameMap) {
	// Let's identify the total generated power vs total power load
	totalGen := 0
	totalLoad := 0
	batteries := make([]*Building, 0)

	isDaytime := gm.Hour >= 6 && gm.Hour <= 18

	for _, b := range gm.Buildings {
		if b.IsBlueprint {
			continue
		}

		// Solar panels output depends on daytime
		if b.Type == BuildingSolarPanel {
			if isDaytime {
				b.PowerOutput = 150
			} else {
				b.PowerOutput = 0
			}
		}

		// Fueled generators require wood fuel
		if b.Type == BuildingGenerator {
			if b.Fuel > 0 {
				b.PowerOutput = 1000
				b.Fuel -= 0.05 // burn rate per hour
			} else {
				b.PowerOutput = 0
			}
		}

		if b.PowerOutput > 0 {
			totalGen += b.PowerOutput
		} else if b.PowerOutput < 0 {
			if b.PowerOn {
				totalLoad += -b.PowerOutput // turn positive
			}
		}

		if b.Type == BuildingBattery {
			batteries = append(batteries, b)
		}
	}

	netPower := totalGen - totalLoad

	// Handle battery charge/discharge
	if len(batteries) > 0 {
		if netPower > 0 {
			// Excess power charging batteries
			chargePerBattery := netPower / len(batteries)
			for _, bat := range batteries {
				bat.BatteryStored += chargePerBattery
				if bat.BatteryStored > bat.BatteryMax {
					bat.BatteryStored = bat.BatteryMax
				}
			}
		} else if netPower < 0 {
			// Deficit power draining batteries
			drainPerBattery := (-netPower) / len(batteries)
			totalDrained := 0
			for _, bat := range batteries {
				if bat.BatteryStored >= drainPerBattery {
					bat.BatteryStored -= drainPerBattery
					totalDrained += drainPerBattery
				} else {
					totalDrained += bat.BatteryStored
					bat.BatteryStored = 0
				}
			}

			// If battery power is empty and deficit continues, some systems turn off
			if totalDrained < (-netPower) {
				// Shut down non-essential consumers
				for _, b := range gm.Buildings {
					if b.PowerOutput < 0 {
						b.PowerOn = false
					}
				}
			} else {
				// Kept alive by batteries
				for _, b := range gm.Buildings {
					if b.PowerOutput < 0 {
						b.PowerOn = true
					}
				}
			}
		}
	} else {
		// No battery buffer: total power load must be <= total generator production
		if netPower < 0 {
			for _, b := range gm.Buildings {
				if b.PowerOutput < 0 {
					b.PowerOn = false
				}
			}
		} else {
			for _, b := range gm.Buildings {
				if b.PowerOutput < 0 {
					b.PowerOn = true
				}
			}
		}
	}
}

// UpdateTemperature runs heat transfer simulation across walls/conduits/adjacent cells.
func UpdateTemperature(gm *GameMap) {
	// A simple cellular automata heat simulation
	// 1. Heaters and Coolers modify local tile temperatures directly
	for _, b := range gm.Buildings {
		if b.IsBlueprint || !b.PowerOn {
			continue
		}

		currentTemp := gm.Grid[b.Pos.Y][b.Pos.X].Temperature
		if b.Type == BuildingHeater {
			if currentTemp < b.TargetTemp {
				gm.Grid[b.Pos.Y][b.Pos.X].Temperature += 3.5
			}
		} else if b.Type == BuildingCooler {
			if currentTemp > b.TargetTemp {
				gm.Grid[b.Pos.Y][b.Pos.X].Temperature -= 4.0
			}
		}
	}

	// 2. Diffuse temperature across adjacent grid tiles
	nextTempGrid := make([][]float64, gm.Height)
	for y := 0; y < gm.Height; y++ {
		nextTempGrid[y] = make([]float64, gm.Width)
		for x := 0; x < gm.Width; x++ {
			nextTempGrid[y][x] = gm.Grid[y][x].Temperature
		}
	}

	for y := 0; y < gm.Height; y++ {
		for x := 0; x < gm.Width; x++ {
			currTile := gm.Grid[y][x]
			neighbors := GetAdjacent(gm, currTile.Pos)
			sumTemp := 0.0
			wallCount := 0

			for _, n := range neighbors {
				// Walls slow down temperature transfer (isolation)
				if b, ok := gm.Buildings[n]; ok && b.Type == BuildingWall {
					sumTemp += gm.Grid[n.Y][n.X].Temperature * 0.1
					wallCount++
				} else {
					sumTemp += gm.Grid[n.Y][n.X].Temperature
				}
			}

			// Add interaction with outdoors temperature
			if !currTile.Roofed {
				// Outdoor temperature has massive direct influence if no roof is present
				nextTempGrid[y][x] = currTile.Temperature + (gm.OutdoorTemp-currTile.Temperature)*0.45
			} else {
				// Roofed: slow decay to outdoor, faster transfer to internal adjacent tiles
				totalWeights := float64(len(neighbors) - wallCount) + float64(wallCount)*0.1
				if totalWeights > 0 {
					avgNeighbourTemp := sumTemp / totalWeights
					nextTempGrid[y][x] = currTile.Temperature + (avgNeighbourTemp-currTile.Temperature)*0.15
					// roofed isolation loss to outdoors
					nextTempGrid[y][x] += (gm.OutdoorTemp - nextTempGrid[y][x]) * 0.015
				}
			}
		}
	}

	// Apply updated values
	for y := 0; y < gm.Height; y++ {
		for x := 0; x < gm.Width; x++ {
			gm.Grid[y][x].Temperature = nextTempGrid[y][x]
		}
	}

	// 3. Spoilage mechanics for stored food based on local tile temperatures
	for pos, item := range gm.Items {
		if item.SpoilsIn > 0 {
			temp := gm.Grid[pos.Y][pos.X].Temperature
			if temp > 0.0 {
				// Warm temp speeds up spoil rate
				spoilSpeed := 1.0
				if temp > 20.0 {
					spoilSpeed = 2.0
				}
				item.SpoilsIn -= (1.0 / 60.0) * spoilSpeed // hours decreased per minute tick
				if item.SpoilsIn <= 0 {
					delete(gm.Items, pos)
					gm.Log(string(item.Type) + " has rotted due to heat.")
				}
			} else {
				// Frozen: spoilage is completely halted!
				item.SpoilsIn = math.Max(item.SpoilsIn, 1.0)
			}
		}
	}
}

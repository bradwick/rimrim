package tui

import (
	"fmt"
	"os"
	"rimworld_tui/game"
)

import (
	"strings"
)

// ProcessRawInput processes string input including keyboard and SGR mouse clicks.
func ProcessRawInput(input string, gm *game.GameMap, cursor *game.Position, activeMenu *string, selectedCol *int, renderMode *string) {
	// 1. Handle SGR Mouse Sequences: e.g. "\033[<0;24;12M" or "\033[<0;24;12m"
	if strings.HasPrefix(input, "\033[<") {
		parts := strings.Split(strings.TrimRight(input, "Mm"), ";")
		if len(parts) >= 3 {
			var x, y int
			// First item in parts is "\033[<0" (the button)
			// Second item is column (X)
			// Third item is line (Y)
			_, _ = fmt.Sscanf(parts[1], "%d", &x)
			_, _ = fmt.Sscanf(parts[2], "%d", &y)

			// Compensate offsets for header rendering offset of 1 row
			mapY := y - 2
			// 16-bit mode uses double-width characters (e.g. 2 columns per character)
			mapX := (x - 1)
			if *renderMode == "16bit" {
				mapX = (x - 1) / 2
			}

			if mapX >= 0 && mapX < gm.Width && mapY >= 0 && mapY < gm.Height {
				cursor.X = mapX
				cursor.Y = mapY
				gm.Log(fmt.Sprintf("Mouse Click registered at [%d, %d]", mapX, mapY))
			}
		}
		return
	}

	// 2. Fallback to processing raw single-rune keyboard inputs
	for _, char := range input {
		ProcessSingleKey(char, gm, cursor, activeMenu, selectedCol, renderMode)
	}
}

// ProcessSingleKey processes individual keystroke commands.
func ProcessSingleKey(char rune, gm *game.GameMap, cursor *game.Position, activeMenu *string, selectedCol *int, renderMode *string) {
	// Global ESC key to close/reset overlays
	if char == 27 { // ESC
		*activeMenu = ""
		return
	}

	if char == 'v' {
		if *renderMode == "16bit" {
			*renderMode = "emoji"
			gm.Log("Graphics display changed to: Emoji Mode")
		} else {
			*renderMode = "16bit"
			gm.Log("Graphics display changed to: 16-Bit Retro Mode")
		}
		return
	}

	// Move cursor keys (Only Vim HJKL keys for simple intuitive map movement)
	// This completely avoids any WASD key collisions inside Architect Mode!
	switch char {
	case 'k':
		if *activeMenu == "" || *activeMenu == "architect" {
			if cursor.Y > 0 {
				cursor.Y--
			}
		}
	case 'j':
		if *activeMenu == "" || *activeMenu == "architect" {
			if cursor.Y < gm.Height-1 {
				cursor.Y++
			}
		}
	case 'h':
		if *activeMenu == "" || *activeMenu == "architect" {
			if cursor.X > 0 {
				cursor.X--
			}
		}
	case 'l':
		if *activeMenu == "" || *activeMenu == "architect" {
			if cursor.X < gm.Width-1 {
				cursor.X++
			}
		}

	// Overlay menu shortcut commands
	case 'A':
		*activeMenu = "architect"
	case 'W':
		*activeMenu = "work"
		*selectedCol = 0
	case 'S':
		*activeMenu = "schedule"
		*selectedCol = 0
	case 'R':
		*activeMenu = "research"

	case 'D':
		// Toggle draft mode for colonist under inspect cursor or first colonist
		colFound := false
		for _, col := range gm.Colonists {
			if col.Pos == *cursor {
				col.Drafted = !col.Drafted
				gm.Log(fmt.Sprintf("%s Draft toggled: %v", col.Name, col.Drafted))
				colFound = true
				break
			}
		}
		if !colFound && len(gm.Colonists) > 0 {
			gm.Colonists[0].Drafted = !gm.Colonists[0].Drafted
			gm.Log(fmt.Sprintf("%s Draft toggled: %v", gm.Colonists[0].Name, gm.Colonists[0].Drafted))
		}

	case ' ':
		// Toggle simulation play/pause speed
		if gm.GameSpeed == 0 {
			gm.GameSpeed = 1
			gm.Log("Simulation resumed.")
		} else {
			gm.GameSpeed = 0
			gm.Log("Simulation paused.")
		}

	// Sub-options inside active menu templates
	default:
		if *activeMenu == "work" {
			// Adjust priority ranks (0-4) using arrow keys and numbers
			if char >= '0' && char <= '4' {
				val := int(char - '0')
				col := gm.Colonists[*selectedCol]
				// Cycle work category update
				for _, wt := range []game.WorkType{game.WorkDoctor, game.WorkHarvest, game.WorkGrow, game.WorkConstruct, game.WorkMine, game.WorkCook, game.WorkResearch, game.WorkHaul} {
					col.Priorities[wt] = val
				}
				gm.Log(fmt.Sprintf("Adjusted %s priority settings to %d", col.Name, val))
			}
			// Cycle colonist selection
			if char == 'j' {
				if *selectedCol < len(gm.Colonists)-1 {
					*selectedCol++
				}
			} else if char == 'k' {
				if *selectedCol > 0 {
					*selectedCol--
				}
			}

		} else if *activeMenu == "schedule" {
			// Schedule Hour type selection
			col := gm.Colonists[*selectedCol]
			if char == '1' {
				for h := 0; h < 24; h++ {
					col.Schedule[h] = game.ScheduleSleep
				}
				gm.Log(fmt.Sprintf("Set schedule for %s to Sleeping", col.Name))
			} else if char == '2' {
				for h := 0; h < 24; h++ {
					col.Schedule[h] = game.ScheduleWork
				}
				gm.Log(fmt.Sprintf("Set schedule for %s to Working", col.Name))
			} else if char == '3' {
				for h := 0; h < 24; h++ {
					col.Schedule[h] = game.ScheduleJoy
				}
				gm.Log(fmt.Sprintf("Set schedule for %s to Joy/Relax", col.Name))
			} else if char == '4' {
				for h := 0; h < 24; h++ {
					col.Schedule[h] = game.ScheduleAnything
				}
				gm.Log(fmt.Sprintf("Set schedule for %s to Anything", col.Name))
			}

			if char == 'j' {
				if *selectedCol < len(gm.Colonists)-1 {
					*selectedCol++
				}
			} else if char == 'k' {
				if *selectedCol > 0 {
					*selectedCol--
				}
			}

		} else if *activeMenu == "research" {
			// Quick unlock selection triggers
			if char == '1' && game.IsTechAvailable(gm, "Electricity") {
				gm.ActiveResearch = "Electricity"
				gm.Log("Active Research set: Electricity")
			} else if char == '2' && game.IsTechAvailable(gm, "Machining") {
				gm.ActiveResearch = "Machining"
				gm.Log("Active Research set: Machining")
			} else if char == '3' && game.IsTechAvailable(gm, "Spaceship") {
				gm.ActiveResearch = "Spaceship"
				gm.Log("Active Research set: Spaceship")
			}

		} else if *activeMenu == "architect" {
			// Map specific blueprints other keys (Do not escape menu so multiple consecutive placements are possible)
			switch char {
			case 'w': // Place Wall Blueprint (perfectly free now!)
				placeBlueprint(gm, *cursor, game.BuildingWall)
			case 'd': // Place Door Blueprint (perfectly free now!)
				placeBlueprint(gm, *cursor, game.BuildingDoor)
			case 'b':
				placeBlueprint(gm, *cursor, game.BuildingBed)
			case 'g':
				placeBlueprint(gm, *cursor, game.BuildingGenerator)
			case 'p':
				placeBlueprint(gm, *cursor, game.BuildingBattery)
			case 's': // Place Solar panel (perfectly free now!)
				placeBlueprint(gm, *cursor, game.BuildingSolarPanel)
			case 'c':
				placeBlueprint(gm, *cursor, game.BuildingCooler)
			case 'h':
				placeBlueprint(gm, *cursor, game.BuildingHeater)
			case 't':
				if gm.TechUnlocked["Machining"] {
					placeBlueprint(gm, *cursor, game.BuildingTurret)
				} else {
					gm.Log("Requires Machining Research unlocked first!")
				}
			case 'r':
				placeBlueprint(gm, *cursor, game.BuildingResBench)
			case 'k':
				placeBlueprint(gm, *cursor, game.BuildingSandbag)
			case 'o':
				// Create stockpile zone
				createZone(gm, *cursor, game.ZoneStockpile)
			case 'z':
				// Create growing zone
				createZone(gm, *cursor, game.ZoneGrowing)
			}
		}
	}
}

func placeBlueprint(gm *game.GameMap, pos game.Position, bType game.BuildingType) {
	// Verify cost requirement
	steelCost := 0
	compCost := 0
	woodCost := 0

	switch bType {
	case game.BuildingWall:
		woodCost = 5
	case game.BuildingDoor:
		woodCost = 10
	case game.BuildingBed:
		woodCost = 35
	case game.BuildingSolarPanel:
		steelCost = 100
		compCost = 3
	case game.BuildingGenerator:
		steelCost = 100
		compCost = 2
	case game.BuildingBattery:
		steelCost = 50
		compCost = 1
	case game.BuildingHeater:
		steelCost = 50
		compCost = 1
	case game.BuildingCooler:
		steelCost = 90
		compCost = 3
	case game.BuildingTurret:
		steelCost = 100
		compCost = 3
	case game.BuildingResBench:
		woodCost = 75
		steelCost = 30
	case game.BuildingSandbag:
		steelCost = 5 // or cloth
	}

	// Scan colony inventory count
	steelCount := countColonyItems(gm, game.ItemSteel)
	compCount := countColonyItems(gm, game.ItemComponents)
	woodCount := countColonyItems(gm, game.ItemWood)

	if steelCount < steelCost || compCount < compCost || woodCount < woodCost {
		gm.Log(fmt.Sprintf("Insufficient resources! Requires %d Steel, %d Comp, %d Wood.", steelCost, compCost, woodCost))
		return
	}

	// Consume resources from stockpiles
	consumeColonyItems(gm, game.ItemSteel, steelCost)
	consumeColonyItems(gm, game.ItemComponents, compCost)
	consumeColonyItems(gm, game.ItemWood, woodCost)

	b := game.GetBuildingTemplate(bType)
	b.Pos = pos
	gm.Buildings[pos] = b
	gm.Log(fmt.Sprintf("Placed Blueprint: %s at [%d, %d]", bType, pos.X, pos.Y))
}

func createZone(gm *game.GameMap, pos game.Position, zType game.ZoneType) {
	// Creates a 3x3 zone of selected zone type centered on pos
	zoneID := gm.ZoneCounter
	gm.ZoneCounter++

	tiles := make([]game.Position, 0)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := pos.X+dx, pos.Y+dy
			if nx >= 0 && nx < gm.Width && ny >= 0 && ny < gm.Height {
				tiles = append(tiles, game.Position{X: nx, Y: ny})
			}
		}
	}

	z := &game.Zone{
		ID:         zoneID,
		Type:       zType,
		Tiles:      tiles,
		CropType:   game.ItemRice, // default crop rice
		AllowItems: make(map[game.ItemType]bool),
	}

	gm.Zones[zoneID] = z
	gm.Log(fmt.Sprintf("Created 3x3 %s Zone at [%d, %d]", zType, pos.X, pos.Y))
}

func countColonyItems(gm *game.GameMap, t game.ItemType) int {
	total := 0
	for _, item := range gm.Items {
		if item.Type == t {
			total += item.Qty
		}
	}
	return total
}

func consumeColonyItems(gm *game.GameMap, t game.ItemType, amount int) {
	if amount <= 0 {
		return
	}
	for pos, item := range gm.Items {
		if item.Type == t {
			if item.Qty >= amount {
				item.Qty -= amount
				if item.Qty == 0 {
					delete(gm.Items, pos)
				}
				return
			} else {
				amount -= item.Qty
				delete(gm.Items, pos)
			}
		}
	}
}

// ReadKey helper gets standard unbuffered keystroke characters
func ReadKey() (rune, error) {
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return 0, err
	}
	return rune(buf[0]), nil
}

package game

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateMap creates a new randomized RimWorld-style game map.
func GenerateMap(w, h int) *GameMap {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	grid := make([][]Tile, h)
	for y := 0; y < h; y++ {
		grid[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			tType := TileSoil
			// Let's make some parts rocky, some parts fertile, and random water lake/river.
			grid[y][x] = Tile{
				Pos:         Position{X: x, Y: y},
				Type:        tType,
				Roofed:      false,
				Temperature: 21.0,
			}
		}
	}

	gm := &GameMap{
		Width:            w,
		Height:           h,
		Grid:             grid,
		Buildings:        make(map[Position]*Building),
		Items:            make(map[Position]*Item),
		Colonists:        make([]*Colonist, 0),
		Enemies:          make([]*Enemy, 0),
		Zones:            make(map[int]*Zone),
		Geysers:          make([]*SteamGeyser, 0),
		Fires:            make(map[Position]*Fire),
		Projectiles:      make([]*Projectile, 0),
		MessageLog:       []string{"Crashlanded! Choose your tasks, construct shelters, and survive."},
		ResearchProgress: make(map[string]int),
		ActiveResearch:   "",
		TechUnlocked:     make(map[string]bool),
		GameSpeed:        1,
		TickCount:        0,
		Day:              1,
		Hour:             8,
		Weather:          "Clear",
		WeatherDuration:  120,
		NextEventTick:    2400, // events every few hours/days
		Storyteller:      "Randy",
		ZoneCounter:      1,
		BuildingCounter:  1,
		ItemCounter:      1,
		EnemyCounter:     1,
		OutdoorTemp:      21.0,
	}

	// Apply natural map generation features:
	// 1. Water lakes
	for i := 0; i < rng.Intn(3)+1; i++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)
		rad := rng.Intn(4) + 3
		for y := cy - rad; y <= cy+rad; y++ {
			for x := cx - rad; x <= cx+rad; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					distSq := (x-cx)*(x-cx) + (y-cy)*(y-cy)
					if distSq < rad*rad {
						gm.Grid[y][x].Type = TileWater
					}
				}
			}
		}
	}

	// 2. Mountains (solid rocky clusters)
	for i := 0; i < rng.Intn(4)+3; i++ {
		cx := rng.Intn(w)
		cy := rng.Intn(h)
		rad := rng.Intn(5) + 4
		for y := cy - rad; y <= cy+rad; y++ {
			for x := cx - rad; x <= cx+rad; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					distSq := (x-cx)*(x-cx) + (y-cy)*(y-cy)
					if distSq < rad*rad {
						if gm.Grid[y][x].Type != TileWater {
							gm.Grid[y][x].Type = TileMountain
						}
					}
				}
			}
		}
	}

	// 3. Fertile soil & Stony soil around mountains/random
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if gm.Grid[y][x].Type == TileSoil {
				noise := rng.Float64()
				if noise < 0.15 {
					gm.Grid[y][x].Type = TileFertileSoil
				} else if noise < 0.30 {
					gm.Grid[y][x].Type = TileStonySoil
				}
			}
		}
	}

	// 4. Geysers
	for i := 0; i < rng.Intn(2)+2; i++ {
		for attempt := 0; attempt < 20; attempt++ {
			gx := rng.Intn(w-4) + 2
			gy := rng.Intn(h-4) + 2
			if gm.Grid[gy][gx].Type == TileSoil || gm.Grid[gy][gx].Type == TileStonySoil || gm.Grid[gy][gx].Type == TileFertileSoil {
				pos := Position{X: gx, Y: gy}
				gm.Geysers = append(gm.Geysers, &SteamGeyser{Pos: pos})
				break
			}
		}
	}

	// 5. Place starter items (Steel, Components, Wood, Meals, Weapons)
	startPos := Position{X: w / 2, Y: h / 2}
	// Clear landing zone of mountains/water
	for y := startPos.Y - 5; y <= startPos.Y+5; y++ {
		for x := startPos.X - 5; x <= startPos.X+5; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				if gm.Grid[y][x].Type == TileMountain || gm.Grid[y][x].Type == TileWater {
					gm.Grid[y][x].Type = TileSoil
				}
			}
		}
	}

	spawnItem(gm, ItemSteel, Position{X: startPos.X - 2, Y: startPos.Y - 2}, 150)
	spawnItem(gm, ItemComponents, Position{X: startPos.X + 2, Y: startPos.Y - 2}, 30)
	spawnItem(gm, ItemWood, Position{X: startPos.X - 2, Y: startPos.Y + 2}, 200)
	spawnItem(gm, ItemMealSimple, Position{X: startPos.X + 2, Y: startPos.Y + 2}, 15)
	spawnItem(gm, ItemPistol, Position{X: startPos.X, Y: startPos.Y - 3}, 1)
	spawnItem(gm, ItemRifle, Position{X: startPos.X, Y: startPos.Y + 3}, 1)
	spawnItem(gm, ItemHealroot, Position{X: startPos.X - 3, Y: startPos.Y}, 10)

	// 6. Generate 3 starting colonists
	names := []string{"Bob", "Alice", "Charly"}
	titles := []string{"Engineer", "Doctor", "Farmer"}
	emojis := []string{"🤠", "👩‍⚕️", "👨‍🌾"}
	backstories := []Backstory{
		{Title: "Industrial Miner", Description: "Grew up in a deep industrial cave colony, skilled at mining and construction."},
		{Title: "Frontier Medic", Description: "Tended to victims of tribal raids, skilled at doctoring and shooting."},
		{Title: "Nursery Assistant", Description: "Raised flora on an agricultural world, skilled at growing and cooking."},
	}

	for i := 0; i < 3; i++ {
		cPos := Position{X: startPos.X + i - 1, Y: startPos.Y}
		colonist := &Colonist{
			ID:        gm.ColonistNextID(),
			Name:      names[i],
			Title:     titles[i],
			Emoji:     emojis[i],
			Pos:       cPos,
			Backstory: backstories[i],
			Hunger:    100.0,
			Rest:      100.0,
			Mood:      75.0,
			Joy:       80.0,
			HealthParts: []*HealthPart{
				{Name: "Torso", HP: 40, MaxHP: 40},
				{Name: "Head", HP: 25, MaxHP: 25},
				{Name: "Left Arm", HP: 30, MaxHP: 30},
				{Name: "Right Arm", HP: 30, MaxHP: 30},
				{Name: "Left Leg", HP: 30, MaxHP: 30},
				{Name: "Right Leg", HP: 30, MaxHP: 30},
			},
			Skills:     make(map[WorkType]*Skill),
			Priorities: make(map[WorkType]int),
			Schedule:   [24]ScheduleHour{},
			Weapon:     "",
			CurrentJob: nil,
			Drafted:    false,
		}

		// Initialize starting skills
		for _, wt := range []WorkType{WorkFirefight, WorkDoctor, WorkBedRest, WorkHarvest, WorkGrow, WorkConstruct, WorkMine, WorkCook, WorkResearch, WorkHaul, WorkClean} {
			lvl := float64(rng.Intn(7) + 3) // random skill levels 3-9
			colonist.Skills[wt] = &Skill{Level: lvl, XP: 0}
			colonist.Priorities[wt] = 3 // default priority
		}

		// Backstory specific adjustments
		if i == 0 { // Bob (Miner/Builder)
			colonist.Skills[WorkMine].Level = 12
			colonist.Skills[WorkConstruct].Level = 10
			colonist.Priorities[WorkMine] = 1
			colonist.Priorities[WorkConstruct] = 1
		} else if i == 1 { // Alice (Doctor/Medic)
			colonist.Skills[WorkDoctor].Level = 14
			colonist.Priorities[WorkDoctor] = 1
			colonist.Weapon = ItemPistol
		} else if i == 2 { // Charly (Farmer/Cook)
			colonist.Skills[WorkGrow].Level = 11
			colonist.Skills[WorkHarvest].Level = 11
			colonist.Skills[WorkCook].Level = 10
			colonist.Priorities[WorkGrow] = 1
			colonist.Priorities[WorkHarvest] = 1
			colonist.Priorities[WorkCook] = 1
		}

		// Standard schedule: 22:00 to 05:00 sleep, 06:00 to 07:00 joy, 08:00 to 17:00 work, rest anything
		for h := 0; h < 24; h++ {
			if h >= 22 || h <= 5 {
				colonist.Schedule[h] = ScheduleSleep
			} else if h == 6 || h == 7 || h == 18 || h == 19 {
				colonist.Schedule[h] = ScheduleJoy
			} else if h >= 8 && h <= 17 {
				colonist.Schedule[h] = ScheduleWork
			} else {
				colonist.Schedule[h] = ScheduleAnything
			}
		}

		gm.Colonists = append(gm.Colonists, colonist)
	}

	return gm
}

func spawnItem(gm *GameMap, t ItemType, pos Position, qty int) {
	gm.Items[pos] = &Item{
		ID:       gm.ItemNextID(),
		Type:     t,
		Pos:      pos,
		Qty:      qty,
		MaxQty:   75,
		HP:       100,
		MaxHP:    100,
		SpoilsIn: -1, // default infinite
	}
	// Food items spoil
	if t == ItemMealSimple || t == ItemMealFine || t == ItemMeat || t == ItemPotato || t == ItemRice {
		gm.Items[pos].SpoilsIn = 72.0 // spoils in 3 days
	}
}

// Helpers for automated ID increments
func (gm *GameMap) ColonistNextID() int {
	gm.EnemyCounter++ // just use a global unified counter or separate
	return len(gm.Colonists) + 1
}

func (gm *GameMap) ItemNextID() int {
	gm.ItemCounter++
	return gm.ItemCounter
}

func (gm *GameMap) BuildingNextID() int {
	gm.BuildingCounter++
	return gm.BuildingCounter
}

func (gm *GameMap) EnemyNextID() int {
	gm.EnemyCounter++
	return gm.EnemyCounter
}

// Log adds messages to the sidebar notification feed
func (gm *GameMap) Log(msg string) {
	gm.MessageLog = append(gm.MessageLog, fmt.Sprintf("[%02d:%02d] %s", gm.Hour, gm.TickCount%60, msg))
	if len(gm.MessageLog) > 100 {
		gm.MessageLog = gm.MessageLog[1:]
	}
}

package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RunStoryteller evaluates timeline ticks and fires random storyteller events.
func RunStoryteller(gm *GameMap) {
	gm.TickCount++
	if gm.TickCount%60 == 0 {
		gm.Hour++
		if gm.Hour >= 24 {
			gm.Hour = 0
			gm.Day++
			gm.Log(fmt.Sprintf("Day %d has begun.", gm.Day))
		}
	}

	// Weather duration decay
	if gm.WeatherDuration > 0 {
		gm.WeatherDuration--
		if gm.WeatherDuration == 0 {
			gm.Weather = "Clear"
			gm.Log("The weather has cleared up.")
		}
	}

	// Trigger storyteller random event
	if gm.TickCount >= gm.NextEventTick {
		triggerRandomEvent(gm)
		// Reset next event timer depending on storyteller
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		interval := 1200 // Phoebe (easy/slow)
		if gm.Storyteller == "Randy" {
			interval = rng.Intn(1600) + 400 // Random & unpredictable
		} else if gm.Storyteller == "Cassandra" {
			interval = rng.Intn(600) + 800 // Moderate/steady build-up
		}
		gm.NextEventTick = gm.TickCount + interval
	}

	// Simulating projectiles flight & combat interactions
	UpdateCombatProjectiles(gm)

	// Simulating random fire spread and combustion
	UpdateFires(gm)
}

func triggerRandomEvent(gm *GameMap) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	events := []string{"Raid", "Blight", "CargoPods", "WanderingJoin", "SolarFlare", "Heatwave"}
	event := events[rng.Intn(len(events))]

	switch event {
	case "Raid":
		spawnRaid(gm)
	case "Blight":
		// Random crops wither/die
		cropCount := 0
		for _, z := range gm.Zones {
			if z.Type == ZoneGrowing {
				for _, t := range z.Tiles {
					if _, ok := gm.Items[t]; ok && rng.Float64() < 0.5 {
						delete(gm.Items, t)
						cropCount++
					}
				}
			}
		}
		gm.Log(fmt.Sprintf("A sudden crop blight has destroyed %d of your plants!", cropCount))

	case "CargoPods":
		// Spawns items in a clear area near center
		spawnPos := Position{X: gm.Width/2 + rng.Intn(8) - 4, Y: gm.Height/2 + rng.Intn(8) - 4}
		if IsWalkable(gm, spawnPos) {
			itemRoll := rng.Intn(3)
			if itemRoll == 0 {
				spawnItem(gm, ItemMealFine, spawnPos, 10)
				gm.Log("Cargo Pods crashed nearby! Yielded: 10x Fine Meals.")
			} else if itemRoll == 1 {
				spawnItem(gm, ItemComponents, spawnPos, 8)
				gm.Log("Cargo Pods crashed nearby! Yielded: 8x Components.")
			} else {
				spawnItem(gm, ItemPlasteel, spawnPos, 25)
				gm.Log("Cargo Pods crashed nearby! Yielded: 25x Plasteel.")
			}
		}

	case "WanderingJoin", "WandererJoin":
		// New colonist joins
		names := []string{"Sparky", "Gronk", "Lumi"}
		emojis := []string{"🧑‍🔧", "👳", "👩‍🦰"}
		cPos := Position{X: 1, Y: 1}
		if IsWalkable(gm, cPos) {
			idx := rng.Intn(len(names))
			c := &Colonist{
				ID:        gm.ColonistNextID(),
				Name:      names[idx],
				Title:     "Wanderer",
				Emoji:     emojis[idx],
				Pos:       cPos,
				Hunger:    100,
				Rest:      100,
				Mood:      60,
				Joy:       50,
				HealthParts: []*HealthPart{
					{Name: "Torso", HP: 40, MaxHP: 40},
					{Name: "Head", HP: 25, MaxHP: 25},
				},
				Skills:     make(map[WorkType]*Skill),
				Priorities: make(map[WorkType]int),
			}
			for _, wt := range []WorkType{WorkFirefight, WorkDoctor, WorkBedRest, WorkHarvest, WorkGrow, WorkConstruct, WorkMine, WorkCook, WorkResearch, WorkHaul, WorkClean} {
				c.Skills[wt] = &Skill{Level: float64(rng.Intn(6) + 2), XP: 0}
				c.Priorities[wt] = 3
			}
			gm.Colonists = append(gm.Colonists, c)
			gm.Log(fmt.Sprintf("A wandering %s named %s has decided to join your colony!", c.Title, c.Name))
		}

	case "SolarFlare":
		gm.Weather = "SolarFlare"
		gm.WeatherDuration = 180 // lasts for 3 game hours
		// Disable power on all electronics
		for _, b := range gm.Buildings {
			if b.PowerOutput != 0 {
				b.PowerOn = false
			}
		}
		gm.Log("ALERT: A solar flare has begun. All electrical systems are temporarily disabled!")

	case "Heatwave":
		gm.Weather = "Heatwave"
		gm.WeatherDuration = 300 // lasts for 5 game hours
		gm.OutdoorTemp = 42.0    // super hot temperature
		gm.Log("ALERT: A dangerous heatwave has struck the region! Cool down your spaces immediately.")
	}
}

func spawnRaid(gm *GameMap) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	raidCount := rng.Intn(2) + 1
	gm.Log(fmt.Sprintf("ALERT: A pirate raid of %d raiders has arrived! Standard combat defenses required.", raidCount))

	for i := 0; i < raidCount; i++ {
		rPos := Position{X: 1, Y: gm.Height - 2 - i}
		if IsWalkable(gm, rPos) {
			enemy := &Enemy{
				ID:        gm.EnemyNextID(),
				Name:      fmt.Sprintf("Raider %c", 'A'+i),
				Pos:       rPos,
				HP:        60,
				MaxHP:     60,
				Weapon:    ItemPistol,
				MoveCool:  0,
				TargetPos: gm.Colonists[0].Pos,
			}
			if i > 0 && rng.Float64() < 0.4 {
				enemy.Weapon = ItemRifle
			}
			gm.Enemies = append(gm.Enemies, enemy)
		}
	}
}

// UpdateCombatProjectiles drives missile flight logic
func UpdateCombatProjectiles(gm *GameMap) {
	remains := make([]*Projectile, 0)
	for _, proj := range gm.Projectiles {
		// Calculate travel progress
		proj.Progress += 0.25
		// Simple direct step hit calculation
		dx := float64(proj.Target.X - proj.Pos.X)
		dy := float64(proj.Target.Y - proj.Pos.Y)
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist <= 1.0 || proj.Progress >= 1.0 {
			// Explode/deal damage on arrival
			dealCombatDamage(gm, proj)
		} else {
			// Smooth intermediate pos steps
			stepX := int(float64(proj.Pos.X) + dx*0.25)
			stepY := int(float64(proj.Pos.Y) + dy*0.25)
			proj.Pos = Position{X: stepX, Y: stepY}
			remains = append(remains, proj)
		}
	}
	gm.Projectiles = remains
}

func dealCombatDamage(gm *GameMap, proj *Projectile) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	if proj.IsEnemy {
		// Target is colonist
		for _, col := range gm.Colonists {
			if col.Pos == proj.Target {
				// Cover sandbag check
				hasSandbag := false
				for _, adj := range GetAdjacent(gm, col.Pos) {
					if b, ok := gm.Buildings[adj]; ok && b.Type == BuildingSandbag {
						hasSandbag = true
						break
					}
				}

				if hasSandbag && rng.Float64() < 0.6 {
					gm.Log(fmt.Sprintf("Bullet hit the defensive sandbag protecting %s!", col.Name))
					return
				}

				// Inflict damage to random health part
				partIdx := rng.Intn(len(col.HealthParts))
				part := col.HealthParts[partIdx]
				part.HP -= proj.Damage
				part.Bleeding += 1.5 // Bleeding rate

				gm.Log(fmt.Sprintf("%s was shot in the %s for %d damage!", col.Name, part.Name, proj.Damage))

				if part.HP <= 0 {
					part.HP = 0
					gm.Log(fmt.Sprintf("CRITICAL: %s's %s was destroyed!", col.Name, part.Name))
				}
				return
			}
		}
	} else {
		// Target is enemy
		for _, enemy := range gm.Enemies {
			if enemy.Pos == proj.Target {
				enemy.HP -= proj.Damage
				enemy.Bleeding += 1.0
				gm.Log(fmt.Sprintf("Raider %s was shot for %d damage!", enemy.Name, proj.Damage))
				if enemy.HP <= 0 {
					enemy.HP = 0
					gm.Log(fmt.Sprintf("Raider %s has died.", enemy.Name))
				}
				return
			}
		}
	}
}

// UpdateFires performs cellular automata combustion
func UpdateFires(gm *GameMap) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Periodic random fire ignition on map during storms or raids
	if gm.Weather == "Rain" {
		// Extinguish fires
		for pos := range gm.Fires {
			delete(gm.Fires, pos)
			gm.Log("Rain extinguished fire.")
		}
		return
	}

	// Update existing fires and spread
	spreads := make(map[Position]*Fire)
	for pos, fire := range gm.Fires {
		fire.Intensity += 0.05
		if fire.Intensity > 5.0 {
			// spread to adjacent
			for _, adj := range GetAdjacent(gm, pos) {
				if IsWalkable(gm, adj) && rng.Float64() < 0.12 {
					if _, exists := gm.Fires[adj]; !exists {
						spreads[adj] = &Fire{Pos: adj, Intensity: 1.0}
					}
				}
			}
		}
		// Damage building or items on the tile
		if b, ok := gm.Buildings[pos]; ok {
			b.HP -= int(fire.Intensity)
			if b.HP <= 0 {
				delete(gm.Buildings, pos)
				gm.Log(string(b.Type) + " burned down!")
			}
		}
		if item, ok := gm.Items[pos]; ok {
			item.HP -= int(fire.Intensity)
			if item.HP <= 0 {
				delete(gm.Items, pos)
			}
		}
	}

	for p, f := range spreads {
		gm.Fires[p] = f
	}
}

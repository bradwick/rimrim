package game

import (
	"math"
)

// DispatchJobs assigns high-priority and relevant jobs to idle colonists.
func DispatchJobs(gm *GameMap) {
	for _, c := range gm.Colonists {
		if c.Drafted {
			// Combat mode: manually controlled or auto-defending closest targets
			c.CurrentJob = nil
			continue
		}

		// Self-preservation: hunger and sleep take top priority if critical
		if c.Hunger < 15 && (c.CurrentJob == nil || c.CurrentJob.Type != JobEat) {
			target, ok := findNearestItem(gm, c.Pos, []ItemType{ItemMealSimple, ItemMealFine, ItemRice, ItemPotato, ItemMeat})
			if ok {
				c.CurrentJob = &Job{
					Type:     JobEat,
					Target:   target,
					WorkLeft: 10,
					Path:     FindPath(gm, c.Pos, target),
				}
				continue
			}
		}

		if c.Rest < 15 && (c.CurrentJob == nil || c.CurrentJob.Type != JobSleep) {
			target, ok := findNearestBedOrFloor(gm, c.Pos)
			if ok {
				c.CurrentJob = &Job{
					Type:     JobSleep,
					Target:   target,
					WorkLeft: 100 - c.Rest,
					Path:     FindPath(gm, c.Pos, target),
				}
				continue
			}
		}

		// Doctoring patient check: If someone is bleeding out or in pain
		if c.Priorities[WorkDoctor] > 0 && (c.CurrentJob == nil || c.CurrentJob.Type == JobNone) {
			for _, patient := range gm.Colonists {
				if patient.ID != c.ID && (isBleeding(patient) || isInjured(patient)) {
					c.CurrentJob = &Job{
						Type:     JobDoctor,
						Target:   patient.Pos,
						WorkLeft: 20,
						Path:     FindPath(gm, c.Pos, patient.Pos),
					}
					break
				}
			}
			if c.CurrentJob != nil {
				continue
			}
		}

		// If colonist already has a job, continue executing it
		if c.CurrentJob != nil && c.CurrentJob.Type != JobNone {
			continue
		}

		// Standard work-priority scanner
		assigned := false
		for wt := 1; wt <= 4; wt++ { // Priority 1 down to 4
			for _, workKind := range []WorkType{WorkFirefight, WorkHarvest, WorkGrow, WorkConstruct, WorkMine, WorkCook, WorkResearch, WorkHaul, WorkClean} {
				if c.Priorities[workKind] == wt {
					if tryAssignJob(gm, c, workKind) {
						assigned = true
						break
					}
				}
			}
			if assigned {
				break
			}
		}

		// Recreation / Idle logic: If no job was assigned
		if !assigned && c.CurrentJob == nil {
			if c.Joy < 55 {
				c.CurrentJob = &Job{
					Type:     JobRecreate,
					Target:   c.Pos, // can recreate on spot
					WorkLeft: 30,
					Path:     []Position{c.Pos},
				}
			} else {
				// Just wander around near starting position
				wanderPos := Position{X: c.Pos.X + (gm.TickCount%5 - 2), Y: c.Pos.Y + (gm.TickCount%3 - 1)}
				if IsWalkable(gm, wanderPos) {
					c.CurrentJob = &Job{
						Type:     JobNone,
						Target:   wanderPos,
						WorkLeft: 5,
						Path:     FindPath(gm, c.Pos, wanderPos),
					}
				}
			}
		}
	}
}

// ExecuteJobs moves colonists along their paths and updates work progression.
func ExecuteJobs(gm *GameMap) {
	for _, c := range gm.Colonists {
		if c.CurrentJob == nil {
			continue
		}

		job := c.CurrentJob

		// Moving along the path
		if job.PathIdx < len(job.Path) {
			nextPos := job.Path[job.PathIdx]
			if IsWalkable(gm, nextPos) {
				c.Pos = nextPos
				job.PathIdx++
			} else {
				// Recalculate path if blocked
				job.Path = FindPath(gm, c.Pos, job.Target)
				job.PathIdx = 0
				if len(job.Path) == 0 {
					// Path is blocked completely, cancel job
					c.CurrentJob = nil
					continue
				}
			}
			continue // Wait next tick to execute work or move further
		}

		// Colonist arrived at target - perform job activity
		switch job.Type {
		case JobEat:
			if item, ok := gm.Items[job.Target]; ok {
				item.Qty--
				if item.Qty <= 0 {
					delete(gm.Items, job.Target)
				}
				c.Hunger += 40
				if c.Hunger > 100 {
					c.Hunger = 100
				}
				gm.Log(c.Name + " ate a meal.")
				c.CurrentJob = nil
			} else {
				c.CurrentJob = nil // Food disappeared
			}

		case JobSleep:
			c.Rest += 2.0
			if c.Rest >= 100 {
				c.Rest = 100
				gm.Log(c.Name + " woke up fully rested.")
				c.CurrentJob = nil
			}

		case JobDoctor:
			var targetPatient *Colonist
			for _, pat := range gm.Colonists {
				if pat.Pos == job.Target {
					targetPatient = pat
					break
				}
			}
			if targetPatient != nil {
				DoctorTend(c, targetPatient)
				gm.Log(c.Name + " tended to " + targetPatient.Name)
				c.CurrentJob = nil
			} else {
				c.CurrentJob = nil
			}

		case JobConstruct:
			if b, ok := gm.Buildings[job.Target]; ok && b.IsBlueprint {
				b.WorkLeft--
				if b.WorkLeft <= 0 {
					b.IsBlueprint = false
					gm.Log(c.Name + " finished constructing " + string(b.Type))
					c.CurrentJob = nil
				}
			} else {
				c.CurrentJob = nil
			}

		case JobMine:
			tile := gm.Grid[job.Target.Y][job.Target.X]
			if tile.Type == TileMountain {
				job.WorkLeft--
				if job.WorkLeft <= 0 {
					gm.Grid[job.Target.Y][job.Target.X].Type = TileStonySoil
					// Drop some steel/components randomly
					roll := gm.TickCount % 10
					if roll < 2 {
						spawnItem(gm, ItemComponents, job.Target, 2)
					} else {
						spawnItem(gm, ItemSteel, job.Target, 35)
					}
					gm.Log(c.Name + " mined mountain rock.")
					c.CurrentJob = nil
				}
			} else {
				c.CurrentJob = nil
			}

		case JobHarvest:
			if tile, ok := isReadyToHarvest(gm, job.Target); ok {
				gm.Grid[job.Target.Y][job.Target.X].Type = TileSoil // reset
				crop := tile.CropType
				yield := 8
				if crop == ItemHealroot {
					yield = 2
				}
				spawnItem(gm, crop, job.Target, yield)
				gm.Log(c.Name + " harvested crops.")
				c.CurrentJob = nil
			} else {
				c.CurrentJob = nil
			}

		case JobSow:
			// Ensure it is a growing zone
			zoneFound := false
			for _, z := range gm.Zones {
				if z.Type == ZoneGrowing {
					for _, t := range z.Tiles {
						if t == job.Target {
							zoneFound = true
							// Clear terrain and mark it sown (simulate simple plant with temporary structural item or flag)
							gm.Grid[job.Target.Y][job.Target.X].Type = TileSoil
							// We will keep a map or flag for crop progression inside main game update loop
							spawnItem(gm, z.CropType, job.Target, 1) // starter seed stack size 1
							gm.Log(c.Name + " sowed crop " + string(z.CropType))
							c.CurrentJob = nil
							break
						}
					}
				}
				if zoneFound {
					break
				}
			}
			c.CurrentJob = nil

		case JobResearch:
			if gm.ActiveResearch != "" {
				gm.ResearchProgress[gm.ActiveResearch]++
				if gm.ResearchProgress[gm.ActiveResearch] >= 100 {
					gm.TechUnlocked[gm.ActiveResearch] = true
					gm.Log("Research Complete: " + gm.ActiveResearch + " unlocked!")
					gm.ActiveResearch = ""
				}
				// Give skill XP
				c.Skills[WorkResearch].XP += 2.0
				if c.Skills[WorkResearch].XP >= 100 {
					c.Skills[WorkResearch].Level++
					c.Skills[WorkResearch].XP = 0
				}
			}
			c.CurrentJob = nil // trigger research evaluation again next tick

		case JobClean:
			// Simply idle on spot and clean dirt
			c.CurrentJob = nil

		case JobRecreate:
			job.WorkLeft--
			if job.WorkLeft <= 0 {
				c.CurrentJob = nil
			}

		default:
			c.CurrentJob = nil
		}
	}
}

// Helpers for scanning items/beds/conditions
func findNearestItem(gm *GameMap, pos Position, types []ItemType) (Position, bool) {
	bestDist := math.MaxFloat64
	bestPos := Position{}
	found := false
	for p, item := range gm.Items {
		for _, t := range types {
			if item.Type == t {
				dist := heuristic(pos, p)
				if dist < bestDist {
					bestDist = dist
					bestPos = p
					found = true
				}
			}
		}
	}
	return bestPos, found
}

func findNearestBedOrFloor(gm *GameMap, pos Position) (Position, bool) {
	bestDist := math.MaxFloat64
	bestPos := pos
	found := false
	for p, b := range gm.Buildings {
		if b.Type == BuildingBed {
			dist := heuristic(pos, p)
			if dist < bestDist {
				bestDist = dist
				bestPos = p
				found = true
			}
		}
	}
	if found {
		return bestPos, true
	}
	return pos, true // fall back to current spot (sleeping on floor)
}

func isBleeding(c *Colonist) bool {
	for _, p := range c.HealthParts {
		if p.Bleeding > 0 {
			return true
		}
	}
	return false
}

func isInjured(c *Colonist) bool {
	for _, p := range c.HealthParts {
		if p.HP < p.MaxHP {
			return true
		}
	}
	return false
}

func isReadyToHarvest(gm *GameMap, pos Position) (*Zone, bool) {
	for _, z := range gm.Zones {
		if z.Type == ZoneGrowing {
			for _, tile := range z.Tiles {
				if tile == pos {
					// Check if item of crop type exists here and is grown (represented by stack qty >= 5)
					if item, ok := gm.Items[pos]; ok && item.Type == z.CropType && item.Qty >= 5 {
						return z, true
					}
				}
			}
		}
	}
	return nil, false
}

func tryAssignJob(gm *GameMap, c *Colonist, wt WorkType) bool {
	switch wt {
	case WorkConstruct:
		for pos, b := range gm.Buildings {
			if b.IsBlueprint {
				c.CurrentJob = &Job{
					Type:     JobConstruct,
					Target:   pos,
					WorkLeft: float64(b.WorkLeft),
					Path:     FindPath(gm, c.Pos, pos),
				}
				return true
			}
		}
	case WorkMine:
		for y := 0; y < gm.Height; y++ {
			for x := 0; x < gm.Width; x++ {
				if gm.Grid[y][x].Type == TileMountain {
					pos := Position{X: x, Y: y}
					dist := heuristic(c.Pos, pos)
					if dist < 15 { // only mine nearby rocks
						c.CurrentJob = &Job{
							Type:     JobMine,
							Target:   pos,
							WorkLeft: 15,
							Path:     FindPath(gm, c.Pos, pos),
						}
						return true
					}
				}
			}
		}
	case WorkResearch:
		if gm.ActiveResearch != "" {
			for pos, b := range gm.Buildings {
				if b.Type == BuildingResBench && !b.IsBlueprint {
					c.CurrentJob = &Job{
						Type:     JobResearch,
						Target:   pos,
						WorkLeft: 1,
						Path:     FindPath(gm, c.Pos, pos),
					}
					return true
				}
			}
		}
	case WorkHarvest:
		for _, z := range gm.Zones {
			if z.Type == ZoneGrowing {
				for _, t := range z.Tiles {
					if _, ok := isReadyToHarvest(gm, t); ok {
						c.CurrentJob = &Job{
							Type:     JobHarvest,
							Target:   t,
							WorkLeft: 5,
							Path:     FindPath(gm, c.Pos, t),
						}
						return true
					}
				}
			}
		}
	case WorkGrow:
		// Sow seeds in growing zones where nothing is growing currently
		for _, z := range gm.Zones {
			if z.Type == ZoneGrowing {
				for _, t := range z.Tiles {
					if _, ok := gm.Items[t]; !ok {
						c.CurrentJob = &Job{
							Type:     JobSow,
							Target:   t,
							WorkLeft: 5,
							Path:     FindPath(gm, c.Pos, t),
						}
						return true
					}
				}
			}
		}
	}
	return false
}

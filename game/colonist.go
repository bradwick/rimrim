package game

// UpdateColonistNeeds simulates hunger, rest, and mood fluctuations.
func UpdateColonistNeeds(gm *GameMap, c *Colonist) {
	if c.Drafted {
		// Colonists burn rest and hunger faster during combat draft
		c.Hunger -= 0.15
		c.Rest -= 0.12
		c.Joy -= 0.1
	} else if c.CurrentJob != nil && c.CurrentJob.Type == JobSleep {
		// Sleeping restores rest
		c.Rest += 1.8
		if c.Rest > 100 {
			c.Rest = 100
		}
		c.Hunger -= 0.03
	} else if c.CurrentJob != nil && c.CurrentJob.Type == JobEat {
		// Eating is handled in job executor
		c.Rest -= 0.04
	} else if c.CurrentJob != nil && c.CurrentJob.Type == JobRecreate {
		// Recreating restores Joy
		c.Joy += 2.0
		if c.Joy > 100 {
			c.Joy = 100
		}
		c.Hunger -= 0.06
		c.Rest -= 0.04
	} else {
		// Idle or working standard decay
		c.Hunger -= 0.08
		c.Rest -= 0.06
		c.Joy -= 0.04
	}

	// Boundary clamping
	if c.Hunger < 0 {
		c.Hunger = 0
	}
	if c.Rest < 0 {
		c.Rest = 0
	}
	if c.Joy < 0 {
		c.Joy = 0
	}

	// Mood calculations (based on hunger, rest, joy, health pain, temperature comfort)
	moodModifier := 0.0
	if c.Hunger < 20 {
		moodModifier -= 25 // starving
	} else if c.Hunger < 40 {
		moodModifier -= 10
	} else {
		moodModifier += 5
	}

	if c.Rest < 15 {
		moodModifier -= 20 // exhausted
	} else if c.Rest < 35 {
		moodModifier -= 8
	} else {
		moodModifier += 4
	}

	if c.Joy < 20 {
		moodModifier -= 10
	} else if c.Joy > 80 {
		moodModifier += 10
	}

	// Temperature discomfort
	tileTemp := gm.Grid[c.Pos.Y][c.Pos.X].Temperature
	if tileTemp < 10 {
		moodModifier -= 12 // freezing
	} else if tileTemp > 35 {
		moodModifier -= 12 // hot
	}

	// Pain and wounds
	hasBleeding := false
	for _, part := range c.HealthParts {
		if part.HP < part.MaxHP {
			moodModifier -= float64(part.MaxHP-part.HP) * 0.2
		}
		if part.Bleeding > 0 {
			hasBleeding = true
		}
	}

	// Final Mood shift (towards baseline + modifiers)
	targetMood := 50.0 + moodModifier
	if targetMood < 0 {
		targetMood = 0
	}
	if targetMood > 100 {
		targetMood = 100
	}

	// Smooth shift towards target
	c.Mood += (targetMood - c.Mood) * 0.05

	// Mental break warning
	if c.Mood < 20 && gm.TickCount%120 == 0 {
		gm.Log(c.Name + " is at risk of a mental break!")
	}

	// Handle Bleeding damage
	if hasBleeding {
		for _, part := range c.HealthParts {
			if part.Bleeding > 0 {
				part.HP -= 1
				if part.HP < 0 {
					part.HP = 0
				}
			}
		}
		if gm.TickCount%120 == 0 {
			gm.Log(c.Name + " is bleeding out! Needs doctoring/bedrest.")
		}
	}
}

// DoctorTend cures bleeding and restores colonist health parts.
func DoctorTend(doctor, patient *Colonist) {
	for _, part := range patient.HealthParts {
		if part.Bleeding > 0 {
			part.Bleeding = 0
			part.Bandaged = true
			return
		}
		if part.HP < part.MaxHP {
			healQty := 5 + int(doctor.Skills[WorkDoctor].Level/3.0)
			part.HP += healQty
			if part.HP > part.MaxHP {
				part.HP = part.MaxHP
			}
			return
		}
	}
}

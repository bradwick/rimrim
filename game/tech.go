package game

// Tech defines details of printable tree levels.
type Tech struct {
	Name         string
	Cost         int
	Description  string
	Requirements []string
}

// GetTechTree returns the overall game progress tiers.
func GetTechTree() map[string]*Tech {
	return map[string]*Tech{
		"Electricity": {
			Name:         "Electricity",
			Cost:         100,
			Description:  "Unlocks power solar panels, heaters, coolers, batteries, and fueled generators.",
			Requirements: []string{},
		},
		"Machining": {
			Name:         "Machining",
			Cost:         150,
			Description:  "Allows manufacturing of automated defensive combat turrets.",
			Requirements: []string{"Electricity"},
		},
		"Mortars": {
			Name:         "Mortars",
			Cost:         250,
			Description:  "Unlocks defensive high-caliber automated artillery units.",
			Requirements: []string{"Machining"},
		},
		"Autocannons": {
			Name:         "Autocannons",
			Cost:         300,
			Description:  "Unlocks heavy high-fire-rate automated defensive turrets.",
			Requirements: []string{"Machining"},
		},
		"Spaceship": {
			Name:         "Spaceship",
			Cost:         500,
			Description:  "Enables construction of escape spacecraft reactor and structural engine units to win the game.",
			Requirements: []string{"Electricity", "Machining"},
		},
	}
}

// IsTechAvailable returns true if all prerequisite parent nodes are fully unlocked.
func IsTechAvailable(gm *GameMap, techName string) bool {
	tree := GetTechTree()
	tech, ok := tree[techName]
	if !ok {
		return false
	}
	for _, req := range tech.Requirements {
		if unlocked := gm.TechUnlocked[req]; !unlocked {
			return false
		}
	}
	return true
}

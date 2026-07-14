package game

// GetBuildingTemplate returns templates with default HP, Power output/consumption values.
func GetBuildingTemplate(bType BuildingType) *Building {
	b := &Building{
		Type:         bType,
		IsBlueprint:  true,
		PowerOn:      true,
		HP:           100,
		MaxHP:        100,
		BatteryMax:   1000,
	}

	switch bType {
	case BuildingWall:
		b.WorkRequired = 5
		b.HP = 250
		b.MaxHP = 250
	case BuildingDoor:
		b.WorkRequired = 8
		b.HP = 150
		b.MaxHP = 150
	case BuildingBed:
		b.WorkRequired = 10
		b.HP = 100
		b.MaxHP = 100
	case BuildingSolarPanel:
		b.WorkRequired = 25
		b.PowerOutput = 150 // generates during day
	case BuildingGenerator:
		b.WorkRequired = 20
		b.PowerOutput = 1000 // fueled
		b.Fuel = 50.0
	case BuildingBattery:
		b.WorkRequired = 15
		b.PowerOutput = 0 // storage
	case BuildingConduit:
		b.WorkRequired = 1
		b.HP = 20
		b.MaxHP = 20
	case BuildingHeater:
		b.WorkRequired = 12
		b.PowerOutput = -100
		b.TargetTemp = 21.0
	case BuildingCooler:
		b.WorkRequired = 15
		b.PowerOutput = -200
		b.TargetTemp = 0.0
	case BuildingTurret:
		b.WorkRequired = 30
		b.HP = 120
		b.MaxHP = 120
		b.PowerOutput = -80
	case BuildingResBench:
		b.WorkRequired = 20
		b.PowerOutput = 0 // pure manual
	case BuildingButcher:
		b.WorkRequired = 15
	case BuildingStove:
		b.WorkRequired = 15
		b.PowerOutput = -350
	case BuildingGeothermal:
		b.WorkRequired = 50
		b.HP = 300
		b.MaxHP = 300
		b.PowerOutput = 3600
	case BuildingSandbag:
		b.WorkRequired = 4
		b.HP = 400
		b.MaxHP = 400
	}

	b.WorkLeft = b.WorkRequired
	return b
}

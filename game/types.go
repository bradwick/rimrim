package game

import "time"

// Position represents 2D grid coordinates.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// TileType classifies the terrain.
type TileType int

const (
	TileSoil TileType = iota
	TileStonySoil
	TileFertileSoil
	TileWater
	TileSand
	TileMountain
	TileFloor
)

// Tile represents a single grid cell.
type Tile struct {
	Pos         Position `json:"pos"`
	Type        TileType `json:"type"`
	Roofed      bool     `json:"roofed"`
	Temperature float64  `json:"temperature"` // in Celsius
}

// BuildingType represents types of structures players can build.
type BuildingType string

const (
	BuildingWall        BuildingType = "wall"
	BuildingDoor        BuildingType = "door"
	BuildingFloor       BuildingType = "floor"
	BuildingBed         BuildingType = "bed"
	BuildingSolarPanel  BuildingType = "solar_panel"
	BuildingGenerator   BuildingType = "generator" // fueled
	BuildingBattery     BuildingType = "battery"
	BuildingConduit     BuildingType = "conduit"
	BuildingHeater      BuildingType = "heater"
	BuildingCooler      BuildingType = "cooler"
	BuildingTurret      BuildingType = "turret"
	BuildingSandbag     BuildingType = "sandbag"
	BuildingResBench    BuildingType = "research_bench"
	BuildingButcher     BuildingType = "butcher_table"
	BuildingStove       BuildingType = "cooking_stove"
	BuildingGeothermal  BuildingType = "geothermal_generator"
)

// Building represents structures on the map.
type Building struct {
	ID             int          `json:"id"`
	Type           BuildingType `json:"type"`
	Pos            Position     `json:"pos"`
	HP             int          `json:"hp"`
	MaxHP          int          `json:"max_hp"`
	IsBlueprint    bool         `json:"is_blueprint"`
	WorkRequired   int          `json:"work_required"`
	WorkLeft       int          `json:"work_left"`
	PowerOutput    int          `json:"power_output"` // positive for gen, negative for consumer
	PowerOn        bool         `json:"power_on"`
	TargetTemp     float64      `json:"target_temp"`  // for heaters/coolers
	BatteryStored  int          `json:"battery_stored"`
	BatteryMax     int          `json:"battery_max"`
	Fuel           float64      `json:"fuel"`         // for fueled generator
	GeyserPos      Position     `json:"geyser_pos"`   // if geothermal
}

// ItemType represents types of resources, food, items.
type ItemType string

const (
	ItemSteel      ItemType = "steel"
	ItemComponents ItemType = "components"
	ItemGold       ItemType = "gold"
	ItemWood       ItemType = "wood"
	ItemRice       ItemType = "rice"
	ItemPotato     ItemType = "potato"
	ItemHealroot   ItemType = "healroot" // herbal medicine
	ItemMealSimple ItemType = "meal_simple"
	ItemMealFine   ItemType = "meal_fine"
	ItemMeat       ItemType = "meat"
	ItemPistol     ItemType = "pistol"
	ItemRifle      ItemType = "rifle"
	ItemPlasteel   ItemType = "plasteel"
)

// Item represents resources lying on the ground.
type Item struct {
	ID        int      `json:"id"`
	Type      ItemType `json:"type"`
	Pos       Position `json:"pos"`
	Qty       int      `json:"qty"`
	MaxQty    int      `json:"max_qty"`
	HP        int      `json:"hp"`
	MaxHP     int      `json:"max_hp"`
	SpoilsIn  float64  `json:"spoils_in"` // hours left before spoiling (-1 for infinite)
}

// Skill represents colonist's expertise.
type Skill struct {
	Level float64 `json:"level"`
	XP    float64 `json:"xp"`
}

// Backstory defines starting stats of a colonist.
type Backstory struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// WorkType is a specific task kind.
type WorkType string

const (
	WorkFirefight WorkType = "Firefight"
	WorkDoctor    WorkType = "Doctor"
	WorkBedRest   WorkType = "BedRest"
	WorkHarvest   WorkType = "Harvest"
	WorkGrow      WorkType = "Grow"
	WorkConstruct WorkType = "Construct"
	WorkMine      WorkType = "Mine"
	WorkCook      WorkType = "Cook"
	WorkResearch  WorkType = "Research"
	WorkHaul      WorkType = "Haul"
	WorkClean     WorkType = "Clean"
)

// ScheduleHour defines the schedule activities: Anything, Work, Joy, Sleep.
type ScheduleHour int

const (
	ScheduleAnything ScheduleHour = iota
	ScheduleWork
	ScheduleJoy
	ScheduleSleep
)

// HealthPart represents body parts and health states.
type HealthPart struct {
	Name     string `json:"name"`
	HP       int    `json:"hp"`
	MaxHP    int    `json:"max_hp"`
	Bleeding float64 `json:"bleeding"` // bleed rate per hour
	Bandaged bool   `json:"bandaged"`
}

// Colonist represents a pawn/character.
type Colonist struct {
	ID          int                     `json:"id"`
	Name        string                  `json:"name"`
	Title       string                  `json:"title"`
	Emoji       string                  `json:"emoji"`
	Pos         Position                `json:"pos"`
	Backstory   Backstory               `json:"backstory"`
	Hunger      float64                 `json:"hunger"` // 0 to 100
	Rest        float64                 `json:"rest"`   // 0 to 100
	Mood        float64                 `json:"mood"`   // 0 to 100
	Joy         float64                 `json:"joy"`    // 0 to 100
	HealthParts []*HealthPart           `json:"health_parts"`
	Skills      map[WorkType]*Skill     `json:"skills"`
	Priorities  map[WorkType]int        `json:"priorities"` // 0 (disabled), 1-4 (highest to lowest)
	Schedule    [24]ScheduleHour        `json:"schedule"`
	Weapon      ItemType                `json:"weapon"` // "" for melee fists, or ItemPistol, ItemRifle
	CurrentJob  *Job                    `json:"current_job"`
	Drafted     bool                    `json:"drafted"`
	MoveCool    float64                 `json:"move_cool"` // travel cooldown
}

// JobType represents the current activity a colonist is performing.
type JobType string

const (
	JobNone        JobType = "None"
	JobMove        JobType = "Move"
	JobSleep       JobType = "Sleep"
	JobEat         JobType = "Eat"
	JobHarvest     JobType = "Harvesting"
	JobSow         JobType = "Sowing"
	JobConstruct   JobType = "Constructing"
	JobMine        JobType = "Mining"
	JobCook        JobType = "Cooking"
	JobButcher     JobType = "Butchering"
	JobResearch    JobType = "Researching"
	JobHaul        JobType = "Hauling"
	JobClean       JobType = "Cleaning"
	JobDoctor      JobType = "Doctoring"
	JobFight       JobType = "Fighting"
	JobRecreate    JobType = "Recreating"
)

// Job contains instructions for colonist behavior.
type Job struct {
	Type     JobType  `json:"type"`
	Target   Position `json:"target"`
	WorkLeft float64  `json:"work_left"`
	Path     []Position `json:"path"`
	PathIdx  int      `json:"path_idx"`
}

// ZoneType defines stock/growing area kinds.
type ZoneType string

const (
	ZoneStockpile ZoneType = "Stockpile"
	ZoneGrowing   ZoneType = "Growing"
)

// Zone represents designated areas on the map.
type Zone struct {
	ID         int          `json:"id"`
	Type       ZoneType     `json:"type"`
	Tiles      []Position   `json:"tiles"`
	CropType   ItemType     `json:"crop_type"`   // for growing zones (rice, potato, healroot)
	AllowItems map[ItemType]bool `json:"allow_items"` // for stockpile zones
}

// Enemy represents raiders or wild hostile beasts.
type Enemy struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Pos       Position `json:"pos"`
	HP        int      `json:"hp"`
	MaxHP     int      `json:"max_hp"`
	Bleeding  float64  `json:"bleeding"`
	Weapon    ItemType `json:"weapon"` // "" or pistol/rifle or claws
	MoveCool  float64  `json:"move_cool"`
	TargetPos Position `json:"target_pos"`
	Fleeing   bool     `json:"fleeing"`
}

// SteamGeyser represents natural vents on the map.
type SteamGeyser struct {
	Pos Position `json:"pos"`
}

// Fire represents a tile that is burning.
type Fire struct {
	Pos      Position `json:"pos"`
	Intensity float64 `json:"intensity"`
}

// Projectile is a combat projectile mid-flight.
type Projectile struct {
	Pos      Position `json:"pos"`
	Target   Position `json:"target"`
	IsEnemy  bool     `json:"is_enemy"`
	Damage   int      `json:"damage"`
	Speed    float64  `json:"speed"` // cool-down counter
	Progress float64  `json:"progress"`
}

// GameMap is the overall world structure.
type GameMap struct {
	Width            int                       `json:"width"`
	Height           int                       `json:"height"`
	Grid             [][]Tile                  `json:"grid"`
	Buildings        map[Position]*Building    `json:"buildings"`
	Items            map[Position]*Item        `json:"items"`
	Colonists        []*Colonist               `json:"colonists"`
	Enemies          []*Enemy                  `json:"enemies"`
	Zones            map[int]*Zone             `json:"zones"`
	Geysers          []*SteamGeyser            `json:"geysers"`
	Fires            map[Position]*Fire        `json:"fires"`
	Projectiles      []*Projectile             `json:"projectiles"`
	MessageLog       []string                  `json:"message_log"`
	ResearchProgress map[string]int            `json:"research_progress"`
	ActiveResearch   string                    `json:"active_research"`
	TechUnlocked     map[string]bool           `json:"tech_unlocked"`
	GameSpeed        int                       `json:"game_speed"` // 0=Paused, 1=Normal, 2=Fast, 3=Ultra
	TickCount        int                       `json:"tick_count"`
	Day              int                       `json:"day"`
	Hour             int                       `json:"hour"`
	Weather          string                    `json:"weather"` // Clear, Rain, Heatwave, ColdSnap, SolarFlare
	WeatherDuration  int                       `json:"weather_duration"`
	NextEventTick    int                       `json:"next_event_tick"`
	Storyteller      string                    `json:"storyteller"` // Cassandra, Randy, Phoebe
	ZoneCounter      int                       `json:"zone_counter"`
	BuildingCounter  int                       `json:"building_counter"`
	ItemCounter      int                       `json:"item_counter"`
	EnemyCounter     int                       `json:"enemy_counter"`
	LastSaveTime     time.Time                 `json:"last_save_time"`
	OutdoorTemp      float64                   `json:"outdoor_temp"`
}

package tui

import (
	"fmt"
	"strings"

	"rimworld_tui/game"
)

// Terminal screen rendering helper functions using clean ANSI escape codes.

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func RenderGame(gm *game.GameMap, cursor game.Position, activeMenu string, selectedCol int, renderMode string) {
	var sb strings.Builder

	// Top title bar / simulation details
	sb.WriteString("\033[H\033[48;5;235m\033[38;5;255m")
	sb.WriteString(fmt.Sprintf("  ▲ RIMWORLD TUI CLONE ▲   Day %d | Hour %02d:00 | Outdoor: %.1f°C | Weather: %s | Speed: %d (1-3, P=Pause)  \n",
		gm.Day, gm.Hour, gm.OutdoorTemp, gm.Weather, gm.GameSpeed))
	sb.WriteString("\033[0m")

	// Pre-compile sprites
	for y := 0; y < gm.Height; y++ {
		for x := 0; x < gm.Width; x++ {
			pos := game.Position{X: x, Y: y}
			tile := gm.Grid[y][x]

			// Highlight selection cursor
			if pos == cursor {
				sb.WriteString("\033[48;5;240m")
			}

			// Render layers: 1) fire 2) projectile 3) colonist/enemy 4) building 5) item 6) terrain
			if _, ok := gm.Fires[pos]; ok {
				sb.WriteString("🔥")
			} else if isProjectileAt(gm, pos) {
				sb.WriteString("💥")
			} else if col, ok := getColonistAt(gm, pos); ok {
				if col.Drafted {
					sb.WriteString("💂") // combat soldier override
				} else {
					if col.Emoji != "" {
						sb.WriteString(col.Emoji)
					} else {
						sb.WriteString("🧑") // regular fallback
					}
				}
			} else if enemy, ok := getEnemyAt(gm, pos); ok {
				_ = enemy
				sb.WriteString("🏴") // pirate / raider
			} else if b, ok := gm.Buildings[pos]; ok {
				if b.IsBlueprint {
					sb.WriteString("📝")
				} else {
					switch b.Type {
					case game.BuildingWall:
						sb.WriteString("🧱")
					case game.BuildingDoor:
						sb.WriteString("🚪")
					case game.BuildingBed:
						sb.WriteString("🛏️")
					case game.BuildingSolarPanel:
						sb.WriteString("☀️")
					case game.BuildingGenerator:
						sb.WriteString("⚙️")
					case game.BuildingBattery:
						sb.WriteString("🔋")
					case game.BuildingConduit:
						sb.WriteString("⚡")
					case game.BuildingHeater:
						sb.WriteString("🔥")
					case game.BuildingCooler:
						sb.WriteString("❄️")
					case game.BuildingTurret:
						sb.WriteString("🔫")
					case game.BuildingResBench:
						sb.WriteString("🔬")
					case game.BuildingSandbag:
						sb.WriteString("🛡️")
					case game.BuildingButcher:
						sb.WriteString("🔪")
					case game.BuildingStove:
						sb.WriteString("🍳")
					default:
						sb.WriteString("🏠")
					}
				}
			} else if item, ok := gm.Items[pos]; ok {
				switch item.Type {
				case game.ItemSteel:
					sb.WriteString("🔩")
				case game.ItemComponents:
					sb.WriteString("⚙️")
				case game.ItemGold:
					sb.WriteString("🪙")
				case game.ItemWood:
					sb.WriteString("🪵")
				case game.ItemRice:
					sb.WriteString("🌾")
				case game.ItemPotato:
					sb.WriteString("🥔")
				case game.ItemHealroot:
					sb.WriteString("🌿")
				case game.ItemMealSimple:
					sb.WriteString("🍲")
				case game.ItemMealFine:
					sb.WriteString("🍱")
				case game.ItemPistol:
					sb.WriteString("🔫")
				case game.ItemRifle:
					sb.WriteString("🔫")
				default:
					sb.WriteString("📦")
				}
			} else {
				// Terrain
				switch tile.Type {
				case game.TileWater:
					sb.WriteString("💧")
				case game.TileMountain:
					sb.WriteString("⛰️")
				case game.TileFertileSoil:
					sb.WriteString("🌱")
				case game.TileStonySoil:
					sb.WriteString("🪨")
				case game.TileFloor:
					sb.WriteString("🪵")
				default:
					sb.WriteString("🟩") // regular soil
				}
			}

			if pos == cursor {
				sb.WriteString("\033[0m")
			}
		}
		sb.WriteString("\n")
	}

	// Bottom command bar
	sb.WriteString("\033[48;5;238m\033[38;5;255m")
	sb.WriteString("  [A] Architect  [W] Work Priority  [S] Schedule  [R] Research Tree  [D] Draft Colonist  [Space] Pause/Play  \n")
	sb.WriteString("\033[0m")

	// Command overlay window / Sidebar options details
	sb.WriteString("\n")
	switch activeMenu {
	case "architect":
		sb.WriteString("\033[1;36m▲ ARCHITECT BUILD MENU ▲\033[0m\n")
		sb.WriteString("Press keys to select blueprint blueprint to place with cursor:\n")
		sb.WriteString("[w] Wall (Steel/Wood)     [d] Door                  [b] Bed\n")
		sb.WriteString("[g] Fueled Gen (Power)    [s] Solar Panel           [p] Battery\n")
		sb.WriteString("[c] Cooler (Temp)         [h] Heater                [t] Auto Turret\n")
		sb.WriteString("[r] Research Bench        [k] Sandbag cover         [u] Butcher Table\n")
		sb.WriteString("[o] Stockpile Zone        [z] Growing Zone (Rice)   [Esc] Close Menu\n")

	case "work":
		sb.WriteString("\033[1;33m▲ COLONIST WORK PRIORITIES (0-4) ▲\033[0m\n")
		sb.WriteString("Press numbers to adjust priorities. Move with arrow keys. [Esc] Close\n")
		sb.WriteString("Name       | Doctor | Harvest | Grow | Construct | Mine | Cook | Research | Haul\n")
		sb.WriteString("---------------------------------------------------------------------------------\n")
		for i, col := range gm.Colonists {
			marker := "  "
			if i == selectedCol {
				marker = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%-8s |   %d    |    %d    |   %d  |     %d     |   %d  |  %d   |    %d     |  %d\n",
				marker, col.Name,
				col.Priorities[game.WorkDoctor],
				col.Priorities[game.WorkHarvest],
				col.Priorities[game.WorkGrow],
				col.Priorities[game.WorkConstruct],
				col.Priorities[game.WorkMine],
				col.Priorities[game.WorkCook],
				col.Priorities[game.WorkResearch],
				col.Priorities[game.WorkHaul],
			))
		}

	case "schedule":
		sb.WriteString("\033[1;35m▲ COLONIST HOURLY SCHEDULE ▲\033[0m\n")
		sb.WriteString("[Esc] Close | Press [1] Sleep (💤) | [2] Work (⚒️) | [3] Joy (🎨) | [4] Anything (❓)\n")
		for i, col := range gm.Colonists {
			marker := "  "
			if i == selectedCol {
				marker = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%-8s: ", marker, col.Name))
			for h := 0; h < 24; h++ {
				switch col.Schedule[h] {
				case game.ScheduleSleep:
					sb.WriteString("💤")
				case game.ScheduleWork:
					sb.WriteString("⚒️")
				case game.ScheduleJoy:
					sb.WriteString("🎨")
				default:
					sb.WriteString("❓")
				}
			}
			sb.WriteString("\n")
		}

	case "research":
		sb.WriteString("\033[1;32m▲ COLONY TECH RESEARCH TREE ▲\033[0m\n")
		sb.WriteString("Press associated number to select active research. [Esc] Close\n")
		tree := game.GetTechTree()
		for name, tech := range tree {
			status := "🔒 Locked"
			if gm.TechUnlocked[name] {
				status = "✅ Unlocked"
			} else if gm.ActiveResearch == name {
				status = fmt.Sprintf("⏳ Active (%d/%d XP)", gm.ResearchProgress[name], tech.Cost)
			} else if game.IsTechAvailable(gm, name) {
				status = "🟢 Available"
			}
			sb.WriteString(fmt.Sprintf("- %-15s [%s]: %s\n", tech.Name, status, tech.Description))
		}

	default:
		// Default HUD details
		sb.WriteString("\033[1;34m▲ COLONY NOTIFICATIONS & MESSAGE FEED ▲\033[0m\n")
		// Draw last 4 message log items
		startIdx := len(gm.MessageLog) - 4
		if startIdx < 0 {
			startIdx = 0
		}
		for i := startIdx; i < len(gm.MessageLog); i++ {
			sb.WriteString(gm.MessageLog[i] + "\n")
		}

		// Selection cursor info box
		sb.WriteString(fmt.Sprintf("\n\033[1mCursor Inspect at [%d, %d]:\033[0m ", cursor.X, cursor.Y))
		tile := gm.Grid[cursor.Y][cursor.X]
		sb.WriteString(fmt.Sprintf("Terrain: %s | Temp: %.1f°C | Roofed: %v ", getTileName(tile.Type), tile.Temperature, tile.Roofed))
		if b, ok := gm.Buildings[cursor]; ok {
			blueprintStr := ""
			if b.IsBlueprint {
				blueprintStr = " (Blueprint)"
			}
			sb.WriteString(fmt.Sprintf("\nStructure: %s%s | HP: %d/%d | PowerOn: %v ", b.Type, blueprintStr, b.HP, b.MaxHP, b.PowerOn))
		}
		if item, ok := gm.Items[cursor]; ok {
			sb.WriteString(fmt.Sprintf("\nItem: %s x%d ", item.Type, item.Qty))
		}
	}

	fmt.Print(sb.String())
}

// Helpers
func isProjectileAt(gm *game.GameMap, pos game.Position) bool {
	for _, p := range gm.Projectiles {
		if p.Pos == pos {
			return true
		}
	}
	return false
}

func getColonistAt(gm *game.GameMap, pos game.Position) (*game.Colonist, bool) {
	for _, col := range gm.Colonists {
		if col.Pos == pos {
			return col, true
		}
	}
	return nil, false
}

func getEnemyAt(gm *game.GameMap, pos game.Position) (*game.Enemy, bool) {
	for _, enemy := range gm.Enemies {
		if enemy.Pos == pos {
			return enemy, true
		}
	}
	return nil, false
}

func getTileName(t game.TileType) string {
	switch t {
	case game.TileWater:
		return "Water"
	case game.TileMountain:
		return "Mountain Mountain Rock"
	case game.TileFertileSoil:
		return "Fertile Soil"
	case game.TileStonySoil:
		return "Stony Soil"
	case game.TileFloor:
		return "Wood Flooring"
	default:
		return "Normal Soil"
	}
}

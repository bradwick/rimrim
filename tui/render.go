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
	sb.WriteString(fmt.Sprintf("  ▲ RIMWORLD 16-BIT RETRO EDITION ▲   Day %d | Hour %02d:00 | Outdoor: %.1f°C | Weather: %s | Mode: %s (v=toggle) | Speed: %d (1-3, Space=Pause)  \n",
		gm.Day, gm.Hour, gm.OutdoorTemp, gm.Weather, renderMode, gm.GameSpeed))
	sb.WriteString("\033[0m")

	// Pre-compile sprites
	for y := 0; y < gm.Height; y++ {
		for x := 0; x < gm.Width; x++ {
			pos := game.Position{X: x, Y: y}
			tile := gm.Grid[y][x]

			// Highlight selection cursor
			isCursor := pos == cursor

			if renderMode == "16bit" {
				// 16-bit retro graphics pixel block using beautiful background/foreground 24-bit TrueColor (RGB) styling
				bgRGB := "38;2;20;80;20" // default soil dark-green
				fgRGB := "38;2;255;255;255"
				charStr := "░░"

				// Determine base tile terrain colors
				switch tile.Type {
				case game.TileWater:
					bgRGB = "48;2;10;50;150" // deep water blue
					fgRGB = "38;2;30;130;255"
					charStr = "~~"
				case game.TileMountain:
					bgRGB = "48;2;80;80;80" // stone grey
					fgRGB = "38;2;150;150;150"
					charStr = "▲▲"
				case game.TileFertileSoil:
					bgRGB = "48;2;30;95;30" // rich green
					fgRGB = "38;2;100;240;100"
					charStr = "🌱"
				case game.TileStonySoil:
					bgRGB = "48;2;60;70;60" // gravel
					fgRGB = "38;2;140;140;140"
					charStr = "::"
				case game.TileFloor:
					bgRGB = "48;2;110;80;40" // rich wood brown
					fgRGB = "38;2;220;180;120"
					charStr = "##"
				default: // TileSoil
					bgRGB = "48;2;30;75;30"
					fgRGB = "38;2;80;140;80"
					charStr = "░░"
				}

				// Apply active dynamic structures layer
				if _, ok := gm.Fires[pos]; ok {
					bgRGB = "48;2;200;50;10"
					fgRGB = "38;2;255;230;30"
					charStr = "🔥"
				} else if isProjectileAt(gm, pos) {
					bgRGB = "48;2;255;100;100"
					fgRGB = "38;2;255;255;255"
					charStr = "**"
				} else if col, ok := getColonistAt(gm, pos); ok {
					if col.Drafted {
						bgRGB = "48;2;200;10;10" // military red
						fgRGB = "38;2;255;255;255"
						charStr = "💂"
					} else {
						bgRGB = "48;2;10;150;180" // teal colonist shirt
						fgRGB = "38;2;255;224;189"
						charStr = "🧑"
					}
				} else if enemy, ok := getEnemyAt(gm, pos); ok {
					_ = enemy
					bgRGB = "48;2;100;0;0" // hostile crimson
					fgRGB = "38;2;255;50;50"
					charStr = "🏴"
				} else if b, ok := gm.Buildings[pos]; ok {
					if b.IsBlueprint {
						bgRGB = "48;2;150;150;100"
						fgRGB = "38;2;255;255;255"
						charStr = "📝"
					} else {
						switch b.Type {
						case game.BuildingWall:
							bgRGB = "48;2;100;100;100" // metallic grey
							fgRGB = "38;2;200;200;200"
							charStr = "🧱"
						case game.BuildingDoor:
							bgRGB = "48;2;120;90;50"
							fgRGB = "38;2;255;255;255"
							charStr = "🚪"
						case game.BuildingBed:
							bgRGB = "48;2;50;50;180"
							fgRGB = "38;2;255;255;255"
							charStr = "🛏️"
						case game.BuildingSolarPanel:
							bgRGB = "48;2;20;30;80"
							fgRGB = "38;2;50;150;255"
							charStr = "☀️"
						case game.BuildingGenerator:
							bgRGB = "48;2;120;60;20"
							fgRGB = "38;2;255;180;0"
							charStr = "⚙️"
						case game.BuildingBattery:
							bgRGB = "48;2;20;120;20"
							fgRGB = "38;2;100;255;100"
							charStr = "🔋"
						case game.BuildingHeater:
							bgRGB = "48;2;150;30;30"
							fgRGB = "38;2;255;100;100"
							charStr = "🔥"
						case game.BuildingCooler:
							bgRGB = "48;2;30;80;180"
							fgRGB = "38;2;100;200;255"
							charStr = "❄️"
						case game.BuildingTurret:
							bgRGB = "48;2;80;20;20"
							fgRGB = "38;2;255;50;50"
							charStr = "🔫"
						default:
							bgRGB = "48;2;80;80;80"
							fgRGB = "38;2;255;255;255"
							charStr = "🏠"
						}
					}
				} else if item, ok := gm.Items[pos]; ok {
					switch item.Type {
					case game.ItemSteel:
						bgRGB = "48;2;50;55;65"
						fgRGB = "38;2;220;225;235"
						charStr = "🔩"
					case game.ItemComponents:
						bgRGB = "48;2;80;40;10"
						fgRGB = "38;2;255;140;0"
						charStr = "⚙️"
					case game.ItemGold:
						bgRGB = "48;2;150;120;10"
						fgRGB = "38;2;255;220;0"
						charStr = "🪙"
					case game.ItemWood:
						bgRGB = "48;2;80;50;20"
						fgRGB = "38;2;180;120;50"
						charStr = "🪵"
					case game.ItemMealSimple:
						bgRGB = "48;2;60;50;30"
						fgRGB = "38;2;255;180;100"
						charStr = "🍲"
					default:
						bgRGB = "48;2;40;40;40"
						fgRGB = "38;2;200;200;200"
						charStr = "📦"
					}
				}

				if isCursor {
					sb.WriteString(fmt.Sprintf("\033[48;2;200;180;50m\033[%s;1m%s\033[0m", fgRGB, charStr))
				} else {
					sb.WriteString(fmt.Sprintf("\033[%s;%sm%s\033[0m", bgRGB, fgRGB, charStr))
				}
			} else {
				// Standard Classic emoji styling
				if isCursor {
					sb.WriteString("\033[48;5;240m")
				}

				if _, ok := gm.Fires[pos]; ok {
					sb.WriteString("🔥")
				} else if isProjectileAt(gm, pos) {
					sb.WriteString("💥")
				} else if col, ok := getColonistAt(gm, pos); ok {
					if col.Drafted {
						sb.WriteString("💂")
					} else {
						if col.Emoji != "" {
							sb.WriteString(col.Emoji)
						} else {
							sb.WriteString("🧑")
						}
					}
				} else if enemy, ok := getEnemyAt(gm, pos); ok {
					_ = enemy
					sb.WriteString("🏴")
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
						sb.WriteString("🟩")
					}
				}

				if isCursor {
					sb.WriteString("\033[0m")
				}
			}
		}
		sb.WriteString("\n")
	}

	// Bottom command bar
	sb.WriteString("\033[48;5;238m\033[38;5;255m")
	sb.WriteString("  [A] Architect  [W] Work Priority  [S] Schedule  [R] Research Tree  [D] Draft Colonist  [v] Toggle graphics  [Space] Pause/Play  \n")
	sb.WriteString("\033[0m")

	// Command overlay window / Sidebar options details
	sb.WriteString("\n")
	switch activeMenu {
	case "architect":
		sb.WriteString("\033[1;36m▲ ARCHITECT BUILD MENU ▲\033[0m\n")
		sb.WriteString("Place blueprints on cursor and move with HJKL. Place multiple blueprints without escaping:\n")
		sb.WriteString("[w] Wall                  [d] Door                  [b] Bed\n")
		sb.WriteString("[g] Fueled Gen (Power)    [s] Solar Panel           [p] Battery\n")
		sb.WriteString("[c] Cooler (Temp)         [h] Heater                [t] Auto Turret\n")
		sb.WriteString("[r] Research Bench        [k] Sandbag cover         [o] Stockpile Zone\n")
		sb.WriteString("[z] Growing Zone (Rice)   [Esc] Done placing blueprints / Exit Menu\n")

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

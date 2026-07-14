package main

import (
	"os"
	"os/exec"
	"time"

	"rimworld_tui/game"
	"rimworld_tui/tui"
)

func main() {
	// Configure terminal to raw mode for fast unbuffered input reading
	setRawMode(true)
	defer setRawMode(false)

	// Enable standard mouse reporting (xterm format)
	os.Stdout.WriteString("\033[?1000h\033[?1006h")
	defer os.Stdout.WriteString("\033[?1000l\033[?1006l")

	gm := game.GenerateMap(50, 30)
	cursor := game.Position{X: 25, Y: 15}
	activeMenu := ""
	selectedCol := 0
	renderMode := "16bit" // Defaults to beautiful 16-bit TrueColor retro graphics!

	// Handle standard input stream reading in non-blocking background thread
	inputChan := make(chan string, 100)
	go func() {
		buf := make([]byte, 128)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				inputChan <- string(buf[:n])
			}
		}
	}()

	// Primary game tick & render loop
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Clear screen before starting
	tui.ClearScreen()

	for {
		select {
		case inputStr := <-inputChan:
			// Check for ESC / Mouse Sequences / Special Keys
			if len(inputStr) > 0 {
				if inputStr == "q" || inputStr == "Q" {
					return
				}
				tui.ProcessRawInput(inputStr, gm, &cursor, &activeMenu, &selectedCol, &renderMode)
			}

		case <-ticker.C:
			// Drive game logic simulation if not paused
			if gm.GameSpeed > 0 {
				ticksPerLoop := gm.GameSpeed
				for i := 0; i < ticksPerLoop; i++ {
					game.RunStoryteller(gm)
					game.DispatchJobs(gm)
					game.ExecuteJobs(gm)
					game.UpdatePower(gm)
					game.UpdateTemperature(gm)
				}
			}

			// Render current status details
			tui.ClearScreen()
			tui.RenderGame(gm, cursor, activeMenu, selectedCol, renderMode)
		}
	}
}

func setRawMode(raw bool) {
	var flag string
	if raw {
		flag = "-icanon"
	} else {
		flag = "icanon"
	}
	cmd := exec.Command("stty", flag)
	cmd.Stdin = os.Stdin
	_ = cmd.Run()

	if raw {
		flag = "-echo"
	} else {
		flag = "echo"
	}
	cmd = exec.Command("stty", flag)
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

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

	gm := game.GenerateMap(50, 30)
	cursor := game.Position{X: 25, Y: 15}
	activeMenu := ""
	selectedCol := 0

	// Handle standard input stream reading in non-blocking background thread
	inputChan := make(chan rune, 100)
	go func() {
		for {
			r, err := tui.ReadKey()
			if err != nil {
				return
			}
			inputChan <- r
		}
	}()

	// Primary game tick & render loop
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Clear screen before starting
	tui.ClearScreen()

	for {
		select {
		case char := <-inputChan:
			if char == 'q' || char == 'Q' {
				// Escape/Exit game sequence
				return
			}
			tui.ProcessInput(char, gm, &cursor, &activeMenu, &selectedCol)

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
			tui.RenderGame(gm, cursor, activeMenu, selectedCol, "emoji")
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

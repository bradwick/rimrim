package game

import "testing"

func TestStorytellerEventsAndCombat(t *testing.T) {
	gm := GenerateMap(20, 20)

	// Inject a raider
	enemyPos := Position{X: 2, Y: 2}
	gm.Enemies = append(gm.Enemies, &Enemy{
		ID:    1,
		Name:  "Test Raider",
		Pos:   enemyPos,
		HP:    100,
		MaxHP: 100,
	})

	// Inject weapon firing projectile
	gm.Projectiles = append(gm.Projectiles, &Projectile{
		Pos:      Position{X: 1, Y: 1},
		Target:   enemyPos,
		IsEnemy:  false,
		Damage:   35,
		Progress: 0.9,
	})

	UpdateCombatProjectiles(gm)

	if gm.Enemies[0].HP != 65 {
		t.Errorf("Expected Raider HP to reduce to 65 on combat projectile hit, got %d", gm.Enemies[0].HP)
	}
}

func TestTechRequirements(t *testing.T) {
	gm := GenerateMap(20, 20)
	gm.TechUnlocked["Electricity"] = false

	if IsTechAvailable(gm, "Machining") {
		t.Error("Machining should be locked when Electricity is locked")
	}

	gm.TechUnlocked["Electricity"] = true
	if !IsTechAvailable(gm, "Machining") {
		t.Error("Machining should be available when Electricity is unlocked")
	}
}

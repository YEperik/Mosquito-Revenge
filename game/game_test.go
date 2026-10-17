package game

import "testing"

func TestDifficultySetsStartingLives(t *testing.T) {
	cases := []struct {
		level difficultyLevel
		want  int
	}{
		{difficultyChill, 4},
		{difficultyNormal, 3},
		{difficultyNightmare, 3},
	}

	for _, tc := range cases {
		t.Run(difficultyLevels[int(tc.level)].name, func(t *testing.T) {
			g := New()
			g.difficulty = tc.level
			g.resetRun()
			if g.lives != tc.want {
				t.Fatalf("starting lives = %d, want %d", g.lives, tc.want)
			}
		})
	}
}

func TestDifficultyScalesStageGoal(t *testing.T) {
	g := New()
	g.difficulty = difficultyChill
	chillGoal := g.goalForStage()
	g.difficulty = difficultyNormal
	normalGoal := g.goalForStage()
	g.difficulty = difficultyNightmare
	nightmareGoal := g.goalForStage()

	if !(chillGoal < normalGoal && normalGoal < nightmareGoal) {
		t.Fatalf("expected chill < normal < nightmare goals, got %.2f, %.2f, %.2f", chillGoal, normalGoal, nightmareGoal)
	}
}

func TestStageRankUsesHitsAndBonusObjectives(t *testing.T) {
	g := New()
	g.stageHitsTaken = 0
	g.stagePowerupsCollected = 2
	g.stagePerfectEscapes = 0
	g.calculateStageStars()
	if g.stageStars != 3 {
		t.Fatalf("clean stage with two pickups rank = %d, want 3", g.stageStars)
	}

	g.stageHitsTaken = 1
	g.stagePowerupsCollected = 0
	g.stagePerfectEscapes = 0
	g.calculateStageStars()
	if g.stageStars != 1 {
		t.Fatalf("hit stage without bonus rank = %d, want 1", g.stageStars)
	}
}

func TestLoseLifeTransitionsToGameOver(t *testing.T) {
	g := New()
	g.state = statePlaying
	g.lives = 1
	g.loseLife("TEST HIT")

	if g.lives != 0 {
		t.Fatalf("lives = %d, want 0", g.lives)
	}
	if g.state != stateGameOver {
		t.Fatalf("state = %d, want game over", g.state)
	}
	if g.totalHitsTaken != 1 || g.stageHitsTaken != 1 {
		t.Fatalf("hit counters = total %d, stage %d; want 1, 1", g.totalHitsTaken, g.stageHitsTaken)
	}
}

func TestFinalStageBossTelegraphsAndRewardsEscape(t *testing.T) {
	g := New()
	g.state = statePlaying
	g.stageIndex = bossStage
	g.player = point{x: 180, y: 500}
	g.bossAttackTimer = 0

	g.updateBossEncounter(1.0 / 60.0)
	if !g.bossAttackActive {
		t.Fatal("boss attack should start when its timer expires")
	}

	// The attack is locked to the position marked at the start, so moving away
	// during the warning window must result in a clean dodge.
	g.player.x = 270
	for i := 0; i < 45; i++ {
		g.updateBossEncounter(1.0 / 60.0)
	}

	if g.bossStrikes != 1 {
		t.Fatalf("boss strikes resolved = %d, want 1", g.bossStrikes)
	}
	if g.score != 65 {
		t.Fatalf("score after boss dodge = %d, want 65", g.score)
	}
	if g.totalPerfectEscapes != 1 {
		t.Fatalf("perfect escapes = %d, want 1", g.totalPerfectEscapes)
	}
}

func TestPauseButtonBounds(t *testing.T) {
	if !inPauseButton(point{x: 330, y: 60}) {
		t.Fatal("pause button should include its center")
	}
	if inPauseButton(point{x: 300, y: 60}) {
		t.Fatal("pause button should not include points outside its left edge")
	}
}

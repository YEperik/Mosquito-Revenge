package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenWidth  = 360
	ScreenHeight = 640

	playerRadius = 8.0
	handX        = 180.0
	handY        = 344.0
	maxLives     = 4
	bossStage    = 4
)

type gameState uint8

const (
	stateMenu gameState = iota
	statePlaying
	stateStageClear
	statePaused
	stateGameOver
	stateVictory
)

type point struct{ x, y float64 }

type difficultyLevel uint8

const (
	difficultyChill difficultyLevel = iota
	difficultyNormal
	difficultyNightmare
)

type difficultyConfig struct {
	name       string
	lives      int
	enemyScale float64
	windScale  float64
	itchRate   float64
	itchDecay  float64
	goalScale  float64
}

var difficultyLevels = []difficultyConfig{
	{name: "CHILL", lives: 4, enemyScale: 0.78, windScale: 0.72, itchRate: 0.78, itchDecay: 1.25, goalScale: 0.90},
	{name: "NORMAL", lives: 3, enemyScale: 1.00, windScale: 1.00, itchRate: 1.00, itchDecay: 1.00, goalScale: 1.00},
	{name: "NIGHTMARE", lives: 3, enemyScale: 1.24, windScale: 1.30, itchRate: 1.28, itchDecay: 0.78, goalScale: 1.12},
}

type stageConfig struct {
	name        string
	description string
	goal        float64
	wind        float64
	accent      color.RGBA
	background  color.RGBA
	grid        color.RGBA
}

var stages = []stageConfig{
	{name: "DORM ROOM", description: "The sleepy human", goal: 58, wind: 0.72, accent: color.RGBA{R: 102, G: 232, B: 190, A: 255}, background: color.RGBA{R: 15, G: 25, B: 43, A: 255}, grid: color.RGBA{R: 31, G: 47, B: 65, A: 255}},
	{name: "HOT KITCHEN", description: "Spider on patrol", goal: 76, wind: 0.95, accent: color.RGBA{R: 255, G: 190, B: 91, A: 255}, background: color.RGBA{R: 38, G: 24, B: 34, A: 255}, grid: color.RGBA{R: 69, G: 43, B: 51, A: 255}},
	{name: "GREENHOUSE", description: "Dragonfly incoming", goal: 90, wind: 1.0, accent: color.RGBA{R: 117, G: 226, B: 143, A: 255}, background: color.RGBA{R: 13, G: 39, B: 42, A: 255}, grid: color.RGBA{R: 31, G: 71, B: 66, A: 255}},
	{name: "NIGHT GARDEN", description: "Watch the beetle charge", goal: 105, wind: 1.18, accent: color.RGBA{R: 175, G: 138, B: 255, A: 255}, background: color.RGBA{R: 23, G: 20, B: 51, A: 255}, grid: color.RGBA{R: 48, G: 42, B: 83, A: 255}},
	{name: "STORMY ROOFTOP", description: "All hunters are awake", goal: 120, wind: 1.48, accent: color.RGBA{R: 255, G: 111, B: 129, A: 255}, background: color.RGBA{R: 31, G: 25, B: 43, A: 255}, grid: color.RGBA{R: 65, G: 46, B: 68, A: 255}},
}

type enemyKind uint8

const (
	enemySpider enemyKind = iota
	enemyDragonfly
	enemyBeetle
)

type enemy struct {
	kind            enemyKind
	pos             point
	origin          point
	phase           float64
	speed           float64
	radius          float64
	chargeCooldown  float64
	chargeRemaining float64
	chargeDir       point
}

type powerupKind uint8

const (
	powerNectar powerupKind = iota
	powerChill
	powerShield
	powerHaste
)

type powerup struct {
	kind powerupKind
	pos  point
	age  float64
}

type particle struct {
	pos      point
	velocity point
	life     float64
	maxLife  float64
	radius   float64
	color    color.RGBA
}

type floatingText struct {
	pos   point
	text  string
	color color.RGBA
	life  float64
}

type Game struct {
	state      gameState
	difficulty difficultyLevel

	player   point
	heading  float64
	elapsed  float64
	fanAngle float64

	stageIndex    int
	stageProgress float64
	stageTimer    float64
	score         int
	scoreExact    float64
	bestScore     int
	blood         float64
	itch          float64
	lives         int
	sucking       bool

	combo        int
	comboTick    float64
	comboTimer   float64
	feedParticle float64

	stageHitsTaken         int
	stagePowerupsCollected int
	stagePerfectEscapes    int
	stageStars             int
	totalHitsTaken         int
	totalPowerupsCollected int
	totalPerfectEscapes    int

	slapActive         bool
	slapElapsed        float64
	slapImpactChecked  bool
	slapCooldown       float64
	escapedFlash       float64
	damageFlash        float64
	invulnerability    float64
	dashTimer          float64
	dashCooldown       float64
	dashDirection      point
	cloakTimer         float64
	cloakCooldown      float64
	pulseTimer         float64
	pulseCooldown      float64
	hasteTimer         float64
	shieldCharges      int
	powerSpawnTimer    float64
	powerSequence      int
	bossAttackTimer    float64
	bossAttackElapsed  float64
	bossAttackActive   bool
	bossImpactChecked  bool
	bossAttackTarget   point
	bossStrikes        int
	message            string
	messageTimer       float64
	gameOverReasonText string

	enemies   []enemy
	powerups  []powerup
	particles []particle
	floaters  []floatingText

	fanPushPoint   point
	lastTouches    map[ebiten.TouchID]bool
	currentTouches map[ebiten.TouchID]bool
}

// New returns a fresh game instance using a fixed 360x640 logical canvas.
func New() *Game {
	g := &Game{
		difficulty:     difficultyNormal,
		lastTouches:    make(map[ebiten.TouchID]bool),
		currentTouches: make(map[ebiten.TouchID]bool),
	}
	g.resetRun()
	g.state = stateMenu
	return g
}

func (g *Game) currentStage() stageConfig {
	if g.stageIndex < 0 || g.stageIndex >= len(stages) {
		return stages[0]
	}
	return stages[g.stageIndex]
}

func (g *Game) difficultyConfig() difficultyConfig {
	index := int(g.difficulty)
	if index < 0 || index >= len(difficultyLevels) {
		index = int(difficultyNormal)
	}
	return difficultyLevels[index]
}

func (g *Game) goalForStage() float64 {
	return g.currentStage().goal * g.difficultyConfig().goalScale
}

func (g *Game) resetRun() {
	g.player = point{x: 180, y: 506}
	g.heading = -math.Pi / 2
	g.elapsed = 0
	g.fanAngle = 0
	g.stageIndex = 0
	g.stageProgress = 0
	g.stageTimer = 0
	g.score = 0
	g.scoreExact = 0
	g.blood = 0
	g.itch = 0
	g.lives = g.difficultyConfig().lives
	g.sucking = false
	g.combo = 0
	g.comboTick = 0
	g.comboTimer = 0
	g.feedParticle = 0
	g.stageHitsTaken = 0
	g.stagePowerupsCollected = 0
	g.stagePerfectEscapes = 0
	g.stageStars = 0
	g.totalHitsTaken = 0
	g.totalPowerupsCollected = 0
	g.totalPerfectEscapes = 0
	g.slapActive = false
	g.slapElapsed = 0
	g.slapImpactChecked = false
	g.slapCooldown = 0
	g.escapedFlash = 0
	g.damageFlash = 0
	g.invulnerability = 0
	g.dashTimer = 0
	g.dashCooldown = 0
	g.dashDirection = point{x: 0, y: -1}
	g.cloakTimer = 0
	g.cloakCooldown = 0
	g.pulseTimer = 0
	g.pulseCooldown = 0
	g.hasteTimer = 0
	g.shieldCharges = 0
	g.powerSpawnTimer = 5
	g.powerSequence = 0
	g.message = ""
	g.messageTimer = 0
	g.gameOverReasonText = ""
	g.enemies = nil
	g.powerups = nil
	g.particles = nil
	g.floaters = nil
	g.fanPushPoint = point{x: 307, y: 112}
	g.setupStage()
}

func (g *Game) setupStage() {
	g.enemies = g.enemies[:0]
	g.powerups = g.powerups[:0]
	g.stageProgress = 0
	g.stageHitsTaken = 0
	g.stagePowerupsCollected = 0
	g.stagePerfectEscapes = 0
	g.stageStars = 0
	g.bossAttackActive = false
	g.bossImpactChecked = false
	g.bossStrikes = 0
	g.bossAttackElapsed = 0
	g.bossAttackTimer = 0
	if g.stageIndex == bossStage {
		g.bossAttackTimer = 4.8 / math.Max(0.75, g.difficultyConfig().enemyScale)
	}
	g.itch = math.Min(g.itch, 24)
	g.blood = math.Max(0, g.blood-28) // A new level starts with a lighter abdomen.
	g.powerSpawnTimer = 4.5
	g.slapActive = false
	g.slapElapsed = 0
	g.slapImpactChecked = false
	g.slapCooldown = math.Max(g.slapCooldown, 0.7)
	g.dashCooldown = math.Min(g.dashCooldown, 0.25)
	g.cloakCooldown = math.Max(0, g.cloakCooldown-1.5)
	g.pulseCooldown = math.Max(0, g.pulseCooldown-1.5)
	g.player = point{x: 180, y: 506}

	add := func(kind enemyKind, x, y, phase, speed float64) {
		radius := 10.0
		if kind == enemyDragonfly {
			radius = 9
		}
		if kind == enemyBeetle {
			radius = 12
		}
		g.enemies = append(g.enemies, enemy{
			kind: kind, pos: point{x: x, y: y}, origin: point{x: x, y: y},
			phase: phase, speed: speed * g.difficultyConfig().enemyScale, radius: radius,
			chargeCooldown: 1.3 + phase*0.2, chargeDir: point{x: -1, y: 0},
		})
	}

	switch g.stageIndex {
	case 0:
		// First stage teaches the core loop without enemy pressure.
	case 1:
		add(enemySpider, 64, 225, 0.3, 55)
	case 2:
		add(enemySpider, 57, 222, 1.0, 60)
		add(enemyDragonfly, 303, 450, 2.1, 70)
	case 3:
		add(enemyDragonfly, 54, 232, 0.4, 82)
		add(enemySpider, 295, 245, 2.2, 66)
		add(enemyBeetle, 87, 490, 1.3, 92)
	case 4:
		add(enemyDragonfly, 54, 218, 0.5, 92)
		add(enemyDragonfly, 305, 423, 2.0, 96)
		add(enemyBeetle, 83, 485, 1.4, 108)
		add(enemySpider, 294, 241, 2.7, 72)
	}
	g.showMessage("STAGE "+itoa(g.stageIndex+1), 1.25, g.currentStage().accent)
}

func (g *Game) Update() error {
	const dt = 1.0 / 60.0
	controls := g.readControls()
	if g.state == statePaused {
		if controls.restartPressed {
			g.resetRun()
			g.state = statePlaying
		} else if controls.pausePressed || controls.confirm {
			g.state = statePlaying
		}
		return nil
	}

	g.elapsed += dt
	g.fanAngle += dt * 5.2
	g.updateTimers(dt)

	switch g.state {
	case stateMenu:
		if controls.hasPress {
			if selected, ok := difficultyButtonAt(controls.pressPos); ok {
				g.difficulty = selected
				return nil
			}
		}
		if controls.confirm || controls.restartPressed {
			g.resetRun()
			g.state = statePlaying
		}
		return nil
	case stateGameOver, stateVictory:
		if controls.confirm {
			g.resetRun()
			g.state = statePlaying
		}
		g.updateParticles(dt)
		return nil
	case stateStageClear:
		g.updateParticles(dt)
		g.stageTimer -= dt
		if controls.confirm || g.stageTimer <= 0 {
			if g.stageIndex+1 >= len(stages) {
				g.state = stateVictory
				g.showMessage("ALL STAGES CLEARED!", 2, color.RGBA{R: 255, G: 215, B: 99, A: 255})
				g.addBurst(point{x: ScreenWidth / 2, y: ScreenHeight / 2}, color.RGBA{R: 255, G: 215, B: 99, A: 255}, 60, 130)
			} else {
				g.stageIndex++
				g.setupStage()
				g.state = statePlaying
			}
		}
		return nil
	}

	if controls.pausePressed {
		g.state = statePaused
		return nil
	}

	if controls.dashPressed && g.dashCooldown <= 0 {
		g.activateDash()
	}
	if controls.cloakPressed && g.cloakCooldown <= 0 {
		g.cloakTimer = 2.4
		g.cloakCooldown = 8.0
		g.showMessage("CLOAK ACTIVE", 1.1, color.RGBA{R: 111, G: 220, B: 255, A: 255})
		g.addBurst(g.player, color.RGBA{R: 111, G: 220, B: 255, A: 255}, 12, 85)
	}
	if controls.pulsePressed && g.pulseCooldown <= 0 {
		g.pulseTimer = 2.8
		g.pulseCooldown = 7.0
		g.showMessage("PHEROMONE PULSE!", 1.1, color.RGBA{R: 190, G: 143, B: 255, A: 255})
		g.addBurst(g.player, color.RGBA{R: 190, G: 143, B: 255, A: 255}, 24, 115)
	}

	g.updatePlayer(controls, dt)
	g.sucking = false
	if g.cloakTimer > 0 {
		g.cloakTimer = math.Max(0, g.cloakTimer-dt)
	}
	if g.pulseTimer > 0 {
		g.pulseTimer = math.Max(0, g.pulseTimer-dt)
	}
	if g.hasteTimer > 0 {
		g.hasteTimer = math.Max(0, g.hasteTimer-dt)
	}

	if g.slapActive {
		g.slapElapsed += dt
		if !g.slapImpactChecked && g.slapElapsed >= 0.43 {
			g.slapImpactChecked = true
			if distance(g.player, point{x: handX, y: handY}) < 101 {
				if g.shieldCharges > 0 {
					g.shieldCharges--
					g.showMessage("SHIELD BLOCK!", 1.1, color.RGBA{R: 100, G: 221, B: 255, A: 255})
					g.addBurst(g.player, color.RGBA{R: 100, G: 221, B: 255, A: 255}, 30, 145)
				} else {
					g.loseLife("KENA TEPUK!")
				}
			} else {
				g.scoreExact += 45
				g.score = int(g.scoreExact)
				g.stagePerfectEscapes++
				g.totalPerfectEscapes++
				g.escapedFlash = 0.7
				g.showMessage("PERFECT ESCAPE +45", 1.0, color.RGBA{R: 102, G: 232, B: 190, A: 255})
				g.addBurst(g.player, color.RGBA{R: 102, G: 232, B: 190, A: 255}, 18, 90)
			}
			g.itch = 18
			g.slapCooldown = 1.4
		}
		if g.slapElapsed >= 0.68 {
			g.slapActive = false
			g.slapElapsed = 0
			g.slapImpactChecked = false
		}
		g.updateParticles(dt)
		g.updateFloaters(dt)
		return nil
	}

	if controls.hold && g.insideHandZone() {
		g.sucking = true
		g.blood = math.Min(100, g.blood+22*dt)
		g.itch = math.Min(100, g.itch+27*dt*g.difficultyConfig().itchRate)
		g.stageProgress += 19 * dt * (1 + float64(g.combo)*0.045)
		g.comboTimer = 1.05
		g.comboTick += dt
		if g.comboTick >= 0.58 {
			g.comboTick -= 0.58
			if g.combo < 5 {
				g.combo++
				g.showMessage("COMBO x"+itoa(g.combo+1), 0.55, color.RGBA{R: 255, G: 209, B: 96, A: 255})
			}
		}
		multiplier := 1 + float64(g.combo)*0.22
		g.scoreExact += 12 * dt * multiplier
		g.score = int(g.scoreExact)
		g.feedParticle -= dt
		if g.feedParticle <= 0 {
			g.feedParticle = 0.12
			g.addParticle(point{x: g.player.x + math.Sin(g.elapsed*19)*5, y: g.player.y - 10}, point{x: math.Sin(g.elapsed*16) * 16, y: -35}, 0.42, 2.2, color.RGBA{R: 255, G: 90, B: 105, A: 235})
		}
	} else {
		g.itch = math.Max(0, g.itch-8.5*dt*g.difficultyConfig().itchDecay)
		g.comboTimer -= dt
		if g.comboTimer <= 0 {
			g.combo = 0
			g.comboTick = 0
		}
	}

	g.updateEnemies(dt)
	g.updateBossEncounter(dt)
	g.updatePowerups(dt)
	g.updateParticles(dt)
	g.updateFloaters(dt)
	g.powerSpawnTimer -= dt
	if g.powerSpawnTimer <= 0 && len(g.powerups) < 3 {
		g.spawnPowerup()
		g.powerSpawnTimer = 8.5 + float64((g.stageIndex+g.powerSequence)%3)*1.4
	}
	if g.itch >= 100 && g.slapCooldown <= 0 {
		g.slapActive = true
		g.slapElapsed = 0
		g.slapImpactChecked = false
		g.showMessage("MOVE! MOVE! MOVE!", 0.55, color.RGBA{R: 255, G: 79, B: 90, A: 255})
	}
	if g.state == statePlaying && g.stageProgress >= g.goalForStage() {
		g.calculateStageStars()
		g.state = stateStageClear
		g.stageTimer = 1.8
		g.addBurst(point{x: ScreenWidth / 2, y: 270}, g.currentStage().accent, 36, 125)
		g.showMessage("STAGE CLEARED!", 1.5, g.currentStage().accent)
	}
	if g.score > g.bestScore {
		g.bestScore = g.score
	}
	return nil
}

func (g *Game) calculateStageStars() {
	g.stageStars = 1
	if g.stageHitsTaken == 0 {
		g.stageStars++
	}
	if g.stagePowerupsCollected >= 2 || g.stagePerfectEscapes > 0 {
		g.stageStars++
	}
	if g.stageStars > 3 {
		g.stageStars = 3
	}
}

// updateBossEncounter runs the final-stage predator's marked-area strike pattern.
// Every strike warns the player first and rewards an evasion with bonus score.
func (g *Game) updateBossEncounter(dt float64) {
	if g.stageIndex != bossStage || g.state != statePlaying || g.bossStrikes >= 3 {
		return
	}
	if !g.bossAttackActive {
		g.bossAttackTimer -= dt
		if g.bossAttackTimer <= 0 {
			g.bossAttackActive = true
			g.bossAttackElapsed = 0
			g.bossImpactChecked = false
			g.bossAttackTarget = g.player
			g.showMessage("GIANT SWATTER INCOMING!", 0.85, color.RGBA{R: 255, G: 80, B: 91, A: 255})
			g.addBurst(g.bossAttackTarget, color.RGBA{R: 255, G: 80, B: 91, A: 255}, 14, 75)
		}
		return
	}

	g.bossAttackElapsed += dt
	if !g.bossImpactChecked && g.bossAttackElapsed >= 0.62 {
		g.bossImpactChecked = true
		protected := g.cloakTimer > 0 || g.dashTimer > 0 || g.invulnerability > 0
		insideStrike := distance(g.player, g.bossAttackTarget) < 58
		if insideStrike && !protected {
			if g.shieldCharges > 0 {
				g.shieldCharges--
				g.showMessage("SHIELD BLOCK!", 0.95, color.RGBA{R: 100, G: 221, B: 255, A: 255})
				g.addBurst(g.player, color.RGBA{R: 100, G: 221, B: 255, A: 255}, 24, 115)
			} else {
				g.loseLife("HIT BY THE GIANT SWATTER!")
			}
		} else {
			g.scoreExact += 65
			g.score = int(g.scoreExact)
			g.stagePerfectEscapes++
			g.totalPerfectEscapes++
			g.showMessage("BOSS DODGED +65", 0.95, color.RGBA{R: 102, G: 232, B: 190, A: 255})
			g.addBurst(g.player, color.RGBA{R: 102, G: 232, B: 190, A: 255}, 18, 95)
		}
		g.bossStrikes++
	}
	if g.bossAttackElapsed >= 0.92 {
		g.bossAttackActive = false
		g.bossAttackElapsed = 0
		g.bossImpactChecked = false
		g.bossAttackTimer = math.Max(3.8, 6.4/g.difficultyConfig().enemyScale)
	}
}

func (g *Game) updateTimers(dt float64) {
	if g.slapCooldown > 0 {
		g.slapCooldown = math.Max(0, g.slapCooldown-dt)
	}
	if g.dashCooldown > 0 {
		g.dashCooldown = math.Max(0, g.dashCooldown-dt)
	}
	if g.dashTimer > 0 {
		g.dashTimer = math.Max(0, g.dashTimer-dt)
	}
	if g.cloakCooldown > 0 {
		g.cloakCooldown = math.Max(0, g.cloakCooldown-dt)
	}
	if g.pulseCooldown > 0 {
		g.pulseCooldown = math.Max(0, g.pulseCooldown-dt)
	}
	if g.invulnerability > 0 {
		g.invulnerability = math.Max(0, g.invulnerability-dt)
	}
	if g.escapedFlash > 0 {
		g.escapedFlash = math.Max(0, g.escapedFlash-dt)
	}
	if g.damageFlash > 0 {
		g.damageFlash = math.Max(0, g.damageFlash-dt)
	}
	if g.messageTimer > 0 {
		g.messageTimer = math.Max(0, g.messageTimer-dt)
	}
}

func (g *Game) activateDash() {
	dx := g.player.x - handX
	dy := g.player.y - handY
	d := math.Hypot(dx, dy)
	if d < 1 {
		dx, dy, d = 0, -1, 1
	}
	g.dashDirection = point{x: dx / d, y: dy / d}
	g.dashTimer = 0.28
	if g.hasteTimer > 0 {
		g.dashTimer = 0.38
	}
	g.dashCooldown = 1.5
	if g.hasteTimer > 0 {
		g.dashCooldown = 0.62
	}
	g.addBurst(g.player, color.RGBA{R: 255, G: 211, B: 99, A: 255}, 10, 90)
}

func (g *Game) updatePlayer(c controlState, dt float64) {
	moveX, moveY := c.moveX, c.moveY
	speed := 246.0
	if c.pointerOn {
		dx := c.pointer.x - g.player.x
		dy := c.pointer.y - g.player.y
		d := math.Hypot(dx, dy)
		if d > 2 {
			moveX, moveY = dx/d, dy/d
		} else {
			moveX, moveY = 0, 0
		}
		speed = 360
	}
	if g.hasteTimer > 0 {
		speed *= 1.24
	}
	if g.dashTimer > 0 {
		moveX, moveY = g.dashDirection.x, g.dashDirection.y
		speed = 790
	}
	if math.Hypot(moveX, moveY) > 0.01 {
		g.heading = math.Atan2(moveY, moveX)
		g.player.x += moveX * speed * dt
		g.player.y += moveY * speed * dt
	}

	// Stage wind is a continuous, gentle force with a stronger vortex close to the fan.
	stage := g.currentStage()
	windScale := stage.wind * g.difficultyConfig().windScale
	windX := (-10.0 + math.Sin(g.elapsed*1.7)*5) * windScale
	windY := (15.0 + math.Cos(g.elapsed*1.3)*4) * windScale
	fanDist := distance(g.player, g.fanPushPoint)
	if fanDist < 112 {
		windX -= (112 - fanDist) * 0.48 * windScale
		windY += (112 - fanDist) * 0.68 * windScale
	}
	g.player.x += windX * dt
	g.player.y += windY * dt
	g.player.x = clamp(g.player.x, 13, ScreenWidth-13)
	g.player.y = clamp(g.player.y, 93, ScreenHeight-95)
}

func (g *Game) updateEnemies(dt float64) {
	hitThisFrame := false
	for i := range g.enemies {
		e := &g.enemies[i]
		if g.pulseTimer > 0 {
			continue
		}
		switch e.kind {
		case enemySpider:
			e.pos.x = e.origin.x + math.Sin(g.elapsed*0.92+e.phase)*58
			e.pos.y = e.origin.y + math.Cos(g.elapsed*1.33+e.phase)*17
		case enemyDragonfly:
			if g.cloakTimer > 0 {
				e.pos.x = e.origin.x + math.Sin(g.elapsed*1.7+e.phase)*25
				e.pos.y = e.origin.y + math.Cos(g.elapsed*1.2+e.phase)*20
			} else {
				dx, dy := g.player.x-e.pos.x, g.player.y-e.pos.y
				d := math.Hypot(dx, dy)
				if d > 1 {
					e.pos.x += dx / d * e.speed * dt
					e.pos.y += dy / d * e.speed * dt
				}
			}
		case enemyBeetle:
			if e.chargeRemaining > 0 {
				e.pos.x += e.chargeDir.x * e.speed * 2.05 * dt
				e.pos.y += e.chargeDir.y * e.speed * 2.05 * dt
				e.chargeRemaining = math.Max(0, e.chargeRemaining-dt)
			} else {
				e.chargeCooldown -= dt
				e.pos.x = e.origin.x + math.Sin(g.elapsed*0.72+e.phase)*16
				e.pos.y = e.origin.y + math.Cos(g.elapsed*1.1+e.phase)*12
				if e.chargeCooldown <= 0 && g.cloakTimer <= 0 {
					dx, dy := g.player.x-e.pos.x, g.player.y-e.pos.y
					d := math.Hypot(dx, dy)
					if d > 1 {
						e.chargeDir = point{x: dx / d, y: dy / d}
					}
					e.chargeRemaining = 0.42
					e.chargeCooldown = 2.0
				}
			}
		}
		e.pos.x = clamp(e.pos.x, 20, ScreenWidth-20)
		e.pos.y = clamp(e.pos.y, 100, ScreenHeight-85)

		protected := g.cloakTimer > 0 || g.pulseTimer > 0 || g.invulnerability > 0 || g.dashTimer > 0
		if !protected && !hitThisFrame && distance(g.player, e.pos) < e.radius+playerRadius+2 {
			hitThisFrame = true
			if g.shieldCharges > 0 {
				g.shieldCharges--
				g.showMessage("SHIELD BLOCK!", 0.9, color.RGBA{R: 100, G: 221, B: 255, A: 255})
				g.addBurst(g.player, color.RGBA{R: 100, G: 221, B: 255, A: 255}, 22, 110)
				g.invulnerability = 0.65
			} else {
				g.loseLife("CAUGHT BY A HUNTER!")
			}
		}
	}
}

func (g *Game) updatePowerups(dt float64) {
	kept := g.powerups[:0]
	for _, p := range g.powerups {
		p.age += dt
		if distance(g.player, p.pos) < 19 {
			g.collectPowerup(p.kind, p.pos)
			continue
		}
		kept = append(kept, p)
	}
	g.powerups = kept
}

func (g *Game) collectPowerup(kind powerupKind, pos point) {
	g.stagePowerupsCollected++
	g.totalPowerupsCollected++
	switch kind {
	case powerNectar:
		g.scoreExact += 150
		g.score = int(g.scoreExact)
		g.blood = math.Min(100, g.blood+18)
		g.stageProgress += 12
		g.showMessage("NECTAR +150", 1.0, color.RGBA{R: 255, G: 211, B: 99, A: 255})
		g.addBurst(pos, color.RGBA{R: 255, G: 211, B: 99, A: 255}, 18, 100)
	case powerChill:
		g.itch = math.Max(0, g.itch-42)
		g.showMessage("COOL MIST -ITCH", 1.0, color.RGBA{R: 106, G: 207, B: 255, A: 255})
		g.addBurst(pos, color.RGBA{R: 106, G: 207, B: 255, A: 255}, 20, 88)
	case powerShield:
		if g.shieldCharges < 2 {
			g.shieldCharges++
		}
		g.showMessage("SHIELD +1", 1.0, color.RGBA{R: 100, G: 221, B: 255, A: 255})
		g.addBurst(pos, color.RGBA{R: 100, G: 221, B: 255, A: 255}, 22, 105)
	case powerHaste:
		g.hasteTimer = 5.0
		g.dashCooldown = 0
		g.showMessage("WING OVERDRIVE!", 1.0, color.RGBA{R: 255, G: 164, B: 88, A: 255})
		g.addBurst(pos, color.RGBA{R: 255, G: 164, B: 88, A: 255}, 24, 120)
	}
}

func (g *Game) spawnPowerup() {
	spots := []point{{x: 49, y: 193}, {x: 309, y: 277}, {x: 50, y: 461}, {x: 307, y: 477}, {x: 180, y: 174}}
	pos := spots[g.powerSequence%len(spots)]
	kind := powerupKind(g.powerSequence % 4)
	g.powerSequence++
	g.powerups = append(g.powerups, powerup{kind: kind, pos: pos})
	g.addBurst(pos, powerupColor(kind), 8, 40)
}

func (g *Game) loseLife(reason string) {
	if g.invulnerability > 0 || g.state != statePlaying {
		return
	}
	g.lives--
	g.stageHitsTaken++
	g.totalHitsTaken++
	g.damageFlash = 0.65
	g.invulnerability = 1.25
	g.itch = 17
	g.slapCooldown = 1.5
	g.escapedFlash = 0.35
	g.gameOverReasonText = reason
	dx, dy := g.player.x-handX, g.player.y-handY
	d := math.Hypot(dx, dy)
	if d < 1 {
		dx, dy, d = 0, 1, 1
	}
	g.player.x = clamp(g.player.x+dx/d*40, 13, ScreenWidth-13)
	g.player.y = clamp(g.player.y+dy/d*40, 93, ScreenHeight-95)
	g.showMessage("OUCH! -1 LIFE", 1.0, color.RGBA{R: 255, G: 83, B: 100, A: 255})
	g.addBurst(g.player, color.RGBA{R: 255, G: 83, B: 100, A: 255}, 26, 130)
	if g.lives <= 0 {
		g.state = stateGameOver
		if g.score > g.bestScore {
			g.bestScore = g.score
		}
	}
}

func (g *Game) insideHandZone() bool {
	dx := (g.player.x - handX) / 91
	dy := (g.player.y - handY) / 70
	return dx*dx+dy*dy < 1
}

func (g *Game) addParticle(pos, velocity point, life, radius float64, c color.RGBA) {
	g.particles = append(g.particles, particle{pos: pos, velocity: velocity, life: life, maxLife: life, radius: radius, color: c})
}

func (g *Game) addBurst(pos point, c color.RGBA, count int, speed float64) {
	for i := 0; i < count; i++ {
		a := (float64(i)/float64(count))*math.Pi*2 + float64(g.stageIndex)*0.21
		v := speed * (0.45 + float64((i%5)+1)*0.11)
		g.addParticle(pos, point{x: math.Cos(a) * v, y: math.Sin(a) * v}, 0.35+float64(i%5)*0.08, 1.6+float64(i%3)*0.8, c)
	}
}

func (g *Game) updateParticles(dt float64) {
	kept := g.particles[:0]
	for _, p := range g.particles {
		p.life -= dt
		if p.life <= 0 {
			continue
		}
		p.pos.x += p.velocity.x * dt
		p.pos.y += p.velocity.y * dt
		p.velocity.x *= 0.985
		p.velocity.y = p.velocity.y*0.985 + 10*dt
		kept = append(kept, p)
	}
	g.particles = kept
}

func (g *Game) showMessage(message string, duration float64, c color.RGBA) {
	g.message = message
	g.messageTimer = duration
	g.floaters = append(g.floaters, floatingText{pos: point{x: ScreenWidth / 2, y: 210}, text: message, color: c, life: duration})
}

func (g *Game) updateFloaters(dt float64) {
	kept := g.floaters[:0]
	for _, f := range g.floaters {
		f.life -= dt
		f.pos.y -= 18 * dt
		if f.life > 0 {
			kept = append(kept, f)
		}
	}
	g.floaters = kept
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

type controlState struct {
	moveX, moveY   float64
	pointer        point
	pointerOn      bool
	hold           bool
	dashPressed    bool
	cloakPressed   bool
	pulsePressed   bool
	pausePressed   bool
	restartPressed bool
	confirm        bool
	pressPos       point
	hasPress       bool
}

func (g *Game) readControls() controlState {
	var c controlState
	mouseX, mouseY := ebiten.CursorPosition()
	mousePos := point{x: float64(mouseX), y: float64(mouseY)}
	mouseJustPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	c.confirm = mouseJustPressed || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace)
	c.hasPress = mouseJustPressed
	c.pressPos = mousePos
	c.pausePressed = inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP)
	c.restartPressed = inpututil.IsKeyJustPressed(ebiten.KeyR)
	c.cloakPressed = inpututil.IsKeyJustPressed(ebiten.KeyE)
	c.pulsePressed = inpututil.IsKeyJustPressed(ebiten.KeyQ)
	c.dashPressed = inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) || inpututil.IsKeyJustPressed(ebiten.KeyShiftRight)

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		c.moveX--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		c.moveX++
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		c.moveY--
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		c.moveY++
	}
	if c.moveX != 0 || c.moveY != 0 {
		n := math.Hypot(c.moveX, c.moveY)
		c.moveX, c.moveY = c.moveX/n, c.moveY/n
	}
	c.hold = ebiten.IsKeyPressed(ebiten.KeySpace)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if mouseJustPressed && inPauseButton(mousePos) {
			c.pausePressed = true
		} else if mouseJustPressed && inAbilityButton(mousePos, 0) {
			c.cloakPressed = true
		} else if mouseJustPressed && inAbilityButton(mousePos, 1) {
			c.pulsePressed = true
		} else if mouseJustPressed && inAbilityButton(mousePos, 2) {
			c.dashPressed = true
		} else if !inAnyAbilityButton(mousePos) && !inPauseButton(mousePos) {
			c.pointer, c.pointerOn, c.hold = mousePos, true, true
		}
	}

	currentTouches := g.currentTouches
	clear(currentTouches)
	for _, id := range ebiten.TouchIDs() {
		x, y := ebiten.TouchPosition(id)
		p := point{x: float64(x), y: float64(y)}
		currentTouches[id] = true
		if !g.lastTouches[id] {
			if !c.hasPress {
				c.pressPos = p
				c.hasPress = true
			}
		}
		if inPauseButton(p) {
			if !g.lastTouches[id] {
				c.pausePressed = true
			}
			continue
		}
		if inAnyAbilityButton(p) {
			if !g.lastTouches[id] {
				switch {
				case inAbilityButton(p, 0):
					c.cloakPressed = true
				case inAbilityButton(p, 1):
					c.pulsePressed = true
				case inAbilityButton(p, 2):
					c.dashPressed = true
				}
			}
			continue
		}
		if !g.lastTouches[id] {
			c.confirm = true
		}
		if !c.pointerOn {
			c.pointer = p
			c.pointerOn = true
			c.hold = true
		}
	}
	g.lastTouches, g.currentTouches = currentTouches, g.lastTouches
	return c
}

func inPauseButton(p point) bool {
	return p.x >= 314 && p.x <= 354 && p.y >= 47 && p.y <= 77
}

func difficultyButtonAt(p point) (difficultyLevel, bool) {
	if p.y < 352 || p.y > 378 {
		return difficultyNormal, false
	}
	switch {
	case p.x >= 37 && p.x <= 128:
		return difficultyChill, true
	case p.x >= 134 && p.x <= 224:
		return difficultyNormal, true
	case p.x >= 230 && p.x <= 323:
		return difficultyNightmare, true
	default:
		return difficultyNormal, false
	}
}

func inAbilityButton(p point, index int) bool {
	switch index {
	case 0:
		return p.x >= 15 && p.x <= 112 && p.y >= 555 && p.y <= 608
	case 1:
		return p.x >= 124 && p.x <= 238 && p.y >= 555 && p.y <= 608
	case 2:
		return p.x >= 249 && p.x <= 347 && p.y >= 555 && p.y <= 608
	default:
		return false
	}
}

func inAnyAbilityButton(p point) bool {
	return inAbilityButton(p, 0) || inAbilityButton(p, 1) || inAbilityButton(p, 2)
}

func distance(a, b point) float64 { return math.Hypot(a.x-b.x, a.y-b.y) }

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

func powerupColor(kind powerupKind) color.RGBA {
	switch kind {
	case powerNectar:
		return color.RGBA{R: 255, G: 211, B: 99, A: 255}
	case powerChill:
		return color.RGBA{R: 106, G: 207, B: 255, A: 255}
	case powerShield:
		return color.RGBA{R: 100, G: 221, B: 255, A: 255}
	default:
		return color.RGBA{R: 255, G: 164, B: 88, A: 255}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

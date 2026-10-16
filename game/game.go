package game

import (
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
)

type gameState uint8

const (
	stateMenu gameState = iota
	statePlaying
	stateGameOver
)

type point struct {
	x float64
	y float64
}

type Game struct {
	state gameState

	player  point
	heading float64
	elapsed float64

	score      int
	scoreExact float64
	bestScore  int
	blood      float64
	itch       float64
	sucking    bool

	slapActive         bool
	slapElapsed        float64
	slapImpactChecked  bool
	slapCooldown       float64
	escapedFlash       float64
	dashTimer          float64
	dashCooldown       float64
	dashDirection      point
	fanAngle           float64
	lastTouches        map[ebiten.TouchID]bool
	gameOverReasonText string
}

// New returns a fresh game instance. The logical canvas is always 360x640.
func New() *Game {
	g := &Game{lastTouches: make(map[ebiten.TouchID]bool)}
	g.resetRun()
	g.state = stateMenu
	return g
}

func (g *Game) resetRun() {
	g.player = point{x: 180, y: 500}
	g.heading = -math.Pi / 2
	g.elapsed = 0
	g.score = 0
	g.scoreExact = 0
	g.blood = 0
	g.itch = 0
	g.sucking = false
	g.slapActive = false
	g.slapElapsed = 0
	g.slapImpactChecked = false
	g.slapCooldown = 0
	g.escapedFlash = 0
	g.dashTimer = 0
	g.dashCooldown = 0
	g.dashDirection = point{x: 0, y: -1}
	g.fanAngle = 0
	g.gameOverReasonText = ""
}

func (g *Game) Update() error {
	const dt = 1.0 / 60.0
	g.elapsed += dt
	g.fanAngle += dt * 5.2
	if g.slapCooldown > 0 {
		g.slapCooldown -= dt
	}
	if g.dashCooldown > 0 {
		g.dashCooldown -= dt
	}
	if g.dashTimer > 0 {
		g.dashTimer -= dt
	}
	if g.escapedFlash > 0 {
		g.escapedFlash -= dt
	}

	controls := g.readControls()
	if g.state == stateMenu {
		if controls.confirm {
			g.resetRun()
			g.state = statePlaying
		}
		return nil
	}
	if g.state == stateGameOver {
		if controls.confirm {
			g.resetRun()
			g.state = statePlaying
		}
		return nil
	}

	// Dash is a short automatic burst away from the hand, so a phone player can
	// flee by tapping the on-screen DASH control with a second finger.
	if controls.dashPressed && g.dashCooldown <= 0 {
		dx := g.player.x - handX
		dy := g.player.y - handY
		d := math.Hypot(dx, dy)
		if d < 1 {
			dx, dy, d = 0, -1, 1
		}
		g.dashDirection = point{x: dx / d, y: dy / d}
		g.dashTimer = 0.27
		g.dashCooldown = 0.62
	}

	g.updatePlayer(controls, dt)
	g.sucking = false

	if g.slapActive {
		g.slapElapsed += dt
		// The hand hits at 0.43s, giving the player a short but readable escape window.
		if !g.slapImpactChecked && g.slapElapsed >= 0.43 {
			g.slapImpactChecked = true
			if distance(g.player, point{x: handX, y: handY}) < 101 {
				g.state = stateGameOver
				g.gameOverReasonText = "KENA TEPUK!"
				if g.score > g.bestScore {
					g.bestScore = g.score
				}
			} else {
				g.itch = 20
				g.slapCooldown = 1.15
				g.escapedFlash = 0.65
			}
		}
		if g.slapElapsed >= 0.68 {
			g.slapActive = false
			g.slapElapsed = 0
			g.slapImpactChecked = false
		}
		return nil
	}

	if controls.hold && g.insideHandZone() {
		g.sucking = true
		g.blood = math.Min(100, g.blood+25*dt)
		g.itch = math.Min(100, g.itch+27*dt)
		g.scoreExact += 12 * dt
		g.score = int(g.scoreExact)
	} else {
		g.itch = math.Max(0, g.itch-7.5*dt)
	}

	if g.itch >= 100 && g.slapCooldown <= 0 {
		g.slapActive = true
		g.slapElapsed = 0
		g.slapImpactChecked = false
	}

	if g.score > g.bestScore {
		g.bestScore = g.score
	}
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

type controlState struct {
	moveX       float64
	moveY       float64
	pointer     point
	pointerOn   bool
	hold        bool
	dashPressed bool
	confirm     bool
}

func (g *Game) readControls() controlState {
	var c controlState
	c.confirm = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace)

	// Keyboard fallback. It remains available while using a mouse or touch screen.
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
		c.moveX /= n
		c.moveY /= n
	}
	c.hold = ebiten.IsKeyPressed(ebiten.KeySpace)
	c.dashPressed = inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) || inpututil.IsKeyJustPressed(ebiten.KeyShiftRight)

	// A held mouse button follows the cursor and also means "suck" when over the hand.
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		c.pointer = point{x: float64(mx), y: float64(my)}
		c.pointerOn = true
		c.hold = true
	}

	// Touch points are tracked independently so one finger can fly while another taps DASH.
	currentTouches := make(map[ebiten.TouchID]bool)
	for _, id := range ebiten.TouchIDs() {
		x, y := ebiten.TouchPosition(id)
		p := point{x: float64(x), y: float64(y)}
		currentTouches[id] = true
		if g.state == statePlaying && inDashButton(p) {
			if !g.lastTouches[id] {
				c.dashPressed = true
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
	g.lastTouches = currentTouches
	return c
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
		speed = 370
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

	// The fan constantly nudges the mosquito, with a stronger draft near its blades.
	windX := -10.0 + math.Sin(g.elapsed*1.7)*5
	windY := 15.0 + math.Cos(g.elapsed*1.3)*4
	fanDist := distance(g.player, point{x: 307, y: 112})
	if fanDist < 112 {
		windX -= (112 - fanDist) * 0.48
		windY += (112 - fanDist) * 0.68
	}
	g.player.x += windX * dt
	g.player.y += windY * dt

	g.player.x = clamp(g.player.x, 13, ScreenWidth-13)
	g.player.y = clamp(g.player.y, 91, ScreenHeight-18)
}

func (g *Game) insideHandZone() bool {
	dx := (g.player.x - handX) / 91
	dy := (g.player.y - handY) / 70
	return dx*dx+dy*dy < 1
}

func inDashButton(p point) bool {
	return p.x >= 261 && p.x <= 347 && p.y >= 553 && p.y <= 604
}

func distance(a, b point) float64 {
	return math.Hypot(a.x-b.x, a.y-b.y)
}

func clamp(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

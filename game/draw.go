package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

type scaledTextKey struct {
	text  string
	color color.RGBA
}

var scaledTextCache = make(map[scaledTextKey]*ebiten.Image)

var (
	colPanel     = color.RGBA{R: 24, G: 39, B: 58, A: 255}
	colCream     = color.RGBA{R: 248, G: 240, B: 217, A: 255}
	colMint      = color.RGBA{R: 102, G: 232, B: 190, A: 255}
	colYellow    = color.RGBA{R: 255, G: 204, B: 92, A: 255}
	colRed       = color.RGBA{R: 255, G: 80, B: 91, A: 255}
	colDark      = color.RGBA{R: 24, G: 25, B: 38, A: 255}
	colSkin      = color.RGBA{R: 237, G: 174, B: 137, A: 255}
	colSkinLight = color.RGBA{R: 255, G: 205, B: 166, A: 255}
	colSkinShade = color.RGBA{R: 190, G: 119, B: 98, A: 255}
)

func (g *Game) Draw(screen *ebiten.Image) {
	drawBackground(screen, g.elapsed, g.stageIndex)
	drawFan(screen, g.fanAngle, g.currentStage().wind*g.difficultyConfig().windScale)
	if g.state == stateMenu {
		g.drawMenu(screen)
		return
	}

	g.drawHand(screen)
	g.drawPowerups(screen)
	g.drawEnemies(screen)
	g.drawParticles(screen)
	g.drawHUD(screen)
	g.drawSlap(screen)
	g.drawBossAttack(screen)
	g.drawMosquito(screen)
	g.drawAbilityEffects(screen)
	g.drawFloaters(screen)

	if g.escapedFlash > 0 {
		alpha := uint8(clamp(g.escapedFlash/0.7, 0, 1) * 170)
		fillRect(screen, 0, 88, ScreenWidth, 460, color.RGBA{R: 67, G: 227, B: 172, A: alpha})
		drawCenteredText(screen, "NICE ESCAPE!", 224, colMint)
	}
	if g.damageFlash > 0 {
		alpha := uint8(clamp(g.damageFlash/0.65, 0, 1) * 95)
		fillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 255, G: 32, B: 66, A: alpha})
	}

	switch g.state {
	case statePaused:
		fillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 5, G: 9, B: 20, A: 190})
		g.drawPauseOverlay(screen)
	case stateStageClear:
		g.drawStageClear(screen)
	case stateGameOver:
		fillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 7, G: 10, B: 21, A: 218})
		g.drawGameOver(screen)
	case stateVictory:
		fillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 7, G: 10, B: 21, A: 208})
		g.drawVictory(screen)
	}
}

func drawBackground(screen *ebiten.Image, elapsed float64, stageIndex int) {
	cfg := stages[0]
	if stageIndex >= 0 && stageIndex < len(stages) {
		cfg = stages[stageIndex]
	}
	screen.Fill(cfg.background)
	for x := 0; x <= ScreenWidth; x += 30 {
		vector.StrokeLine(screen, float32(x), 0, float32(x), ScreenHeight, 1, cfg.grid, false)
	}
	for y := 0; y <= ScreenHeight; y += 32 {
		vector.StrokeLine(screen, 0, float32(y), ScreenWidth, float32(y), 1, cfg.grid, false)
	}
	bandY := math.Mod(elapsed*16, ScreenHeight+55) - 55
	fillRect(screen, 0, bandY, ScreenWidth, 2, color.RGBA{R: 89, G: 156, B: 177, A: 45})

	switch stageIndex {
	case 1: // warm kitchen tiles
		for x := 16; x < ScreenWidth; x += 64 {
			for y := 92; y < ScreenHeight-76; y += 64 {
				vector.StrokeLine(screen, float32(x), float32(y), float32(x+48), float32(y), 1, color.RGBA{R: 255, G: 153, B: 91, A: 26}, false)
			}
		}
	case 2: // greenhouse leaves
		for i := 0; i < 11; i++ {
			x := float64((i*43 + 19) % ScreenWidth)
			y := 125 + math.Mod(float64(i*59)+elapsed*7, 390)
			angle := math.Sin(elapsed*0.8+float64(i)) * 0.4
			vector.StrokeLine(screen, float32(x), float32(y), float32(x+math.Sin(angle)*13), float32(y-17), 2, color.RGBA{R: 88, G: 210, B: 133, A: 38}, true)
		}
	case 3: // slow night fireflies
		for i := 0; i < 15; i++ {
			x := 18 + math.Mod(float64(i*71)+elapsed*5, 324)
			y := 115 + math.Mod(float64(i*47)+math.Sin(elapsed+float64(i))*16, 408)
			vector.FillCircle(screen, float32(x), float32(y), 1.4, color.RGBA{R: 190, G: 156, B: 255, A: 110}, true)
		}
	case 4: // diagonal storm streaks
		for i := 0; i < 13; i++ {
			x := math.Mod(float64(i*41)+elapsed*58, 420) - 30
			y := math.Mod(float64(i*67)+elapsed*92, 620)
			vector.StrokeLine(screen, float32(x), float32(y), float32(x-8), float32(y+19), 1, color.RGBA{R: 142, G: 175, B: 255, A: 48}, false)
		}
	}
	fillRect(screen, 0, 0, ScreenWidth, 96, color.RGBA{R: 6, G: 11, B: 22, A: 232})
	fillRect(screen, 0, 615, ScreenWidth, 25, color.RGBA{R: 6, G: 11, B: 22, A: 218})
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	fillRoundRect(screen, 21, 80, 318, 481, 22, colPanel)
	strokeRoundRect(screen, 21, 80, 318, 481, 22, 2, color.RGBA{R: 62, G: 91, B: 110, A: 255})
	drawCenteredScaledText(screen, "MOSQUITO", 131, 2.0, colCream)
	drawCenteredScaledText(screen, "REVENGE", 169, 2.0, colYellow)
	drawCenteredText(screen, "TINY WINGS. BIG PROBLEMS.", 196, colMint)
	g.drawHandAt(screen, 180, 274, 0.60)
	g.drawMosquitoAt(screen, point{x: 108, y: 266}, 0.1, 22)
	vector.StrokeCircle(screen, 180, 275, float32(70+math.Sin(g.elapsed*3)*3), 2, color.RGBA{R: 102, G: 232, B: 190, A: 100}, true)

	drawCenteredText(screen, "CHOOSE YOUR DIFFICULTY", 338, colCream)
	drawDifficultyButton(screen, 37, 346, 91, 30, difficultyChill, g.difficulty)
	drawDifficultyButton(screen, 134, 346, 90, 30, difficultyNormal, g.difficulty)
	drawDifficultyButton(screen, 230, 346, 93, 30, difficultyNightmare, g.difficulty)
	fillRoundRect(screen, 57, 388, 246, 45, 13, colMint)
	drawCenteredScaledText(screen, "TAP TO START", 417, 1.35, colDark)
	drawCenteredText(screen, "5 STAGES  /  3 ENEMY TYPES  /  4 POWER-UPS", 454, colCream)
	drawCenteredText(screen, "DRAG TO FLY  -  HOLD ON HAND TO FEED", 475, color.RGBA{R: 184, G: 204, B: 210, A: 255})
	drawCenteredText(screen, "DASH: SHIFT  /  CLOAK: E  /  PULSE: Q", 496, colYellow)
	drawCenteredText(screen, "Collect glowing items. Avoid hunters.", 517, color.RGBA{R: 169, G: 193, B: 202, A: 255})
	drawCenteredText(screen, "ESCAPE THE RED TARGET AREAS!", 538, colRed)
	drawCenteredText(screen, "Touch controls available on mobile", 584, color.RGBA{R: 147, G: 172, B: 184, A: 255})
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	stage := g.currentStage()
	fillRect(screen, 0, 0, ScreenWidth, 96, color.RGBA{R: 6, G: 11, B: 22, A: 232})
	drawText(screen, "SCORE", 12, 18, color.RGBA{R: 185, G: 204, B: 209, A: 255})
	drawText(screen, fmt.Sprintf("%05d", g.score), 12, 37, colYellow)
	drawText(screen, fmt.Sprintf("STAGE %d/%d", g.stageIndex+1, len(stages)), 118, 18, stage.accent)
	drawText(screen, stage.name, 118, 37, colCream)
	drawText(screen, "HP", 280, 18, color.RGBA{R: 185, G: 204, B: 209, A: 255})
	for i := 0; i < maxLives; i++ {
		c := color.RGBA{R: 57, G: 70, B: 82, A: 255}
		if i < g.lives {
			c = color.RGBA{R: 255, G: 76, B: 98, A: 255}
		}
		x := float32(286 + i*17)
		vector.FillCircle(screen, x-3, 30, 4.5, c, true)
		vector.FillCircle(screen, x+3, 30, 4.5, c, true)
		vector.StrokeLine(screen, x-7, 31, x, 39, 5, c, true)
		vector.StrokeLine(screen, x+7, 31, x, 39, 5, c, true)
	}

	drawText(screen, "BLOOD", 12, 59, colCream)
	drawBar(screen, 55, 48, 72, 13, g.blood/100, colRed)
	drawText(screen, fmt.Sprintf("%d%%", int(g.blood)), 58, 78, color.RGBA{R: 255, G: 132, B: 132, A: 255})
	drawText(screen, "ITCH", 138, 59, colCream)
	itchColor := colMint
	if g.itch >= 68 {
		itchColor = colRed
	} else if g.itch >= 38 {
		itchColor = colYellow
	}
	drawBar(screen, 174, 48, 108, 13, g.itch/100, itchColor)
	drawText(screen, fmt.Sprintf("%d%%", int(g.itch)), 288, 59, itchColor)
	if g.sucking {
		drawText(screen, "SLURP!", 12, 87, colMint)
	} else {
		drawText(screen, "OBJECTIVE", 12, 87, color.RGBA{R: 180, G: 199, B: 207, A: 255})
	}
	drawBar(screen, 78, 77, 204, 9, g.stageProgress/g.goalForStage(), stage.accent)
	drawPauseButton(screen)
	if g.combo > 0 {
		drawText(screen, fmt.Sprintf("COMBO x%d", g.combo+1), 286, 87, colYellow)
	}
	if g.itch >= 72 && !g.slapActive {
		drawCenteredText(screen, "WARNING: HAND GETTING ITCHY!", 111, colRed)
	}
}

func (g *Game) drawSlap(screen *ebiten.Image) {
	if !g.slapActive {
		return
	}
	p := g.slapElapsed / 0.68
	pulse := 1 + math.Sin(g.slapElapsed*46)*0.055
	radius := 86.0 + p*30
	vector.StrokeCircle(screen, handX, handY, float32(radius*pulse), 5, color.RGBA{R: 255, G: 64, B: 81, A: 255}, true)
	vector.StrokeCircle(screen, handX, handY, float32((radius-12)*pulse), 2, color.RGBA{R: 255, G: 128, B: 112, A: 255}, true)
	if g.slapElapsed < 0.43 {
		drawCenteredScaledText(screen, "SLAP INCOMING!", 202, 1.0, colRed)
		progress := clamp(g.slapElapsed/0.43, 0, 1)
		g.drawHandAt(screen, handX, handY-170+progress*170, 0.60)
	} else {
		progress := clamp((g.slapElapsed-0.43)/0.25, 0, 1)
		g.drawHandAt(screen, handX, handY+progress*25, 0.92)
		vector.StrokeLine(screen, 94, 259, 267, 430, 8, color.RGBA{R: 255, G: 225, B: 205, A: 255}, true)
		vector.StrokeLine(screen, 266, 259, 93, 430, 8, color.RGBA{R: 255, G: 225, B: 205, A: 255}, true)
	}
}

func (g *Game) drawMosquito(screen *ebiten.Image) {
	if g.invulnerability > 0 && int(g.elapsed*18)%2 == 0 {
		return
	}
	if g.cloakTimer > 0 {
		alpha := uint8(80 + 55*math.Sin(g.elapsed*15))
		vector.StrokeCircle(screen, float32(g.player.x), float32(g.player.y), 20, 2, color.RGBA{R: 111, G: 220, B: 255, A: alpha}, true)
	}
	if g.shieldCharges > 0 {
		vector.StrokeCircle(screen, float32(g.player.x), float32(g.player.y), 16+float32(math.Sin(g.elapsed*7)), 1.5, color.RGBA{R: 100, G: 221, B: 255, A: 180}, true)
	}
	g.drawMosquitoAt(screen, g.player, g.heading, g.blood)
}

func (g *Game) drawMosquitoAt(screen *ebiten.Image, p point, heading, blood float64) {
	backX := p.x - math.Cos(heading)*8
	backY := p.y - math.Sin(heading)*8
	bellyR := 4.5 + 0.095*blood
	bellyColor := color.RGBA{R: 90, G: 42, B: 48, A: 255}
	if blood > 0 {
		bellyColor = color.RGBA{R: uint8(125 + blood*1.15), G: uint8(36 + blood*0.14), B: uint8(52 + blood*0.10), A: 255}
	}
	wingAlpha := uint8(145 + (math.Sin(g.elapsed*43)+1)*45)
	vector.FillCircle(screen, float32(p.x-3), float32(p.y-8), 4.5, color.RGBA{R: 172, G: 226, B: 232, A: wingAlpha}, true)
	vector.FillCircle(screen, float32(p.x+4), float32(p.y-7), 4.5, color.RGBA{R: 172, G: 226, B: 232, A: wingAlpha}, true)
	vector.StrokeLine(screen, float32(p.x-2), float32(p.y+2), float32(p.x-8), float32(p.y+8), 1.5, colDark, true)
	vector.StrokeLine(screen, float32(p.x+1), float32(p.y+2), float32(p.x+7), float32(p.y+8), 1.5, colDark, true)
	vector.FillCircle(screen, float32(backX), float32(backY), float32(bellyR), bellyColor, true)
	vector.FillCircle(screen, float32(p.x), float32(p.y), 6.2, color.RGBA{R: 39, G: 47, B: 59, A: 255}, true)
	vector.FillCircle(screen, float32(p.x+1.5), float32(p.y-1.5), 1.5, colCream, true)
	needleX := p.x + math.Cos(heading)*18
	needleY := p.y + math.Sin(heading)*18
	vector.StrokeLine(screen, float32(p.x+4*math.Cos(heading)), float32(p.y+4*math.Sin(heading)), float32(needleX), float32(needleY), 1.8, color.RGBA{R: 230, G: 235, B: 219, A: 255}, true)
	if g.sucking {
		phase := math.Sin(g.elapsed*22) * 7
		vector.FillCircle(screen, float32(p.x+phase), float32(p.y-18), 2.5, colRed, true)
		vector.FillCircle(screen, float32(p.x-phase), float32(p.y-24), 1.7, color.RGBA{R: 255, G: 117, B: 107, A: 255}, true)
	}
	if g.dashTimer > 0 {
		vector.StrokeLine(screen, float32(p.x-g.dashDirection.x*10), float32(p.y-g.dashDirection.y*10), float32(p.x-g.dashDirection.x*24), float32(p.y-g.dashDirection.y*24), 3, color.RGBA{R: 255, G: 211, B: 99, A: 190}, true)
	}
}

func (g *Game) drawHand(screen *ebiten.Image) {
	g.drawHandAt(screen, handX, handY, 1)
	for i := 0; i < 32; i++ {
		a := float64(i) * math.Pi * 2 / 32
		x := handX + math.Cos(a)*98
		y := handY + math.Sin(a)*76
		c := color.RGBA{R: 102, G: 232, B: 190, A: 95}
		if g.itch > 65 {
			c = color.RGBA{R: 255, G: 80, B: 91, A: 185}
		}
		vector.FillCircle(screen, float32(x), float32(y), 1.7, c, true)
	}
}

func (g *Game) drawHandAt(screen *ebiten.Image, cx, cy, scale float64) {
	s := func(v float64) float64 { return v * scale }
	fillRoundRect(screen, cx-s(31), cy+s(27), s(62), s(83), s(20), colSkinShade)
	fillRoundRect(screen, cx-s(28), cy+s(24), s(56), s(82), s(18), colSkin)
	fillRoundRect(screen, cx-s(28), cy+s(24), s(56), s(24), s(10), color.RGBA{R: 222, G: 151, B: 124, A: 255})
	fillRect(screen, cx-s(43), cy-s(27), s(86), s(61), colSkin)
	vector.FillCircle(screen, float32(cx-s(27)), float32(cy-s(10)), float32(s(29)), colSkin, true)
	vector.FillCircle(screen, float32(cx+s(27)), float32(cy-s(10)), float32(s(29)), colSkin, true)
	vector.FillCircle(screen, float32(cx), float32(cy+s(1)), float32(s(39)), colSkin, true)
	fillCapsuleV(screen, cx-s(37), cy-s(78), s(17), s(71), s(8), colSkinLight)
	fillCapsuleV(screen, cx-s(17), cy-s(94), s(17), s(86), s(8), colSkinLight)
	fillCapsuleV(screen, cx+s(3), cy-s(90), s(17), s(82), s(8), colSkinLight)
	fillCapsuleV(screen, cx+s(23), cy-s(75), s(16), s(68), s(8), colSkinLight)
	fillRoundRect(screen, cx-s(34), cy-s(74), s(11), s(14), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	fillRoundRect(screen, cx-s(14), cy-s(89), s(11), s(14), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	fillRoundRect(screen, cx+s(6), cy-s(85), s(11), s(14), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	fillRoundRect(screen, cx+s(26), cy-s(70), s(10), s(13), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	vector.StrokeLine(screen, float32(cx-s(18)), float32(cy+s(7)), float32(cx+s(2)), float32(cy+s(14)), 1.5, colSkinShade, true)
	vector.StrokeLine(screen, float32(cx+s(4)), float32(cy-s(3)), float32(cx+s(23)), float32(cy+s(1)), 1.5, colSkinShade, true)
	vector.StrokeLine(screen, float32(cx+s(29)), float32(cy+s(9)), float32(cx+s(49)), float32(cy-s(2)), float32(s(18)), colSkin, true)
	vector.FillCircle(screen, float32(cx+s(48)), float32(cy-s(3)), float32(s(9)), colSkinLight, true)
}

func drawFan(screen *ebiten.Image, angle, strength float64) {
	cx, cy := 307.0, 129.0
	vector.FillCircle(screen, float32(cx), float32(cy), 35, color.RGBA{R: 35, G: 59, B: 74, A: 255}, true)
	vector.StrokeCircle(screen, float32(cx), float32(cy), 34, 3, color.RGBA{R: 106, G: 145, B: 158, A: 255}, true)
	for i := 0; i < 3; i++ {
		a := angle + float64(i)*math.Pi*2/3
		x1, y1 := cx+math.Cos(a)*5, cy+math.Sin(a)*5
		x2, y2 := cx+math.Cos(a+0.37)*26, cy+math.Sin(a+0.37)*26
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 9, color.RGBA{R: 131, G: 172, B: 178, A: 255}, true)
	}
	vector.FillCircle(screen, float32(cx), float32(cy), 6, colYellow, true)
	drawCenteredText(screen, "FAN", 177, color.RGBA{R: 133, G: 171, B: 185, A: 255})
	if strength > 1.1 {
		for i := 0; i < 3; i++ {
			y := 178 + float64(i)*12 + math.Sin(angle*2+float64(i))*3
			vector.StrokeLine(screen, 284, float32(y), 270, float32(y+5), 1.5, color.RGBA{R: 128, G: 184, B: 212, A: 115}, true)
		}
	}
}

func (g *Game) drawEnemies(screen *ebiten.Image) {
	for i := range g.enemies {
		e := &g.enemies[i]
		p := e.pos
		if g.pulseTimer > 0 {
			vector.StrokeCircle(screen, float32(p.x), float32(p.y), float32(e.radius+7), 2, color.RGBA{R: 190, G: 143, B: 255, A: 150}, true)
		}
		switch e.kind {
		case enemySpider:
			webX := e.origin.x
			vector.StrokeLine(screen, float32(webX), 97, float32(p.x), float32(p.y-e.radius), 1, color.RGBA{R: 190, G: 199, B: 212, A: 75}, true)
			for leg := 0; leg < 4; leg++ {
				side := -1.0
				if leg%2 == 1 {
					side = 1
				}
				offsetY := float64(leg/2)*5 - 2
				vector.StrokeLine(screen, float32(p.x+side*4), float32(p.y+offsetY), float32(p.x+side*12), float32(p.y+offsetY-4), 1.7, color.RGBA{R: 24, G: 25, B: 38, A: 255}, true)
				vector.StrokeLine(screen, float32(p.x+side*12), float32(p.y+offsetY-4), float32(p.x+side*14), float32(p.y+offsetY+3), 1.7, color.RGBA{R: 24, G: 25, B: 38, A: 255}, true)
			}
			vector.FillCircle(screen, float32(p.x), float32(p.y), 9.2, color.RGBA{R: 33, G: 32, B: 48, A: 255}, true)
			vector.FillCircle(screen, float32(p.x-3), float32(p.y-1), 1.8, colRed, true)
			vector.FillCircle(screen, float32(p.x+3), float32(p.y-1), 1.8, colRed, true)
		case enemyDragonfly:
			wing := color.RGBA{R: 164, G: 237, B: 238, A: 165}
			vector.StrokeLine(screen, float32(p.x-2), float32(p.y), float32(p.x-11), float32(p.y-10+math.Sin(g.elapsed*23+e.phase)*2), 4, wing, true)
			vector.StrokeLine(screen, float32(p.x+2), float32(p.y), float32(p.x+11), float32(p.y-10+math.Cos(g.elapsed*22+e.phase)*2), 4, wing, true)
			vector.StrokeLine(screen, float32(p.x), float32(p.y+2), float32(p.x), float32(p.y+11), 3, color.RGBA{R: 70, G: 183, B: 171, A: 255}, true)
			vector.FillCircle(screen, float32(p.x), float32(p.y-2), 5.5, color.RGBA{R: 85, G: 213, B: 196, A: 255}, true)
			vector.FillCircle(screen, float32(p.x-2), float32(p.y-4), 1.5, colCream, true)
			vector.FillCircle(screen, float32(p.x+2), float32(p.y-4), 1.5, colCream, true)
		case enemyBeetle:
			if e.chargeRemaining > 0 {
				vector.StrokeLine(screen, float32(p.x-e.chargeDir.x*24), float32(p.y-e.chargeDir.y*24), float32(p.x-e.chargeDir.x*7), float32(p.y-e.chargeDir.y*7), 4, color.RGBA{R: 255, G: 100, B: 93, A: 180}, true)
			}
			vector.FillCircle(screen, float32(p.x), float32(p.y), 12, color.RGBA{R: 127, G: 62, B: 63, A: 255}, true)
			vector.FillCircle(screen, float32(p.x-2), float32(p.y-2), 8, color.RGBA{R: 197, G: 88, B: 77, A: 255}, true)
			vector.StrokeLine(screen, float32(p.x), float32(p.y-8), float32(p.x+1), float32(p.y+6), 1.5, color.RGBA{R: 61, G: 38, B: 49, A: 255}, true)
			for leg := -1; leg <= 1; leg++ {
				vector.StrokeLine(screen, float32(p.x-7), float32(p.y+float64(leg)*4), float32(p.x-13), float32(p.y+float64(leg)*6), 1.8, colDark, true)
				vector.StrokeLine(screen, float32(p.x+7), float32(p.y+float64(leg)*4), float32(p.x+13), float32(p.y+float64(leg)*6), 1.8, colDark, true)
			}
		}
	}
}

func (g *Game) drawPowerups(screen *ebiten.Image) {
	for _, p := range g.powerups {
		c := powerupColor(p.kind)
		y := p.pos.y + math.Sin(p.age*4)*3
		r := 12 + math.Sin(p.age*5)*1.4
		vector.FillCircle(screen, float32(p.pos.x), float32(y), float32(r+4), color.RGBA{R: c.R, G: c.G, B: c.B, A: 35}, true)
		vector.FillCircle(screen, float32(p.pos.x), float32(y), float32(r), color.RGBA{R: c.R, G: c.G, B: c.B, A: 220}, true)
		vector.StrokeCircle(screen, float32(p.pos.x), float32(y), float32(r), 2, colCream, true)
		label := "N"
		switch p.kind {
		case powerChill:
			label = "C"
		case powerShield:
			label = "S"
		case powerHaste:
			label = ">"
		}
		drawText(screen, label, int(p.pos.x)-3, int(y)+4, colDark)
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		alpha := uint8(clamp(p.life/p.maxLife, 0, 1) * float64(p.color.A))
		c := color.RGBA{R: p.color.R, G: p.color.G, B: p.color.B, A: alpha}
		vector.FillCircle(screen, float32(p.pos.x), float32(p.pos.y), float32(p.radius), c, true)
	}
}

func (g *Game) drawFloaters(screen *ebiten.Image) {
	for _, f := range g.floaters {
		c := f.color
		c.A = uint8(clamp(f.life/1.1, 0.15, 1) * float64(c.A))
		drawCenteredText(screen, f.text, int(f.pos.y), c)
	}
}

func (g *Game) drawAbilityEffects(screen *ebiten.Image) {
	if g.pulseTimer > 0 {
		progress := 1 - g.pulseTimer/2.8
		r := 18 + progress*125
		alpha := uint8(180 * (1 - progress))
		vector.StrokeCircle(screen, float32(g.player.x), float32(g.player.y), float32(r), 3, color.RGBA{R: 190, G: 143, B: 255, A: alpha}, true)
		vector.StrokeCircle(screen, float32(g.player.x), float32(g.player.y), float32(r*0.72), 1.5, color.RGBA{R: 190, G: 143, B: 255, A: alpha}, true)
	}
	if g.hasteTimer > 0 {
		for i := 0; i < 3; i++ {
			y := g.player.y - 18 - float64(i)*7
			vector.StrokeLine(screen, float32(g.player.x-8), float32(y), float32(g.player.x+8), float32(y-3), 1.5, color.RGBA{R: 255, G: 164, B: 88, A: 155}, true)
		}
	}
	if g.messageTimer > 0 && g.message != "" {
		// The individual floating text provides the animation; the top banner keeps feedback readable.
		fillRoundRect(screen, 60, 117, 240, 25, 8, color.RGBA{R: 8, G: 16, B: 29, A: 190})
		drawCenteredText(screen, g.message, 134, colCream)
	}
	g.drawActionButtons(screen)
}

func (g *Game) drawActionButtons(screen *ebiten.Image) {
	drawAbilityButton(screen, 15, 555, 97, 53, "CLOAK", "E", g.cloakCooldown, 8.0, color.RGBA{R: 71, G: 173, B: 219, A: 255}, g.cloakTimer > 0)
	drawAbilityButton(screen, 124, 555, 114, 53, "PULSE", "Q", g.pulseCooldown, 7.0, color.RGBA{R: 171, G: 125, B: 239, A: 255}, g.pulseTimer > 0)
	drawAbilityButton(screen, 249, 555, 98, 53, "DASH", "SHIFT", g.dashCooldown, 1.5, colYellow, g.dashTimer > 0)
}

func drawDifficultyButton(screen *ebiten.Image, x, y, w, h float64, level, selected difficultyLevel) {
	cfg := difficultyLevels[int(level)]
	bg := color.RGBA{R: 31, G: 46, B: 64, A: 245}
	border := color.RGBA{R: 84, G: 107, B: 124, A: 255}
	textColor := color.RGBA{R: 170, G: 190, B: 202, A: 255}
	if level == selected {
		bg = color.RGBA{R: 35, G: 82, B: 75, A: 255}
		border = colMint
		textColor = colCream
	}
	fillRoundRect(screen, x, y, w, h, 8, bg)
	strokeRoundRect(screen, x, y, w, h, 8, 1.5, border)
	drawText(screen, cfg.name, int(x+(w-float64(len(cfg.name)*7))/2), int(y+19), textColor)
}

func drawPauseButton(screen *ebiten.Image) {
	fillRoundRect(screen, 314, 47, 40, 29, 7, color.RGBA{R: 31, G: 46, B: 64, A: 245})
	strokeRoundRect(screen, 314, 47, 40, 29, 7, 1.5, color.RGBA{R: 126, G: 151, B: 168, A: 255})
	drawCenteredTextAt(screen, "II", 327, 66, colCream)
}

func drawCenteredTextAt(screen *ebiten.Image, s string, x, y int, c color.Color) {
	drawText(screen, s, x-(len(s)*7)/2, y, c)
}

func drawAbilityButton(screen *ebiten.Image, x, y, w, h float64, label, key string, cooldown, cooldownMax float64, accent color.RGBA, active bool) {
	bg := color.RGBA{R: 31, G: 46, B: 64, A: 245}
	if cooldown <= 0 || active {
		bg = color.RGBA{R: accent.R / 3, G: accent.G / 3, B: accent.B / 3, A: 250}
	}
	fillRoundRect(screen, x, y, w, h, 10, bg)
	border := accent
	if cooldown > 0 && !active {
		border.A = 100
	}
	strokeRoundRect(screen, x, y, w, h, 10, 1.5, border)
	labelText := label
	labelColor := colCream
	if active {
		labelText += "!"
		labelColor = accent
	} else if cooldown > 0 {
		labelText += " " + fmt.Sprintf("%.0f", math.Ceil(cooldown))
		labelColor = color.RGBA{R: 152, G: 171, B: 184, A: 255}
	}
	drawText(screen, labelText, int(x+(w-float64(len(labelText)*7))/2), int(y+22), labelColor)
	progress := 1 - clamp(cooldown/cooldownMax, 0, 1)
	fillRoundRect(screen, x+5, y+h-9, (w-10)*progress, 3, 1.5, accent)
	drawText(screen, key, int(x+(w-float64(len(key)*7))/2), int(y+h-7), color.RGBA{R: 137, G: 157, B: 172, A: 255})
}

func (g *Game) drawBossAttack(screen *ebiten.Image) {
	if g.stageIndex != bossStage || !g.bossAttackActive {
		return
	}
	t := g.bossAttackElapsed
	target := g.bossAttackTarget
	if t < 0.62 {
		pulse := 1 + 0.06*math.Sin(g.elapsed*38)
		alpha := uint8(100 + 85*math.Abs(math.Sin(g.elapsed*12)))
		drawCenteredText(screen, fmt.Sprintf("SWATTER STRIKE %d/3", g.bossStrikes+1), 159, colRed)
		vector.FillCircle(screen, float32(target.x), float32(target.y), float32(53*pulse), color.RGBA{R: 255, G: 48, B: 65, A: 35}, true)
		vector.StrokeCircle(screen, float32(target.x), float32(target.y), float32(53*pulse), 3, color.RGBA{R: 255, G: 64, B: 81, A: alpha}, true)
		vector.StrokeCircle(screen, float32(target.x), float32(target.y), 35, 1.5, color.RGBA{R: 255, G: 128, B: 112, A: alpha}, true)
		vector.StrokeLine(screen, float32(target.x-13), float32(target.y), float32(target.x+13), float32(target.y), 2, colCream, true)
		vector.StrokeLine(screen, float32(target.x), float32(target.y-13), float32(target.x), float32(target.y+13), 2, colCream, true)
		drawCenteredTextAt(screen, "MOVE!", int(target.x), int(target.y-63), colRed)
	} else {
		vector.FillCircle(screen, float32(target.x), float32(target.y), 58, color.RGBA{R: 255, G: 48, B: 65, A: 105}, true)
		vector.StrokeCircle(screen, float32(target.x), float32(target.y), 58, 5, color.RGBA{R: 255, G: 210, B: 198, A: 240}, true)
		vector.StrokeLine(screen, float32(target.x+70), float32(target.y-100), float32(target.x+18), float32(target.y-19), 9, color.RGBA{R: 210, G: 186, B: 147, A: 255}, true)
		drawCenteredTextAt(screen, "WHAM!", int(target.x), int(target.y-67), colCream)
	}
}

func (g *Game) drawPauseOverlay(screen *ebiten.Image) {
	fillRoundRect(screen, 33, 220, 294, 200, 22, colPanel)
	strokeRoundRect(screen, 33, 220, 294, 200, 22, 2, colMint)
	drawCenteredScaledText(screen, "PAUSED", 272, 1.7, colCream)
	drawCenteredText(screen, "Take a breath, tiny hunter.", 309, color.RGBA{R: 184, G: 204, B: 210, A: 255})
	fillRoundRect(screen, 69, 331, 222, 42, 12, colMint)
	drawCenteredText(screen, "TAP / P / ESC TO RESUME", 357, colDark)
	drawCenteredText(screen, "PRESS R TO RESTART RUN", 397, colYellow)
}

func (g *Game) drawStageClear(screen *ebiten.Image) {
	fillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 7, G: 13, B: 27, A: 205})
	fillRoundRect(screen, 27, 200, 306, 242, 20, colPanel)
	strokeRoundRect(screen, 27, 200, 306, 242, 20, 2, g.currentStage().accent)
	drawCenteredScaledText(screen, "STAGE CLEAR!", 250, 1.45, g.currentStage().accent)
	drawCenteredText(screen, g.currentStage().name, 279, colCream)
	drawCenteredText(screen, g.currentStage().description, 299, color.RGBA{R: 173, G: 194, B: 205, A: 255})
	drawCenteredText(screen, fmt.Sprintf("SCORE  %05d", g.score), 323, colYellow)
	drawCenteredText(screen, fmt.Sprintf("LIVES  %d/%d", g.lives, g.difficultyConfig().lives), 346, colMint)
	drawCenteredText(screen, fmt.Sprintf("STAGE RANK  %d/3", g.stageStars), 369, colCream)
	for i := 0; i < 3; i++ {
		c := color.RGBA{R: 67, G: 83, B: 97, A: 255}
		if i < g.stageStars {
			c = colYellow
		}
		vector.FillCircle(screen, float32(154+i*26), 381, 5.5, c, true)
	}
	if g.stageIndex+1 < len(stages) {
		drawCenteredText(screen, "NEXT: "+stages[g.stageIndex+1].name, 404, color.RGBA{R: 185, G: 204, B: 210, A: 255})
		drawCenteredText(screen, "TAP TO CONTINUE", 427, colCream)
	} else {
		drawCenteredText(screen, "FINAL STAGE COMPLETE!", 404, colYellow)
		drawCenteredText(screen, "TAP TO CELEBRATE", 427, colCream)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	fillRoundRect(screen, 29, 176, 302, 286, 22, colPanel)
	strokeRoundRect(screen, 29, 176, 302, 286, 22, 2, color.RGBA{R: 255, G: 80, B: 91, A: 255})
	drawCenteredScaledText(screen, "SPLAT!", 229, 2.1, colRed)
	drawCenteredText(screen, g.gameOverReasonText, 260, colCream)
	drawCenteredText(screen, fmt.Sprintf("SCORE  %05d", g.score), 309, colYellow)
	drawCenteredText(screen, fmt.Sprintf("BEST   %05d", g.bestScore), 338, colMint)
	drawCenteredText(screen, fmt.Sprintf("REACHED STAGE %d/%d  |  %s", g.stageIndex+1, len(stages), g.difficultyConfig().name), 359, color.RGBA{R: 183, G: 202, B: 211, A: 255})
	drawCenteredText(screen, fmt.Sprintf("PICKUPS %d  /  CLEAN ESCAPES %d", g.totalPowerupsCollected, g.totalPerfectEscapes), 380, color.RGBA{R: 183, G: 202, B: 211, A: 255})
	fillRoundRect(screen, 76, 390, 208, 49, 13, colMint)
	drawCenteredText(screen, "TAP TO RETRY", 420, colDark)
	drawCenteredText(screen, "Use cloak, pulse and shields wisely.", 453, color.RGBA{R: 167, G: 184, B: 193, A: 255})
}

func (g *Game) drawVictory(screen *ebiten.Image) {
	fillRoundRect(screen, 25, 157, 310, 345, 23, colPanel)
	strokeRoundRect(screen, 25, 157, 310, 345, 23, 2, colYellow)
	drawCenteredScaledText(screen, "MOSQUITO", 205, 1.65, colCream)
	drawCenteredScaledText(screen, "CHAMPION!", 245, 1.65, colYellow)
	drawCenteredText(screen, "ALL 5 STAGES CLEARED", 282, colMint)
	drawCenteredText(screen, fmt.Sprintf("FINAL SCORE  %05d", g.score), 318, colCream)
	drawCenteredText(screen, fmt.Sprintf("BEST SCORE   %05d", g.bestScore), 344, colYellow)
	drawCenteredText(screen, fmt.Sprintf("HITS %d  /  POWER-UPS %d", g.totalHitsTaken, g.totalPowerupsCollected), 371, color.RGBA{R: 183, G: 202, B: 211, A: 255})
	drawCenteredText(screen, fmt.Sprintf("PERFECT ESCAPES %d  /  MODE %s", g.totalPerfectEscapes, g.difficultyConfig().name), 392, color.RGBA{R: 183, G: 202, B: 211, A: 255})
	fillRoundRect(screen, 76, 412, 208, 49, 13, colMint)
	drawCenteredText(screen, "PLAY AGAIN", 442, colDark)
	drawCenteredText(screen, "The tiniest terror wins again.", 480, color.RGBA{R: 167, G: 184, B: 193, A: 255})
}

func drawBar(screen *ebiten.Image, x, y, w, h int, value float64, fill color.Color) {
	fillRoundRect(screen, float64(x), float64(y), float64(w), float64(h), float64(h/2), color.RGBA{R: 47, G: 63, B: 77, A: 255})
	inner := float64(w-4) * clamp(value, 0, 1)
	if inner > 0 {
		fillRoundRect(screen, float64(x+2), float64(y+2), inner, float64(h-4), float64((h-4)/2), fill)
	}
}

func drawText(screen *ebiten.Image, s string, x, y int, c color.Color) {
	text.Draw(screen, s, basicfont.Face7x13, x, y, c)
}

func drawCenteredText(screen *ebiten.Image, s string, y int, c color.Color) {
	x := (ScreenWidth - len(s)*7) / 2
	drawText(screen, s, x, y, c)
}

func drawCenteredScaledText(screen *ebiten.Image, s string, baselineY int, scale float64, c color.RGBA) {
	key := scaledTextKey{text: s, color: c}
	glyphs, ok := scaledTextCache[key]
	if !ok {
		glyphs = ebiten.NewImage(len(s)*7+4, 16)
		text.Draw(glyphs, s, basicfont.Face7x13, 2, 13, c)
		scaledTextCache[key] = glyphs
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	x := (float64(ScreenWidth) - float64(glyphs.Bounds().Dx())*scale) / 2
	y := float64(baselineY) - 13*scale
	op.GeoM.Translate(x, y)
	screen.DrawImage(glyphs, op)
}

func fillRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	if w <= 0 || h <= 0 {
		return
	}
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, true)
}

func fillRoundRect(dst *ebiten.Image, x, y, w, h, r float64, c color.Color) {
	if r <= 0 {
		fillRect(dst, x, y, w, h, c)
		return
	}
	r = math.Min(r, math.Min(w/2, h/2))
	fillRect(dst, x+r, y, w-2*r, h, c)
	fillRect(dst, x, y+r, w, h-2*r, c)
	vector.FillCircle(dst, float32(x+r), float32(y+r), float32(r), c, true)
	vector.FillCircle(dst, float32(x+w-r), float32(y+r), float32(r), c, true)
	vector.FillCircle(dst, float32(x+r), float32(y+h-r), float32(r), c, true)
	vector.FillCircle(dst, float32(x+w-r), float32(y+h-r), float32(r), c, true)
}

func strokeRoundRect(dst *ebiten.Image, x, y, w, h, r, stroke float64, c color.Color) {
	vector.StrokeLine(dst, float32(x+r), float32(y), float32(x+w-r), float32(y), float32(stroke), c, true)
	vector.StrokeLine(dst, float32(x+r), float32(y+h), float32(x+w-r), float32(y+h), float32(stroke), c, true)
	vector.StrokeLine(dst, float32(x), float32(y+r), float32(x), float32(y+h-r), float32(stroke), c, true)
	vector.StrokeLine(dst, float32(x+w), float32(y+r), float32(x+w), float32(y+h-r), float32(stroke), c, true)
	vector.StrokeCircle(dst, float32(x+r), float32(y+r), float32(r), float32(stroke), c, true)
	vector.StrokeCircle(dst, float32(x+w-r), float32(y+r), float32(r), float32(stroke), c, true)
	vector.StrokeCircle(dst, float32(x+r), float32(y+h-r), float32(r), float32(stroke), c, true)
	vector.StrokeCircle(dst, float32(x+w-r), float32(y+h-r), float32(r), float32(stroke), c, true)
}

func fillCapsuleV(dst *ebiten.Image, x, y, w, h, r float64, c color.Color) {
	if r > w/2 {
		r = w / 2
	}
	fillRect(dst, x, y+r, w, h-2*r, c)
	vector.FillCircle(dst, float32(x+w/2), float32(y+r), float32(r), c, true)
	vector.FillCircle(dst, float32(x+w/2), float32(y+h-r), float32(r), c, true)
}

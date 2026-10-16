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
	colBackground = color.RGBA{R: 15, G: 25, B: 43, A: 255}
	colGrid       = color.RGBA{R: 31, G: 47, B: 65, A: 255}
	colPanel      = color.RGBA{R: 24, G: 39, B: 58, A: 255}
	colCream      = color.RGBA{R: 248, G: 240, B: 217, A: 255}
	colMint       = color.RGBA{R: 102, G: 232, B: 190, A: 255}
	colYellow     = color.RGBA{R: 255, G: 204, B: 92, A: 255}
	colRed        = color.RGBA{R: 255, G: 80, B: 91, A: 255}
	colDark       = color.RGBA{R: 24, G: 25, B: 38, A: 255}
	colSkin       = color.RGBA{R: 237, G: 174, B: 137, A: 255}
	colSkinLight  = color.RGBA{R: 255, G: 205, B: 166, A: 255}
	colSkinShade  = color.RGBA{R: 190, G: 119, B: 98, A: 255}
)

func (g *Game) Draw(screen *ebiten.Image) {
	drawBackground(screen, g.elapsed)
	drawFan(screen, g.fanAngle)

	switch g.state {
	case stateMenu:
		g.drawMenu(screen)
	case statePlaying:
		g.drawHand(screen)
		g.drawHUD(screen)
		g.drawSlap(screen)
		g.drawMosquito(screen)
		if g.escapedFlash > 0 {
			drawCenteredText(screen, "NICE ESCAPE!", 217, colMint)
		}
	case stateGameOver:
		g.drawHand(screen)
		g.drawHUD(screen)
		g.drawMosquito(screen)
		fillRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{R: 7, G: 10, B: 21, A: 218})
		g.drawGameOver(screen)
	}
}

func drawBackground(screen *ebiten.Image, elapsed float64) {
	screen.Fill(colBackground)
	for x := 0; x <= ScreenWidth; x += 30 {
		vector.StrokeLine(screen, float32(x), 0, float32(x), ScreenHeight, 1, colGrid, false)
	}
	for y := 0; y <= ScreenHeight; y += 32 {
		vector.StrokeLine(screen, 0, float32(y), ScreenWidth, float32(y), 1, colGrid, false)
	}
	// Subtle moving scan band gives the arena a little arcade character.
	bandY := math.Mod(elapsed*16, ScreenHeight+55) - 55
	fillRect(screen, 0, bandY, ScreenWidth, 2, color.RGBA{R: 59, G: 97, B: 113, A: 65})
	fillRect(screen, 0, 0, ScreenWidth, 76, color.RGBA{R: 10, G: 17, B: 31, A: 230})
	fillRect(screen, 0, 618, ScreenWidth, 22, color.RGBA{R: 10, G: 17, B: 31, A: 225})
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	fillRoundRect(screen, 25, 102, 310, 434, 22, colPanel)
	strokeRoundRect(screen, 25, 102, 310, 434, 22, 2, color.RGBA{R: 62, G: 91, B: 110, A: 255})
	drawCenteredScaledText(screen, "MOSQUITO", 157, 2.1, colCream)
	drawCenteredScaledText(screen, "REVENGE", 197, 2.1, colYellow)
	drawCenteredText(screen, "TINY WINGS. BIG PROBLEMS.", 228, colMint)
	g.drawHandAt(screen, 180, 357, 0.72)
	g.drawMosquitoAt(screen, point{x: 107, y: 345}, 0.1, 22)
	fillRoundRect(screen, 79, 448, 202, 55, 14, colMint)
	drawCenteredScaledText(screen, "TAP TO START", 482, 1.45, colDark)
	drawCenteredText(screen, "DRAG TO FLY  /  HOLD TO FEED", 521, colCream)
	drawCenteredText(screen, "ESCAPE THE SLAP ZONE!", 541, colYellow)
	drawCenteredText(screen, "PC: WASD / ARROWS, SPACE, SHIFT", 586, color.RGBA{R: 150, G: 172, B: 184, A: 255})
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	fillRect(screen, 0, 0, ScreenWidth, 78, color.RGBA{R: 10, G: 17, B: 31, A: 235})
	drawText(screen, "SCORE", 14, 22, colCream)
	drawText(screen, fmt.Sprintf("%05d", g.score), 14, 43, colYellow)
	drawText(screen, "BLOOD", 147, 22, colCream)
	drawBar(screen, 147, 31, 91, 12, g.blood/100, colRed)
	drawText(screen, fmt.Sprintf("%d%%", int(g.blood)), 248, 42, colRed)
	drawText(screen, "ITCH", 147, 59, colCream)
	itchColor := colMint
	if g.itch >= 68 {
		itchColor = colRed
	} else if g.itch >= 38 {
		itchColor = colYellow
	}
	drawBar(screen, 184, 49, 133, 12, g.itch/100, itchColor)
	if g.sucking {
		drawText(screen, "SLURP!", 18, 66, colMint)
	} else {
		drawText(screen, "STAY SNEAKY", 18, 66, color.RGBA{R: 155, G: 177, B: 192, A: 255})
	}

	if g.itch >= 72 && !g.slapActive {
		drawCenteredText(screen, "WARNING: HAND GETTING ITCHY!", 103, colRed)
	}

	if g.dashCooldown > 0 {
		cooldownProgress := 1 - clamp(g.dashCooldown/0.62, 0, 1)
		fillRoundRect(screen, 261, 562, 86, 42, 10, color.RGBA{R: 35, G: 51, B: 67, A: 245})
		fillRoundRect(screen, 263, 597, 82*cooldownProgress, 5, 2, colMint)
		drawText(screen, "DASH", 288, 586, color.RGBA{R: 167, G: 184, B: 193, A: 255})
	} else {
		fillRoundRect(screen, 261, 562, 86, 42, 10, colYellow)
		drawText(screen, "DASH!", 284, 586, colDark)
	}
	drawText(screen, "SHIFT", 287, 617, color.RGBA{R: 135, G: 159, B: 178, A: 255})
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
		drawCenteredText(screen, "SLAP INCOMING!", 197, colRed)
		startY := handY - 170
		progress := clamp(g.slapElapsed/0.43, 0, 1)
		strikeY := startY + progress*170
		g.drawHandAt(screen, handX, strikeY, 0.60)
	} else {
		// A fast descending hand and twin motion streaks sell the slap impact.
		progress := clamp((g.slapElapsed-0.43)/0.25, 0, 1)
		g.drawHandAt(screen, handX, handY+progress*25, 0.92)
		vector.StrokeLine(screen, 94, 259, 267, 430, 8, color.RGBA{R: 255, G: 225, B: 205, A: 255}, true)
		vector.StrokeLine(screen, 266, 259, 93, 430, 8, color.RGBA{R: 255, G: 225, B: 205, A: 255}, true)
	}
}

func (g *Game) drawMosquito(screen *ebiten.Image) {
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
	// Wings are translucent geometry; the body and abdomen scale as blood is consumed.
	vector.FillCircle(screen, float32(p.x-3), float32(p.y-8), 4.5, color.RGBA{R: 172, G: 226, B: 232, A: 185}, true)
	vector.FillCircle(screen, float32(p.x+4), float32(p.y-7), 4.5, color.RGBA{R: 172, G: 226, B: 232, A: 185}, true)
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
}

func (g *Game) drawHand(screen *ebiten.Image) {
	g.drawHandAt(screen, handX, handY, 1)
	// A dotted ellipse makes the target zone legible without external textures.
	for i := 0; i < 32; i++ {
		a := float64(i) * math.Pi * 2 / 32
		x := handX + math.Cos(a)*98
		y := handY + math.Sin(a)*76
		c := color.RGBA{R: 102, G: 232, B: 190, A: 115}
		if g.itch > 65 {
			c = color.RGBA{R: 255, G: 80, B: 91, A: 190}
		}
		vector.FillCircle(screen, float32(x), float32(y), 1.7, c, true)
	}
}

func (g *Game) drawHandAt(screen *ebiten.Image, cx, cy, scale float64) {
	s := func(v float64) float64 { return v * scale }
	// Wrist and forearm.
	fillRoundRect(screen, cx-s(31), cy+s(27), s(62), s(83), s(20), colSkinShade)
	fillRoundRect(screen, cx-s(28), cy+s(24), s(56), s(82), s(18), colSkin)
	fillRoundRect(screen, cx-s(28), cy+s(24), s(56), s(24), s(10), color.RGBA{R: 222, G: 151, B: 124, A: 255})
	// Palm: overlapping circles with a rectangle make a soft hand silhouette.
	fillRect(screen, cx-s(43), cy-s(27), s(86), s(61), colSkin)
	vector.FillCircle(screen, float32(cx-s(27)), float32(cy-s(10)), float32(s(29)), colSkin, true)
	vector.FillCircle(screen, float32(cx+s(27)), float32(cy-s(10)), float32(s(29)), colSkin, true)
	vector.FillCircle(screen, float32(cx), float32(cy+s(1)), float32(s(39)), colSkin, true)
	// Four upright fingers with rounded tips.
	fillCapsuleV(screen, cx-s(37), cy-s(78), s(17), s(71), s(8), colSkinLight)
	fillCapsuleV(screen, cx-s(17), cy-s(94), s(17), s(86), s(8), colSkinLight)
	fillCapsuleV(screen, cx+s(3), cy-s(90), s(17), s(82), s(8), colSkinLight)
	fillCapsuleV(screen, cx+s(23), cy-s(75), s(16), s(68), s(8), colSkinLight)
	// Nails and stylized crease marks.
	fillRoundRect(screen, cx-s(34), cy-s(74), s(11), s(14), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	fillRoundRect(screen, cx-s(14), cy-s(89), s(11), s(14), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	fillRoundRect(screen, cx+s(6), cy-s(85), s(11), s(14), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	fillRoundRect(screen, cx+s(26), cy-s(70), s(10), s(13), s(5), color.RGBA{R: 255, G: 227, B: 206, A: 255})
	vector.StrokeLine(screen, float32(cx-s(18)), float32(cy+s(7)), float32(cx+s(2)), float32(cy+s(14)), 1.5, colSkinShade, true)
	vector.StrokeLine(screen, float32(cx+s(4)), float32(cy-s(3)), float32(cx+s(23)), float32(cy+s(1)), 1.5, colSkinShade, true)
	// Thumb sticks out to the right.
	vector.StrokeLine(screen, float32(cx+s(29)), float32(cy+s(9)), float32(cx+s(49)), float32(cy-s(2)), float32(s(18)), colSkin, true)
	vector.FillCircle(screen, float32(cx+s(48)), float32(cy-s(3)), float32(s(9)), colSkinLight, true)
}

func drawFan(screen *ebiten.Image, angle float64) {
	cx, cy := 307.0, 112.0
	vector.FillCircle(screen, float32(cx), float32(cy), 39, color.RGBA{R: 35, G: 59, B: 74, A: 255}, true)
	vector.StrokeCircle(screen, float32(cx), float32(cy), 38, 3, color.RGBA{R: 106, G: 145, B: 158, A: 255}, true)
	for i := 0; i < 3; i++ {
		a := angle + float64(i)*math.Pi*2/3
		x1, y1 := cx+math.Cos(a)*5, cy+math.Sin(a)*5
		x2, y2 := cx+math.Cos(a+0.37)*29, cy+math.Sin(a+0.37)*29
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 10, color.RGBA{R: 131, G: 172, B: 178, A: 255}, true)
	}
	vector.FillCircle(screen, float32(cx), float32(cy), 7, colYellow, true)
	drawCenteredText(screen, "FAN", 163, color.RGBA{R: 133, G: 171, B: 185, A: 255})
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	fillRoundRect(screen, 29, 176, 302, 286, 22, colPanel)
	strokeRoundRect(screen, 29, 176, 302, 286, 22, 2, color.RGBA{R: 255, G: 80, B: 91, A: 255})
	drawCenteredScaledText(screen, "SPLAT!", 229, 2.1, colRed)
	drawCenteredText(screen, g.gameOverReasonText, 260, colCream)
	drawCenteredText(screen, fmt.Sprintf("SCORE  %05d", g.score), 309, colYellow)
	drawCenteredText(screen, fmt.Sprintf("BEST   %05d", g.bestScore), 338, colMint)
	fillRoundRect(screen, 76, 371, 208, 54, 13, colMint)
	drawCenteredText(screen, "TAP TO RETRY", 403, colDark)
	drawCenteredText(screen, "A human hand never forgets...", 445, color.RGBA{R: 167, G: 184, B: 193, A: 255})
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

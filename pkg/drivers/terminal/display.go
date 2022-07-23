package terminal

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/tvarney/grogue/pkg/game"
	"github.com/tvarney/grogue/pkg/game/chunk"
	"github.com/tvarney/grogue/pkg/game/color"
	"github.com/tvarney/grogue/pkg/game/tile"
)

var (
	titleStyle    = tcell.StyleDefault.Bold(true).Underline(true)
	optionStyle   = tcell.StyleDefault
	selectedStyle = tcell.StyleDefault.Underline(true).Foreground(tcell.ColorTeal)

	grassStyles = []tcell.Style{
		tcell.StyleDefault.Foreground(tcell.NewHexColor(int32(color.DarkGreen.Value()))),
		tcell.StyleDefault.Foreground(tcell.NewHexColor(int32(color.Green.Value()))),
	}
	emptyStyle = tcell.StyleDefault.
			Foreground(tcell.NewHexColor(0x001111)).
			Background(tcell.NewHexColor(0x002222))
	playerStyle = tcell.StyleDefault.Bold(true)
)

const (
	cursor      = "* "
	eraseCursor = "  "
)

func (d *Driver) Clear() {
	d.screen.Clear()
}

func (d *Driver) Draw(app *game.Application) {
	menu := app.GetMenu()
	if menu != nil {
		d.drawMenu(app, menu)
		return
	}

	if !app.InGame {
		app.Quit()
		return
	}
	d.drawGame(app)
}

func (d *Driver) drawGame(app *game.Application) {
	game := app.Game
	currChunk := game.ActiveChunks[(game.Player.Chunk.X+1)+(game.Player.Chunk.Y+1)*3]

	for y := uint16(0); y < chunk.Length; y++ {
		if int(y) >= d.height {
			break
		}
		for x := uint16(0); x < chunk.Width; x++ {
			if int(x) >= d.width {
				break
			}

			t := currChunk.Get(int(x), int(y), game.Player.Z)

			block := t.Block.Definition
			floor := t.Floor.Definition
			if block == tile.BlockEmpty {
				if floor == tile.FloorEmpty {
					d.screen.SetContent(int(x), int(y), '.', nil, emptyStyle)
				} else {
					if t.Flags&tile.HasGrass != 0 {
						// Tile selection from a random value uses the random value as-is.
						// This decouples the value from that selection.
						r := (t.Random >> 16) | (t.Random << 16)
						gc := grassStyles[int(r)%len(grassStyles)]
						d.screen.SetContent(int(x), int(y), d.grass.Rune(0, 0, x, y, t), nil, gc)
					} else {
						mat := game.Materials[t.Floor.Material]
						d.screen.SetContent(
							int(x), int(y),
							d.floors[floor].Rune(0, 0, x, y, t), nil,
							tcell.StyleDefault.Foreground(tcell.NewHexColor(int32(mat.Solid.Color.Value()))),
						)
					}
				}
			} else {
				mat := game.Materials[t.Block.Material]
				d.screen.SetContent(
					int(x), int(y),
					d.blocks[block].Rune(0, 0, x, y, t), nil,
					tcell.StyleDefault.Foreground(tcell.NewHexColor(int32(mat.Solid.Color.Value()))),
				)
			}
		}
		d.screen.SetContent(game.Player.X, game.Player.Y, d.player.Rune(0, 0, 0, 0, nil), nil, playerStyle)
		currTile := currChunk.Get(game.Player.X, game.Player.Y, game.Player.Z)
		d.drawString(0, chunk.Length, fmt.Sprintf("Z: %2d | Random: 0x%08X", game.Player.Z, currTile.Random), tcell.StyleDefault)
		d.clearLine(chunk.Length + 1)
		d.drawString(0, chunk.Length+1, fmt.Sprintf("Tile: %s", currTile.Describe(game.Blocks, game.Floors, game.Materials)), tcell.StyleDefault)

		d.screen.Show()
	}
}

func (d *Driver) drawMenu(app *game.Application, menu game.Menu) {
	d.drawStringCentered(0, menu.GetTitle(), titleStyle)

	opts := menu.GetOptions()
	maxlen := 0
	for _, o := range opts {
		if len(o) > maxlen {
			maxlen = len(o)
		}
	}
	opt_x := (d.width - (maxlen + len(cursor))) / 2

	selected := menu.GetOption()
	for i, option := range menu.GetOptions() {
		if i == selected {
			d.drawString(opt_x, 2+i, cursor+option, selectedStyle)
		} else {
			d.drawString(opt_x, 2+i, eraseCursor+option, optionStyle)
		}
	}

	d.screen.Show()
}

func (d *Driver) drawString(x, y int, str string, style tcell.Style) {
	if y >= d.height {
		return
	}

	// Figure out _where_ in the string to start
	start := 0
	if x < 0 {
		start = -x
	}
	// If that's past the end, nothing to draw
	if start > len(str) {
		return
	}

	// Calculate how much of the string to draw; if x is negative, this will
	// 'draw' more than the screen width by skipping the start bits.
	n := d.width - x
	if len(str) < n {
		n = len(str)
	}

	for i, r := range str[start:n] {
		d.screen.SetContent(i+x, y, r, nil, style)
	}
}

func (d *Driver) drawStringCentered(y int, str string, style tcell.Style) {
	d.drawString((d.width-len(str))/2, y, str, style)
}

func (d *Driver) clearLine(y int) {
	if y >= d.height {
		return
	}
	for x := 0; x < d.width; x++ {
		d.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault)
	}
}

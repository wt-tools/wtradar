package ui

import (
	"context"
	"image/color"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/grafov/kiwi"
	"github.com/wt-tools/wtscope/input/gamechat"
)

type gameChat struct {
	w          *app.Window
	th         *material.Theme
	cfg        configurator
	log        *kiwi.Logger
	list       widget.List
	rows       []gamechat.Message
	tropes     map[string]int
	latestTime time.Duration
}

func newGameChat(cfg configurator, log *kiwi.Logger) *gameChat {
	var w app.Window
	w.Option(app.Title("WT Scope: Game Chat"))
	return &gameChat{
		w:      &w,
		th:     material.NewTheme(),
		tropes: make(map[string]int),
		cfg:    cfg,
		log:    log,
	}
}

func (g *gui) UpdateGameChat(ctx context.Context, gamechat *gamechat.Service) {
	l := g.log.New()
	go func() {
		for {
			select {
			case data := <-gamechat.Messages:
				if len(g.gc.rows) > 0 && g.gc.latestTime > data.At {
					// Reset chat on a new battle session.
					l.Log("new battle new talks")
					g.gc.rows = nil
				}
				g.gc.latestTime = data.At
				g.gc.rows = append(g.gc.rows, data)
				g.gc.w.Invalidate()
				l.Log("game chat", data)
			}
		}
	}()
}

func (b *gameChat) panel() error {
	var ops op.Ops
	b.list.Axis = layout.Vertical
	b.list.ScrollToEnd = true
	for {
		switch e := b.w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			paint.FillShape(gtx.Ops, bgColor, clip.Rect{Max: gtx.Constraints.Max}.Op())
			layout.Flex{
				Alignment: layout.Start,
				Axis:      layout.Vertical,
				Spacing:   layout.SpaceEvenly,
			}.Layout(gtx,
				layout.Rigid(b.header("Game chat")),
				layout.Rigid(b.chatLayout),
			)
			e.Frame(gtx.Ops)
		}
	}
}

func (b *gameChat) header(title string) func(C) D {
	return func(gtx C) D {
		return layout.UniformInset(10).Layout(gtx,
			material.Label(b.th, unit.Sp(20), title).Layout,
		)
	}
}

func (b *gameChat) chatLayout(gtx C) D {
	return material.List(b.th, &b.list).Layout(gtx, len(b.rows), func(gtx layout.Context, i int) layout.Dimensions {
		var (
			text string
			act  chatRow
		)
		switch {
		case len(b.rows) == 0:
			text = "no battle log yet"
			return material.Label(b.th, unit.Sp(16), text).Layout(gtx)
		case i > len(b.rows): // TODO broken case, handle this in other way
			// text = fmtAction(b.rows[len(b.rows)-1])
			act = chatRow(b.rows[len(b.rows)-1])
		default:
			// text = fmtAction(b.rows[i])
			act = chatRow(b.rows[i])
		}
		return act.rowDisplay(gtx, i, b.cfg.PlayerName(), b.th)
	})
}

type chatRow gamechat.Message

func (r chatRow) rowDisplay(gtx C, num int, playerName string, th *material.Theme) D {
	const offset = 3
	border := widget.Border{
		Color: thinLineColor,
		Width: unit.Dp(1),
	}
	rowTableColor := evenRowColor
	if num%2 == 0 {
		rowTableColor = oddRowColor
	}
	return border.Layout(gtx, func(gtx C) D {
		return layout.UniformInset(1).Layout(gtx,
			func(gtx C) D {
				paint.FillShape(gtx.Ops, rowTableColor, clip.Rect{Max: gtx.Constraints.Max}.Op())
				return layout.Flex{
					Alignment: layout.Start,
					Axis:      layout.Horizontal,
					Spacing:   layout.SpaceEvenly,
				}.Layout(gtx,
					// Timestamp
					layout.Rigid(
						func(gtx C) D {
							return layout.UniformInset(offset).Layout(gtx,
								material.Label(th, unit.Sp(12), r.At.String()).Layout,
							)
						},
					),
					layout.Flexed(0.1,
						func(gtx C) D {
							return layout.UniformInset(offset).Layout(gtx, material.Label(th, unit.Sp(14), strings.ToUpper(r.Mode)).Layout)
						},
					),
					layout.Flexed(0.15,
						func(gtx C) D {
							playerInfo := material.Label(th, unit.Sp(14), r.Sender)
							playerInfo.Color = color.NRGBA{0, 0, 0, 255} // black
							if r.Sender == playerName {
								playerInfo.Color = color.NRGBA{0, 255, 0, 255} // green
							}
							return layout.UniformInset(offset).Layout(gtx, playerInfo.Layout)
						},
					),
					layout.Flexed(0.75,
						func(gtx C) D {
							return layout.UniformInset(offset).Layout(gtx, material.Label(th, unit.Sp(14), r.Msg).Layout)
						},
					),
				)
			},
		)
	})
}

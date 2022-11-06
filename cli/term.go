package cli

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/enescakir/emoji"
)

func Start() {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.Cursor = true
	g.InputEsc = true

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlL, gocui.ModNone, quit); err != nil {
		panic(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func layout(g *gocui.Gui) error {

	maxX, maxY := g.Size()
	if v, err := g.SetView("cmd", 1, maxY-3, maxX, maxY, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = true
		v.Editor = gocui.EditorFunc(editor)
		v.Clear()
		flag, _ := emoji.CountryFlag("ch")
		prompt := fmt.Sprintf("mp %s > ", flag)
		v.Write([]byte(prompt))
		v.MoveCursor(len(prompt)+1, 0)
	}
	v, err := g.SetView("out", 0, 0, maxX-1, maxY-3, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Frame = true
		v.FrameColor = gocui.ColorGreen
		// v.Editor = gocui.EditorFunc(editor)
		// v.Editable = true
	}

	curView := "cmd"
	if _, err := g.SetCurrentView(curView); err != nil {
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

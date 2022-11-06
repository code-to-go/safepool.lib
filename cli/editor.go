package cli

import (
	"strings"

	"github.com/awesome-gocui/gocui"
)

func editor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	if ch != 0 && mod == 0 {
		v.EditWrite(ch)
	}

	switch key {
	// Space, backspace, Del
	case gocui.KeySpace:
		v.EditWrite(' ')
	case gocui.KeyBackspace, gocui.KeyBackspace2:
		v.EditDelete(true)
		moveAhead(v)
	case gocui.KeyDelete:
		v.EditDelete(false)

	// Cursor movement
	case gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0)
		moveAhead(v)
	case gocui.KeyArrowRight:
		x, _ := v.Cursor()
		x2, _ := v.Origin()
		x += x2
		buf := v.Buffer()
		// Position of cursor should be on space that gocui adds at the end if at end
		if buf != "" && len(strings.TrimRight(buf, "\r\n")) > x {
			v.MoveCursor(1, 0)
		}

	case gocui.KeyEnter:
		buf := v.Buffer()
		v.Clear()
		v.SetCursor(0, 0)

		if buf != "" {
			buf = buf[:len(buf)-1]
		}
		if strings.TrimSpace(buf) != "" {
			processLine(buf)
		}

		//		enterActions[state.Mode](buf)
	}
}

func processLine(but string) {

}

// func setText(v *gocui.View, text string) {
// 	v.Clear()
// 	v.Write([]byte(text))
// 	v.SetCursor(len(text), 0)
// }

// moveAhead displays the next 10 characters when moving backwards,
// in order to see where we're moving or what we're deleting.
func moveAhead(v *gocui.View) {
	cX, _ := v.Cursor()
	oX, _ := v.Origin()
	if cX < 10 && oX > 0 {
		newOX := oX - 10
		forward := 10
		if newOX < 0 {
			forward += newOX
			newOX = 0
		}
		v.SetOrigin(newOX, 0)
		v.MoveCursor(forward, 0)
	}
}

// func moveDown(v *gocui.View) {
// 	_, yPos := v.Cursor()
// 	if _, err := v.Line(yPos + 1); err == nil {
// 		v.MoveCursor(0, 1, false)
// 	}
// }

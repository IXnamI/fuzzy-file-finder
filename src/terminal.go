package fff

import (
	"fmt"
	"fuzzy-file-finder/src/algos"
	"github.com/gdamore/tcell/v2"
)

type Terminal struct {
	Screen              tcell.Screen
	CachedResults       []algos.MatchResult
	CurrentLineSelected int
	ResultsStart        int
	ResultsEnd          int
	Width               int
	Height              int
	PointerStyle        tcell.Style
	QueryStyle          tcell.Style
	ResultsStyle        tcell.Style
	InfoStyle           tcell.Style
	DebugStyle          tcell.Style
	SelectedStyle       tcell.Style
	isResultsDisplayed  bool
	isCursorDisplayed   bool
}

const queryLinePosition = 0
const infoLinePosition = 1
const resultLineBuffer = 2
const defaultHorizontalStart = 1

type Direction bool

const (
	DirectionUp   Direction = true
	DirectionDown Direction = false
)

func NewTerminal() *Terminal {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	term := &Terminal{
		Screen:              screen,
		CachedResults:       nil,
		CurrentLineSelected: 2,
		ResultsStart:        0,
		ResultsEnd:          0,
		PointerStyle:        tcell.StyleDefault.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.ColorRed),
		QueryStyle:          tcell.StyleDefault.Background(tcell.NewRGBColor(122, 162, 247)).Foreground(tcell.ColorBlack),
		ResultsStyle:        tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.NewRGBColor(115, 218, 202)),
		InfoStyle:           tcell.StyleDefault.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.NewRGBColor(158, 206, 106)),
		DebugStyle:          tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite),
		SelectedStyle:       tcell.StyleDefault.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.NewRGBColor(115, 218, 202)),
		isResultsDisplayed:  false,
		isCursorDisplayed:   false,
	}
	term.Width, term.Height = screen.Size()
	term.initTerminalStyling()
	return term
}

func (t *Terminal) Stop() {
	t.Screen.Fini()
}

func (t *Terminal) Resize() {
	t.Width, t.Height = t.Screen.Size()
}

func (t *Terminal) DrawQuery(query string) {
	t.clearRow(queryLinePosition, t.QueryStyle)
	t.drawText(defaultHorizontalStart, queryLinePosition, query, t.QueryStyle)
}

func (t *Terminal) DrawInfo(info string) {
	t.clearRow(infoLinePosition, t.InfoStyle)
	t.drawText(defaultHorizontalStart, infoLinePosition, info, t.InfoStyle)
}

func (t *Terminal) DrawDebug(debug string) {
	t.clearRow(infoLinePosition, t.DebugStyle)
	t.drawText(defaultHorizontalStart, infoLinePosition, debug, t.DebugStyle)
}

func (t *Terminal) AppendToResults(slice []algos.MatchResult) {
	t.CachedResults = t.CachedResults[:0]
	t.CachedResults = append(t.CachedResults, slice...)
	t.ResultsStart = 0
	t.ResultsEnd = min(len(slice), t.Height-resultLineBuffer)
}

func (t *Terminal) DrawResults() {
	t.ClearResults()
	y := resultLineBuffer
	currentResults := t.CachedResults[t.ResultsStart:t.ResultsEnd]
	for _, res := range currentResults {
		t.drawText(defaultHorizontalStart, y, fmt.Sprintf("[%d] %s", res.Score, res.Candidate), t.ResultsStyle)
		t.drawText(0, y, " ", t.InfoStyle)
		y++
	}
	t.drawPointer(t.CurrentLineSelected)
	t.isResultsDisplayed = true
}

func (t *Terminal) DrawSelected(prevSelected int) {
	textPrev := t.CachedResults[t.ResultsStart+prevSelected-resultLineBuffer]
	text := t.CachedResults[t.ResultsStart+t.CurrentLineSelected-resultLineBuffer]
	t.drawText(defaultHorizontalStart, prevSelected, fmt.Sprintf("[%d] %s", textPrev.Score, textPrev.Candidate), t.ResultsStyle)
	t.drawText(defaultHorizontalStart, t.CurrentLineSelected, fmt.Sprintf("[%d] %s", text.Score, text.Candidate), t.SelectedStyle)
}

func (t *Terminal) MoveResults(direction Direction) {
	if direction == DirectionDown {
		t.ResultsStart++
		t.ResultsEnd++
	} else {
		t.ResultsStart--
		t.ResultsEnd--
	}
	t.DrawResults()
}

func (t *Terminal) MovePointer(direction Direction) {
	if !t.isResultsDisplayed {
		return
	}
	if direction == DirectionDown && t.ResultsEnd < len(t.CachedResults) && t.CurrentLineSelected == t.Height-resultLineBuffer+1 {
		t.MoveResults(DirectionDown)
		return
	} else if direction == DirectionUp && t.ResultsStart > 0 && t.CurrentLineSelected == resultLineBuffer {
		t.MoveResults(DirectionUp)
		return
	}
	if direction == DirectionDown && (t.CurrentLineSelected <= t.Height-resultLineBuffer && t.CurrentLineSelected < t.ResultsEnd+resultLineBuffer-1) {
		t.CurrentLineSelected++
		t.drawPointer(t.CurrentLineSelected - 1)
	} else if direction == DirectionUp && t.CurrentLineSelected > resultLineBuffer {
		t.CurrentLineSelected--
		t.drawPointer(t.CurrentLineSelected + 1)
	}
}

func (t *Terminal) drawPointer(prevLine int) {
	t.drawText(0, prevLine, " ", t.PointerStyle)
	t.drawText(0, t.CurrentLineSelected, fmt.Sprintf(">"), t.PointerStyle)
	t.DrawSelected(prevLine)
	t.isCursorDisplayed = true
}

func (t *Terminal) ClearResults() {
	for i := range t.Height {
		if i == 0 || i == 1 {
			continue
		}
		t.clearRow(i, t.ResultsStyle)
	}
	t.isResultsDisplayed = false
}

func (t *Terminal) ClearScreen() {
	t.Screen.Clear()
}
func (t *Terminal) clearRow(row int, style tcell.Style) {
	for x := range t.Width {
		t.Screen.SetContent(x, row, ' ', nil, style)
	}
}

func (t *Terminal) drawText(xCell int, yCell int, text string, style tcell.Style) {
	for i, ch := range text {
		t.Screen.SetContent(xCell+i, yCell, ch, nil, style)
	}
}

func (t *Terminal) initTerminalStyling() {
	t.setRowStyle(queryLinePosition, t.QueryStyle)
	t.setRowStyle(infoLinePosition, t.InfoStyle)
	for i := range t.Height {
		if i == 0 || i == 1 {
			continue
		}
		t.setRowStyle(i, t.ResultsStyle)
	}
}

func (t *Terminal) setRowStyle(y int, style tcell.Style) {
	for x := range t.Width {
		t.Screen.SetContent(x, y, ' ', nil, style)
	}
}

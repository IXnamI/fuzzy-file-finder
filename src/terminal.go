package fff

import (
	"fmt"
	"fuzzy-file-finder/src/algos"
	"log/slog"

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
	IsResultsDisplayed  bool
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
		CurrentLineSelected: 0,
		ResultsStart:        0,
		ResultsEnd:          0,
		PointerStyle:        tcell.StyleDefault.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.ColorRed),
		QueryStyle:          tcell.StyleDefault.Background(tcell.NewRGBColor(122, 162, 247)).Foreground(tcell.ColorBlack),
		ResultsStyle:        tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.NewRGBColor(115, 218, 202)),
		InfoStyle:           tcell.StyleDefault.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.NewRGBColor(158, 206, 106)),
		DebugStyle:          tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite),
		SelectedStyle:       tcell.StyleDefault.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.NewRGBColor(115, 218, 202)),
		IsResultsDisplayed:  false,
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
	t.clearRow(t.Height-1, t.DebugStyle)
	t.drawText(defaultHorizontalStart, t.Height-1, debug, t.DebugStyle)
}

func (t *Terminal) AppendToResults(slice []algos.MatchResult) {
	slog.Info("AppendToResults() called")
	t.CachedResults = slice
	t.ResultsStart = 0
	t.ResultsEnd = min(len(slice), t.Height-resultLineBuffer)
	if t.CurrentLineSelected > t.ResultsEnd-1 {
		t.CurrentLineSelected = max(t.ResultsEnd-1, 0)
		slog.Info("Modyfying curr line selected", "t.CurrentLineSelected", t.CurrentLineSelected, "End of slice", t.ResultsEnd-1)
		slog.Info("Slice info", "len(slice)", len(slice), "slice content", t.CachedResults)
	}
}

func (t *Terminal) DrawResults() {
	t.ClearResults()
	if len(t.CachedResults) == 0 {
		return
	}
	y := resultLineBuffer
	currentResults := t.CachedResults[t.ResultsStart:t.ResultsEnd]
	for _, res := range currentResults {
		t.drawText(defaultHorizontalStart, y, fmt.Sprintf("[%d] %s", res.Score, res.Candidate), t.ResultsStyle)
		t.drawText(0, y, " ", t.InfoStyle)
		y++
	}
	t.IsResultsDisplayed = true
	t.drawPointer(t.CurrentLineSelected)
}

func (t *Terminal) DrawSelected(prevSelected int) {
	slog.Info("Attempting to grab prev result", "line", prevSelected)
	textPrev := t.GetResultAt(prevSelected)
	slog.Info("Attempting to grab result", "line", t.CurrentLineSelected)
	text := t.GetResultAt(t.CurrentLineSelected)
	t.drawText(defaultHorizontalStart, prevSelected+resultLineBuffer, fmt.Sprintf("[%d] %s", textPrev.Score, textPrev.Candidate), t.ResultsStyle)
	t.drawText(defaultHorizontalStart, t.CurrentLineSelected+resultLineBuffer, fmt.Sprintf("[%d] %s", text.Score, text.Candidate), t.SelectedStyle)
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
	if !t.IsResultsDisplayed {
		return
	}
	if direction == DirectionDown && t.ResultsEnd < len(t.CachedResults) && t.CurrentLineSelected == t.Height-resultLineBuffer-1 {
		slog.Info("Moving results down")
		t.MoveResults(DirectionDown)
		return
	} else if direction == DirectionUp && t.ResultsStart > 0 && t.CurrentLineSelected == 0 {
		slog.Info("Moving results up")
		t.MoveResults(DirectionUp)
		return
	}
	if direction == DirectionDown && (t.CurrentLineSelected < t.Height-resultLineBuffer-1 && t.CurrentLineSelected < t.ResultsEnd-1) {
		slog.Info("Moving pointer down")
		t.CurrentLineSelected++
		t.drawPointer(t.CurrentLineSelected - 1)
	} else if direction == DirectionUp && t.CurrentLineSelected > 0 {
		slog.Info("Moving pointer up")
		t.CurrentLineSelected--
		t.drawPointer(t.CurrentLineSelected + 1)
	}
}

func (t *Terminal) drawPointer(prevLine int) {
	slog.Info("Attempting to draw pointer", "line", t.CurrentLineSelected)
	t.drawText(0, prevLine+resultLineBuffer, " ", t.PointerStyle)
	t.drawText(0, t.CurrentLineSelected+resultLineBuffer, fmt.Sprintf(">"), t.PointerStyle)
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
	t.IsResultsDisplayed = false
}

func (t *Terminal) ClearScreen() {
	t.Screen.Clear()
}

func (t *Terminal) GetCurrentSelected() *algos.MatchResult {
	return t.GetResultAt(t.CurrentLineSelected)
}

func (t *Terminal) GetResultAt(index int) *algos.MatchResult {
	if !t.IsResultsDisplayed {
		return nil
	}
	if t.ResultsEnd <= t.ResultsStart+index {
		slog.Error("Attempting to grab impossible index", "t.CurrentLineSelected", t.CurrentLineSelected, "results", t.CachedResults)
	}
	return &t.CachedResults[t.ResultsStart+index]
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

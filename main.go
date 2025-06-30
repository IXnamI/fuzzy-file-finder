package main

import (
	"context"
	"fmt"
	fff "fuzzy-file-finder/src"
	"fuzzy-file-finder/src/algos"
	"fuzzy-file-finder/src/fileTree"
	"github.com/eiannone/keyboard"
	"github.com/gdamore/tcell/v2"
	"sync"
	"time"
	"unicode/utf8"
)

var wg sync.WaitGroup
var screen tcell.Screen
var queryStyle tcell.Style
var resultsStyle tcell.Style

func main() {
	//Keyboard setup
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	//Screen setup
	screen, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	w, h := screen.Size()
	defer screen.Fini()
	clearScreen()
	queryStyle = queryStyle.Background(tcell.NewRGBColor(122, 162, 247)).Foreground(tcell.ColorBlack)
	resultsStyle = resultsStyle.Background(tcell.NewRGBColor(59, 66, 97)).Foreground(tcell.NewRGBColor(115, 218, 202))
	initTerminalStyling(w, h)

	//Matcher and scorer setup
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan string, 400000)
	lastQueryLen := 0
	query := ""
	fs := fileTree.CreateDirTreeStruct()
	root := "C:\\"
	ticker := time.NewTicker(300 * time.Millisecond)

	fileTree.CreateAsyncJob(root, jobs, fs, 10)
	drawText(1, 0, fmt.Sprintf("Query: %s", query), queryStyle)
	screen.Show()

	// Poll for updates from results
	go func() {
		for {
			select {
			case <-ticker.C:
				if len(query) == 0 {
					clearResults(w)
					screen.Show()
					continue
				}
				if len(query) == lastQueryLen {
					continue
				}
				cancel()
				resultsChan := make(chan algos.MatchResult, 10000)
				ctx, cancel = context.WithCancel(context.Background())
				algos.CreateNewWorker(ctx, fs, resultsChan, 4, query, &wg)
				go func() {
					defer close(resultsChan)
					wg.Wait()
				}()
				scoreResults := fff.NewResultArray(20)
				for res := range resultsChan {
					scoreResults.Add(res)
				}
				y := 1
				clearResults(w)
				for _, res := range scoreResults.Holder {
					drawText(1, y, fmt.Sprintf("[%d] %s", res.Score, res.Candidate), resultsStyle)
					y++
				}
				lastQueryLen = len(query)
				screen.Show()
			}
		}
	}()

	// Reading user inputs key-by-key
	for {
		ev := screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, _ = screen.Size()
			initTerminalStyling(w, h)
			screen.Show()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyESC || ev.Key() == tcell.KeyCtrlC {
				ticker.Stop()
				return
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if len(query) > 0 {
					_, size := utf8.DecodeLastRuneInString(query)
					query = query[:len(query)-size]
					lastQueryLen = len(query) + 1
				}
			} else if ev.Key() == tcell.KeyRune {
				query += string(ev.Rune())
				lastQueryLen = len(query) - 1
			}
		}
		clearRow(0, w, queryStyle)
		drawText(1, 0, fmt.Sprintf("Query: %s", query), queryStyle)
		screen.Show()
	}
}

func clearScreen() {
	screen.Clear()
}

func drawText(x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		screen.SetContent(x+i, y, ch, nil, style)
	}
}

func clearRow(y int, width int, style tcell.Style) {
	for x := range width {
		screen.SetContent(x, y, ' ', nil, style)
	}
}

func setRowStyle(y int, width int, style tcell.Style) {
	for x := range width {
		screen.SetContent(x, y, ' ', nil, style)
	}
}

func initTerminalStyling(w int, h int) {
	setRowStyle(0, w, queryStyle)
	for i := range h {
		if i == 0 {
			continue
		}
		setRowStyle(i, w, resultsStyle)
	}
}

func clearResults(width int) {
	for i := range 21 {
		if i == 0 {
			continue
		}
		clearRow(i, width, resultsStyle)
	}
}

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
	w, _ := screen.Size()
	defer screen.Fini()
	clearScreen()
	queryStyle = queryStyle.Background(tcell.Color111).Foreground(tcell.ColorBlack)
	initTerminalStyling(w)

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
				if len(query) == 0 || len(query) == lastQueryLen {
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
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		if key == keyboard.KeyEsc {
			ticker.Stop()
			break
		}
		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			if len(query) > 0 {
				_, size := utf8.DecodeLastRuneInString(query)
				query = query[:len(query)-size]
				lastQueryLen = len(query) + 1
			}
		} else if char >= 32 && char <= 126 {
			query += string(char)
			lastQueryLen = len(query) - 1
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

func initTerminalStyling(w int) {
	setRowStyle(0, w, queryStyle)
}

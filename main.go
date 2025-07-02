package main

import (
	"context"
	"fmt"
	fff "fuzzy-file-finder/src"
	"fuzzy-file-finder/src/algos"
	"fuzzy-file-finder/src/fileTree"
	"github.com/gdamore/tcell/v2"
	"sync"
	"time"
	"unicode/utf8"
)

var wg sync.WaitGroup

func main() {
	term := fff.NewTerminal()
	defer term.Stop()
	term.ClearScreen()

	//Matcher and scorer setup
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan string, 400000)
	lastQueryLen := 0
	query := ""
	fs := fileTree.CreateDirTreeStruct()
	root := "C:\\"
	ticker := time.NewTicker(300 * time.Millisecond)

	fileTree.CreateAsyncJob(root, jobs, fs, 10)
	term.DrawQuery(fmt.Sprintf("Query: %s", query))
	term.Screen.Show()

	// Poll for updates from results
	go func() {
		for {
			select {
			case <-ticker.C:
				term.DrawInfo(fmt.Sprintf("Indexed %d files", len(fs.GetSnapShot())))
				if len(query) == 0 {
					term.ClearResults()
					term.Screen.Show()
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
				scoreResults := fff.NewResultArray(100)
				for res := range resultsChan {
					scoreResults.Add(res)
				}
				term.AppendToResults(scoreResults.Holder)
				term.DrawResults()
				lastQueryLen = len(query)
				term.Screen.Show()
			}
		}
	}()

	term.Screen.Sync()
	// Reading user inputs key-by-key
	for {
		ev := term.Screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			term.Resize()
			term.Screen.Show()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyESC, tcell.KeyCtrlC:
				term.DrawDebug(fmt.Sprintf("Escape sequence pressed"))
				term.Screen.Show()
				ticker.Stop()
				return
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				if len(query) > 0 {
					_, size := utf8.DecodeLastRuneInString(query)
					query = query[:len(query)-size]
					lastQueryLen = len(query) + 1
				}
			case tcell.KeyRune:
				query += string(ev.Rune())
				lastQueryLen = len(query) - 1
			case tcell.KeyUp:
				term.MovePointer(fff.DirectionUp)
			case tcell.KeyDown:
				term.MovePointer(fff.DirectionDown)
			}
		}
		term.DrawQuery(fmt.Sprintf("Query: %s", query))
		term.Screen.Show()
	}
}

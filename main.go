package main

import (
	"context"
	"fmt"
	fff "fuzzy-file-finder/src"
	"fuzzy-file-finder/src/algos"
	"fuzzy-file-finder/src/fileTree"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

var wg sync.WaitGroup

func main() {
	f, err := os.Create("output.log")
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	handler := slog.NewTextHandler(f, nil)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	term := fff.NewTerminal()
	defer term.Stop()
	term.ClearScreen()

	//inputStream := fff.NewEventStream(100)

	//Matcher and scorer setup
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan string, 800000)
	lastQuery := ""
	query := ""
	fs := fileTree.CreateDirTreeStruct()
	root := getDrives()
	ticker := time.NewTicker(200 * time.Millisecond)
	prevResultsLength := 0

	fileTree.CreateAsyncJob(root, jobs, fs, 10)

	term.DrawQuery(fmt.Sprintf("Query: %s", query))
	term.Screen.Show()

	// Poll for updates from results
	go func() {
		for {
			select {
			case <-ticker.C:
				if term.IsResultsDisplayed {
					term.DrawInfo(fmt.Sprintf("Showing %d/%d", prevResultsLength, len(fs.GetSnapShot())))
					term.Screen.Show()
				}
				if !term.IsResultsDisplayed {
					term.DrawInfo(fmt.Sprintf("Indexed %d files", len(fs.GetSnapShot())))
					term.Screen.Show()
				}
				if len(query) == 0 {
					term.ClearResults()
					term.Screen.Show()
					continue
				}
				if strings.Compare(query, lastQuery) == 0 {
					continue
				}
				lastQuery = query
				cancel()
				resultsChan := make(chan algos.MatchResult, 10000)
				ctx, cancel = context.WithCancel(context.Background())
				algos.CreateNewWorker(ctx, fs, resultsChan, 4, query, &wg)
				go func() {
					defer close(resultsChan)
					wg.Wait()
				}()
				scoreResults := fff.NewResultArray(1000)
				for res := range resultsChan {
					scoreResults.Add(res)
				}
				term.AppendToResults(scoreResults.Holder)
				term.DrawResults()
				term.DrawInfo(fmt.Sprintf("Showing %d/%d", len(scoreResults.Holder), len(fs.GetSnapShot())))
				prevResultsLength = len(scoreResults.Holder)
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
		case *tcell.EventMouse:
			btn := ev.Buttons()
			if btn&tcell.Button1 != 0 {
				slog.Info("Left click was received")
				_, y := ev.Position()
				term.SetSelected(y)
				term.Screen.Show()
			}
			if btn&tcell.WheelUp != 0 {
				slog.Info("Scroll up was received")
				term.MovePointer(fff.DirectionUp)
				term.Screen.Show()
			}
			if btn&tcell.WheelDown != 0 {
				slog.Info("Scroll down was received")
				term.MovePointer(fff.DirectionDown)
				term.Screen.Show()
			}
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyESC, tcell.KeyCtrlC:
				term.Screen.Show()
				ticker.Stop()
				return
			case tcell.KeyEnter:
				result := term.GetCurrentSelected()
				if result == nil {
					continue
				}
				err := clipboard.WriteAll(result.Candidate)
				if err != nil {
					slog.Error("Error copying to clipboard", "Error: ", err)
				}
				err = openInExplorer(result.Candidate)
				if err != nil {
					slog.Error("Error opening file in Explorer", "Error: ", err)
				}
			case tcell.KeyBackspace, tcell.KeyBackspace2:
				lastQuery = query
				if len(query) > 0 {
					_, size := utf8.DecodeLastRuneInString(query)
					query = query[:len(query)-size]
				}
			case tcell.KeyRune:
				lastQuery = query
				query += string(ev.Rune())
			case tcell.KeyUp:
				term.MovePointer(fff.DirectionUp)
			case tcell.KeyDown:
				term.MovePointer(fff.DirectionDown)
			}
		default:
			slog.Info("Something was received", "ev", ev)
		}
		term.DrawQuery(fmt.Sprintf("Query: %s", query))
		term.Screen.Show()
	}
}

func getDrives() []string {
	kernel32, _ := syscall.LoadLibrary("kernel32.dll")
	getLogicalDrivesHandle, _ := syscall.GetProcAddress(kernel32, "GetLogicalDrives")

	var drives []string

	if ret, _, callErr := syscall.SyscallN(uintptr(getLogicalDrivesHandle), 0, 0, 0, 0); callErr != 0 {
		slog.Error("Fail to get drives")
	} else {
		drives = bitsToDrives(uint32(ret))
	}
	return drives
}

func bitsToDrives(bitMap uint32) (drives []string) {
	availableDrives := []string{"A:\\", "B:\\", "C:\\", "D:\\", "E:\\", "F:\\", "G:\\", "H:\\", "I:\\", "J:\\", "K:\\", "L:\\", "M:\\", "N:\\", "O:\\", "P:\\", "Q:\\", "R:\\", "S:\\", "T:\\", "U:\\", "V:\\", "W:\\", "X:\\", "Y:\\", "Z:\\"}

	for i := range availableDrives {
		if bitMap&1 == 1 {
			drives = append(drives, availableDrives[i])
		}
		bitMap >>= 1
	}

	return
}

func openInExplorer(path string) error {
	cmd := exec.Command("explorer", "/select,"+path)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

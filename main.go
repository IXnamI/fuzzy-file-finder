package main

import (
	"bufio"
	"fmt"
	fff "fuzzy-file-finder/src"
	"fuzzy-file-finder/src/algos"
	"fuzzy-file-finder/src/fileTree"
	"os"
	"sync"
	"unicode/utf8"
)

var wg sync.WaitGroup

func main() {
	jobs := make(chan string, 400000)
	resultsChan := make(chan algos.MatchResult, 10000)
	query := ""
	fs := fileTree.CreateDirTreeStruct()
	root := "C:\\"
	reader := bufio.NewReader(os.Stdin)

	fileTree.CreateAsyncJob(root, jobs, fs, 10)

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			break
		}

		if char == 127 && len(query) > 0 {
			_, size := utf8.DecodeLastRuneInString(query)
			query = query[:len(query)-size]
		} else if char == 10 {
			clearScreen()
			fmt.Printf("Query: %s\n", query)
			algos.CreateNewWorker(fs, resultsChan, 4, query, &wg)
			go func() {
				defer close(resultsChan)
				wg.Wait()
			}()
			scoreResults := fff.NewResultArray(20)
			for res := range resultsChan {
				scoreResults.Add(res)
			}
			for _, result := range scoreResults.Holder {
				fmt.Printf("[%d] %s \n", result.Score, result.Candidate)
			}
			break
		} else if char >= 32 && char <= 126 {
			query += string(char)
		}
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

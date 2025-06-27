package main

import (
	"context"
	"fmt"
	fff "fuzzy-file-finder/src"
	"fuzzy-file-finder/src/algos"
	"fuzzy-file-finder/src/fileTree"
	"sync"
	"unicode/utf8"

	"github.com/eiannone/keyboard"
)

var wg sync.WaitGroup

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan string, 400000)
	query := ""
	fs := fileTree.CreateDirTreeStruct()
	root := "C:\\"

	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	fileTree.CreateAsyncJob(root, jobs, fs, 10)

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		if key == keyboard.KeyEsc {
			break
		}
		if key == keyboard.KeyBackspace || key == keyboard.KeyBackspace2 {
			if len(query) > 0 {
				_, size := utf8.DecodeLastRuneInString(query)
				query = query[:len(query)-size]
			}
		} else if char >= 32 && char <= 126 {
			query += string(char)
			cancel()
			resultsChan := make(chan algos.MatchResult, 10000)
			clearScreen()
			fmt.Printf("Query: %s\n", query)
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
			for _, result := range scoreResults.Holder {
				fmt.Printf("[%d] %s \n", result.Score, result.Candidate)
			}
		}
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

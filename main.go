package main

import (
	"fmt"
	fff "fuzzy-file-finder/src"
	"fuzzy-file-finder/src/algos"
	"fuzzy-file-finder/src/fileTree"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	start := time.Now()
	jobs := make(chan string, 400000)
	resultsChan := make(chan algos.MatchResult, 10000)
	fs := fileTree.CreateDirTreeStruct()
	root := "E:\\"
	query := "Appdata\\Local"

	fileTree.CreateAsyncJob(root, jobs, fs, 10)
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
	elapsed := time.Since(start)
	fmt.Printf("Time taken: %s\n", elapsed)
}

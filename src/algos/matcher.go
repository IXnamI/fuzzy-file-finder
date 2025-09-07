package algos

import (
	"context"
	"fuzzy-file-finder/src/fileTree"
	"strings"
	"sync"
)

func CreateNewWorker(ctx context.Context, fs *fileTree.DirTreeHolder, outputChannel chan MatchResult, numWorkers int, query string, wg *sync.WaitGroup) {
	fileTreeResults := make(chan string, 100000)
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileTreeResults {
				select {
				case <-ctx.Done():
					return
				default:
					if match(query, path) {
						outputChannel <- Score(query, path)
					}
				}
			}
		}()
	}
	go func() {
		for _, path := range fs.GetSnapShot() {
			fileTreeResults <- path
		}
		close(fileTreeResults)
	}()
}

func match(query, candidate string) bool {
	q := []rune(strings.ToLower(query))
	c := []rune(strings.ToLower(candidate))

	i := 0
	for _, ch := range c {
		if i < len(q) && q[i] == ch {
			i++
		}
	}
	return i == len(q)
}

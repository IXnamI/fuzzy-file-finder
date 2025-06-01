package algos

import (
	"fuzzy-file-finder/src/fileTree"
	"strings"
	"sync"
)

func CreateNewWorker(fs *fileTree.DirTreeHolder, outputChannel chan MatchResult, numWorkers int, query string, wg *sync.WaitGroup) {
	fileTreeResults := make(chan string, 10000)
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileTreeResults {
				if match(query, path) {
					outputChannel <- Score(query, path)
				}
			}
		}()
	}
	for _, path := range fs.GetSnapShot() {
		fileTreeResults <- path
	}
	close(fileTreeResults)
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

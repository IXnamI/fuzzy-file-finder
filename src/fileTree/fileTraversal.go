package fileTree

import (
	"os"
	"path/filepath"
	"sync"
)

var internalWg sync.WaitGroup

func CreateAsyncJob(root []string, jobs chan string, fs *DirTreeHolder, numWorkers int) {
	output := make(chan string, 10000)
	for _, path := range root {
		internalWg.Add(1)
		jobs <- path
	}
	for range numWorkers {
		go func() {
			for path := range jobs {
				entries, err := os.ReadDir(path)
				if err != nil {
					internalWg.Done()
					continue
				}

				for _, entry := range entries {
					fullPath := filepath.Join(path, entry.Name())
					if entry.IsDir() {
						internalWg.Add(1)
						jobs <- fullPath
					} else {
						output <- fullPath
					}
				}
				internalWg.Done()
			}
		}()
	}
	go func() {
		defer close(jobs)
		defer close(output)
		internalWg.Wait()
	}()
	go func() {
		for path := range output {
			fs.Add(path)
		}
	}()
}

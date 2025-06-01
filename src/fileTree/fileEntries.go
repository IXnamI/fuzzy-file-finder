package fileTree

import (
	"sync"
)

type DirTreeHolder struct {
	DirTreeArray []string
	Mu           *sync.Mutex
}

func CreateDirTreeStruct() *DirTreeHolder {
	return &DirTreeHolder{DirTreeArray: make([]string, 0, 100000), Mu: &sync.Mutex{}}
}

func (dt *DirTreeHolder) Add(elem string) {
	defer dt.Mu.Unlock()
	dt.Mu.Lock()
	dt.DirTreeArray = append(dt.DirTreeArray, elem)
	return
}

func (dt *DirTreeHolder) GetSnapShot() []string {
	defer dt.Mu.Unlock()
	dt.Mu.Lock()
	snapshot := make([]string, len(dt.DirTreeArray))
	copy(snapshot, dt.DirTreeArray)
	return snapshot
}

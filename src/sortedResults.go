package fff

import (
	"fuzzy-file-finder/src/algos"
	"math"
	"slices"
)

type ResultArray struct {
	Capacity int
	Holder   []algos.MatchResult
	MinScore int
}

func NewResultArray(capacity int) *ResultArray {
	return &ResultArray{Capacity: capacity, MinScore: math.MaxInt16}
}

func (ra *ResultArray) Add(elem algos.MatchResult) {
	sliceLength := len(ra.Holder)
	if sliceLength < ra.Capacity {
		add(ra, elem)
		return
	}
	if elem.Score <= ra.MinScore {
		return
	}
	ra.Holder = slices.Delete(ra.Holder, 0, 1)
	add(ra, elem)
}

func add(ra *ResultArray, elem algos.MatchResult) {
	insertIndex, _ := slices.BinarySearchFunc(ra.Holder, elem, func(a, b algos.MatchResult) int {
		if a.Score > b.Score {
			return -1
		}
		if a.Score < b.Score {
			return 1
		}
		return 0
	})
	ra.Holder = slices.Insert(ra.Holder, insertIndex, elem)
	ra.MinScore = ra.Holder[0].Score
}

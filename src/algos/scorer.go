package algos

import (
	"strings"
)

type MatchResult struct {
	Candidate  string
	Score      int
	MatchIndex []int
}

// Naive scoring, should still be relatively fast due to O(n)
func Score(query, candidate string) MatchResult {

	q := strings.ToLower(query)
	t := strings.ToLower(candidate)
	score := 0
	queryIndex := 0
	matchIndex := make([]int, 0, len(query))

	for i := range len(t) {
		if i <= len(t)-len(q) {
			slice := t[i:(i + len(q))]
			if strings.Compare(slice, q) == 0 {
				return MatchResult{Score: 20 * len(q), Candidate: candidate, MatchIndex: spread(i, i+len(q))}
			}
		}
		if queryIndex == len(q) {
			continue
		}
		if q[queryIndex] == t[i] {
			score += 10
			queryIndex++
			matchIndex = append(matchIndex, i)
		}
	}
	//slog.Info("returning matched result", "Candidate: ", candidate, "matchIndex", matchIndex)
	return MatchResult{Score: score, Candidate: candidate, MatchIndex: matchIndex}
}

func spread(start int, end int) []int {
	local := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		local = append(local, i)
	}
	return local
}

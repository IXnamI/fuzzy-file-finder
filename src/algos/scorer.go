package algos

import (
	"strings"
)

type MatchResult struct {
	Score     int
	Candidate string
}

// Naive scoring, should still be relatively fast due to O(n)
func Score(query, candidate string) MatchResult {

	q := strings.ToLower(query)
	t := strings.ToLower(candidate)

	score := 0
	queryIndex := 0
	for i := range len(t) {
		if i < len(t)-len(q) {
			slice := t[i:(i + len(q))]
			if strings.Compare(slice, q) == 0 {
				return MatchResult{Score: 20 * len(q), Candidate: candidate}
			}
		}
		if queryIndex == len(q) {
			continue
		}
		if q[queryIndex] == t[i] {
			score += 10
			queryIndex++
		}
	}
	return MatchResult{Score: score, Candidate: candidate}
}

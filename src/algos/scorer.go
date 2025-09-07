package algos

import (
	"strings"
	"unicode"
)

type MatchResult struct {
	Score      int
	Candidate  string
	MatchIndex []int
	FirstMatch int
}

func Score(query, candidate string) MatchResult {
	q := strings.ToLower(query)
	t := strings.ToLower(candidate)

	queryIndex := 0
	lastMatchIndex := -2
	matchIndex := make([]int, 0, len(q))
	score := 0
	firstMatch := -1

	if strings.Contains(t, q) {
		i := strings.Index(t, q)
		return MatchResult{
			Score:      100,
			Candidate:  candidate,
			MatchIndex: spread(i, i+len(q)),
			FirstMatch: i,
		}
	}

	for i := 0; i < len(t); i++ {
		if queryIndex >= len(q) {
			break
		}
		if t[i] == q[queryIndex] {
			matchIndex = append(matchIndex, i)

			if lastMatchIndex == i-1 {
				score += 15
			} else {
				score += 10
			}

			if i == 0 {
				score += 10
			}

			if i > 0 && !unicode.IsLetter(rune(t[i-1])) {
				score += 5
			}

			if lastMatchIndex != -2 && i-lastMatchIndex > 1 {
				score -= (i - lastMatchIndex) / 5
			}

			if firstMatch == -1 {
				firstMatch = i
			}

			lastMatchIndex = i
			queryIndex++
		}
	}

	if len(matchIndex) > 0 {
		chunkLen := 1
		maxChunk := 1
		for i := 1; i < len(matchIndex); i++ {
			if matchIndex[i] == matchIndex[i-1]+1 {
				chunkLen++
			} else {
				if chunkLen > maxChunk {
					maxChunk = chunkLen
				}
				chunkLen = 1
			}
		}
		if chunkLen > maxChunk {
			maxChunk = chunkLen
		}
		if maxChunk >= 4 {
			score += maxChunk * 4
		}
	}

	lengthPenalty := len(candidate) - len(query)
	if lengthPenalty > 0 {
		score -= lengthPenalty / 4
	}

	maxScore := 15*len(q) + 10 + 5
	if score < 0 {
		score = 0
	}
	if score > maxScore {
		score = maxScore
	}
	normalized := int(float64(score) / float64(maxScore) * 100)

	if normalized >= 100 {
		normalized = 99
	}

	return MatchResult{
		Score:      normalized,
		Candidate:  candidate,
		MatchIndex: matchIndex,
		FirstMatch: firstMatch,
	}
}

func spread(start int, end int) []int {
	local := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		local = append(local, i)
	}
	return local
}

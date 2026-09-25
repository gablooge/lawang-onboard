package tools

import (
	"math"
	"sort"
	"strings"

	"github.com/gablooge/lawang-onboard/internal/corpus"
	"github.com/gablooge/lawang-onboard/internal/filter"
	"github.com/gablooge/lawang-onboard/internal/roles"
)

// SearchResult is the response from the search tool.
type SearchResult struct {
	Items    []SearchHit    `json:"items"`
	Withheld map[string]int `json:"withheld"`
}

// SearchHit is one ranked result.
type SearchHit struct {
	ID    string  `json:"id"`
	Kind  string  `json:"kind"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

// Search returns corpus items visible to role that match the query, ranked by
// a simple BM25-style TF-IDF score. A hidden item never affects results or
// scores. At most maxResults results are returned; pass 0 for the default (20).
func Search(role roles.Role, items []corpus.Item, query string, maxResults int) SearchResult {
	if maxResults <= 0 {
		maxResults = 20
	}

	w := filter.Withheld{}
	terms := tokenize(query)
	if len(terms) == 0 {
		return SearchResult{Items: []SearchHit{}, Withheld: map[string]int{}}
	}

	// Collect visible items and compute per-item scores.
	type scored struct {
		item  corpus.Item
		score float64
	}

	var visible []corpus.Item
	for _, item := range items {
		if filter.Visible(role, item) {
			visible = append(visible, item)
		} else {
			w.Record(role, item)
		}
	}

	n := len(visible)
	if n == 0 {
		return SearchResult{Items: []SearchHit{}, Withheld: w.Summary()}
	}

	// BM25 parameters.
	const k1 = 1.2
	const b = 0.75

	// Compute average document length (in tokens).
	totalLen := 0
	for _, item := range visible {
		totalLen += len(tokenize(item.Title + " " + item.Text))
	}
	avgLen := float64(totalLen) / float64(n)
	if avgLen == 0 {
		avgLen = 1
	}

	// Compute document frequency for each query term over visible items only.
	df := make(map[string]int, len(terms))
	for _, item := range visible {
		docTerms := termSet(item.Title + " " + item.Text)
		for _, t := range terms {
			if docTerms[t] {
				df[t]++
			}
		}
	}

	var hits []scored
	for _, item := range visible {
		docTokens := tokenize(item.Title + " " + item.Text)
		docLen := float64(len(docTokens))

		// Term frequency map.
		tf := make(map[string]int, len(docTokens))
		for _, t := range docTokens {
			tf[t]++
		}

		score := 0.0
		for _, t := range terms {
			tfv := float64(tf[t])
			if tfv == 0 {
				continue
			}
			idf := math.Log((float64(n)-float64(df[t])+0.5)/(float64(df[t])+0.5) + 1)
			norm := tfv * (k1 + 1) / (tfv + k1*(1-b+b*docLen/avgLen))
			score += idf * norm
		}

		if score > 0 {
			hits = append(hits, scored{item, score})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		return hits[i].score > hits[j].score
	})

	if len(hits) > maxResults {
		hits = hits[:maxResults]
	}

	out := make([]SearchHit, len(hits))
	for i, h := range hits {
		out[i] = SearchHit{
			ID:    h.item.ID,
			Kind:  h.item.Kind,
			Title: h.item.Title,
			Score: h.score,
		}
	}

	return SearchResult{Items: out, Withheld: w.Summary()}
}

// tokenize splits text into lowercase word tokens.
func tokenize(s string) []string {
	s = strings.ToLower(s)
	var tokens []string
	var cur strings.Builder
	for _, r := range s {
		if isWordChar(r) {
			cur.WriteRune(r)
		} else if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

// termSet returns a set of unique tokens for quick membership tests.
func termSet(s string) map[string]bool {
	tokens := tokenize(s)
	m := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		m[t] = true
	}
	return m
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}

package core

import (
	"math"
	"sort"
	"sync"
)

type MemoryIndex struct {
	index     map[string][]int
	totalDocs int
	mu        sync.RWMutex
}

func NewMemoryIndex() *MemoryIndex {
	return &MemoryIndex{
		index: make(map[string][]int),
	}
}

func (m *MemoryIndex) Set(word string, id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.index[word] = append(m.index[word], id)
}

func (m *MemoryIndex) SetTotal(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalDocs = n
}

func (m *MemoryIndex) Gets(words []string) []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	score := make(map[int]float64)
	for _, word := range words {
		if ids, ok := m.index[word]; ok {
			idf := math.Log(float64(m.totalDocs) / float64(len(ids)))
			for _, id := range ids {
				score[id] += idf
			}
		}
	}

	result := make([]int, 0, len(score))
	for id := range score {
		result = append(result, id)
	}

	sort.Slice(result, func(i, j int) bool {
		return score[result[i]] > score[result[j]]
	})

	return result
}

func (m *MemoryIndex) Drop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	clear(m.index)
	m.totalDocs = 0
}

func (m *MemoryIndex) Rebuild(newIndex map[string][]int, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.index = newIndex
	m.totalDocs = total
}

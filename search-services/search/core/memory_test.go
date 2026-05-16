package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func setupIndex(t *testing.T) *MemoryIndex {
	t.Helper()

	index := NewMemoryIndex()
	index.Set("apple", 1)
	index.Set("gopher", 2)
	index.SetTotal(2)
	return index
}

func TestSearchIndex(t *testing.T) {
	t.Run("init index struct", TestIndexInit)
	t.Run("append method to index", TestIndexSearchSet)
	t.Run("set total method", TestIndexSearchSetTotal)
	t.Run("method to gets ids", TestIndexSearchGets)
	t.Run("method to drop index", TestIndexSearchDrop)
	t.Run("method to rebuild index", TestIndexSearchRebuild)
}

func TestIndexInit(t *testing.T) {
	index := setupIndex(t)
	require.NotNil(t, index.index)
	require.Equal(t, 2, index.totalDocs)
}

func TestIndexSearchSet(t *testing.T) {
	index := setupIndex(t)
	index.Set("roman", 3)
	require.Equal(t, 3, len(index.index), "need to append to index")
}

func TestIndexSearchSetTotal(t *testing.T) {
	index := setupIndex(t)
	require.Equal(t, 2, index.totalDocs, "should be zero totalDocs at start")
	index.SetTotal(1)
	require.Equal(t, 1, index.totalDocs, "shouldn't be zero after change")
}

func TestIndexSearchGets(t *testing.T) {
	index := setupIndex(t)
	index.Set("roman", 3)
	index.Set("poli", 4)
	result := index.Gets([]string{"roman", "romeo"})
	require.NotEmpty(t, result, "should return some ids")
	require.Contains(t, result, 3, "should to return 1 in result")
}

func TestIndexSearchDrop(t *testing.T) {
	index := setupIndex(t)
	index.Drop()
	require.Equal(t, 0, len(index.index), "should be zero index after drop")
	require.Equal(t, 0, index.totalDocs, "should be zero totalDocs after drop")
}

func TestIndexSearchRebuild(t *testing.T) {
	index := setupIndex(t)
	testIndex := setupIndex(t)
	testIndex.Set("test", 3)
	index.Rebuild(testIndex.index, 3)
	require.Equal(t, testIndex.index, index.index, "index after rebuild should be new")
	require.Equal(t, 3, index.totalDocs)
}

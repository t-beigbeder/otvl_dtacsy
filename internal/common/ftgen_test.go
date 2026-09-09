package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileTreeGenerate(t *testing.T) {
	td := t.TempDir()
	require.NoError(t, FileTreeGenerate(td, 10, 20, 1, 1024, true, 0))
	require.NoError(t, FileTreeGenerate(td, 1000, 100000, 2, 1024, true, 0))
	require.NoError(t, FileTreeGenerate(td, 10, 20, 1, 1024, false, 0))
	td = t.TempDir()
	require.NoError(t, FileTreeGenerate(td, 100, 10000, 2, 1024, false, 0))
	td = t.TempDir()
	require.NoError(t, FileTreeGenerate(td, 250, 25000, 2, 32768, false, 3))
	require.True(t, true)
}

package readline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReadline(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, rl)
	defer rl.Close()

	assert.NotNil(t, rl.instance)
}

func TestSetPrompt(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	defer rl.Close()

	rl.SetPrompt("new> ")
	assert.Equal(t, "new> ", rl.config.Prompt)
}

func TestWriteString(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	defer rl.Close()

	n, err := rl.WriteString("test output\n")
	assert.NoError(t, err)
	assert.Greater(t, n, 0)
}

func TestPrintln(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	defer rl.Close()

	rl.Println("test output")
}

func TestPrintf(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	defer rl.Close()

	rl.Printf("test %s\n", "output")
}

func TestStdout(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	defer rl.Close()

	assert.NotNil(t, rl.Stdout())
}

func TestStderr(t *testing.T) {
	cfg := Config{
		Prompt:      "test> ",
		HistoryFile: "",
	}

	rl, err := New(cfg)
	require.NoError(t, err)
	defer rl.Close()

	assert.NotNil(t, rl.Stderr())
}

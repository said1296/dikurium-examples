package chainlogger

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"regexp"
	"testing"
)

func TestGetLogger(t *testing.T) {
	expectedChain := "eth_main"
	id := 1
	expectedColorId := 31 + id

	buf := &bytes.Buffer{}

	// Redirect STDOUT to a buffer
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("Failed to redirect STDOUT")
	}
	os.Stdout = w

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			buf.WriteString(scanner.Text())
		}
	}()

	// Force log
	logger := GetLogger(expectedChain, id)
	logger.Info("test")

	// Block until log is scanned
	for buf.Len() == 0 {
	}

	// Close writer and reset stdout
	w.Close()
	os.Stdout = stdout

	// Test output
	reg, err := regexp.Compile(
		fmt.Sprintf("\u001B\\[%dm\\[%s].*INFO.*chainlogger/chainlogger_test.*test", expectedColorId, expectedChain),
	)
	assert.NoError(t, err)
	s := reg.FindString(buf.String())
	assert.NotEmpty(t, s)
}

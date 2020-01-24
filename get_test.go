package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoopbackTestCommand(t *testing.T) {
	assert.Equal(t, []byte{187, 0, 0, 68}, loopbackTestCommand())
}

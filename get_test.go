package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteLoopbackTest(t *testing.T) {
	var buffer bytes.Buffer
	err := writeLoopbackTest(&buffer)
	assert.Nil(t, err)
	assert.Equal(t, []byte{187, 0, 0, 68}, buffer.Bytes())
}

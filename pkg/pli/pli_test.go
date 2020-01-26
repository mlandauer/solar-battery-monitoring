package pli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteLoopbackTest(t *testing.T) {
	var buffer bytes.Buffer
	err := commandLoopbackTest(&buffer)
	assert.Nil(t, err)
	assert.Equal(t, []byte{187, 0, 0, 68}, buffer.Bytes())
}

func TestExtractNibbles(t *testing.T) {
	// 00110001 = 24V system running Prog 3
	msn, lsn := extractNibbles(0x31)
	assert.Equal(t, byte(3), msn)
	assert.Equal(t, byte(1), lsn)
}

package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyPrefix(t *testing.T) {
	assert.True(t, filepath.HasPrefix("pumpcraft/blah.tgz", "pumpcraft"))
	assert.False(t, filepath.HasPrefix("pumpcraft/blah.tgz", "newworld"))
}

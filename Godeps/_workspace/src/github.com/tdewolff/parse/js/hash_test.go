package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashTable(t *testing.T) {
	assert.Equal(t, Break, ToHash([]byte("break")), "'break' must resolve to hash.Break")
	assert.Equal(t, "break", Break.String(), "hash.Break must resolve to 'break'")
	assert.Equal(t, Hash(0), ToHash([]byte("")), "empty string must resolve to zero")
	assert.Equal(t, "", Hash(0xffffff).String(), "Hash(0xffffff) must resolve to empty string")
	assert.Equal(t, Hash(0), ToHash([]byte("breaks")), "'breaks' must resolve to zero")
}

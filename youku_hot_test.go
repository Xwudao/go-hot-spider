package hotspider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYoukuHot_Televisions(t *testing.T) {
	hot := NewYoukuHot()

	words, err := hot.Televisions()

	assert.Nilf(t, err, "Televisions() error = %v", err)
	assert.NotEmpty(t, words)
	t.Logf("Televisions() returned %d words, %v", len(words), words)
}

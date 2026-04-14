package hotspider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectYoukuSearchKeywords(t *testing.T) {
	const payload = `{
		"data": {
			"data": {
				"热门搜索": {
					"nodes": [
						{"data": {"keyword": "月鳞绮纪"}},
						{"nodes": [
							{"data": {"keyword": "重案解密"}},
							{"data": {"keyword": "正义女神"}}
						]}
					]
				}
			}
		}
	}`

	var resp youkuSearchResponse
	err := json.Unmarshal([]byte(payload), &resp)
	require.NoError(t, err)

	words := extractYoukuSearchHotWords(resp.Data.Data, 10)
	assert.Equal(t, []string{"月鳞绮纪", "重案解密", "正义女神"}, words)
}

func TestYoukuHot_Televisions(t *testing.T) {
	hot := NewYoukuHot()

	words, err := hot.Televisions()

	assert.Nilf(t, err, "Televisions() error = %v", err)
	assert.NotEmpty(t, words)
	t.Logf("Televisions() returned %d words, %v", len(words), words)
}

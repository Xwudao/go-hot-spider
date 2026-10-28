package hotspider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractYoukuTabKeywords(t *testing.T) {
	const payload = `{
		"data": {
			"data": {
				"nodes": [{"data": {"tabDataMap": {
					"热门搜索": {"nodes": [{"nodes": [
						{"data": {"keyword": "嫁金钗", "videoInfo": "电视剧·古装·爱情", "showId": "736526", "searchIndexValue": "96.4w"}},
						{"data": {"keyword": "空枪", "videoInfo": "电影·犯罪", "showId": "789832", "searchIndexValue": "90.3w"}},
						{"data": {"keyword": "小猪佩奇 第十二季", "videoInfo": "少儿", "showId": "886756", "searchIndexValue": "61.0w"}},
						{"data": "unexpected"}
					]}]},
					"动漫": {"nodes": [{"nodes": [
						{"data": {"keyword": "师兄啊师兄", "videoInfo": "动漫·新国风", "showId": "554099", "searchIndexValue": "72.3w"}}
					]}]},
					"微短剧": {"nodes": [{"nodes": [
						{"data": {"keyword": "空难当前，前女友拒绝迫降", "videoInfo": "电视剧·都市", "showId": "973248", "searchIndexValue": ""}}
					]}]}
				}}}]
			}
		}
	}`

	var resp youkuSearchResponse
	err := json.Unmarshal([]byte(payload), &resp)
	require.NoError(t, err)

	assert.Equal(t, []string{"嫁金钗", "空枪"}, extractYoukuTabKeywords(resp.Data, youkuHotAllTab, 10))
	assert.Equal(t, []string{"师兄啊师兄"}, extractYoukuTabKeywords(resp.Data, "动漫", 10))
	assert.Empty(t, extractYoukuTabKeywords(resp.Data, "微短剧", 10))
}

func TestYoukuHot_Televisions(t *testing.T) {
	hot := NewYoukuHot()

	words, err := hot.Televisions()

	assert.Nilf(t, err, "Televisions() error = %v", err)
	assert.NotEmpty(t, words)
	t.Logf("Televisions() returned %d words, %v", len(words), words)
}

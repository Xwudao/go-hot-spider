package hotspider

import (
	"regexp"
	"time"

	"github.com/imroc/req/v3"
)

// BaiduSuggestion 百度搜索建议抓取器。
// API: https://suggestion.baidu.com/su?cb=jsonp&wd=<keyword>
type BaiduSuggestion struct {
	r *req.Client
}

// NewBaiduSuggestion 创建百度搜索建议抓取器。
func NewBaiduSuggestion() *BaiduSuggestion {
	var r = req.NewClient().SetTimeout(time.Second * 10).
		SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36").
		SetBaseURL("https://suggestion.baidu.com")
	return &BaiduSuggestion{
		r: r,
	}
}

// GetSuggestion 获取百度搜索建议词列表。
func (b *BaiduSuggestion) GetSuggestion(wd string) ([]string, error) {
	var query = map[string]string{
		"cb": "jsonp",
		"wd": wd,
	}

	resp, err := b.r.R().SetQueryParams(query).Get("/su")
	if err != nil {
		return nil, err
	}

	return b.findWords(resp.String()), nil
}

// jsonp({q:"你好",p:false,s:["你好 李焕英","你好星期六",...]});
var findWordRE = regexp.MustCompile(`(?m)"(.*?)",?`)

func (b *BaiduSuggestion) findWords(s string) []string {
	words := findWordRE.FindAllStringSubmatch(s, -1)

	if len(words) <= 1 {
		return nil
	}

	var rtn []string
	for _, word := range words {
		if len(word) > 1 {
			rtn = append(rtn, word[1])
		}
	}

	return rtn
}

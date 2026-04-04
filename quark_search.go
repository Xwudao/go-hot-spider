package hotspider

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

// QuarkSearch 通过夸克搜索页获取资源内容。
type QuarkSearch struct {
	r *req.Client
}

// NewQuarkSearch 创建夸克搜索抓取器。
func NewQuarkSearch() *QuarkSearch {
	var r = req.NewClient().ImpersonateChrome().SetTimeout(time.Second * 10)

	return &QuarkSearch{
		r: r,
	}
}

// Search 搜索关键词并返回页面主体文本。
func (b *QuarkSearch) Search(q string) (string, error) {
	var api = `https://m.quark.cn/s`
	var qm = map[string]string{
		"q": q,
	}
	resp, err := b.r.R().SetQueryParams(qm).Get(api)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return "", err
	}

	text := doc.Find("#main").Text()
	return text, nil
}

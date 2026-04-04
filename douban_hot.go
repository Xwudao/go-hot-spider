package hotspider

import (
	"errors"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

// DoubanHot 抓取豆瓣电影首页热门影视词。
type DoubanHot struct {
	r *req.Client
}

var doubanSupportedCategories = []VideoCategory{
	VideoCategoryMovie,
}

// NewDoubanHot 创建豆瓣热门词抓取器。
func NewDoubanHot() *DoubanHot {
	r := req.NewClient().SetTimeout(time.Second * 10).
		SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	return &DoubanHot{r: r}
}

// SupportedCategories 返回豆瓣热词当前支持的类目。
func (d *DoubanHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(doubanSupportedCategories)
}

// HotByCategory 按类目返回豆瓣热词。
func (d *DoubanHot) HotByCategory(category VideoCategory) ([]string, error) {
	normalized, ok := normalizeVideoCategory(category)
	if !ok || !supportsVideoCategory(doubanSupportedCategories, normalized) {
		return nil, unsupportedCategoryError("douban hot", category)
	}

	return d.Televisions()
}

// Movies 返回豆瓣电影热词。
func (d *DoubanHot) Movies() ([]string, error) {
	return d.HotByCategory(VideoCategoryMovie)
}

// Teleplays 返回豆瓣电视剧热词。
func (d *DoubanHot) Teleplays() ([]string, error) {
	return d.HotByCategory(VideoCategoryTeleplay)
}

// VarietyShows 返回豆瓣综艺热词。
func (d *DoubanHot) VarietyShows() ([]string, error) {
	return d.HotByCategory(VideoCategoryVariety)
}

// Animations 返回豆瓣动漫热词。
func (d *DoubanHot) Animations() ([]string, error) {
	return d.HotByCategory(VideoCategoryAnimation)
}

// Televisions 返回豆瓣电影首页可见的热门影视词。
func (d *DoubanHot) Televisions() ([]string, error) {
	resp, err := d.r.R().Get("https://movie.douban.com/")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() {
		return nil, errors.New("douban hot fail")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	return d.findWords(doc), nil
}

func (d *DoubanHot) findWords(doc *goquery.Document) []string {
	hotTitles := make([]string, 0, 24)
	seen := make(map[string]struct{})

	appendTitle := func(raw string) {
		title := removeChars(strings.TrimSpace(raw))
		if title == "" {
			return
		}
		if _, ok := seen[title]; ok {
			return
		}

		seen[title] = struct{}{}
		hotTitles = append(hotTitles, title)
	}

	doc.Find(".screening-bd img[alt]").Each(func(_ int, selection *goquery.Selection) {
		appendTitle(selection.AttrOr("alt", ""))
	})
	doc.Find(".billboard-bd a").Each(func(_ int, selection *goquery.Selection) {
		appendTitle(selection.Text())
	})

	return hotTitles
}

package hotspider

import (
	"errors"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

// NovelQuickAppHot 抓取红果短剧官网首页热门短剧词。
type NovelQuickAppHot struct {
	r *req.Client
}

var novelQuickAppSupportedCategories []VideoCategory

// NewNovelQuickAppHot 创建红果短剧热门词抓取器。
func NewNovelQuickAppHot() *NovelQuickAppHot {
	r := req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &NovelQuickAppHot{r: r}
}

// SupportedCategories 返回红果短剧当前支持的类目。
func (n *NovelQuickAppHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(novelQuickAppSupportedCategories)
}

// HotByCategory 按类目返回红果短剧热词。
func (n *NovelQuickAppHot) HotByCategory(category VideoCategory) ([]string, error) {
	return nil, unsupportedCategoryError("novelquickapp hot", category)
}

// Movies 返回红果短剧电影热词。
func (n *NovelQuickAppHot) Movies() ([]string, error) {
	return n.HotByCategory(VideoCategoryMovie)
}

// Teleplays 返回红果短剧电视剧热词。
func (n *NovelQuickAppHot) Teleplays() ([]string, error) {
	return n.HotByCategory(VideoCategoryTeleplay)
}

// VarietyShows 返回红果短剧综艺热词。
func (n *NovelQuickAppHot) VarietyShows() ([]string, error) {
	return n.HotByCategory(VideoCategoryVariety)
}

// Animations 返回红果短剧动漫热词。
func (n *NovelQuickAppHot) Animations() ([]string, error) {
	return n.HotByCategory(VideoCategoryAnimation)
}

// Televisions 返回红果短剧官网首页热门短剧词。
func (n *NovelQuickAppHot) Televisions() ([]string, error) {
	resp, err := n.r.R().Get("https://novelquickapp.com/")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() {
		return nil, errors.New("novelquickapp hot fail")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	words := n.findWords(doc)
	if len(words) == 0 {
		return nil, errors.New("novelquickapp hot fail")
	}

	return words, nil
}

func (n *NovelQuickAppHot) findWords(doc *goquery.Document) []string {
	hotTitles := make([]string, 0, 32)
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

	doc.Find(`a[href*="/detail?series_id="]`).Each(func(_ int, selection *goquery.Selection) {
		paragraphs := selection.Find("p")
		if paragraphs.Length() < 2 {
			return
		}

		appendTitle(paragraphs.Eq(1).Text())
	})

	return hotTitles
}
package hotspider

import (
	"bytes"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
)

// QQHot 抓取腾讯视频热门搜索词。
type QQHot struct {
	r *req.Client
}

type qqHotSection struct {
	channel string
}

const qqHotRanksURL = "https://v.qq.com/biu/ranks/?t=hotsearch"

var qqSupportedCategories = []VideoCategory{
	VideoCategoryMovie,
	VideoCategoryTeleplay,
	VideoCategoryVariety,
	VideoCategoryAnimation,
}

var qqHotSections = map[VideoCategory]qqHotSection{
	VideoCategoryMovie: {
		channel: "1",
	},
	VideoCategoryTeleplay: {
		channel: "2",
	},
	VideoCategoryVariety: {
		channel: "10",
	},
	VideoCategoryAnimation: {
		channel: "3",
	},
}

// NewQQHot 创建腾讯视频热门搜索词抓取器。
func NewQQHot() *QQHot {
	r := req.NewClient().SetTimeout(time.Second*10).ImpersonateChrome().
		SetCommonHeader("Referer", qqHotRanksURL)
	return &QQHot{r: r}
}

// SupportedCategories 返回腾讯视频当前支持的类目。
func (q *QQHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(qqSupportedCategories)
}

// HotByCategory 按类目返回腾讯视频热词。
func (q *QQHot) HotByCategory(category VideoCategory) ([]string, error) {
	normalized, ok := normalizeVideoCategory(category)
	if !ok || !supportsVideoCategory(qqSupportedCategories, normalized) {
		return nil, unsupportedCategoryError("qq hot", category)
	}

	section := qqHotSections[normalized]
	words, err := q.fetchWords(section.channel)
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, errors.New("qq hot fail")
	}

	return words, nil
}

// Movies 返回腾讯视频电影热词。
func (q *QQHot) Movies() ([]string, error) {
	return q.HotByCategory(VideoCategoryMovie)
}

// Teleplays 返回腾讯视频电视剧热词。
func (q *QQHot) Teleplays() ([]string, error) {
	return q.HotByCategory(VideoCategoryTeleplay)
}

// VarietyShows 返回腾讯视频综艺热词。
func (q *QQHot) VarietyShows() ([]string, error) {
	return q.HotByCategory(VideoCategoryVariety)
}

// Animations 返回腾讯视频动漫热词。
func (q *QQHot) Animations() ([]string, error) {
	return q.HotByCategory(VideoCategoryAnimation)
}

// Televisions 返回腾讯视频热门搜索词。
func (q *QQHot) Televisions() ([]string, error) {
	words, err := q.fetchWords("0")
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, errors.New("qq hot fail")
	}

	return words, nil
}

func (q *QQHot) fetchWords(channel string) ([]string, error) {
	html, err := q.fetchPage(qqHotRanksURL)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
	if err != nil {
		return nil, err
	}

	words := q.findWords(doc, channel)
	if len(words) > 0 {
		return words, nil
	}

	return q.findWordsFromSelection(doc.Selection), nil
}

func (q *QQHot) fetchPage(pageURL string) (string, error) {
	resp, err := q.r.R().Get(pageURL)
	if err != nil {
		return "", err
	}
	if !resp.IsSuccessState() {
		return "", errors.New("qq hot fail")
	}

	body, err := resp.ToString()
	if err != nil {
		return "", err
	}

	return body, nil
}

func (q *QQHot) findWords(doc *goquery.Document, channel string) []string {
	words := make([]string, 0, 16)
	target := strings.TrimSpace(channel)

	doc.Find(".mod_rank_figure").EachWithBreak(func(_ int, section *goquery.Selection) bool {
		href, ok := section.Find(".mod_rank_title .link_more").First().Attr("href")
		if !ok || qqHotChannelFromHref(href) != target {
			return true
		}

		words = q.findWordsFromSelection(section)
		return false
	})

	return words
}

func (q *QQHot) findWordsFromSelection(selection *goquery.Selection) []string {
	hotTitles := make([]string, 0, 16)
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

	selection.Find(".mod_rank_search_list .hotlist a").Each(func(_ int, item *goquery.Selection) {
		appendTitle(qqHotWord(item))
	})

	return hotTitles
}

func qqHotWord(selection *goquery.Selection) string {
	if href, ok := selection.Attr("href"); ok {
		if title := qqHotWordFromHref(href); title != "" {
			return title
		}
	}

	if title, ok := selection.Attr("title"); ok {
		if trimmed := strings.TrimSpace(title); trimmed != "" {
			return trimmed
		}
	}

	if text := strings.TrimSpace(selection.Find(".name").First().Text()); text != "" {
		return text
	}

	return strings.TrimSpace(selection.Text())
}

func qqHotWordFromHref(rawURL string) string {
	href := strings.TrimSpace(rawURL)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	query := parsed.Query()
	return strings.TrimSpace(query.Get("q"))
}

func qqHotChannelFromHref(rawURL string) string {
	href := strings.TrimSpace(rawURL)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "//") {
		href = "https:" + href
	} else if strings.HasPrefix(href, "/") {
		href = "https://v.qq.com" + href
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(parsed.Query().Get("channel"))
}

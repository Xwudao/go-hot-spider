package hotspider

import (
	"errors"
	"time"

	"github.com/imroc/req/v3"
)

// QuarkHot 抓取夸克影视排行榜（电影 / 电视剧）热词。
type QuarkHot struct {
	r *req.Client
}

var quarkSupportedCategories = []VideoCategory{
	VideoCategoryMovie,
	VideoCategoryTeleplay,
	VideoCategoryVariety,
	VideoCategoryAnimation,
}

// NewQuarkHot 创建夸克热门搜索词抓取器。
func NewQuarkHot() *QuarkHot {
	var r = req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &QuarkHot{
		r: r,
	}
}

// SupportedCategories 返回夸克热榜当前支持的类目。
func (q *QuarkHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(quarkSupportedCategories)
}

// HotByCategory 按类目返回夸克热词。
func (q *QuarkHot) HotByCategory(category VideoCategory) ([]string, error) {
	normalized, ok := normalizeVideoCategory(category)
	if !ok || !supportsVideoCategory(quarkSupportedCategories, normalized) {
		return nil, unsupportedCategoryError("quark hot", category)
	}

	return q.fetchChannel(string(normalized))
}

// Movies 返回夸克电影热词。
func (q *QuarkHot) Movies() ([]string, error) {
	return q.HotByCategory(VideoCategoryMovie)
}

// Teleplays 返回夸克电视剧热词。
func (q *QuarkHot) Teleplays() ([]string, error) {
	return q.HotByCategory(VideoCategoryTeleplay)
}

// VarietyShows 返回夸克综艺热词。
func (q *QuarkHot) VarietyShows() ([]string, error) {
	return q.HotByCategory(VideoCategoryVariety)
}

// Animations 返回夸克动漫热词。
func (q *QuarkHot) Animations() ([]string, error) {
	return q.HotByCategory(VideoCategoryAnimation)
}

// Televisions 返回夸克电影 / 电视剧榜单热词。
func (q *QuarkHot) Televisions() ([]string, error) {
	var channels = []string{"电影", "电视剧"}

	var hotTitles []string

	for _, channel := range channels {
		words, err := q.fetchChannel(channel)
		if err != nil {
			return nil, err
		}
		hotTitles = append(hotTitles, words...)
	}

	return hotTitles, nil
}

func (q *QuarkHot) fetchChannel(channel string) ([]string, error) {
	var qm = map[string]string{
		"channel":      channel,
		"rank_type":    "最热",
		"second_tag":   "true",
		"start":        "0",
		"hit":          "10",
		"area":         "全部",
		"year":         "全部",
		"cate":         "全部",
		"uc_param_str": "dnfrpfbivessbtbmnilauputogpintnwmtsvcppcprsnnnchmicckpgixsnx",
		"from":         "hot_page",
		"belong":       "quark",
	}

	var url = "https://biz.quark.cn/api/trending/ranking/getYingshiRanking"
	var respModel QuarkTelevisionsModel

	resp, err := q.r.R().SetSuccessResult(&respModel).SetQueryParams(qm).Get(url)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() {
		return nil, errors.New(resp.String())
	}

	hotTitles := make([]string, 0, len(respModel.Data.Hits.Hit.Item))
	seen := make(map[string]struct{}, len(respModel.Data.Hits.Hit.Item))
	for _, item := range respModel.Data.Hits.Hit.Item {
		title := removeChars(item.Title)
		if title == "" {
			continue
		}
		if _, ok := seen[title]; ok {
			continue
		}

		seen[title] = struct{}{}
		hotTitles = append(hotTitles, title)
	}

	return hotTitles, nil
}

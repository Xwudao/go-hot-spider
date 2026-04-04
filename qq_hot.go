package hotspider

import (
	"errors"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// QQHot 抓取腾讯视频热门搜索词。
type QQHot struct {
	r *req.Client
}

type qqHotResponse struct {
	Data struct {
		ErrorCode   int `json:"errorCode"`
		HotWordList []struct {
			SearchWord string `json:"searchWord"`
		} `json:"hotWordList"`
	} `json:"data"`
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}

var qqSupportedCategories []VideoCategory

// NewQQHot 创建腾讯视频热门搜索词抓取器。
func NewQQHot() *QQHot {
	r := req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &QQHot{r: r}
}

// SupportedCategories 返回腾讯视频当前支持的类目。
func (q *QQHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(qqSupportedCategories)
}

// HotByCategory 按类目返回腾讯视频热词。
func (q *QQHot) HotByCategory(category VideoCategory) ([]string, error) {
	return nil, unsupportedCategoryError("qq hot", category)
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
	var respModel qqHotResponse
	query := map[string]string{
		"appID":     "3172",
		"appKey":    "lGhFIPeD3HsO9xEp",
		"platform":  "2",
		"channelID": "0",
		"v":         "2958812",
	}

	resp, err := q.r.R().SetSuccessResult(&respModel).SetQueryParams(query).Get("https://pbaccess.video.qq.com/trpc.universal_backend_service.hot_word_info.HttpHotWordRecall/GetHotWords")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() || respModel.Ret != 0 || respModel.Data.ErrorCode != 0 {
		return nil, errors.New("qq hot fail")
	}

	return q.findWords(&respModel), nil
}

func (q *QQHot) findWords(respModel *qqHotResponse) []string {
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

	for _, item := range respModel.Data.HotWordList {
		appendTitle(item.SearchWord)
	}

	return hotTitles
}

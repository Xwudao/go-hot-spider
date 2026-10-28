package hotspider

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// QQHot 抓取腾讯视频热门搜索词。
type QQHot struct {
	r *req.Client
}

// qqHotRanksURL 是腾讯视频排行榜页面地址，仅用于设置 Referer。
const qqHotRanksURL = "https://v.qq.com/ranks/pages/index.html?t=hotsearch"

// qqHotListAPI 是排行榜页面实际使用的热词接口。
const qqHotListAPI = "https://pbaccess.video.qq.com/trpc.tencentvideohotlistdata.hotlistdata.DoSearchHotListHttp/getSearchHotListHttp"

const (
	qqHotAllChannel = ""
	qqHotAllPageLen = 10
	qqHotCatPageLen = 50
)

// qqHotResponse 是热词接口的响应结构。
type qqHotResponse struct {
	Data struct {
		ErrCode int    `json:"ErrCode"`
		ErrMsg  string `json:"ErrMsg"`
		Results []struct {
			ChannelID   string `json:"ChannelID"`
			ChannelName string `json:"ChannelName"`
			Result      struct {
				TotalNum int `json:"TotalNum"`
				ListInfo []struct {
					Query       string `json:"Query"`
					HotLevel    int64  `json:"HotLevel"`
					Trend       int    `json:"Trend"`
					ID          string `json:"ID"`
					SecondTitle string `json:"SecondTitle"`
					HzImgURL    string `json:"HzImgURL"`
					VideoType   string `json:"VideoType"`
				} `json:"ListInfo"`
			} `json:"Result"`
		} `json:"Results"`
	} `json:"data"`
}

var qqSupportedCategories = []VideoCategory{
	VideoCategoryMovie,
	VideoCategoryTeleplay,
	VideoCategoryVariety,
	VideoCategoryAnimation,
}

var qqHotSections = map[VideoCategory]string{
	VideoCategoryMovie:     "1",
	VideoCategoryTeleplay:  "2",
	VideoCategoryVariety:   "10",
	VideoCategoryAnimation: "3",
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

	channel := qqHotSections[normalized]
	words, err := q.fetchWords(channel, qqHotCatPageLen)
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
	words, err := q.fetchWords(qqHotAllChannel, qqHotAllPageLen)
	if err != nil {
		return nil, err
	}
	if len(words) == 0 {
		return nil, errors.New("qq hot fail")
	}

	return words, nil
}

func (q *QQHot) fetchWords(channel string, pageSize int) ([]string, error) {
	respModel, err := q.fetchResponse(channel, pageSize)
	if err != nil {
		return nil, err
	}

	words := q.findWords(respModel, channel)
	if len(words) == 0 && channel != qqHotAllChannel {
		// 接口偶尔忽略 ChannelID，退化为取首个榜单项。
		words = q.findWords(respModel, "")
	}

	return words, nil
}

func (q *QQHot) fetchResponse(channel string, pageSize int) (*qqHotResponse, error) {
	query := map[string]string{
		"ChannelID": channel,
		"PageSize":  strconv.Itoa(pageSize),
		"PageNum":   "0",
		"SceneID":   "2",
	}

	var respModel qqHotResponse
	resp, err := q.r.R().SetQueryParams(query).SetSuccessResult(&respModel).Get(qqHotListAPI)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() || respModel.Data.ErrCode != 0 {
		return nil, errors.New("qq hot fail")
	}

	return &respModel, nil
}

func (q *QQHot) findWords(respModel *qqHotResponse, channel string) []string {
	words := make([]string, 0, 16)
	seen := make(map[string]struct{})

	target := strings.TrimSpace(channel)

	appendTitle := func(raw string) {
		if isNoiseTitle(raw) {
			return
		}

		title := removeChars(strings.TrimSpace(raw))
		if title == "" {
			return
		}
		if _, ok := seen[title]; ok {
			return
		}

		seen[title] = struct{}{}
		words = append(words, title)
	}

	for _, result := range respModel.Data.Results {
		if target != "" && strings.TrimSpace(result.ChannelID) != target {
			continue
		}

		for _, item := range result.Result.ListInfo {
			appendTitle(item.Query)
		}
	}

	return words
}

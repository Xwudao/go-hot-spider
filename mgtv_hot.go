package hotspider

import (
	"errors"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// MGTVHot 抓取芒果搜索页热门搜索词。
type MGTVHot struct {
	r *req.Client
}

type mgtvSuggestResponse struct {
	Code int `json:"code"`
	Data struct {
		TopList []struct {
			Label string `json:"label"`
			Data  []struct {
				Name string `json:"name"`
			} `json:"data"`
		} `json:"topList"`
	} `json:"data"`
}

var mgtvSupportedCategories = []VideoCategory{
	VideoCategoryMovie,
	VideoCategoryTeleplay,
	VideoCategoryVariety,
	VideoCategoryAnimation,
}

// NewMGTVHot 创建芒果热门搜索词抓取器。
func NewMGTVHot() *MGTVHot {
	r := req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &MGTVHot{r: r}
}

// SupportedCategories 返回芒果热词当前支持的类目。
func (m *MGTVHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(mgtvSupportedCategories)
}

// HotByCategory 按类目返回芒果热词。
func (m *MGTVHot) HotByCategory(category VideoCategory) ([]string, error) {
	normalized, ok := normalizeVideoCategory(category)
	if !ok || !supportsVideoCategory(mgtvSupportedCategories, normalized) {
		return nil, unsupportedCategoryError("mgtv hot", category)
	}

	respModel, err := m.fetchHotResponse()
	if err != nil {
		return nil, err
	}

	words := m.findWordsByGroup(respModel, string(normalized))
	if len(words) == 0 {
		return nil, errors.New("mgtv hot fail")
	}

	return words, nil
}

// Movies 返回芒果电影热词。
func (m *MGTVHot) Movies() ([]string, error) {
	return m.HotByCategory(VideoCategoryMovie)
}

// Teleplays 返回芒果电视剧热词。
func (m *MGTVHot) Teleplays() ([]string, error) {
	return m.HotByCategory(VideoCategoryTeleplay)
}

// VarietyShows 返回芒果综艺热词。
func (m *MGTVHot) VarietyShows() ([]string, error) {
	return m.HotByCategory(VideoCategoryVariety)
}

// Animations 返回芒果动漫热词。
func (m *MGTVHot) Animations() ([]string, error) {
	return m.HotByCategory(VideoCategoryAnimation)
}

// Televisions 返回芒果搜索页的热门搜索词。
func (m *MGTVHot) Televisions() ([]string, error) {
	respModel, err := m.fetchHotResponse()
	if err != nil {
		return nil, err
	}

	return m.findWords(respModel), nil
}

func (m *MGTVHot) fetchHotResponse() (*mgtvSuggestResponse, error) {
	var respModel mgtvSuggestResponse
	query := map[string]string{
		"allowedRC": "1",
		"src":       "mgtv",
		"pc":        "1",
		"q":         "",
		"_support":  "10000000",
	}

	resp, err := m.r.R().SetSuccessResult(&respModel).SetQueryParams(query).Get("https://mobileso.bz.mgtv.com/pc/suggest/v1")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() || respModel.Code != 200 {
		return nil, errors.New("mgtv hot fail")
	}

	return &respModel, nil
}

func (m *MGTVHot) findWords(respModel *mgtvSuggestResponse) []string {
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

	for _, group := range respModel.Data.TopList {
		for _, item := range group.Data {
			appendTitle(item.Name)
		}
	}

	return hotTitles
}

func (m *MGTVHot) findWordsByGroup(respModel *mgtvSuggestResponse, label string) []string {
	words := make([]string, 0, 16)
	seen := make(map[string]struct{})

	for _, group := range respModel.Data.TopList {
		if strings.TrimSpace(group.Label) != label {
			continue
		}

		for _, item := range group.Data {
			title := removeChars(strings.TrimSpace(item.Name))
			if title == "" {
				continue
			}
			if _, ok := seen[title]; ok {
				continue
			}

			seen[title] = struct{}{}
			words = append(words, title)
		}
	}

	return words
}

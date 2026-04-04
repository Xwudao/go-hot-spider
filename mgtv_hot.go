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

// NewMGTVHot 创建芒果热门搜索词抓取器。
func NewMGTVHot() *MGTVHot {
	r := req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &MGTVHot{r: r}
}

// Televisions 返回芒果搜索页的热门搜索词。
func (m *MGTVHot) Televisions() ([]string, error) {
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

	return m.findWords(&respModel), nil
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

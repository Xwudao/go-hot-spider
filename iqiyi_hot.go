package hotspider

import (
	"errors"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// IQiyiHot 抓取爱奇艺热门搜索词。
type IQiyiHot struct {
	r *req.Client
}

type iQiyiHotItem struct {
	Title  string `json:"title"`
	Tag    string `json:"tag"`
	QipuID int64  `json:"qipuId"`
}

type iQiyiHotResponse struct {
	HotQuery []struct {
		Title string         `json:"title"`
		Items []iQiyiHotItem `json:"items"`
	} `json:"hotQuery"`
}

var iqiyiSupportedCategories = map[string]struct{}{
	"电影":  {},
	"电视剧": {},
	"综艺":  {},
	"动漫":  {},
}

var iqiyiSupplementGroups = []string{"电视剧", "电影", "综艺", "动漫"}

// NewIQiyiHot 创建爱奇艺热门搜索词抓取器。
func NewIQiyiHot() *IQiyiHot {
	r := req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &IQiyiHot{r: r}
}

// Televisions 返回爱奇艺热门搜索词。
func (i *IQiyiHot) Televisions() ([]string, error) {
	var respModel iQiyiHotResponse
	resp, err := i.r.R().SetSuccessResult(&respModel).SetQueryParam("v", "17.041.24982").Get("https://mesh.if.iqiyi.com/portal/lw/search/keywords/hotList")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() || len(respModel.HotQuery) == 0 {
		return nil, errors.New("iqiyi hot fail")
	}

	words := i.findWords(&respModel)
	if len(words) == 0 {
		return nil, errors.New("iqiyi hot fail")
	}

	return words, nil
}

func (i *IQiyiHot) findWords(respModel *iQiyiHotResponse) []string {
	hotTitles := make([]string, 0, 20)
	seen := make(map[string]struct{})
	groups := make(map[string][]iQiyiHotItem, len(respModel.HotQuery))

	appendItem := func(item iQiyiHotItem) {
		if item.QipuID <= 0 || !i.supportsCategory(i.categoryOf(item.Tag)) {
			return
		}

		title := removeChars(strings.TrimSpace(item.Title))
		if title == "" {
			return
		}
		if _, ok := seen[title]; ok {
			return
		}

		seen[title] = struct{}{}
		hotTitles = append(hotTitles, title)
	}

	for _, group := range respModel.HotQuery {
		groups[group.Title] = group.Items
	}

	for _, item := range groups["热搜"] {
		appendItem(item)
	}

	for _, groupTitle := range iqiyiSupplementGroups {
		if len(hotTitles) >= cap(hotTitles) {
			break
		}

		for _, item := range groups[groupTitle] {
			appendItem(iQiyiHotItem{
				Title:  item.Title,
				Tag:    groupTitle,
				QipuID: item.QipuID,
			})
			if len(hotTitles) >= cap(hotTitles) {
				break
			}
		}
	}

	return hotTitles
}

func (i *IQiyiHot) categoryOf(tag string) string {
	part, _, _ := strings.Cut(tag, "/")
	return strings.TrimSpace(part)
}

func (i *IQiyiHot) supportsCategory(category string) bool {
	_, ok := iqiyiSupportedCategories[strings.TrimSpace(category)]
	return ok
}

package hotspider

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// YoukuHot 抓取优酷搜索页热门搜索词。
type YoukuHot struct {
	r *req.Client
}

const (
	youkuRecommendAppKey = "24679788"
	youkuSearchAppKey    = "23774304"
	youkuDefaultUtdid    = "homepage_empty_cna"
	youkuSearchHotLimit  = 12
	youkuMtopBaseURL     = "https://acs.youku.com/h5/"
)

type youkuRecommendData struct {
	AppID      string `json:"appid"`
	MTopParams string `json:"mtopParams"`
	Utdid      string `json:"utdid"`
}

type youkuAuthContext struct {
	Token string
	Utdid string
}

type youkuSearchData struct {
	Pg         string  `json:"pg"`
	Pz         string  `json:"pz"`
	AppScene   string  `json:"appScene"`
	AppCaller  string  `json:"appCaller"`
	SearchFrom string  `json:"searchFrom"`
	Utdid      *string `json:"utdId"`
	YkPid      string  `json:"ykPid"`
}

type youkuSearchResponse struct {
	// Data 同时包含 data（搜索框配置）与 nodes（热搜榜单 tabDataMap），整体保留供递归查找。
	Data json.RawMessage `json:"data"`
	Ret  []string        `json:"ret"`
}

type youkuSearchNode struct {
	// Data 结构随节点类型变化，按需逐个解析，避免单个异常节点导致整体解析失败。
	Data  json.RawMessage   `json:"data"`
	Nodes []youkuSearchNode `json:"nodes"`
}

type youkuHotItem struct {
	Keyword          string `json:"keyword"`
	VideoInfo        string `json:"videoInfo"`
	ShowID           string `json:"showId"`
	SearchIndexValue string `json:"searchIndexValue"`
}

// youkuHotAllTab 是热搜榜的综合榜单；同一响应中还有 少儿/人文/纪录片/微短剧 等榜单，均不抓取。
const youkuHotAllTab = "热门搜索"

var youkuSupportedCategories = []VideoCategory{
	VideoCategoryMovie,
	VideoCategoryTeleplay,
	VideoCategoryVariety,
	VideoCategoryAnimation,
}

var youkuHotTabs = map[VideoCategory]string{
	VideoCategoryMovie:     "电影",
	VideoCategoryTeleplay:  "电视剧",
	VideoCategoryVariety:   "综艺",
	VideoCategoryAnimation: "动漫",
}

// NewYoukuHot 创建优酷热门搜索词抓取器。
func NewYoukuHot() *YoukuHot {
	r := req.NewClient().SetTimeout(time.Second*10).ImpersonateChrome().
		SetCommonHeader("Referer", "https://so.youku.com/search/q_")
	return &YoukuHot{r: r}
}

// SupportedCategories 返回优酷当前支持的类目。
func (y *YoukuHot) SupportedCategories() []VideoCategory {
	return copyVideoCategories(youkuSupportedCategories)
}

// HotByCategory 按类目返回优酷热词。
func (y *YoukuHot) HotByCategory(category VideoCategory) ([]string, error) {
	normalized, ok := normalizeVideoCategory(category)
	if !ok || !supportsVideoCategory(youkuSupportedCategories, normalized) {
		return nil, unsupportedCategoryError("youku hot", category)
	}

	return y.fetchTabWords(youkuHotTabs[normalized])
}

// Movies 返回优酷电影热词。
func (y *YoukuHot) Movies() ([]string, error) {
	return y.HotByCategory(VideoCategoryMovie)
}

// Teleplays 返回优酷电视剧热词。
func (y *YoukuHot) Teleplays() ([]string, error) {
	return y.HotByCategory(VideoCategoryTeleplay)
}

// VarietyShows 返回优酷综艺热词。
func (y *YoukuHot) VarietyShows() ([]string, error) {
	return y.HotByCategory(VideoCategoryVariety)
}

// Animations 返回优酷动漫热词。
func (y *YoukuHot) Animations() ([]string, error) {
	return y.HotByCategory(VideoCategoryAnimation)
}

// Televisions 返回优酷热门搜索词。
func (y *YoukuHot) Televisions() ([]string, error) {
	return y.fetchTabWords(youkuHotAllTab)
}

// fetchTabWords 抓取热搜榜中指定榜单的词。
// 不再使用 ykrec 推荐接口：它是个性化的“猜你想搜”，混有人名、儿歌、体育等非影视词。
func (y *YoukuHot) fetchTabWords(tab string) ([]string, error) {
	recommendData, err := y.requestRecommendData(youkuDefaultUtdid)
	if err != nil {
		return nil, err
	}

	auth, err := y.bootstrapAuth(recommendData)
	if err != nil {
		return nil, err
	}

	data, err := y.requestSearchHotData(auth)
	if err != nil {
		return nil, err
	}

	words := y.findWords(extractYoukuTabKeywords(data, tab, youkuSearchHotLimit))
	if len(words) == 0 {
		return nil, errors.New("youku search hot fail")
	}

	return words, nil
}

func (y *YoukuHot) requestRecommendData(utdid string) (string, error) {
	payload := youkuRecommendData{
		AppID:      "14177",
		MTopParams: fmt.Sprintf(`{"count":"10","channel":"PC","fr":"pc","app_source":"main_page","x_utdid":"%s"}`, utdid),
		Utdid:      utdid,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (y *YoukuHot) requestSearchData(utdid *string) (string, error) {
	payload := youkuSearchData{
		Pg:         "1",
		Pz:         "10",
		AppScene:   "default_page",
		AppCaller:  "youku-search-sdk",
		SearchFrom: "home",
		Utdid:      utdid,
		YkPid:      "",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (y *YoukuHot) bootstrapAuth(data string) (*youkuAuthContext, error) {
	auth := &youkuAuthContext{Utdid: youkuDefaultUtdid}
	_, _ = y.r.R().Get("https://so.youku.com/search/q_")

	searchData, err := y.requestSearchData(nil)
	if err == nil {
		resp, searchErr := y.r.R().SetQueryParams(map[string]string{
			"jsv":            "2.7.5",
			"appKey":         youkuSearchAppKey,
			"t":              strconv.FormatInt(time.Now().UnixMilli(), 10),
			"sign":           "",
			"api":            "mtop.youku.soku.yksearch",
			"type":           "jsonp",
			"v":              "2.0",
			"ecode":          "1",
			"dataType":       "jsonp",
			"jsonpIncPrefix": "soukuheaderSearch",
			"callback":       "mtopjsonpsoukuheaderSearch1",
			"data":           searchData,
		}).Get(youkuMtopBaseURL + "mtop.youku.soku.yksearch/2.0/")
		if searchErr == nil {
			for _, cookie := range resp.Cookies() {
				switch cookie.Name {
				case "_m_h5_tk":
					auth.Token = strings.SplitN(cookie.Value, "_", 2)[0]
				case "cna":
					if strings.TrimSpace(cookie.Value) != "" {
						auth.Utdid = cookie.Value
					}
				}
			}
		}
	}
	if auth.Token == "" {
		resp, err := y.r.R().SetQueryParams(map[string]string{
			"jsv":      "2.7.5",
			"appKey":   youkuRecommendAppKey,
			"t":        strconv.FormatInt(time.Now().UnixMilli(), 10),
			"sign":     "",
			"api":      "mtop.ykrec.RecommendService.recommend",
			"v":        "1.0",
			"dataType": "json",
			"type":     "originaljson",
			"data":     data,
		}).Get(youkuMtopBaseURL + "mtop.ykrec.recommendservice.recommend/1.0/")
		if err != nil {
			return nil, err
		}

		for _, cookie := range resp.Cookies() {
			switch cookie.Name {
			case "_m_h5_tk":
				auth.Token = strings.SplitN(cookie.Value, "_", 2)[0]
			case "cna":
				if strings.TrimSpace(cookie.Value) != "" {
					auth.Utdid = cookie.Value
				}
			}
		}
	}
	if auth.Utdid == youkuDefaultUtdid {
		auth.Utdid = y.findCookieValue("https://acs.youku.com/", "cna")
	}
	if auth.Utdid == "" {
		auth.Utdid = y.findCookieValue("https://so.youku.com/search/q_", "cna")
	}
	if auth.Utdid == "" {
		_, _ = y.r.R().Get("https://so.youku.com/search/q_")
		auth.Utdid = y.findCookieValue("https://so.youku.com/search/q_", "cna")
	}
	if auth.Utdid == "" {
		auth.Utdid = youkuDefaultUtdid
	}
	if auth.Token == "" {
		return nil, errors.New("youku mtop token missing")
	}

	return auth, nil
}

func (y *YoukuHot) findCookieValue(rawURL, name string) string {
	cookies, err := y.r.GetCookies(rawURL)
	if err != nil {
		return ""
	}

	for _, cookie := range cookies {
		if cookie.Name == name && strings.TrimSpace(cookie.Value) != "" {
			return cookie.Value
		}
	}

	return ""
}

func (y *YoukuHot) requestSearchHotData(auth *youkuAuthContext) (json.RawMessage, error) {
	// utdId 为 null 时接口不返回 tabDataMap 热搜榜单。
	data, err := y.requestSearchData(&auth.Utdid)
	if err != nil {
		return nil, err
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(auth.Token+"&"+timestamp+"&"+youkuSearchAppKey+"&"+data)))

	resp, err := y.r.R().SetQueryParams(map[string]string{
		"jsv":            "2.7.5",
		"appKey":         youkuSearchAppKey,
		"t":              timestamp,
		"sign":           sign,
		"api":            "mtop.youku.soku.yksearch",
		"type":           "jsonp",
		"v":              "2.0",
		"ecode":          "1",
		"dataType":       "jsonp",
		"jsonpIncPrefix": "soukuheaderSearch",
		"callback":       "mtopjsonpsoukuheaderSearch1",
		"data":           data,
	}).Get(youkuMtopBaseURL + "mtop.youku.soku.yksearch/2.0/")
	if err != nil {
		return nil, err
	}

	payload, err := unwrapYoukuJSONP(resp.String())
	if err != nil {
		return nil, err
	}

	var respModel youkuSearchResponse
	if err := json.Unmarshal([]byte(payload), &respModel); err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() || len(respModel.Data) == 0 {
		return nil, errors.New("youku search hot fail")
	}

	return respModel.Data, nil
}

func unwrapYoukuJSONP(body string) (string, error) {
	start := strings.IndexByte(body, '(')
	end := strings.LastIndexByte(body, ')')
	if start < 0 || end <= start {
		return "", errors.New("invalid youku jsonp payload")
	}

	return body[start+1 : end], nil
}

// extractYoukuTabKeywords 从热搜榜响应的 tabDataMap 中取出指定榜单的影视词。
func extractYoukuTabKeywords(data json.RawMessage, tab string, limit int) []string {
	if limit <= 0 || len(data) == 0 {
		return nil
	}

	var tabs map[string]youkuSearchNode
	if err := json.Unmarshal(findYoukuSection(data, "tabDataMap"), &tabs); err != nil {
		return nil
	}

	root, ok := tabs[tab]
	if !ok {
		return nil
	}

	return collectYoukuSearchKeywords(root, limit)
}

func findYoukuSection(data json.RawMessage, target string) json.RawMessage {
	var search func(raw json.RawMessage) json.RawMessage
	search = func(raw json.RawMessage) json.RawMessage {
		if len(raw) == 0 {
			return nil
		}

		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err == nil {
			if section, ok := object[target]; ok {
				return section
			}
			for _, value := range object {
				if section := search(value); len(section) > 0 {
					return section
				}
			}
			return nil
		}

		var array []json.RawMessage
		if err := json.Unmarshal(raw, &array); err == nil {
			for _, value := range array {
				if section := search(value); len(section) > 0 {
					return section
				}
			}
		}

		return nil
	}

	return search(data)
}

func collectYoukuSearchKeywords(root youkuSearchNode, limit int) []string {
	if limit <= 0 {
		return nil
	}

	words := make([]string, 0, limit)
	var walk func(node youkuSearchNode)
	walk = func(node youkuSearchNode) {
		if len(words) >= limit {
			return
		}
		var item youkuHotItem
		if len(node.Data) > 0 && json.Unmarshal(node.Data, &item) == nil && item.isVideo() {
			words = append(words, item.Keyword)
			if len(words) >= limit {
				return
			}
		}

		for _, child := range node.Nodes {
			walk(child)
			if len(words) >= limit {
				return
			}
		}
	}

	walk(root)
	return words
}

// isVideo 判断热搜项是否为正经影视条目：需有节目 ID 和搜索指数（微短剧榜单没有搜索指数），
// 且 videoInfo 首段（如“电视剧·古装·爱情”中的“电视剧”）属于支持的类目，排除少儿/文化/纪录片等。
func (item youkuHotItem) isVideo() bool {
	if strings.TrimSpace(item.Keyword) == "" || item.ShowID == "" || strings.TrimSpace(item.SearchIndexValue) == "" {
		return false
	}

	kind, _, _ := strings.Cut(item.VideoInfo, "·")
	category, ok := normalizeVideoCategoryString(kind)
	return ok && supportsVideoCategory(youkuSupportedCategories, category)
}

func (y *YoukuHot) findWords(words []string) []string {
	hotTitles := make([]string, 0, 16)
	seen := make(map[string]struct{})

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
		hotTitles = append(hotTitles, title)
	}

	for _, word := range words {
		appendTitle(word)
	}

	return hotTitles
}

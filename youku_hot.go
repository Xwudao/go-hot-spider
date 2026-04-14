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

type youkuRecommendResponse struct {
	Status int      `json:"status"`
	Q      []string `json:"q"`
	Ret    []string `json:"ret"`
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
	Data struct {
		Data json.RawMessage `json:"data"`
	} `json:"data"`
	Ret []string `json:"ret"`
}

type youkuSearchNode struct {
	Data  *youkuSearchNodeData `json:"data"`
	Nodes []youkuSearchNode    `json:"nodes"`
}

type youkuSearchNodeData struct {
	Keyword string `json:"keyword"`
}

var youkuSupportedCategories []VideoCategory

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
	return nil, unsupportedCategoryError("youku hot", category)
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
	recommendData, err := y.requestRecommendData(youkuDefaultUtdid)
	if err != nil {
		return nil, err
	}

	auth, err := y.bootstrapAuth(recommendData)
	if err != nil {
		return nil, err
	}

	merged := make([]string, 0, 24)
	errs := make([]error, 0, 2)

	recommendWords, err := y.requestRecommendWords(auth.Token, recommendData)
	if err != nil {
		errs = append(errs, fmt.Errorf("youku recommend words: %w", err))
	} else {
		merged = append(merged, recommendWords...)
	}

	searchHotWords, err := y.requestSearchHotWords(auth)
	if err != nil {
		errs = append(errs, fmt.Errorf("youku search hot words: %w", err))
	} else {
		merged = append(merged, searchHotWords...)
	}

	merged = y.findWords(merged)
	if len(merged) > 0 {
		return merged, nil
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return nil, errors.New("youku hot fail")
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

func (y *YoukuHot) requestSearchData() (string, error) {
	payload := youkuSearchData{
		Pg:         "1",
		Pz:         "10",
		AppScene:   "default_page",
		AppCaller:  "youku-search-sdk",
		SearchFrom: "home",
		Utdid:      nil,
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

	searchData, err := y.requestSearchData()
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

func (y *YoukuHot) requestRecommendWords(token, data string) ([]string, error) {
	var respModel youkuRecommendResponse
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(token+"&"+timestamp+"&"+youkuRecommendAppKey+"&"+data)))

	resp, err := y.r.R().SetSuccessResult(&respModel).SetQueryParams(map[string]string{
		"jsv":      "2.7.5",
		"appKey":   youkuRecommendAppKey,
		"t":        timestamp,
		"sign":     sign,
		"api":      "mtop.ykrec.RecommendService.recommend",
		"v":        "1.0",
		"dataType": "json",
		"type":     "originaljson",
		"data":     data,
	}).Get(youkuMtopBaseURL + "mtop.ykrec.recommendservice.recommend/1.0/")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() || respModel.Status != 0 || len(respModel.Q) == 0 {
		return nil, errors.New("youku recommend hot fail")
	}

	return respModel.Q, nil
}

func (y *YoukuHot) requestSearchHotWords(auth *youkuAuthContext) ([]string, error) {
	data, err := y.requestSearchData()
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

	words := extractYoukuSearchHotWords(respModel.Data.Data, youkuSearchHotLimit)
	if !resp.IsSuccessState() || len(words) == 0 {
		return nil, errors.New("youku search hot fail")
	}

	return words, nil
}

func unwrapYoukuJSONP(body string) (string, error) {
	start := strings.IndexByte(body, '(')
	end := strings.LastIndexByte(body, ')')
	if start < 0 || end <= start {
		return "", errors.New("invalid youku jsonp payload")
	}

	return body[start+1 : end], nil
}

func extractYoukuSearchHotWords(data json.RawMessage, limit int) []string {
	if limit <= 0 || len(data) == 0 {
		return nil
	}

	for _, sectionName := range []string{"热门搜索", "电视剧"} {
		section := findYoukuSection(data, sectionName)
		if len(section) == 0 {
			continue
		}

		var node youkuSearchNode
		if err := json.Unmarshal(section, &node); err != nil {
			continue
		}

		if words := collectYoukuSearchKeywords(node, limit); len(words) > 0 {
			return words
		}
	}

	var node youkuSearchNode
	if err := json.Unmarshal(data, &node); err != nil {
		return nil
	}

	return collectYoukuSearchKeywords(node, limit)
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
		if node.Data != nil && strings.TrimSpace(node.Data.Keyword) != "" {
			words = append(words, node.Data.Keyword)
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

func (y *YoukuHot) findWords(words []string) []string {
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

	for _, word := range words {
		appendTitle(word)
	}

	return hotTitles
}

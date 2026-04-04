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

// NewYoukuHot 创建优酷热门搜索词抓取器。
func NewYoukuHot() *YoukuHot {
	r := req.NewClient().SetTimeout(time.Second*10).ImpersonateChrome().
		SetCommonHeader("Referer", "https://so.youku.com/search/q_")
	return &YoukuHot{r: r}
}

// Televisions 返回优酷热门搜索词。
func (y *YoukuHot) Televisions() ([]string, error) {
	data, err := y.requestData()
	if err != nil {
		return nil, err
	}

	token, err := y.bootstrapToken(data)
	if err != nil {
		return nil, err
	}

	var respModel youkuRecommendResponse
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(token+"&"+timestamp+"&24679788&"+data)))

	resp, err := y.r.R().SetSuccessResult(&respModel).SetQueryParams(map[string]string{
		"jsv":      "2.7.5",
		"appKey":   "24679788",
		"t":        timestamp,
		"sign":     sign,
		"api":      "mtop.ykrec.RecommendService.recommend",
		"v":        "1.0",
		"dataType": "json",
		"type":     "originaljson",
		"data":     data,
	}).Get("https://acs.youku.com/h5/mtop.ykrec.recommendservice.recommend/1.0/")
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccessState() || respModel.Status != 0 || len(respModel.Q) == 0 {
		return nil, errors.New("youku hot fail")
	}

	return y.findWords(respModel.Q), nil
}

func (y *YoukuHot) requestData() (string, error) {
	payload := youkuRecommendData{
		AppID:      "14177",
		MTopParams: `{"count":"10","channel":"PC","fr":"pc","app_source":"main_page","x_utdid":"homepage_empty_cna"}`,
		Utdid:      "homepage_empty_cna",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (y *YoukuHot) bootstrapToken(data string) (string, error) {
	resp, err := y.r.R().SetQueryParams(map[string]string{
		"jsv":      "2.7.5",
		"appKey":   "24679788",
		"t":        strconv.FormatInt(time.Now().UnixMilli(), 10),
		"sign":     "",
		"api":      "mtop.ykrec.RecommendService.recommend",
		"v":        "1.0",
		"dataType": "json",
		"type":     "originaljson",
		"data":     data,
	}).Get("https://acs.youku.com/h5/mtop.ykrec.recommendservice.recommend/1.0/")
	if err != nil {
		return "", err
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_m_h5_tk" {
			return strings.SplitN(cookie.Value, "_", 2)[0], nil
		}
	}

	return "", errors.New("youku mtop token missing")
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

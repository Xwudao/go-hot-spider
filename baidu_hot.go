package hotspider

import (
	"regexp"
	"time"

	"github.com/imroc/req/v3"
	json "github.com/json-iterator/go"
)

// BaiduHot 抓取百度热榜（电影 / 电视剧）。
type BaiduHot struct {
	r *req.Client
}

// NewBaiduHot 创建百度热榜抓取器。
func NewBaiduHot() *BaiduHot {
	var r = req.NewClient().SetTimeout(time.Second * 10).
		SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36").
		SetBaseURL("https://top.baidu.com/board")
	return &BaiduHot{
		r: r,
	}
}

// GetMovie 获取百度电影榜热词。
func (b *BaiduHot) GetMovie() ([]*HotData, error) {
	return b.GetHotByType("movie")
}

// GetTeleplay 获取百度电视剧榜热词。
func (b *BaiduHot) GetTeleplay() ([]*HotData, error) {
	return b.GetHotByType("teleplay")
}

// Televisions 返回百度电影榜和电视剧榜的热词。
func (b *BaiduHot) Televisions() ([]string, error) {
	movie, err := b.GetMovie()
	if err != nil {
		return nil, err
	}

	teleplay, err := b.GetTeleplay()
	if err != nil {
		return nil, err
	}

	hotTitles := make([]string, 0, len(movie)+len(teleplay))
	for _, item := range movie {
		hotTitles = append(hotTitles, removeChars(item.Word))
	}
	for _, item := range teleplay {
		hotTitles = append(hotTitles, removeChars(item.Word))
	}

	return hotTitles, nil
}

// GetHotByType 按 tab 类型获取热榜数据。
func (b *BaiduHot) GetHotByType(tab string) ([]*HotData, error) {
	var query = map[string]string{
		"tab": tab,
	}

	resp, err := b.r.R().SetQueryParams(query).Get("")
	if err != nil {
		return nil, err
	}

	return b.findWords(resp.String()), nil
}

var wordRE = regexp.MustCompile(`(?m)<!--s-data:(.*?)-->`)

func (b *BaiduHot) findWords(text string) []*HotData {
	words := wordRE.FindStringSubmatch(text)

	if len(words) <= 1 {
		return nil
	}

	var data BaiduHotData
	err := json.UnmarshalFromString(words[1], &data)
	if err != nil {
		return nil
	}

	var hotData []*HotData
	if len(data.Data.Cards) == 0 {
		return nil
	}
	for _, card := range data.Data.Cards[0].Content {
		hotData = append(hotData, &HotData{
			Desc:  card.Desc,
			Word:  card.Word,
			Show:  card.Show,
			Index: card.Index,
			Image: card.Img,
		})
	}

	return hotData
}

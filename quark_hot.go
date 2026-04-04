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

// NewQuarkHot 创建夸克热门搜索词抓取器。
func NewQuarkHot() *QuarkHot {
	var r = req.NewClient().SetTimeout(time.Second * 10).ImpersonateChrome()
	return &QuarkHot{
		r: r,
	}
}

// Televisions 返回夸克电影 / 电视剧榜单热词。
func (q *QuarkHot) Televisions() ([]string, error) {
	var channels = []string{"电影", "电视剧"}

	var hotTitles []string

	for _, channel := range channels {
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

		for _, s := range respModel.Data.Hits.Hit.Item {
			hotTitles = append(hotTitles, removeChars(s.Title))
		}
	}

	return hotTitles, nil
}

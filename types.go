package hotspider

import "time"

// HotData 热门内容数据
type HotData struct {
	Desc  string   `json:"desc"`
	Word  string   `json:"word"`
	Show  []string `json:"show"`
	Index int      `json:"index"`
	Image string   `json:"image"`

	DoubanID string   `json:"douban_id"`
	Keywords []string `json:"-"`

	CreateTime time.Time `json:"create_time"`
}

// BaiduHotData 百度热榜原始响应数据
type BaiduHotData struct {
	Data struct {
		Cards []struct {
			Component string `json:"component"`
			Content   []struct {
				AppUrl    string   `json:"appUrl"`
				Desc      string   `json:"desc"`
				HotChange string   `json:"hotChange"`
				HotScore  string   `json:"hotScore"`
				Img       string   `json:"img"`
				Index     int      `json:"index"`
				IndexUrl  string   `json:"indexUrl"`
				Query     string   `json:"query"`
				RawUrl    string   `json:"rawUrl"`
				Show      []string `json:"show"`
				Url       string   `json:"url"`
				Word      string   `json:"word"`
			} `json:"content"`
			More       int    `json:"more"`
			MoreAppUrl string `json:"moreAppUrl"`
			MoreUrl    string `json:"moreUrl"`
			Text       string `json:"text"`
			TopContent any    `json:"topContent"`
			TypeName   string `json:"typeName"`
		} `json:"cards"`
		CurBoardName string `json:"curBoardName"`
		Logid        string `json:"logid"`
		Platform     string `json:"platform"`
		TabBoard     []struct {
			Index    int    `json:"index"`
			Text     string `json:"text"`
			TypeName string `json:"typeName"`
		} `json:"tabBoard"`
		Tag []struct {
			TypeName string   `json:"typeName"`
			Text     string   `json:"text"`
			Content  []string `json:"content"`
			CurIndex int      `json:"curIndex"`
		} `json:"tag"`
	} `json:"data"`
	View   string `json:"view"`
	Tab    string `json:"tab"`
	Config struct {
	} `json:"config"`
	RootWrapperClass string `json:"rootWrapperClass"`
	ShowScrollToTop  bool   `json:"showScrollToTop"`
}

# go-hot-spider

A Go library for scraping hot-search keywords and trending content from major Chinese video platforms and search engines.

## Installation

```bash
go get github.com/Xwudao/go-hot-spider
```

## Supported Platforms

| Platform | Struct | Description |
|----------|--------|-------------|
| 百度热榜 | `BaiduHot` | 电影 / 电视剧热榜词 |
| 百度搜索建议 | `BaiduSuggestion` | 搜索词联想建议 |
| 爱奇艺 | `IQiyiHot` | 热门搜索词（电影/电视剧/综艺/动漫） |
| 芒果TV | `MGTVHot` | 热门搜索词 |
| 腾讯视频 | `QQHot` | 热门搜索词 |
| 夸克视频 | `QuarkHot` | 影视排行榜热词 |
| 夸克搜索 | `QuarkSearch` | 网盘搜索结果文本 |
| 豆瓣 | `DoubanHot` | 电影首页热门影视词 |
| 优酷 | `YoukuHot` | 热门搜索词 |

## Usage

Each scraper exposes a `Televisions() ([]string, error)` method that returns a deduplicated list of hot keyword strings. `BaiduSuggestion` instead exposes `GetSuggestion(wd string) ([]string, error)`.

```go
package main

import (
    "fmt"
    hotspider "github.com/Xwudao/go-hot-spider"
)

func main() {
    // 百度热榜
    bh := hotspider.NewBaiduHot()
    words, err := bh.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("百度热榜:", words)

    // 爱奇艺热搜
    iq := hotspider.NewIQiyiHot()
    words, err = iq.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("爱奇艺热搜:", words)

    // 腾讯视频热搜
    qq := hotspider.NewQQHot()
    words, err = qq.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("腾讯视频热搜:", words)

    // 芒果TV热搜
    mg := hotspider.NewMGTVHot()
    words, err = mg.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("芒果TV热搜:", words)

    // 夸克热榜
    qk := hotspider.NewQuarkHot()
    words, err = qk.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("夸克热榜:", words)

    // 豆瓣首页热词
    db := hotspider.NewDoubanHot()
    words, err = db.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("豆瓣热词:", words)

    // 优酷热搜
    yk := hotspider.NewYoukuHot()
    words, err = yk.Televisions()
    if err != nil {
        panic(err)
    }
    fmt.Println("优酷热搜:", words)

    // 百度搜索建议
    bs := hotspider.NewBaiduSuggestion()
    suggestions, err := bs.GetSuggestion("三体")
    if err != nil {
        panic(err)
    }
    fmt.Println("百度建议:", suggestions)

    // 百度热榜详细数据（含描述、图片等）
    movie, err := bh.GetMovie()
    if err != nil {
        panic(err)
    }
    for _, item := range movie {
        fmt.Printf("Word: %s, Desc: %s, Image: %s\n", item.Word, item.Desc, item.Image)
    }
}
```

## Types

### HotData

```go
type HotData struct {
    Desc       string    `json:"desc"`
    Word       string    `json:"word"`
    Show       []string  `json:"show"`
    Index      int       `json:"index"`
    Image      string    `json:"image"`
    DoubanID   string    `json:"douban_id"`
    Keywords   []string  `json:"-"`
    CreateTime time.Time `json:"create_time"`
}
```

`BaiduHot.GetMovie()` and `BaiduHot.GetTeleplay()` return `[]*HotData` with richer metadata. All other scrapers return `[]string` via `Televisions()`.

## License

MIT

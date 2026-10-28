package hotspider

import (
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

// NoiseFilter 过滤非正经影视词（短剧、网文式标题、儿歌等），作用于抓取到的原始标题。
type NoiseFilter struct {
	// Keywords 标题包含任一关键词即被过滤。
	Keywords []string
	// CommaTitleMinLen 含逗号且去标点后长度不少于该值的标题被过滤，<=0 表示不启用。
	// 如“十八岁太奶奶驾到，重整家族荣耀”。
	CommaTitleMinLen int
	// PronounTitleMinLen 含“我/你”且去标点后长度不少于该值的标题被过滤，<=0 表示不启用。
	// 如“身为魔王的我娶了精灵为妻该如何表白我的爱”。
	PronounTitleMinLen int
	// Custom 自定义判断，返回 true 即过滤，可为 nil。
	Custom func(title string) bool
}

// DefaultNoiseFilter 返回默认过滤规则。
func DefaultNoiseFilter() *NoiseFilter {
	return &NoiseFilter{
		Keywords:           []string{"儿歌", "动态漫", "漫剧", "短剧"},
		CommaTitleMinLen:   11,
		PronounTitleMinLen: 12,
	}
}

// IsNoise 判断原始标题是否应被过滤，需传入 removeChars 之前的标题（依赖标点判断）。
func (f *NoiseFilter) IsNoise(raw string) bool {
	if f == nil {
		return false
	}

	title := strings.TrimSpace(raw)
	for _, keyword := range f.Keywords {
		if keyword != "" && strings.Contains(title, keyword) {
			return true
		}
	}

	length := utf8.RuneCountInString(removeChars(title))
	if f.CommaTitleMinLen > 0 && strings.ContainsAny(title, ",，") && length >= f.CommaTitleMinLen {
		return true
	}
	if f.PronounTitleMinLen > 0 && strings.ContainsAny(title, "我你") && length >= f.PronounTitleMinLen {
		return true
	}

	return f.Custom != nil && f.Custom(title)
}

var noiseFilter atomic.Pointer[NoiseFilter]

func init() {
	noiseFilter.Store(DefaultNoiseFilter())
}

// SetNoiseFilter 替换全局过滤规则，对之后的所有抓取生效；传 nil 关闭过滤。
func SetNoiseFilter(f *NoiseFilter) {
	noiseFilter.Store(f)
}

// CurrentNoiseFilter 返回当前生效的过滤规则，未启用时为 nil。
func CurrentNoiseFilter() *NoiseFilter {
	return noiseFilter.Load()
}

func isNoiseTitle(raw string) bool {
	return noiseFilter.Load().IsNoise(raw)
}

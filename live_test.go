package hotspider

import (
	"errors"
	"strings"
	"testing"
)

type liveWordFetcher interface {
	Televisions() ([]string, error)
}

type categoryWordFetcher interface {
	HotByCategory(category VideoCategory) ([]string, error)
}

type categorySupportFetcher interface {
	SupportedCategories() []VideoCategory
}

func TestLiveTelevisions(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	tests := []struct {
		name    string
		fetcher liveWordFetcher
	}{
		{name: "baidu", fetcher: NewBaiduHot()},
		{name: "iqiyi", fetcher: NewIQiyiHot()},
		{name: "mgtv", fetcher: NewMGTVHot()},
		{name: "qq", fetcher: NewQQHot()},
		{name: "quark", fetcher: NewQuarkHot()},
		{name: "douban", fetcher: NewDoubanHot()},
		{name: "youku", fetcher: NewYoukuHot()},
		{name: "novelquickapp", fetcher: NewNovelQuickAppHot()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, err := tt.fetcher.Televisions()
			if err != nil {
				t.Fatalf("Televisions() error = %v", err)
			}

			assertLiveWords(t, words)
		})
	}
}

func TestBaiduHotLiveData(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	bh := NewBaiduHot()

	tests := []struct {
		name  string
		fetch func() ([]*HotData, error)
	}{
		{name: "movie", fetch: bh.GetMovie},
		{name: "teleplay", fetch: bh.GetTeleplay},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := tt.fetch()
			if err != nil {
				t.Fatalf("fetch error = %v", err)
			}
			if len(items) == 0 {
				t.Fatal("expected at least one item")
			}

			for _, item := range items {
				if item == nil {
					t.Fatal("expected non-nil hot item")
				}
				if strings.TrimSpace(item.Word) == "" {
					t.Fatal("expected hot item word to be non-empty")
				}
			}
		})
	}
}

func TestBaiduSuggestionLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	suggestions, err := NewBaiduSuggestion().GetSuggestion("三体")
	if err != nil {
		t.Fatalf("GetSuggestion() error = %v", err)
	}

	assertLiveWords(t, suggestions)
}

func TestHotByCategoryLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	tests := []struct {
		name     string
		fetcher  categoryWordFetcher
		category VideoCategory
	}{
		{name: "baidu-movie", fetcher: NewBaiduHot(), category: VideoCategoryMovie},
		{name: "baidu-teleplay", fetcher: NewBaiduHot(), category: VideoCategoryTeleplay},
		{name: "iqiyi-movie", fetcher: NewIQiyiHot(), category: VideoCategoryMovie},
		{name: "iqiyi-teleplay", fetcher: NewIQiyiHot(), category: VideoCategoryTeleplay},
		{name: "iqiyi-variety", fetcher: NewIQiyiHot(), category: VideoCategoryVariety},
		{name: "iqiyi-animation", fetcher: NewIQiyiHot(), category: VideoCategoryAnimation},
		{name: "mgtv-movie", fetcher: NewMGTVHot(), category: VideoCategoryMovie},
		{name: "mgtv-teleplay", fetcher: NewMGTVHot(), category: VideoCategoryTeleplay},
		{name: "mgtv-variety", fetcher: NewMGTVHot(), category: VideoCategoryVariety},
		{name: "mgtv-animation", fetcher: NewMGTVHot(), category: VideoCategoryAnimation},
		{name: "quark-movie", fetcher: NewQuarkHot(), category: VideoCategoryMovie},
		{name: "quark-teleplay", fetcher: NewQuarkHot(), category: VideoCategoryTeleplay},
		{name: "quark-variety", fetcher: NewQuarkHot(), category: VideoCategoryVariety},
		{name: "quark-animation", fetcher: NewQuarkHot(), category: VideoCategoryAnimation},
		{name: "douban-movie", fetcher: NewDoubanHot(), category: VideoCategoryMovie},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, err := tt.fetcher.HotByCategory(tt.category)
			if err != nil {
				t.Fatalf("HotByCategory(%q) error = %v", tt.category, err)
			}

			assertLiveWords(t, words)
		})
	}
}

func TestHotByCategoryUnsupported(t *testing.T) {
	tests := []struct {
		name     string
		fetcher  categoryWordFetcher
		category VideoCategory
	}{
		{name: "qq-movie", fetcher: NewQQHot(), category: VideoCategoryMovie},
		{name: "youku-teleplay", fetcher: NewYoukuHot(), category: VideoCategoryTeleplay},
		{name: "novelquickapp-movie", fetcher: NewNovelQuickAppHot(), category: VideoCategoryMovie},
		{name: "douban-variety", fetcher: NewDoubanHot(), category: VideoCategoryVariety},
		{name: "baidu-animation", fetcher: NewBaiduHot(), category: VideoCategoryAnimation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fetcher.HotByCategory(tt.category)
			if !errors.Is(err, ErrCategoryNotSupported) {
				t.Fatalf("expected ErrCategoryNotSupported, got %v", err)
			}
		})
	}
}

func TestSupportedCategories(t *testing.T) {
	tests := []struct {
		name     string
		fetcher  categorySupportFetcher
		expected []VideoCategory
	}{
		{name: "baidu", fetcher: NewBaiduHot(), expected: []VideoCategory{VideoCategoryMovie, VideoCategoryTeleplay}},
		{name: "iqiyi", fetcher: NewIQiyiHot(), expected: []VideoCategory{VideoCategoryMovie, VideoCategoryTeleplay, VideoCategoryVariety, VideoCategoryAnimation}},
		{name: "mgtv", fetcher: NewMGTVHot(), expected: []VideoCategory{VideoCategoryMovie, VideoCategoryTeleplay, VideoCategoryVariety, VideoCategoryAnimation}},
		{name: "qq", fetcher: NewQQHot(), expected: nil},
		{name: "quark", fetcher: NewQuarkHot(), expected: []VideoCategory{VideoCategoryMovie, VideoCategoryTeleplay, VideoCategoryVariety, VideoCategoryAnimation}},
		{name: "douban", fetcher: NewDoubanHot(), expected: []VideoCategory{VideoCategoryMovie}},
		{name: "youku", fetcher: NewYoukuHot(), expected: nil},
		{name: "novelquickapp", fetcher: NewNovelQuickAppHot(), expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categories := tt.fetcher.SupportedCategories()
			if len(categories) != len(tt.expected) {
				t.Fatalf("SupportedCategories() len = %d, want %d", len(categories), len(tt.expected))
			}

			for index, category := range tt.expected {
				if categories[index] != category {
					t.Fatalf("SupportedCategories()[%d] = %q, want %q", index, categories[index], category)
				}
			}
		})
	}
}

func TestQuarkSearchLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	text, err := NewQuarkSearch().Search("三体")
	if err != nil {
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "captcha") || strings.Contains(errText, "blocked") {
			return
		}

		t.Fatalf("Search() error = %v", err)
	}
	if strings.TrimSpace(text) == "" {
		t.Fatal("expected search result text to be non-empty")
	}
}

func assertLiveWords(t *testing.T, words []string) {
	t.Helper()

	if len(words) == 0 {
		t.Fatal("expected at least one word")
	}

	nonEmpty := 0
	seen := make(map[string]struct{}, len(words))
	for _, word := range words {
		cleanWord := strings.TrimSpace(word)
		if cleanWord == "" {
			continue
		}

		nonEmpty++
		seen[cleanWord] = struct{}{}
	}

	if nonEmpty == 0 {
		t.Fatal("expected at least one non-empty word")
	}
	if len(seen) == 0 {
		t.Fatal("expected at least one distinct word")
	}
}

package hotspider

import (
	"strings"
	"testing"
)

type liveWordFetcher interface {
	Televisions() ([]string, error)
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

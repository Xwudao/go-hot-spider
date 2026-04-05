package hotspider

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestNovelQuickAppLiveTelevisions(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	words, err := NewNovelQuickAppHot().Televisions()
	if err != nil {
		t.Fatalf("Televisions() error = %v", err)
	}

	assertLiveWords(t, words)
	t.Logf("live words: %v", words)
}

func TestNovelQuickAppFindWords(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
		<html><body>
			<a href="/detail?series_id=1">
				<p>全76集</p>
				<p>战龙令</p>
			</a>
			<a href="/detail?series_id=2">
				<p>全65集</p>
				<p>绯色过浓</p>
			</a>
			<a href="/detail?series_id=2">
				<p>全65集</p>
				<p>绯色过浓</p>
			</a>
			<a href="/detail?series_id=3">
				<p>播放正片</p>
			</a>
		</body></html>
	`))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}

	words := NewNovelQuickAppHot().findWords(doc)
	want := []string{"战龙令", "绯色过浓"}
	if len(words) != len(want) {
		t.Fatalf("len(words) = %d, want %d; words = %v", len(words), len(want), words)
	}

	for index, word := range want {
		if words[index] != word {
			t.Fatalf("words[%d] = %q, want %q", index, words[index], word)
		}
	}
}

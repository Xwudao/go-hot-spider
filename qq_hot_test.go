package hotspider

import (
	"reflect"
	"testing"
)

func TestQQHotSupportedCategories(t *testing.T) {
	got := NewQQHot().SupportedCategories()
	t.Logf("QQ SupportedCategories(): %v", got)

	want := []VideoCategory{
		VideoCategoryMovie,
		VideoCategoryTeleplay,
		VideoCategoryVariety,
		VideoCategoryAnimation,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedCategories() = %v, want %v", got, want)
	}
}

func TestQQHotLiveData(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live HTTP tests in short mode")
	}

	fetcher := NewQQHot()

	televisions, err := fetcher.Televisions()
	if err != nil {
		t.Fatalf("Televisions() error = %v", err)
	}
	t.Logf("QQ Televisions(): %v", televisions)
	assertLiveWords(t, televisions)

	tests := []struct {
		name     string
		category VideoCategory
	}{
		{name: "movie", category: VideoCategoryMovie},
		{name: "teleplay", category: VideoCategoryTeleplay},
		{name: "variety", category: VideoCategoryVariety},
		{name: "animation", category: VideoCategoryAnimation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, err := fetcher.HotByCategory(tt.category)
			if err != nil {
				t.Fatalf("HotByCategory(%q) error = %v", tt.category, err)
			}

			t.Logf("QQ %s: %v", tt.category, words)
			assertLiveWords(t, words)
		})
	}
}

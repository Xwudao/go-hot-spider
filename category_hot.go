package hotspider

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// VideoCategory 表示影视热词分类。
type VideoCategory string

const (
	VideoCategoryMovie     VideoCategory = "电影"
	VideoCategoryTeleplay  VideoCategory = "电视剧"
	VideoCategoryVariety   VideoCategory = "综艺"
	VideoCategoryAnimation VideoCategory = "动漫"
)

// ErrCategoryNotSupported 表示当前站点不支持指定类目。
var ErrCategoryNotSupported = errors.New("hot category not supported")

var videoCategoryAliases = map[string]VideoCategory{
	"电影":           VideoCategoryMovie,
	"movie":        VideoCategoryMovie,
	"movies":       VideoCategoryMovie,
	"电视剧":          VideoCategoryTeleplay,
	"teleplay":     VideoCategoryTeleplay,
	"teleplays":    VideoCategoryTeleplay,
	"tv":           VideoCategoryTeleplay,
	"tvseries":     VideoCategoryTeleplay,
	"tvshow":       VideoCategoryTeleplay,
	"tvshows":      VideoCategoryTeleplay,
	"series":       VideoCategoryTeleplay,
	"综艺":           VideoCategoryVariety,
	"variety":      VideoCategoryVariety,
	"varieties":    VideoCategoryVariety,
	"varietyshow":  VideoCategoryVariety,
	"varietyshows": VideoCategoryVariety,
	"show":         VideoCategoryVariety,
	"shows":        VideoCategoryVariety,
	"动漫":           VideoCategoryAnimation,
	"animation":    VideoCategoryAnimation,
	"animations":   VideoCategoryAnimation,
	"anime":        VideoCategoryAnimation,
	"cartoon":      VideoCategoryAnimation,
	"cartoons":     VideoCategoryAnimation,
}

func normalizeVideoCategory(category VideoCategory) (VideoCategory, bool) {
	return normalizeVideoCategoryString(string(category))
}

func normalizeVideoCategoryString(raw string) (VideoCategory, bool) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.ReplaceAll(key, " ", "")

	category, ok := videoCategoryAliases[key]
	return category, ok
}

func supportsVideoCategory(categories []VideoCategory, target VideoCategory) bool {
	return slices.Contains(categories, target)
}

func copyVideoCategories(categories []VideoCategory) []VideoCategory {
	if len(categories) == 0 {
		return nil
	}

	out := make([]VideoCategory, len(categories))
	copy(out, categories)
	return out
}

func unsupportedCategoryError(provider string, category VideoCategory) error {
	return fmt.Errorf("%w: %s does not support %s", ErrCategoryNotSupported, provider, category)
}

func wordsFromHotData(items []*HotData) []string {
	words := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		if item == nil {
			continue
		}

		title := removeChars(strings.TrimSpace(item.Word))
		if title == "" {
			continue
		}
		if _, ok := seen[title]; ok {
			continue
		}

		seen[title] = struct{}{}
		words = append(words, title)
	}

	return words
}

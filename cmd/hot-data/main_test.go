package main

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	hotspider "github.com/Xwudao/go-hot-spider"
)

type stubFetcher struct {
	categories []hotspider.VideoCategory
	television []string
	byCategory map[hotspider.VideoCategory][]string
	errs       map[string]error
	calls      []string
}

func (s *stubFetcher) Televisions() ([]string, error) {
	s.calls = append(s.calls, "televisions")
	if err := s.errs["televisions"]; err != nil {
		return nil, err
	}

	return s.television, nil
}

func (s *stubFetcher) SupportedCategories() []hotspider.VideoCategory {
	return append([]hotspider.VideoCategory(nil), s.categories...)
}

func (s *stubFetcher) HotByCategory(category hotspider.VideoCategory) ([]string, error) {
	s.calls = append(s.calls, string(category))
	if err := s.errs[string(category)]; err != nil {
		return nil, err
	}

	return s.byCategory[category], nil
}

func TestBuildTasksUsesCategoriesAndTelevisions(t *testing.T) {
	categorized := &stubFetcher{
		categories: []hotspider.VideoCategory{hotspider.VideoCategoryMovie, hotspider.VideoCategoryTeleplay},
		byCategory: map[hotspider.VideoCategory][]string{
			hotspider.VideoCategoryMovie:    {"流浪地球"},
			hotspider.VideoCategoryTeleplay: {"三体"},
		},
	}
	uncategorized := &stubFetcher{television: []string{"庆余年"}}

	tasks := buildTasks([]sourceConfig{
		{name: "categorized", fetcher: categorized},
		{name: "uncategorized", fetcher: uncategorized},
	})

	labels := make([]string, 0, len(tasks))
	for _, task := range tasks {
		labels = append(labels, task.label)
		if _, err := task.fetch(); err != nil {
			t.Fatalf("task %s returned error: %v", task.label, err)
		}
	}

	wantLabels := []string{"categorized/电影", "categorized/电视剧", "uncategorized"}
	if !reflect.DeepEqual(labels, wantLabels) {
		t.Fatalf("labels = %v, want %v", labels, wantLabels)
	}
	if !reflect.DeepEqual(categorized.calls, []string{"电影", "电视剧"}) {
		t.Fatalf("categorized calls = %v", categorized.calls)
	}
	if !reflect.DeepEqual(uncategorized.calls, []string{"televisions"}) {
		t.Fatalf("uncategorized calls = %v", uncategorized.calls)
	}
}

func TestCollectWordsDeduplicatesAndSleepsBetweenTasks(t *testing.T) {
	tasks := []fetchTask{
		{label: "one", fetch: func() ([]string, error) { return []string{"三体", "繁花"}, nil }},
		{label: "two", fetch: func() ([]string, error) { return []string{"繁花", "庆余年"}, nil }},
		{label: "three", fetch: func() ([]string, error) { return nil, errors.New("boom") }},
	}

	var sleeps []time.Duration
	history := newWordHistory(filepath.Join(t.TempDir(), "data", "movies.txt"), nil)
	errList, err := collectWords(tasks, 1500*time.Millisecond, func(delay time.Duration) {
		sleeps = append(sleeps, delay)
	}, history.addAndFlush)

	if err != nil {
		t.Fatalf("collectWords() error = %v", err)
	}
	if len(errList) != 1 {
		t.Fatalf("errs len = %d, want 1", len(errList))
	}
	if !reflect.DeepEqual(history.words, []string{"三体", "繁花", "庆余年"}) {
		t.Fatalf("history words = %v", history.words)
	}
	if !reflect.DeepEqual(sleeps, []time.Duration{1500 * time.Millisecond, 1500 * time.Millisecond}) {
		t.Fatalf("sleeps = %v", sleeps)
	}
}

func TestWordHistoryAddAndFlushWritesIncrementally(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "movies.txt")
	history := newWordHistory(path, []string{"三体"})

	added, err := history.addAndFlush([]string{"三体", "繁花"})
	if err != nil {
		t.Fatalf("addAndFlush() error = %v", err)
	}
	if added != 1 {
		t.Fatalf("added = %d, want 1", added)
	}

	gotWords, err := readWords(path)
	if err != nil {
		t.Fatalf("readWords() error = %v", err)
	}
	if !reflect.DeepEqual(gotWords, []string{"三体", "繁花"}) {
		t.Fatalf("gotWords = %v", gotWords)
	}
}

func TestMergeWordsPreservesExistingOrder(t *testing.T) {
	merged, newCount := mergeWords(
		[]string{"三体", "繁花"},
		[]string{"繁花", "庆余年", " ", "猎罪图鉴"},
	)

	if !reflect.DeepEqual(merged, []string{"三体", "繁花", "庆余年", "猎罪图鉴"}) {
		t.Fatalf("merged = %v", merged)
	}
	if newCount != 2 {
		t.Fatalf("newCount = %d, want 2", newCount)
	}
}

func TestReadAndWriteWords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "movies.txt")
	wantWords := []string{"三体", "繁花", "庆余年"}

	if err := writeWords(path, wantWords); err != nil {
		t.Fatalf("writeWords() error = %v", err)
	}

	gotWords, err := readWords(path)
	if err != nil {
		t.Fatalf("readWords() error = %v", err)
	}
	if !reflect.DeepEqual(gotWords, wantWords) {
		t.Fatalf("gotWords = %v, want %v", gotWords, wantWords)
	}
}

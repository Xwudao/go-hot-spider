package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	hotspider "github.com/Xwudao/go-hot-spider"
)

type wordFetcher interface {
	Televisions() ([]string, error)
	SupportedCategories() []hotspider.VideoCategory
	HotByCategory(category hotspider.VideoCategory) ([]string, error)
}

type sourceConfig struct {
	name    string
	fetcher wordFetcher
}

type fetchTask struct {
	label string
	fetch func() ([]string, error)
}

func main() {
	var outputPath string
	var delay time.Duration

	flag.StringVar(&outputPath, "output", filepath.Join("data", "movies.txt"), "output file path")
	flag.DurationVar(&delay, "delay", 2*time.Second, "delay between upstream requests")
	flag.Parse()

	if err := run(outputPath, delay); err != nil {
		log.Fatal(err)
	}
}

func run(outputPath string, delay time.Duration) error {
	existingWords, err := readWords(outputPath)
	if err != nil {
		return fmt.Errorf("read existing words: %w", err)
	}

	history := newWordHistory(outputPath, existingWords)
	tasks := buildTasks(defaultSources())
	fetchErrs, err := collectWords(tasks, delay, time.Sleep, history.addAndFlush)
	if err != nil {
		return err
	}

	if len(history.words) == 0 && len(fetchErrs) > 0 {
		return errors.Join(fetchErrs...)
	}

	log.Printf("saved %d words to %s (%d new)", len(history.words), outputPath, history.newCount)
	if len(fetchErrs) > 0 {
		return fmt.Errorf("completed with %d fetch errors: %w", len(fetchErrs), errors.Join(fetchErrs...))
	}

	return nil
}

type wordHistory struct {
	path     string
	words    []string
	seen     map[string]struct{}
	newCount int
}

func newWordHistory(path string, existingWords []string) *wordHistory {
	seen := make(map[string]struct{}, len(existingWords))
	words := appendUniqueWords(nil, seen, existingWords)

	return &wordHistory{
		path:  path,
		words: words,
		seen:  seen,
	}
}

func defaultSources() []sourceConfig {
	return []sourceConfig{
		{name: "baidu", fetcher: hotspider.NewBaiduHot()},
		{name: "iqiyi", fetcher: hotspider.NewIQiyiHot()},
		{name: "mgtv", fetcher: hotspider.NewMGTVHot()},
		{name: "qq", fetcher: hotspider.NewQQHot()},
		{name: "quark", fetcher: hotspider.NewQuarkHot()},
		{name: "douban", fetcher: hotspider.NewDoubanHot()},
		{name: "youku", fetcher: hotspider.NewYoukuHot()},
	}
}

func buildTasks(sources []sourceConfig) []fetchTask {
	tasks := make([]fetchTask, 0, len(sources)*2)

	for _, source := range sources {
		categories := source.fetcher.SupportedCategories()
		if len(categories) == 0 {
			source := source
			tasks = append(tasks, fetchTask{
				label: source.name,
				fetch: func() ([]string, error) {
					return source.fetcher.Televisions()
				},
			})
			continue
		}

		for _, category := range categories {
			source := source
			category := category
			tasks = append(tasks, fetchTask{
				label: fmt.Sprintf("%s/%s", source.name, category),
				fetch: func() ([]string, error) {
					return source.fetcher.HotByCategory(category)
				},
			})
		}
	}

	return tasks
}

func collectWords(tasks []fetchTask, delay time.Duration, sleep func(time.Duration), flush func([]string) (int, error)) ([]error, error) {
	var fetchErrs []error

	for index, task := range tasks {
		log.Printf("fetching %s", task.label)

		words, err := task.fetch()
		if err != nil {
			fetchErrs = append(fetchErrs, fmt.Errorf("%s: %w", task.label, err))
		} else {
			added, flushErr := flush(words)
			if flushErr != nil {
				return fetchErrs, fmt.Errorf("flush words after %s: %w", task.label, flushErr)
			}

			log.Printf("fetched %d words from %s (%d new)", len(words), task.label, added)
		}

		if index < len(tasks)-1 && delay > 0 {
			sleep(delay)
		}
	}

	return fetchErrs, nil
}

func (h *wordHistory) addAndFlush(words []string) (int, error) {
	before := len(h.words)
	h.words = appendUniqueWords(h.words, h.seen, words)
	added := len(h.words) - before
	if added == 0 {
		return 0, nil
	}

	if err := writeWords(h.path, h.words); err != nil {
		return 0, err
	}

	h.newCount += added
	return added, nil
}

func mergeWords(existingWords, latestWords []string) ([]string, int) {
	merged := make([]string, 0, len(existingWords)+len(latestWords))
	seen := make(map[string]struct{}, len(existingWords)+len(latestWords))

	merged = appendUniqueWords(merged, seen, existingWords)
	before := len(merged)
	merged = appendUniqueWords(merged, seen, latestWords)

	return merged, len(merged) - before
}

func appendUniqueWords(dst []string, seen map[string]struct{}, words []string) []string {
	for _, word := range words {
		cleanWord := strings.TrimSpace(word)
		if cleanWord == "" {
			continue
		}
		if _, ok := seen[cleanWord]; ok {
			continue
		}

		seen[cleanWord] = struct{}{}
		dst = append(dst, cleanWord)
	}

	return dst
}

func readWords(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue
		}

		words = append(words, word)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return words, nil
}

func writeWords(path string, words []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content := strings.Join(words, "\n")
	if content != "" {
		content += "\n"
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

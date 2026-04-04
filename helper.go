package hotspider

import "regexp"

var charsRE = regexp.MustCompile(`[:：'"\s*!！,，·]`)

func removeChars(s string) string {
	return charsRE.ReplaceAllString(s, "")
}

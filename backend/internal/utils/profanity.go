package utils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	profaneWords     []string
	profaneWordsOnce sync.Once
	profaneRegex     *regexp.Regexp
)

// LoadProfanityList loads the profanity word list
func LoadProfanityList() error {
	var err error
	profaneWordsOnce.Do(func() {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, httpErr := client.Get("https://raw.githubusercontent.com/zautomz/bad-words/master/lib/lang.json")
		if httpErr != nil {
			// Fallback to basic list if download fails
			profaneWords = []string{
				"fuck", "shit", "bitch", "ass", "damn", "crap", "piss", "dick", "cock",
				"pussy", "bastard", "slut", "whore", "fag", "nigger", "cunt",
			}
			return
		}
		defer resp.Body.Close()

		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			profaneWords = []string{"fuck", "shit", "bitch", "damn"}
			return
		}

		var wordMap map[string][]string
		if jsonErr := json.Unmarshal(body, &wordMap); jsonErr == nil {
			if words, ok := wordMap["en"]; ok {
				profaneWords = words
			}
		}

		// Build regex pattern
		if len(profaneWords) > 0 {
			patterns := make([]string, len(profaneWords))
			for i, word := range profaneWords {
				// Escape special regex characters and add word boundaries
				patterns[i] = `\b` + regexp.QuoteMeta(word) + `\b`
			}
			pattern := `(?i)` + strings.Join(patterns, "|")
			profaneRegex = regexp.MustCompile(pattern)
		}
	})

	return err
}

// ContainsProfanity checks if text contains profanity
func ContainsProfanity(text string) bool {
	if profaneRegex == nil {
		LoadProfanityList()
	}
	if profaneRegex == nil {
		return false
	}
	return profaneRegex.MatchString(strings.ToLower(text))
}

// FilterProfanity replaces profane words with asterisks
func FilterProfanity(text string) string {
	if profaneRegex == nil {
		LoadProfanityList()
	}
	if profaneRegex == nil {
		return text
	}

	return profaneRegex.ReplaceAllStringFunc(text, func(match string) string {
		if len(match) <= 2 {
			return strings.Repeat("*", len(match))
		}
		return string(match[0]) + strings.Repeat("*", len(match)-2) + string(match[len(match)-1])
	})
}

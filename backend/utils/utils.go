package utils

import (
	"fmt"
	"strings"
	"time"
)

func GetTodayDate() string {
	t := time.Now()
	return t.Format("2006-01-02")
}

func WrapError(msg string, err error) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func RemoveStringFromSlice(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func SanitizeString(s string) string {
	// basic sanitization
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, ".", "")

	// remove trailing spaces
	s = strings.TrimSpace(s)

	return s
}

func InList(s string, list []string) bool {
	for _, l := range list {
		if l == s {
			return true
		}
	}
	return false
}

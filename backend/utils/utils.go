package utils

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Returns the "real" user IP from common proxy headers (or fallbackIp if none is found).
//
// The returned IP value shouldn't be trusted if not behind a trusted reverse proxy!
func RealUserIp(r *http.Request, fallbackIp string) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ipsList := r.Header.Get("X-Forwarded-For"); ipsList != "" {
		ips := strings.Split(ipsList, ",")
		// extract the rightmost ip
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip != "" {
				return ip
			}
		}
	}

	return fallbackIp
}

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

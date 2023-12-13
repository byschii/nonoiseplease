package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/routine"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cast"
)

func ActivityLoggerWithPostAndAuthSupport(app core.App) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			httpRequest := c.Request()
			httpResponse := c.Response()
			status := httpResponse.Status
			meta := types.JsonMap{}

			meta["Authorization"] = httpRequest.Header.Get("Authorization")
			// if not GET and body is not empty
			if httpRequest.Method != http.MethodGet && httpRequest.Method != http.MethodDelete {
				// read request body
				var body map[string]interface{}
				if err := c.Bind(&body); err != nil {
					log.Printf("failed to read request body, %v for req %s %s", err, httpRequest.Method, httpRequest.URL.RequestURI())
				}
				// convert body back to byte
				bodyBytes, err := json.Marshal(body)
				if err != nil {
					log.Printf("failed to marshal request body, %v for req %s %s", err, httpRequest.Method, httpRequest.URL.RequestURI())
				}
				httpRequest.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

				meta["body"] = body
			}

			// faccio tutte le chiamate
			err := next(c)

			// no logs retention
			if app.Settings().Logs.MaxDays == 0 {
				return err
			}

			if err != nil {
				switch v := err.(type) {
				case *echo.HTTPError:
					status = v.Code
					meta["errorMessage"] = v.Message
					meta["errorDetails"] = fmt.Sprint(v.Internal)
				case *apis.ApiError:
					status = v.Code
					meta["errorMessage"] = v.Message
					meta["errorDetails"] = fmt.Sprint(v.RawData())
				default:
					status = http.StatusBadRequest
					meta["errorMessage"] = v.Error()
				}
			}

			requestAuth := models.RequestAuthGuest
			if c.Get(apis.ContextAuthRecordKey) != nil {
				requestAuth = models.RequestAuthRecord
			} else if c.Get(apis.ContextAdminKey) != nil {
				requestAuth = models.RequestAuthAdmin
			}

			ip, _, _ := net.SplitHostPort(httpRequest.RemoteAddr)

			model := &models.Request{
				Url:       httpRequest.URL.RequestURI(),
				Method:    strings.ToLower(httpRequest.Method),
				Status:    status,
				Auth:      requestAuth,
				UserIp:    realUserIp(httpRequest, ip),
				RemoteIp:  ip,
				Referer:   httpRequest.Referer(),
				UserAgent: httpRequest.UserAgent(),
				Meta:      meta,
			}
			// set timestamp fields before firing a new go routine
			model.RefreshCreated()
			model.RefreshUpdated()

			routine.FireAndForget(func() {
				if err := app.LogsDao().SaveRequest(model); err != nil && app.IsDebug() {
					log.Println("Log save failed:", err)
				}

				// Delete old request logs
				// ---
				now := time.Now()
				lastLogsDeletedAt := cast.ToTime(app.Cache().Get("lastLogsDeletedAt"))
				daysDiff := now.Sub(lastLogsDeletedAt).Hours() * 24

				if daysDiff > float64(app.Settings().Logs.MaxDays) {
					deleteErr := app.LogsDao().DeleteOldRequests(now.AddDate(0, 0, -1*app.Settings().Logs.MaxDays))
					if deleteErr == nil {
						app.Cache().Set("lastLogsDeletedAt", now)
					} else if app.IsDebug() {
						log.Println("Logs delete failed:", deleteErr)
					}
				}
			})

			return err
		}
	}
}

// Returns the "real" user IP from common proxy headers (or fallbackIp if none is found).
//
// The returned IP value shouldn't be trusted if not behind a trusted reverse proxy!
func realUserIp(r *http.Request, fallbackIp string) string {
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

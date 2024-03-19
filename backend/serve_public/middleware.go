package servestatic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/routine"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cast"

	controllers "be/controllers"
	util "be/utils"
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
					log.Debug().Msgf("failed to read request body, %v for req %s %s", err, httpRequest.Method, httpRequest.URL.RequestURI())
				}
				// convert body back to byte
				bodyBytes, err := json.Marshal(body)
				if err != nil {
					log.Debug().Msgf("failed to marshal request body, %v for req %s %s", err, httpRequest.Method, httpRequest.URL.RequestURI())
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
				UserIp:    util.RealUserIp(httpRequest, ip),
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
					log.Error().Msgf("Log save failed: %v", err)
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
						log.Debug().Msgf("Logs delete failed: %v", deleteErr)
					}
				}
			})

			return err
		}
	}
}

// handels static files from 'fileSystem'
// if indexFallback is true, it will try to serve index.html if file not found
// dao and tokenSecret are used for templating
// also '.html' will be added to the end of name if it doesnt ends with '.html'
func StaticDirectoryHandlerWOptionalHTML(fileSystem fs.FS, indexFallback bool, uc controllers.UserControllerInterface, ac controllers.AuthControllerInterface) echo.HandlerFunc {
	return StaticDirectoryHandlerWHTMLAdder(fileSystem, indexFallback, true, uc, ac)
}

func StaticDirectoryHandlerWHTMLAdder(fileSystem fs.FS, indexFallback bool, autoAddHtml bool, uc controllers.UserControllerInterface, ac controllers.AuthControllerInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		p := c.PathParam("*")

		// escape url path
		tmpPath, err := url.PathUnescape(p)
		if err != nil {
			return fmt.Errorf("failed to unescape path variable: %w", err)
		}
		p = tmpPath

		// fs.FS.Open() already assumes that file names are relative to FS root path and considers name with prefix `/` as invalid
		name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))
		// if name doesnt ends with '.html' and autoAddHtml is true, add '.html' to the end of name
		if name != "." && autoAddHtml {
			name = completeName(fileSystem, name)
		}

		// parse and evaluate eventual template
		// array of strings
		for _, templatedPage := range getTemplatedPages(fileSystem) {
			// if static file is a "templated"
			if name == templatedPage.TemplateName {
				// get go template
				pageTemplate := templatedPage.ParsedTemplate
				// try to extract user from request
				user, err := ac.FindUserFromJWTInContext(c)
				data := interface{}(nil)
				// if no user found, use simple DataRetriever
				if err != nil {
					log.Error().Msgf("templating: no user found")
					data = templatedPage.DataRetriever(uc)
				} else { // if user found, use DataRetrieverWithUser
					log.Error().Msgf("templating: user found")
					data = templatedPage.DataRetrieverWithUser(uc, user.Id)
				}
				// build page with data and put it in response
				pageTemplate.Execute(c.Response().Writer, data)
				return nil
			}
		}

		// try to respond with file
		fileErr := c.FileFS(name, fileSystem)
		if fileErr != nil && indexFallback && errors.Is(fileErr, echo.ErrNotFound) {
			return c.FileFS("index.html", fileSystem)
		}

		return fileErr
	}
}

func completeName(fileSystem fs.FS, name string) string {
	if !strings.HasSuffix(name, ".html") {
		// search file
		file, err := fileSystem.Open(name)
		// check if file is directory
		if err == nil {
			defer file.Close()
			fileInfo, err := file.Stat()
			if err == nil && fileInfo.IsDir() {
				name += ".html"
			}
		} else {
			if errors.Is(err, fs.ErrNotExist) {
				// if file not found
				// add '.html' to the end of name
				name += ".html"
			}
		}
	}
	return name
}

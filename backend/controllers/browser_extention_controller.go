package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type BrowserExtentionController struct {
}

func AddEndpointBookmarkSyncFromExtention(
	app *pocketbase.PocketBase,
	e *core.ServeEvent,
	authController *AuthController,
	method string,
	path string) error {

	e.Router.AddRoute(echo.Route{
		Method: method,
		Path:   path,
		Handler: func(c echo.Context) error {
			pocketBaseEndpoint := c.Request().Host
			reqData := c.Request().Form
			if !reqData.Has("e") || !reqData.Has("p") || !reqData.Has("b") {
				return c.String(http.StatusBadRequest, "missing parameters")
			}
			authEndpoint := "http://" + pocketBaseEndpoint + "/api/collections/users/auth-with-password"
			err := authController.CheckAuthCredentials(reqData.Get("e"), reqData.Get("p"), authEndpoint)
			if err != nil {
				return c.String(http.StatusUnauthorized, err.Error())
			}
			stringifiedBookmarks := reqData.Get("b")
			if stringifiedBookmarks == "" {
				return c.String(http.StatusBadRequest, "no bookmarks provided")
			}
			bookmarks := map[string]interface{}{}
			err = json.Unmarshal([]byte(stringifiedBookmarks), &bookmarks)
			if err != nil {
				return c.String(http.StatusBadRequest, "invalid bookmarks")
			}
			fmt.Println(bookmarks) // TODO: save bookmarks by crawling them
			return c.String(http.StatusOK, "ok")

		},
		Middlewares: []echo.MiddlewareFunc{
			apis.ActivityLogger(app),
		},
	})
	return nil
}

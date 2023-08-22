package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func AddEndpointBookmarkSyncFromExtention(app *pocketbase.PocketBase, e *core.ServeEvent, method string, path string) error {

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
			err := CheckAuthCredentials(reqData.Get("e"), reqData.Get("p"), authEndpoint)
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
			fmt.Println(bookmarks)
			return c.String(http.StatusOK, "ok")

		},
		Middlewares: []echo.MiddlewareFunc{
			apis.ActivityLogger(app),
		},
	})
	return nil
}

func CheckAuthCredentials(email string, password string, endpoint string) error {

	authCheckRequest, err := http.NewRequest(
		http.MethodPost,
		"",
		bytes.NewReader([]byte(
			`{"identity": "`+email+`", "password": "`+password+`"}`)),
	)
	if err != nil {
		return errors.New("cannot check auth")
	}
	authResp, err := http.DefaultClient.Do(authCheckRequest)
	if err != nil {
		return errors.New("cannot check auth")
	}
	// parse response to json
	authRespJson := map[string]interface{}{}
	err = json.NewDecoder(authResp.Body).Decode(&authRespJson)
	if err != nil {
		return errors.New("error parsing auth response")
	}
	// check if auth was successful
	if authRespJson["verified"] != true {
		return errors.New("auth failed")
	}
	return nil
}

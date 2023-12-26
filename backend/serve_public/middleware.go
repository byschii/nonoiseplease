package servestatic

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"

	controllers "be/controllers"
)

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
			if err != nil && errors.Is(err, fs.ErrNotExist) {
				// if file not found
				// add '.html' to the end of name
				name += ".html"
			}
		}
	}
	return name
}

// handels static files from 'fileSystem'
// if indexFallback is true, it will try to serve index.html if file not found
// dao and tokenSecret are used for templating
// also '.html' will be added to the end of name if it doesnt ends with '.html'
func StaticDirectoryHandlerWOptionalHTML(fileSystem fs.FS, indexFallback bool, uc controllers.UserController, ac controllers.AuthController) echo.HandlerFunc {
	return StaticDirectoryHandlerWHTMLAdder(fileSystem, indexFallback, true, uc, ac)
}

func StaticDirectoryHandlerWHTMLAdder(fileSystem fs.FS, indexFallback bool, autoAddHtml bool, uc controllers.UserController, ac controllers.AuthController) echo.HandlerFunc {
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
					log.Print("templating: no user found")
					data = templatedPage.DataRetriever(uc)
				} else { // if user found, use DataRetrieverWithUser
					log.Print("templating: user found")
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

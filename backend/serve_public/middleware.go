package servestatic

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"

	controllers "be/controllers"
)

// handels static files from 'fileSystem'
// if indexFallback is true, it will try to serve index.html if file not found
// dao and tokenSecret are used for templating
// also '.html' will be added to the end of name if it doesnt ends with '.html'
func StaticDirectoryHandlerWOptionalHTML(
	fileSystem fs.FS,
	indexFallback bool,
	uc controllers.UserControllerInterface,
	ac controllers.AuthControllerInterface,
	stateController controllers.AppStateControllerInterface) echo.HandlerFunc {
	log.Debug().Msgf("StaticDirectoryHandlerWOptionalHTML %+v", fileSystem)
	return StaticDirectoryHandlerWHTMLAdder(fileSystem, indexFallback, true, uc, ac, stateController)
}

func StaticDirectoryHandlerWHTMLAdder(
	fileSystem fs.FS,
	indexFallback bool,
	autoAddHtml bool,
	uc controllers.UserControllerInterface, ac controllers.AuthControllerInterface, stateController controllers.AppStateControllerInterface) echo.HandlerFunc {
	return func(c echo.Context) error {
		p := c.PathParam("*")

		// escape url path
		tmpPath, err := url.PathUnescape(p)
		if err != nil {
			return fmt.Errorf("failed to unescape path variable: %w", err)
		}
		p = tmpPath

		log.Debug().Msgf("StaticDirectoryHandlerWHTMLAdder %+v", p)
		name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))
		// if name doesnt ends with '.html' and autoAddHtml is true, add '.html' to the end of name
		if name != "." && autoAddHtml {
			name = completeName(fileSystem, name)
		}

		// parse and evaluate eventual template
		// array of strings
		for _, templatedPage := range getTemplatedPages() {
			// if static file is a "templated"
			if name == templatedPage.TemplateName {
				// get go template
				// pageTemplate := templatedPage.ParsedTemplate
				// try to extract user from request
				user, err := ac.FindUserFromJWTInContext(c)
				data := interface{}(nil)
				// if no user found, use simple DataRetriever
				if err != nil {
					data = templatedPage.DataRetriever(uc, stateController)
					log.Debug().Msgf("templating: no user found %+v", data)
				} else { // if user found, use DataRetrieverWithUser
					data = templatedPage.DataRetrieverWithUser(uc, user.Id, stateController)
					log.Debug().Msgf("templating: user found %+v", data)

				}

				// build page with data and put it in response
				templatedPage.ParsedTemplate.Execute(c.Response().Writer, data)
				return nil
			}
		}

		// try to respond with file
		fileErr := c.FileFS(name, fileSystem)
		if fileErr != nil && errors.Is(fileErr, echo.ErrNotFound) {
			if indexFallback {
				return c.FileFS("index.html", fileSystem)
			} else {
				apis.NewNotFoundError("not found", "not found")
			}
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

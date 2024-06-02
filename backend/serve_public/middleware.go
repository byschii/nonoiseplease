package servestatic

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v5"

	controllers "be/controllers"
)

// handels static files from 'fileSystem'
// if indexFallback is true, it will try to serve index.html if file not found
// dao and tokenSecret are used for templating
// also '.html' will be added to the end of name if it doesnt ends with '.html'
func StaticDirectoryHandlerWOptionalHTML(
	fileSystem fs.FS,
	templateFileSystem fs.FS,
	templateParts []string,
	uc controllers.UserControllerInterface,
	ac controllers.AuthControllerInterface,
	stateController controllers.AppStateControllerInterface) echo.HandlerFunc {

	return StaticDirectoryHandlerWHTMLAdder(fileSystem, templateFileSystem, templateParts, uc, ac, stateController)
}

func StaticDirectoryHandlerWHTMLAdder(
	fileSystem fs.FS,
	templateFileSystem fs.FS,
	templateParts []string,
	uc controllers.UserControllerInterface, ac controllers.AuthControllerInterface, stateController controllers.AppStateControllerInterface) echo.HandlerFunc {
	return func(c echo.Context) error {

		// calculate re quested path, cause path can miss html
		name, err := getPath(c, fileSystem)
		if err != nil {
			return fmt.Errorf("failed to get path: %w", err)
		}
		if !strings.HasSuffix(name, ".html") {
			// insta serve file
			return c.FileFS(name, fileSystem)
		}

		// try to template every html
		templatePage, err := template.ParseFS(fileSystem, name)
		templatePage.ParseFiles(templateParts...)
		if err != nil {
			return fmt.Errorf("failed to parse template : %w", err)
		}

		// calc templated data
		data := getTemplateData(
			getTemplatedPages()[name],
			c,
			uc, ac, stateController)

		// execute template in saparate string
		var b strings.Builder
		err = templatePage.Execute(&b, data)
		if err != nil {
			return fmt.Errorf("failed to execute template : %w", err)
		}
		return c.HTML(http.StatusOK, b.String())
	}
}

func getPath(echoContext echo.Context, fileSystem fs.FS) (string, error) {
	p := echoContext.PathParam("*")

	// escape url path
	tmpPath, err := url.PathUnescape(p)
	if err != nil {
		return "", fmt.Errorf("failed to unescape path variable: %w", err)
	}
	p = tmpPath

	name := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "/")))
	// if name doesnt ends with '.html' add '.html' to the end of name
	if name != "." {
		name = completeName(fileSystem, name)
	}
	return name, nil
}

func getTemplateData(
	tmpl *TemplateRenderer,
	echoContext echo.Context,
	uc controllers.UserControllerInterface,
	ac controllers.AuthControllerInterface,
	stateController controllers.AppStateControllerInterface) interface{} {

	if tmpl == nil {
		return interface{}(nil)
	}

	user, err := ac.FindUserFromJWTInContext(echoContext)
	data := interface{}(nil)
	// if no user found, use simple DataRetriever
	if err != nil {
		data = tmpl.DataRetriever(uc, stateController)
		log.Debug().Msgf("templating: no user found %+v", data)
	} else { // if user found, use DataRetrieverWithUser
		data = tmpl.DataRetrieverWithUser(uc, user.Id, stateController)
		log.Debug().Msgf("templating: user found %+v", data)
	}

	return data
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

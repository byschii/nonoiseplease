package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "be/migrations"

	controllers "be/controllers"
	conf "be/model/config"
	rest "be/rest"
	servepublic "be/serve_public"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models/settings"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/viper"
)

/*func setResponseACAOHeaderFromRequest(req http.Request, resp echo.Response) {
	resp.Header().Set(echo.HeaderAccessControlAllowOrigin,
		req.Header.Get(echo.HeaderOrigin))
}

func ACAOHeaderOverwriteMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		ctx.Response().Before(func() {
			setResponseACAOHeaderFromRequest(*ctx.Request(), *ctx.Response())
		})
		return next(ctx)
	}
}*/

func main() {

	viper.AddConfigPath(".") // optionally look for config in the working directory
	viper.SetConfigFile("./config.yml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	log.Printf("config: %+v", viper.AllSettings())

	FRONTEND_FOLDER := viper.GetString("frontend_folder")
	LOG_FOLDER := viper.GetString("log_folder")
	LOG_FILENAME := viper.GetString("log_filename")
	APP_URL := viper.GetString("app_url")
	MEILI_MASTER_KEY := viper.GetString("meili_master_key")
	MAIL_USERNAME := viper.GetString("mail_username")
	MAIL_PASSWORD := viper.GetString("mail_password")
	MEILI_HOST_ADDRESS := viper.GetString("meili_host_address")
	MAX_SCRAPE_PER_MONTH := viper.GetInt("max_scrape_per_month")
	VERSION := "0.0.1"
	if os.Getenv("VERSION") != "" {
		VERSION = os.Getenv("VERSION")
	}
	if os.Getenv("RUNNING") == "PUBLIC" {
		APP_URL = "https://nonoiseplease.com"
		MEILI_HOST_ADDRESS = "http://0.0.0.0:7700"

		// set up logging
		log.SetFlags(log.LstdFlags | log.Lshortfile)

		// create log folder if not exists
		if _, err := os.Stat(LOG_FOLDER); os.IsNotExist(err) {
			os.Mkdir(LOG_FOLDER, 0755)
		}
		// open a file
		f, err := os.OpenFile(LOG_FOLDER+"/"+LOG_FILENAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		log.SetOutput(f)
	}

	app := pocketbase.New()
	meiliClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   MEILI_HOST_ADDRESS,
		APIKey: MEILI_MASTER_KEY,
	})

	migratecmd.MustRegister(app, app.RootCmd, &migratecmd.Options{
		Automigrate: true, // auto creates migration files when making collection changes
	})

	app.OnMailerBeforeRecordChangeEmailSend().Add(func(e *core.MailerRecordEvent) error {
		log.Println("sending mail for registration to", e.Record.Email())
		return nil
	})
	app.OnMailerAfterRecordChangeEmailSend().Add(func(e *core.MailerRecordEvent) error {
		log.Println("mail for registration sent to", e.Record.Email())
		return nil
	})

	c := controllers.NewConfigController(MAX_SCRAPE_PER_MONTH)

	authController := controllers.AuthController{
		App:         app,
		TokenSecret: app.Settings().RecordAuthToken.Secret,
	}
	userController := controllers.SimpleUserController{
		App:            app,
		MeiliClient:    meiliClient,
		AuthController: authController,
	}
	categoryController := controllers.CategoryController{
		PBDao: app.Dao(),
	}
	fulltextsearchController := controllers.FTSController{
		PBDao:       app.Dao(),
		MeiliClient: meiliClient,
	}
	pageController := controllers.PageController{
		PBDao:              app.Dao(),
		MeiliClient:        meiliClient,
		CategoryController: &categoryController,
		FTSController:      &fulltextsearchController,
	}

	appController := controllers.WebController{
		PageController:   pageController,
		UserController:   userController,
		ConfigController: c,
	}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		log.Println("lessgoozz!!")
		userController.App = app
		authController.App = app
		authController.TokenSecret = app.Settings().RecordAuthToken.Secret
		categoryController.PBDao = app.Dao()
		fulltextsearchController.PBDao = app.Dao()
		pageController.PBDao = app.Dao()

		// SETUP SERVER
		conf.InitConfig(app.Dao())

		e.Router.GET("/*", servepublic.StaticDirectoryHandlerWOptionalHTML(
			echo.MustSubFS(e.Router.Filesystem, FRONTEND_FOLDER),
			false,
			userController,
			authController),
		)

		middlewares := []echo.MiddlewareFunc{
			apis.RequireRecordAuth("users"),
			servepublic.ActivityLoggerWithPostAndAuthSupport(app),
		}

		middlewaresNoAuths := []echo.MiddlewareFunc{
			servepublic.ActivityLoggerWithPostAndAuthSupport(app),
		}

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/version",
			Handler: func(c echo.Context) error {
				return c.String(http.StatusOK, VERSION)
			},
			Middlewares: middlewaresNoAuths,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/url/scrape",
			Handler:     appController.PostUrlScrape,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodGet,
			Path:        "/api/page-manage",
			Handler:     rest.GetPagemanage(pageController),
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/page-manage/load",
			Handler:     rest.PostPagemanageLoad(pageController, authController, app.Settings().RecordAuthToken.Secret),
			Middlewares: middlewaresNoAuths,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/page-manage/category",
			Handler:     rest.PostPagemanageCategory(pageController),
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodDelete,
			Path:        "/api/page-manage/category",
			Handler:     rest.DeletePagemanageCategory(pageController),
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodDelete,
			Path:        "/api/page-manage/page",
			Handler:     rest.DeletePagemanagePage(pageController),
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodDelete,
			Path:        "/api/drop-account",
			Handler:     appController.DeleteAccount,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodGet,
			Path:        "/api/search/info",
			Handler:     appController.GetSearchInfo,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodGet,
			Path:        "/api/search",
			Handler:     appController.GetSearch,
			Middlewares: middlewares,
		})

		// SETUP CONFIG

		// set application smtp
		app.Settings().Smtp = settings.SmtpConfig{
			Enabled:  true,
			Host:     "mail.smtp2go.com",
			Port:     587,
			Username: MAIL_USERNAME,
			Password: MAIL_PASSWORD,
		}

		// set application email
		app.Settings().Meta.AppUrl = APP_URL
		app.Settings().Meta.AppName = "nonoiseplease"
		app.Settings().Meta.SenderName = "Support"
		app.Settings().Meta.SenderAddress = "support@nonoiseplease.com"

		// log retention
		app.Settings().Logs.MaxDays = 14

		return nil
	})

	app.OnRecordAuthRequest().Add(func(e *core.RecordAuthEvent) error {
		log.Println("authenticating user")
		removeToken := !e.Record.Verified() && conf.IsRequireMailVerification(app.Dao())
		if removeToken {
			log.Println(" user not verified, removing token")
			e.Token = ""
		} else {
			e.Record.SetVerified(true)
			err := app.Dao().SaveRecord(e.Record)
			if err != nil {
				log.Println("error saving record - " + err.Error())
				return err
			}
		}
		return nil
	})

	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "users" {
			if conf.IsGreatWallEnabled(app.Dao()) {
				log.Print("great wall enabled, aborting user creation")
				return fmt.Errorf("great wall enabled, aborting user creation")
			}
		}
		return nil
	})

	app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "users" {
			log.Print("creating user")
			err := userController.StoreNewUserDetails(e.Record.Id, "")
			if err != nil {
				log.Print("error creating user details or important data ")
				log.Println(err)
				return err
			}
			log.Print("StoreNewUserDetails ok")

			err = fulltextsearchController.CreateNewFTSIndex(e.Record.Id, 2)
			if err != nil {
				log.Println("error creating user index ", err)
				return err
			}
			log.Print("CreateNewFTSIndex ok")
		}

		return nil
	})

	app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "users" {
			log.Print("deleting user, reflect on fts")
			_, err := meiliClient.DeleteIndex(e.Record.Id)
			if err != nil {
				log.Println("error deleting user index ", err)
			}
		}
		return nil
	})

	app.OnRecordBeforeDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "pages" {
			err := fulltextsearchController.RemoveDocFTSIndex(e.Record.Id)
			if err != nil {
				log.Print("error deleting page ", err)
			}
		}
		return nil
	})
	app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "pages" {
			err := rest.AfterRemovePage(categoryController)
			if err != nil {
				log.Print("error deleting page ", err)
			}
		}
		return nil
	})

	app.OnRecordBeforeUpdateRequest().Add(func(e *core.RecordUpdateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "user_important_data" {
			log.Print("editing user_important_data")
		}

		return nil
	})

	err = app.Start()
	if err != nil {
		fmt.Print(err)
		panic(err)
	}

}

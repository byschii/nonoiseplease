package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "be/migrations"
	"be/pkg/jobs"

	controllers "be/controllers"
	categories "be/pkg/categories"
	conf "be/pkg/config"
	servepublic "be/serve_public"

	"github.com/labstack/echo/v5"
	"github.com/meilisearch/meilisearch-go"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models/settings"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/spf13/viper"
)

func main() {

	viper.AddConfigPath(".") // optionally look for config in the working directory
	viper.SetConfigFile("./config.yml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	print(fmt.Sprintf("config: %+v \n", viper.AllSettings()))

	FRONTEND_FOLDER := viper.GetString("frontend_folder")
	LOG_FOLDER := viper.GetString("log_folder")
	LOG_FILENAME := viper.GetString("log_filename")
	APP_URL := viper.GetString("app_url")
	MEILI_MASTER_KEY := viper.GetString("meili_master_key")
	MAIL_HOST := viper.GetString("mail_host")
	MAIL_PORT := viper.GetInt("mail_port")
	MAIL_USERNAME := viper.GetString("mail_username")
	MAIL_PASSWORD := viper.GetString("mail_password")
	MEILI_HOST_ADDRESS := viper.GetString("meili_host_address")
	LOG_MAX_DAYS := viper.GetInt("log_max_days")
	// get interface slice from config as 'default_config'
	// [ {}, {}... ]
	INITIAL_DB_CONFIGS := viper.Get("default_config").([]interface{})
	PROXIES := viper.Get("proxies").([]interface{})
	print(fmt.Sprintf("INITIAL_DB_CONFIGS: %+v \n", INITIAL_DB_CONFIGS))

	VERSION := "0.0.1"

	if os.Getenv("VERSION") != "" {
		VERSION = os.Getenv("VERSION")
	}

	logDestination := os.Stdout
	if os.Getenv("RUNNING") == "PUBLIC" {
		APP_URL = "https://nonoiseplease.com"
		MEILI_HOST_ADDRESS = "http://0.0.0.0:7700"

		// create log folder if not exists
		if _, err := os.Stat(LOG_FOLDER); os.IsNotExist(err) {
			print("creating log folder")
			os.Mkdir(LOG_FOLDER, 0755)
		} else {
			print("log folder exists")
		}
		// open a file
		logDestination, err = os.OpenFile(LOG_FOLDER+"/"+LOG_FILENAME, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			// format error and panic
			panic(fmt.Sprintf("error opening file: %v", err))
		}
	}
	log.Logger = zerolog.New(logDestination).With().Timestamp().Caller().Logger()

	app := pocketbase.New()

	meiliClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   MEILI_HOST_ADDRESS,
		APIKey: MEILI_MASTER_KEY,
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true, // auto creates migration files when making collection changes
	})

	app.OnMailerBeforeRecordChangeEmailSend().Add(func(e *core.MailerRecordEvent) error {
		log.Debug().Msgf("sending mail for registration to %s", e.Record.Email())
		return nil
	})
	app.OnMailerAfterRecordChangeEmailSend().Add(func(e *core.MailerRecordEvent) error {
		log.Debug().Msgf("mail for registration sent to %s", e.Record.Email())
		return nil
	})

	confController := controllers.NewConfigController(app.Dao())
	authController := controllers.NewAuthController(app)
	userController := controllers.NewUserController(app, meiliClient, authController)
	categoryController := controllers.NewCategoryController(app.Dao())
	fulltextsearchController := controllers.NewFTSController(app.Dao(), meiliClient)
	pageController := controllers.NewPageController(app.Dao(), meiliClient, categoryController, fulltextsearchController)
	appController := controllers.NewNoNoiseInterface(pageController, userController, confController)

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()

		scheduler.MustAdd("ScrapeBufferedPages", "* * * * *", func() {
			_ = jobs.ScrapeBufferedPages(app.Dao(), meiliClient)
		})
		scheduler.Start()
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		log.Debug().Msgf("lessgoozz!!")
		userController.SetApp(app)
		authController.SetApp(app)
		confController.SetPBDAO(app.Dao())
		categoryController.SetPBDAO(app.Dao())
		fulltextsearchController.SetPBDAO(app.Dao())
		pageController.SetPBDAO(app.Dao())

		// SETUP SERVER
		conf.InitConfigFromYaml(app.Dao(), INITIAL_DB_CONFIGS, PROXIES)

		e.Router.GET("/*", servepublic.StaticDirectoryHandlerWOptionalHTML(
			echo.MustSubFS(e.Router.Filesystem, FRONTEND_FOLDER),
			false,
			userController,
			authController,
			confController),
		)

		middlewares := []echo.MiddlewareFunc{
			apis.RequireRecordAuth("users"),
		}

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/version",
			Handler: func(c echo.Context) error {
				return c.String(http.StatusOK, VERSION)
			},
		})

		// pageManage := e.Router.Group("/api/page-manage", middlewares...)

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/url/scrape",
			Handler:     appController.PostUrlScrape,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodGet,
			Path:        "/api/page-manage",
			Handler:     appController.GetPagemanage,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/page-manage/load",
			Handler:     appController.PostPagemanageLoad,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/page-manage/category",
			Handler:     appController.PostPagemanageCategory,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodDelete,
			Path:        "/api/page-manage/category",
			Handler:     appController.DeletePagemanageCategory,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodDelete,
			Path:        "/api/page-manage/page",
			Handler:     appController.DeletePagemanagePage,
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

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodGet,
			Path:        "/api/search/html",
			Handler:     appController.GetSearchExtentionHtml,
			Middlewares: middlewares,
		})

		e.Router.AddRoute(echo.Route{
			Method:      http.MethodPost,
			Path:        "/api/bookmark/scrape",
			Handler:     appController.PostBookmarkScrape,
			Middlewares: middlewares,
		})

		// SETUP CONFIG

		// set application smtp
		app.Settings().Smtp = settings.SmtpConfig{
			Enabled:  true,
			Host:     MAIL_HOST,
			Port:     MAIL_PORT,
			Username: MAIL_USERNAME,
			Password: MAIL_PASSWORD,
		}

		// set application email
		app.Settings().Meta.AppUrl = APP_URL
		app.Settings().Meta.AppName = "nonoiseplease"
		app.Settings().Meta.SenderName = "Support"
		app.Settings().Meta.SenderAddress = "support@nonoiseplease.com"

		// log retention
		app.Settings().Logs.MaxDays = LOG_MAX_DAYS

		return nil
	})

	app.OnRecordAuthRequest().Add(func(e *core.RecordAuthEvent) error {
		log.Debug().Msgf("authenticating user")
		removeToken := !e.Record.Verified() && confController.IsRequireMailVerification()
		if removeToken {
			log.Debug().Msgf(" user not verified, removing token")
			e.Token = ""
		} else {
			e.Record.SetVerified(true)            // suppose we verified the user (cause RequireMailVerification false )
			err := app.Dao().SaveRecord(e.Record) // then save it
			if err != nil {
				log.Debug().Msgf("error saving record - " + err.Error())
				return err
			}
		}
		return nil
	})

	app.OnRecordBeforeCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "users" {
			if confController.IsGreatWallEnabled() {
				log.Debug().Msgf("great wall enabled, aborting user creation")
				return fmt.Errorf("great wall enabled, aborting user creation")
			}
		}
		return nil
	})

	app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "users" {
			log.Debug().Msgf("creating user")
			err := userController.StoreNewUserDetails(e.Record.Id, "")
			if err != nil {
				log.Debug().Msg("error creating user details or important data ")
				log.Error().Err(err)
				return err
			}
			log.Debug().Msgf("StoreNewUserDetails ok")

			err = fulltextsearchController.CreateNewFTSIndex(e.Record.Id, 2)
			if err != nil {
				log.Debug().Msg("error creating user index ")
				log.Error().Err(err)
				return err
			}
			log.Debug().Msgf("CreateNewFTSIndex ok")
		}

		return nil
	})

	app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "users" {
			log.Debug().Msgf("deleting user, reflect on fts")
			_, err := meiliClient.DeleteIndex(e.Record.Id)
			if err != nil {
				log.Debug().Msg("error deleting user index ")
				log.Error().Err(err)
			}
		}
		return nil
	})

	app.OnRecordBeforeDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "pages" {
			err := fulltextsearchController.RemoveDocFTSIndex(e.Record.Id)
			if err != nil {
				log.Debug().Msg("error deleting page ")
				log.Error().Err(err)
			}
		}
		return nil
	})
	app.OnRecordAfterDeleteRequest().Add(func(e *core.RecordDeleteEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "pages" {
			// get all categories
			categories, err := categories.GetAllCategories(categoryController.DAO())
			if err != nil {
				log.Error().Msgf("failed to get all categories, %v\n", err)
				return err
			}

			// delete orphan categories
			for _, category := range categories {
				err = categoryController.RemoveOrphanCategory(&category)
				if err != nil {
					log.Error().Msgf("failed to delete orphan category, %v\n", err)
					return err
				}
			}
		}
		return nil
	})

	app.OnRecordBeforeUpdateRequest().Add(func(e *core.RecordUpdateEvent) error {
		collectionName := e.Record.Collection().Name
		if collectionName == "user_important_data" {
			log.Debug().Msgf("editing user_important_data")
		}

		return nil
	})

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		// quickly create a file names as the signal
		os.Create(fmt.Sprintf("./tmp/%s", s.String()))
	}()

	err = app.Start()
	if err != nil {
		fmt.Print(err)
		panic(err)
	}

}

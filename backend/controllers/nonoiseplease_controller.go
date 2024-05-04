package controllers

import (
	"github.com/labstack/echo/v5"
)

type NoNoiseInterface interface {
	GetSearch(c echo.Context) error
	GetSearchExtentionHtml(c echo.Context) error
	GetSearchInfo(c echo.Context) error
	DeleteAccount(c echo.Context) error
	PostUrlScrape(c echo.Context) error
	PostPagemanageCategory(c echo.Context) error
	GetPagemanage(c echo.Context) error
	DeletePagemanageCategory(c echo.Context) error
	PostPagemanageLoad(c echo.Context) error
	DeletePagemanagePage(c echo.Context) error
	PostBookmarkUpload(c echo.Context) error
}

func NewNoNoiseInterface(pageController PageControllerInterface, userController UserControllerInterface, state AppStateControllerInterface) NoNoiseInterface {
	return &WebController{
		PageController:   pageController,
		UserController:   userController,
		ConfigController: state,
	}
}

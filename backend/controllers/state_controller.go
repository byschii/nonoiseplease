package controllers

import (
	"be/model/config"
	users "be/model/users"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/rs/zerolog/log"
)

type AppStateController struct {
	PBDao             *daos.Dao
	maxScrapePerMonth int
}

type AppStateControllerInterface interface {
	CommonController
	SetPBDAO(dao *daos.Dao)
	UserList() ([]users.Users, error)
	MaxScrapePerMonth() int
	IsGreatWallEnabled() bool
	IsRequireMailVerification() bool
	GetConfigUseProxyProbability() float32
}

func NewConfigController(dao *daos.Dao, maxScrapePerMonth int) AppStateControllerInterface {
	return &AppStateController{
		PBDao:             dao,
		maxScrapePerMonth: maxScrapePerMonth,
	}
}

// AppDao implements AppStateControllerInterface.
func (c AppStateController) AppDao() *daos.Dao {
	return c.PBDao
}

// SetPBDAO implements AppStateControllerInterface.
func (c *AppStateController) SetPBDAO(dao *daos.Dao) {
	c.PBDao = dao
}

// if any error
// return 0 and please use default value, do not handle error
func (c AppStateController) GetConfigUseProxyProbability() float32 {
	config, err := c.getConfigByKey(config.UseProxyProb)
	if err != nil || !config.BooleanValue {
		return 0
	}

	value := config.FloatValue
	return value
}

func (c AppStateController) UserList() ([]users.Users, error) {
	var u []users.Users
	err := c.AppDao().ModelQuery(&users.Users{}).
		All(&u)

	return u, err
}

func (c AppStateController) MaxScrapePerMonth() int {
	return c.maxScrapePerMonth
}

func (c AppStateController) IsGreatWallEnabled() bool {
	return c.getConfigGreatWallEnabled()
}

func (c AppStateController) IsRequireMailVerification() bool {
	return c.getConfigMailVerificationRequired()
}

// if any error, return true
func (c AppStateController) getConfigMailVerificationRequired() bool {
	config, err := c.getConfigByKey(config.MailVerificationRequired)
	if err != nil {
		log.Error().Err(err).Msg("get mail verification required config error")
		return true
	}

	value := config.BooleanValue
	return value
}

// if any error, return false
// great wall is used to block new user actions (like register, ecc)
func (c AppStateController) getConfigGreatWallEnabled() bool {
	config, err := c.getConfigByKey(config.GreatWallEnabled)
	if err != nil {
		log.Error().Err(err).Msg("get great wall enabled config error")
		return false
	}

	value := config.BooleanValue
	return value
}

func (c AppStateController) getConfigByKey(key config.AvailableConfig) (config.Config, error) {
	var conf config.Config
	err := c.AppDao().ModelQuery(&config.Config{}).
		AndWhere(dbx.HashExp{"key": key}).
		One(&conf)

	return conf, err
}

package config

import (
	"math/rand"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/rs/zerolog/log"
)

func GetRandomProxy(dao *daos.Dao) (ProxyConnection, error) {
	var proxy ProxyConnection
	err := dao.ModelQuery(&ProxyConnection{}).
		AndWhere(dbx.HashExp{"enabled": true}).
		OrderBy("RANDOM()").
		One(&proxy)
	return proxy, err
}

func GetProxyByAddress(dao *daos.Dao, address string) (ProxyConnection, error) {
	var proxy ProxyConnection
	err := dao.ModelQuery(&ProxyConnection{}).
		AndWhere(dbx.HashExp{"address": address}).
		One(&proxy)
	return proxy, err
}

func getConfigByKey(dao *daos.Dao, key AvailableConfig) (Config, error) {
	var config Config
	err := dao.ModelQuery(&Config{}).
		AndWhere(dbx.HashExp{"key": key}).
		One(&config)

	return config, err
}

// if any error, return true
func IsMailVerificationRequired(dao *daos.Dao) bool {
	config, err := getConfigByKey(dao, MailVerificationRequired)
	if err != nil {
		log.Error().Err(err).Msg("get mail verification required config error")
		return true
	}

	value := config.BooleanValue
	return value
}

func CountMaxScrapePerMonth(dao *daos.Dao) int {
	config, err := getConfigByKey(dao, MaxScrapePerMonth)
	if err != nil {
		log.Error().Err(err).Msg("get max scrape per month config error")
		return 0
	}

	value := int(config.FloatValue)
	return value
}

func CleanProxy(dao *daos.Dao) error {
	// get all proxies
	var proxy []ProxyConnection
	err := dao.ModelQuery(&ProxyConnection{}).
		All(&proxy)
	if err != nil {
		return err
	}

	// delete all proxies
	for _, p := range proxy {
		dao.Delete(&p)
	}

	return err
}

// if any error, return false
// great wall is used to block new user actions (like register, ecc)
func IsGreatWallEnabled(dao *daos.Dao) bool {
	config, err := getConfigByKey(dao, GreatWallEnabled)
	if err != nil {
		log.Error().Err(err).Msg("get great wall enabled config error")
		return false
	}

	value := config.BooleanValue
	return value
}

// return 0 and please use default value, do not handle error
func UseProxy(dao *daos.Dao) bool {
	return rand.Float32() < ProxyProb(dao)
}

// return 0 and please use default value, do not handle error
func ProxyProb(dao *daos.Dao) float32 {
	config, err := getConfigByKey(dao, UseProxyProb)
	if err != nil || !config.BooleanValue {
		return 0
	}

	return config.FloatValue

}

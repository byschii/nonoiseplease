package config

import (
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

func configQuery(dao *daos.Dao) *dbx.SelectQuery {
	return dao.ModelQuery(&Config{})
}

func getConfigByKey(dao *daos.Dao, key AvailableConfig) (Config, error) {
	var config Config
	err := configQuery(dao).
		AndWhere(dbx.HashExp{"key": key}).
		One(&config)

	return config, err
}

// if any error
// return 0 and please use default value, do not handle error
func GetConfigUseProxyProbability(dao *daos.Dao) float32 {
	config, err := getConfigByKey(dao, UseProxyProb)
	if err != nil || !config.BooleanValue {
		return 0
	}

	value := config.FloatValue
	return value
}

// if any error, return true
func getConfigMailVerificationRequired(dao *daos.Dao) bool {
	config, err := getConfigByKey(dao, MailVerificationRequired)
	if err != nil {
		return true
	}

	value := config.BooleanValue
	return value
}
func IsRequireMailVerification(dao *daos.Dao) bool {
	return getConfigMailVerificationRequired(dao)
}

// if any error, return false
// great wall is used to block new user actions (like register, ecc)
func getConfigGreatWallEnabled(dao *daos.Dao) bool {
	config, err := getConfigByKey(dao, GreatWallEnabled)
	if err != nil {
		return false
	}

	value := config.BooleanValue
	return value
}
func IsGreatWallEnabled(dao *daos.Dao) bool {
	return getConfigGreatWallEnabled(dao)
}

func InitConfigFromYaml(dao *daos.Dao, configMap []interface{}, proxyMap []interface{}) error {
	log.Debug().Msg("init config from yaml")

	for _, config := range configMap {
		// every config is a map "string" -> any
		configEntity := Config{
			Key:          config.(map[string]interface{})["key"].(string),
			TextValue:    config.(map[string]interface{})["text_value"].(string),
			FloatValue:   float32(config.(map[string]interface{})["float_value"].(float64)),
			BooleanValue: config.(map[string]interface{})["boolean_value"].(bool),
			Note:         config.(map[string]interface{})["note"].(string),
		}
		log.Debug().Msgf("config: %p -> %+v", &configEntity, configEntity)

		_, err := getConfigByKey(dao, AvailableConfig(configEntity.Key))
		if err != nil {
			err := dao.Save(&configEntity)
			if err != nil {
				return err
			}
		}
	}

	for _, proxy := range proxyMap {
		// every proxy is a map "string" -> any (address->string, port->int)
		proxyEntity := ProxyConnection{
			Enabled: true,
			Address: proxy.(map[string]interface{})["address"].(string),
			Port:    int(proxy.(map[string]interface{})["port"].(int)),
		}
		log.Debug().Msgf("proxy: %p -> %+v", &proxyEntity, proxyEntity)

		_, err := GetProxyByAddress(dao, proxyEntity.Address)
		if err != nil {
			err := dao.Save(&proxyEntity)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

// init config if not exists
func InitConfig(dao *daos.Dao) error {
	// use proxy prob
	_, err := getConfigByKey(dao, UseProxyProb)
	if err != nil {
		config := Config{
			Key:          string(UseProxyProb),
			TextValue:    "",
			FloatValue:   0.0,
			BooleanValue: false,
			Note:         "use proxy prob",
		}
		err = dao.Save(&config)
		if err != nil {
			return err
		}
	}

	// mail verification required
	_, err = getConfigByKey(dao, MailVerificationRequired)
	if err != nil {
		config := Config{
			Key:          string(MailVerificationRequired),
			TextValue:    "",
			FloatValue:   0.0,
			BooleanValue: true,
			Note:         "mail verification required",
		}
		err = dao.Save(&config)
		if err != nil {
			return err
		}
	}

	// great wall enabled
	_, err = getConfigByKey(dao, GreatWallEnabled)
	if err != nil {
		config := Config{
			Key:          string(GreatWallEnabled),
			TextValue:    "",
			FloatValue:   0.0,
			BooleanValue: false,
			Note:         "great wall enabled",
		}
		err = dao.Save(&config)
		if err != nil {
			return err
		}
	}

	return nil
}

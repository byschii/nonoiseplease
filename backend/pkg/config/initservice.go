package config

import (
	"github.com/pocketbase/pocketbase/daos"
	"github.com/rs/zerolog/log"
)

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

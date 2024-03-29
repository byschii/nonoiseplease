package config

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

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

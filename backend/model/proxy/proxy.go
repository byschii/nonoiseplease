package proxy

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
)

func GetRandomProxy(dao *daos.Dao) (ProxyConnection, error) {
	var proxy ProxyConnection
	err := dao.ModelQuery(&ProxyConnection{}).
		AndWhere(dbx.HashExp{"enabled": true}).
		OrderBy("RANDOM()").
		One(&proxy)
	return proxy, err
}

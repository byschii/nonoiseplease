package config

import (
	"github.com/pocketbase/pocketbase/models"
)

// https://scrapingant.com/free-proxies/
type ProxyConnection struct {
	models.BaseModel

	Enabled bool   `db:"enabled" json:"enabled"`
	Address string `db:"address" json:"address"`
	Port    int    `db:"port" json:"port"`
}

func (p *ProxyConnection) TableName() string {
	return "proxy_connections"
}

var _ models.Model = (*ProxyConnection)(nil)

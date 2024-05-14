package config

import (
	"github.com/pocketbase/pocketbase/models"
)

// activity to category mapping
type AvailableConfig string

const (
	UseProxyProb             AvailableConfig = "USE_PROXY_PROB"
	MailVerificationRequired AvailableConfig = "MAIL_VERIFICATION_REQUIRED"
	GreatWallEnabled         AvailableConfig = "GREAT_WALL_ENABLED"
	MaxScrapePerMonth        AvailableConfig = "MAX_SCRAPE_PER_MONTH"
)

// https://scrapingant.com/free-proxies/
type Config struct {
	models.BaseModel

	Key          string  `db:"key" json:"key"`
	TextValue    string  `db:"text_value" json:"text_value"`
	FloatValue   float32 `db:"float_value" json:"float_value"`
	BooleanValue bool    `db:"boolean_value" json:"boolean_value"`
	Note         string  `db:"note" json:"note"`
}

func (p *Config) TableName() string {
	return "config"
}

var _ models.Model = (*Config)(nil)

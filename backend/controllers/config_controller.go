package controllers

type ConfigController struct {
	maxScrapePerMonth int
}

func NewConfigController(maxScrapePerMonth int) ConfigController {
	return ConfigController{
		maxScrapePerMonth: maxScrapePerMonth,
	}
}

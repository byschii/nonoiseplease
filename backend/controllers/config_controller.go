package controllers

type ConfigController struct {
	maxScrapePerMonth int
}

type ConfigControllerInterface interface {
	MaxScrapePerMonth() int
}

func NewConfigController(maxScrapePerMonth int) ConfigControllerInterface {
	return ConfigController{
		maxScrapePerMonth: maxScrapePerMonth,
	}
}

func (c ConfigController) MaxScrapePerMonth() int {
	return c.maxScrapePerMonth
}

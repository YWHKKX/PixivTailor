package crawler

import (
	"os"
	"path/filepath"
	"time"
)

type Mode string

const (
	MODE_SAFE Mode = "safe"
	MODE_R18  Mode = "r18"
	MODE_ALL  Mode = "all"
)

type Order string

const (
	ORDER_POPULAR_D Order = "popular_d"
	ORDER_DATE_D    Order = "date_d"
)

type config struct {
	tag                   string
	order                 Order
	mode                  Mode
	cookie, agent, accept string
	savePath              string
	delay                 time.Duration
}

func InitConfig(tag string, order Order, mode Mode, savePaths ...string) config {
	currentPath, _ := os.Getwd()
	savePath := filepath.Join(currentPath, "images")
	if len(savePaths) > 0 {
		savePath = savePaths[0]
	}
	return config{
		tag:      tag,
		order:    order,
		mode:     mode,
		savePath: savePath,
	}
}

func (c *config) GetTag() string {
	return c.tag
}

func (c *config) SetTag(tag string) {
	c.tag = tag
}

func (c *config) GetCookie() string {
	return c.cookie
}

func (c *config) SetCookie(cookie string) {
	c.cookie = cookie
}

func (c *config) GetAgent() string {
	return c.agent
}

func (c *config) SetAgent(agent string) {
	c.agent = agent
}

func (c *config) GetAccept() string {
	return c.accept
}

func (c *config) SetAccept(accept string) {
	c.accept = accept
}

func (c *config) GetSavePath() string {
	return c.savePath
}

func (c *config) SetSavePath(savePath string) {
	c.savePath = savePath
}

func (c *config) GetDelay() time.Duration {
	return c.delay
}

func (c *config) SetDelay(delay time.Duration) {
	c.delay = delay
}

package crawler

import (
	"os"
	"path/filepath"
	"time"
)

// pixiv api文档: https://github.com/daydreamer-json/pixiv-ajax-api-docs

type ConfigType int

const (
	SEARCH_BY_TAG ConfigType = iota
	SEARCH_BY_USER
)

type Mode string
type Order string

const (
	MODE_SAFE Mode = "safe"
	MODE_R18  Mode = "r18"
	MODE_ALL  Mode = "all"
)

const (
	ORDER_POPULAR_D Order = "popular_d"
	ORDER_DATE_D    Order = "date_d"
)

type config struct {
	configType            ConfigType
	tag                   string
	user                  int
	order                 Order
	mode                  Mode
	cookie, agent, accept string
	savePath              string
	delay                 time.Duration
}

// https://www.pixiv.net/ajax/search/artworks/{tag}?word={tag}&order={order}&mode={mode}&p=1&s_mode=s_tag_full&type={mode}&lang=zh
func InitTagConfig(tag string, order Order, mode Mode, savePaths ...string) config {
	currentPath, _ := os.Getwd()
	savePath := filepath.Join(currentPath, "images")
	if len(savePaths) > 0 {
		savePath = savePaths[0]
	}
	return config{
		configType: SEARCH_BY_TAG,
		tag:        tag,
		order:      order,
		mode:       mode,
		savePath:   savePath,
	}
}

// https://www.pixiv.net/ajax/user/{userid}/works/latest
func InitUserConfig(user int, savePaths ...string) config {
	currentPath, _ := os.Getwd()
	savePath := filepath.Join(currentPath, "images")
	if len(savePaths) > 0 {
		savePath = savePaths[0]
	}
	return config{
		configType: SEARCH_BY_USER,
		user:       user,
		savePath:   savePath,
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

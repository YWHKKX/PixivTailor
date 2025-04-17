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

type searchConfig struct {
	tag   string
	user  int
	order Order
	mode  Mode
}

type requestConfig struct {
	cookie, agent, accept string
}

type downloadConfig struct {
	savePath string
	saveName string
	delay    time.Duration
	limit    int
}

type config struct {
	configType     ConfigType
	searchConfig   searchConfig
	requestConfig  requestConfig
	downloadConfig downloadConfig
}

// https://www.pixiv.net/ajax/search/artworks/{tag}?word={tag}&order={order}&mode={mode}&p=1&s_mode=s_tag_full&type={mode}&lang=zh
func InitTagConfig(tag string, order Order, mode Mode, basePaths ...string) config {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	return config{
		configType: SEARCH_BY_TAG,
		searchConfig: searchConfig{
			tag:   tag,
			order: order,
			mode:  mode,
		},
		downloadConfig: downloadConfig{
			savePath: filepath.Join(basePath, "images"),
		},
	}
}

// https://www.pixiv.net/ajax/user/{userid}/works/latest
func InitUserConfig(user int, basePaths ...string) config {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	return config{
		configType: SEARCH_BY_USER,
		searchConfig: searchConfig{
			user: user,
		},
		downloadConfig: downloadConfig{
			savePath: filepath.Join(basePath, "images"),
		},
	}
}

func (c *config) GetTag() string {
	return c.searchConfig.tag
}

func (c *config) SetTag(tag string) {
	c.searchConfig.tag = tag
}

func (c *config) GetOrder() Order {
	return c.searchConfig.order
}

func (c *config) GetMode() Mode {
	return c.searchConfig.mode
}
func (c *config) getUser() int {
	return c.searchConfig.user
}

func (c *config) GetCookie() string {
	return c.requestConfig.cookie
}

func (c *config) SetCookie(cookie string) {
	c.requestConfig.cookie = cookie
}

func (c *config) GetAgent() string {
	return c.requestConfig.agent
}

func (c *config) SetAgent(agent string) {
	c.requestConfig.agent = agent
}

func (c *config) GetAccept() string {
	return c.requestConfig.accept
}

func (c *config) SetAccept(accept string) {
	c.requestConfig.accept = accept
}

func (c *config) GetSavePath() string {
	return c.downloadConfig.savePath
}

func (c *config) SetSavePath(savePath string) {
	c.downloadConfig.savePath = savePath
}

func (c *config) GetSaveName() string {
	return c.downloadConfig.saveName
}

func (c *config) SetSaveName(saveName string) {
	c.downloadConfig.saveName = saveName
}

func (c *config) GetLimit() int {
	return c.downloadConfig.limit
}

func (c *config) SetLimit(limit int) {
	c.downloadConfig.limit = limit
}

func (c *config) CheckLimit(index int) bool {
	if index >= c.GetLimit() && c.GetLimit() != 0 {
		return false
	}
	return true
}

func (c *config) GetDelay() time.Duration {
	return c.downloadConfig.delay
}

func (c *config) SetDelay(delay time.Duration) {
	c.downloadConfig.delay = delay
}

func (c *config) GetConfigType() ConfigType {
	return c.configType
}

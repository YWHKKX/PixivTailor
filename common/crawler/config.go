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
	SEARCH_BY_Illust
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
	tag    string
	user   int
	illust int
	order  Order
	mode   Mode
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

type CrawlerConfig struct {
	configType     ConfigType
	searchConfig   searchConfig
	requestConfig  requestConfig
	downloadConfig downloadConfig
}

// https://www.pixiv.net/ajax/search/artworks/{tag}?word={tag}&order={order}&mode={mode}&p=1&s_mode=s_tag_full&type={mode}&lang=zh
func InitTagConfig(tag string, order Order, mode Mode, basePaths ...string) *CrawlerConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	return &CrawlerConfig{
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
func InitUserConfig(user int, basePaths ...string) *CrawlerConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	return &CrawlerConfig{
		configType: SEARCH_BY_USER,
		searchConfig: searchConfig{
			user: user,
		},
		downloadConfig: downloadConfig{
			savePath: filepath.Join(basePath, "images"),
		},
	}
}

// https://www.pixiv.net/ajax/illust/{illustid}/pages
func InitIllustConfig(illust int, basePaths ...string) *CrawlerConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	return &CrawlerConfig{
		configType: SEARCH_BY_Illust,
		searchConfig: searchConfig{
			illust: illust,
		},
		downloadConfig: downloadConfig{
			savePath: filepath.Join(basePath, "images"),
		},
	}
}

func (c *CrawlerConfig) GetTag() string {
	return c.searchConfig.tag
}

func (c *CrawlerConfig) SetTag(tag string) {
	c.searchConfig.tag = tag
}

func (c *CrawlerConfig) GetOrder() Order {
	return c.searchConfig.order
}

func (c *CrawlerConfig) GetMode() Mode {
	return c.searchConfig.mode
}
func (c *CrawlerConfig) GetUser() int {
	return c.searchConfig.user
}

func (c *CrawlerConfig) GetIllust() int {
	return c.searchConfig.illust
}

func (c *CrawlerConfig) GetCookie() string {
	return c.requestConfig.cookie
}

func (c *CrawlerConfig) SetCookie(cookie string) {
	c.requestConfig.cookie = cookie
}

func (c *CrawlerConfig) GetAgent() string {
	return c.requestConfig.agent
}

func (c *CrawlerConfig) SetAgent(agent string) {
	c.requestConfig.agent = agent
}

func (c *CrawlerConfig) GetAccept() string {
	return c.requestConfig.accept
}

func (c *CrawlerConfig) SetAccept(accept string) {
	c.requestConfig.accept = accept
}

func (c *CrawlerConfig) GetSavePath() string {
	return c.downloadConfig.savePath
}

func (c *CrawlerConfig) SetSavePath(savePath string) {
	c.downloadConfig.savePath = savePath
}

func (c *CrawlerConfig) GetSaveName() string {
	return c.downloadConfig.saveName
}

func (c *CrawlerConfig) SetSaveName(saveName string) {
	c.downloadConfig.saveName = saveName
}

func (c *CrawlerConfig) GetLimit() int {
	return c.downloadConfig.limit
}

func (c *CrawlerConfig) SetLimit(limit int) {
	c.downloadConfig.limit = limit
}

func (c *CrawlerConfig) CheckLimit(index int) bool {
	if index >= c.GetLimit() && c.GetLimit() != 0 {
		return true
	}
	return false
}

func (c *CrawlerConfig) GetDelay() time.Duration {
	return c.downloadConfig.delay
}

func (c *CrawlerConfig) SetDelay(delay time.Duration) {
	c.downloadConfig.delay = delay
}

func (c *CrawlerConfig) GetConfigType() ConfigType {
	return c.configType
}

package main

import (
	"time"

	"github.com/GolangProject/PixivCrawler/common/crawler"
)

func do_crawler() {
	cookie := ""
	agent := ""
	accept := ""

	index := 0

	var config crawler.CrawlerConfig
	switch index {
	case 0:
		config = crawler.InitTagConfig("", crawler.ORDER_DATE_D, crawler.MODE_SAFE)
		config.SetCookie(cookie)
		config.SetAgent(agent)
		config.SetAccept(accept)
		config.SetDelay(2 * time.Second)
		config.SetLimit(10)
	case 1:
		config = crawler.InitUserConfig(0)
		config.SetCookie(cookie)
		config.SetAgent(agent)
		config.SetAccept(accept)
		config.SetDelay(2 * time.Second)
		config.SetLimit(10)
	case 2:
		config = crawler.InitIllustConfig(0)
		config.SetCookie(cookie)
		config.SetAgent(agent)
		config.SetAccept(accept)
		config.SetDelay(2 * time.Second)
	}

	crawler := crawler.NewCrawler(config)
	crawler.Run()
}

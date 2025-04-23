package crawler

import (
	"github.com/GolangProject/PixivCrawler/common/utils"
)

type Crawler struct {
	*CrawlerConfig
}

func NewCrawler(config *CrawlerConfig) *Crawler {
	return &Crawler{CrawlerConfig: config}
}

func (c *Crawler) Run() {
	utils.Info("Start crawler url")
	mangas := Collector(c.CrawlerConfig)

	for _, manga := range mangas {
		for i, url := range manga.urls {
			ok := Download(manga.id, url, i, c.CrawlerConfig)
			if !ok {
				utils.Errorf("Download error: %s", url)
				continue
			}
		}
	}
}

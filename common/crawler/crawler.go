package crawler

import (
	"github.com/GolangProject/PixivCrawler/common/utils"
)

type Crawler struct {
	config
}

func NewCrawler(config config) *Crawler {
	return &Crawler{config: config}
}

func (c *Crawler) Run() {
	utils.Info("Start crawler url")
	mangas := Collector(c.config)

	for _, manga := range mangas {
		for i, url := range manga.urls {
			ok := Download(manga.id, url, i, c.config)
			if !ok {
				utils.Errorf("Download error: %s", url)
				continue
			}
		}
	}
}

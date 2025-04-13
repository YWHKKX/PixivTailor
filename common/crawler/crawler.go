package crawler

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type Crawler struct {
	config
}

func NewCrawler(config config) *Crawler {
	return &Crawler{config: config}
}

func (c *Crawler) Run() {
	utils.Info("start crawler url")
	mangas := Collector(c.config)

	for _, manga := range mangas {
		index := 0
		re := regexp.MustCompile(`_p(\d)_`)
		matches := re.FindStringSubmatch(manga.url)
		if len(matches) > 1 {
			index, _ = strconv.Atoi(matches[1])
		}
		Download(manga.id, manga.url, index, c.config, false)

		for i := index + 1; ; i++ {
			newURL := re.ReplaceAllString(manga.url, fmt.Sprintf("_p%d_", i))
			if !Download(manga.id, newURL, i, c.config, true) {
				break
			}
		}

	}
}

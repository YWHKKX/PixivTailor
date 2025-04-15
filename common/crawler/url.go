package crawler

import (
	"net/http"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

var Global_Root string = "https://www.pixiv.net"
var Global_Search_Tag string = "https://www.pixiv.net/ajax/search/artworks"
var Global_User string = "https://www.pixiv.net/ajax/user"
var Global_Illust string = "https://www.pixiv.net/ajax/illust"
var Global_Artworks string = "https://www.pixiv.net/artworks"

func makeRequest(target, cookie, agent, accept string) *http.Request {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		utils.Errorf("http.NewRequest error: %v")
	}
	req.Header = http.Header{
		"Accept":          []string{accept},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6"},
		"Cookie":          []string{cookie},
		"User-Agent":      []string{agent},
		"Referer":         []string{"https://www.pixiv.net/"},
	}
	return req
}

package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type Response struct {
	Error bool `json:"error"`
	Body  struct {
		IllustManga struct {
			Data []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
				URL   string `json:"url"`
			} `json:"data"`
		} `json:"illustManga"`
	} `json:"body"`
}

type Manga struct {
	title string
	id    string
	url   string
}

func NewManga(title, id, url string) *Manga {
	return &Manga{
		title: title,
		id:    id,
		url:   url,
	}
}

func GetList(input string, isDebug bool) (ret []*Manga) {
	var resp Response
	err := json.Unmarshal([]byte(input), &resp)
	if err != nil {
		panic(err)
	}

	for _, d := range resp.Body.IllustManga.Data {
		ret = append(ret, NewManga(d.Title, d.ID, d.URL))
	}
	if isDebug {
		utils.Infof("GetList len: %d", len(ret))
		for _, i := range ret {
			utils.Infof("{%s %s %s}", i.id, i.title, fmt.Sprintf("%s/%s", Global_Artworks, i.id))
		}
	}
	return
}

// https://www.pixiv.net/ajax/search/artworks/{tag}?word={tag}&order={order}&mode={mode}&p=1&s_mode=s_tag_full&type={mode}&lang=zh
func Collector(config config) []*Manga {
	target := fmt.Sprintf("%s/%s?word=%s&order=%s&mode=%s&p=1&s_mode=s_tag_full&type=%s&lang=zh",
		Global_Search, config.tag, config.tag, config.order, config.mode, config.mode)
	req := makeRequest(target, config.cookie, config.agent, config.accept)

	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	utils.Infof("Collector artworks from: %s", target)
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		utils.Errorf("Request error(%d): %v", err, res.StatusCode)
		return []*Manga{}
	}
	if res.Header.Get("X-Userid") == "" {
		utils.Warn("Cookie invalid, please update")
	}
	defer res.Body.Close()

	bytes := decodeZip(res)
	utf8Body := decodeToUTF8(res, bytes)
	fmt.Println(string(utf8Body))

	return GetList(string(utf8Body), true)
}

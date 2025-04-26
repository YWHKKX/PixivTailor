package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type TagResponse struct {
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

type UserResponse struct {
	Error bool `json:"error"`
	Body  struct {
		Illusts map[string]struct {
			ID    string   `json:"id"`
			Title string   `json:"title"`
			URL   string   `json:"url"`
			Tags  []string `json:"tags"`
		} `json:"illusts"`
	} `json:"body"`
}

type ImageResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    []struct {
		URLs struct {
			ThumbMini string `json:"thumb_mini"`
			Small     string `json:"small"`
			Regular   string `json:"regular"`
			Original  string `json:"original"`
		} `json:"urls"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"body"`
}

type Manga struct {
	title string
	id    string
	urls  []string
}

func NewManga(title, id string, url []string) *Manga {
	return &Manga{
		title: title,
		id:    id,
		urls:  url,
	}
}

func GetImageUrl(id string, config *CrawlerConfig) []string {
	var ret []string

	target := fmt.Sprintf("%s/%s/pages", Global_Illust, id)
	req := makeRequest(target, config.GetCookie(), config.GetAgent(), config.GetAccept())

	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	res, err := client.Do(req)
	if err != nil {
		utils.Errorf("Request error: %v", err)
		return []string{}
	}
	if res.StatusCode != http.StatusOK {
		utils.Errorf("Response StatusCode: %d", res.StatusCode)
		return []string{}
	}
	defer res.Body.Close()

	var resp ImageResponse
	err = json.Unmarshal([]byte(decodeToUTF8(res, decodeZip(res))), &resp)
	if err != nil {
		utils.Errorf("Function Unmarshal Error: %v", err)
		return []string{}
	}

	for _, d := range resp.Body {
		ret = append(ret, d.URLs.Original)
	}

	return ret
}

func GetMangaIllus(config *CrawlerConfig, isDebug bool) (ret []*Manga) {
	id := fmt.Sprintf("%d", config.GetIllust())
	urls := GetImageUrl(id, config)
	ret = append(ret, NewManga("", id, urls))
	if isDebug {
		utils.Infof("{%s\t%s\t%s}", id, "", fmt.Sprintf("%s/%s", Global_Artworks, id))
	}

	if isDebug {
		utils.Infof("GetList len: %d", len(ret))
	}
	return
}

func GetMangaTag(input string, config *CrawlerConfig, isDebug bool) (ret []*Manga) {
	var resp TagResponse
	err := json.Unmarshal([]byte(input), &resp)
	if err != nil {
		utils.Errorf("Function Unmarshal Error: %v", err)
		return
	}

	for index, d := range resp.Body.IllustManga.Data {
		if config.CheckLimit(index) {
			utils.Warnf("Limit reached: %d", config.GetLimit())
			break
		}

		urls := GetImageUrl(d.ID, config)
		ret = append(ret, NewManga(d.Title, d.ID, urls))
		if isDebug {
			utils.Infof("{%s\t%s\t%s}", d.ID, d.Title, fmt.Sprintf("%s/%s", Global_Artworks, d.ID))
		}
	}
	if isDebug {
		utils.Infof("GetList len: %d", len(ret))
	}
	return
}

func GetMangaUser(input string, config *CrawlerConfig, isDebug bool) (ret []*Manga) {
	var resp UserResponse
	err := json.Unmarshal([]byte(input), &resp)
	if err != nil {
		utils.Errorf("Function Unmarshal Error: %v", err)
		return
	}

	index := 0
	for i, d := range resp.Body.Illusts {
		if config.CheckLimit(index) {
			utils.Warnf("Limit reached: %d", config.GetLimit())
			break
		}
		index++

		urls := GetImageUrl(i, config)
		ret = append(ret, NewManga(d.Title, i, urls))
		if isDebug {
			title := d.Title
			if title == "" {
				title = "No Title"
			}
			utils.Infof("{%s\t%s\t%s}", i, title, fmt.Sprintf("%s/%s", Global_Artworks, i))
		}
	}
	if isDebug {
		utils.Infof("GetList len: %d", len(ret))
	}
	return
}

func Collector(config *CrawlerConfig) []*Manga {
	var ret []*Manga

	getBody := func(target string) []byte {
		req := makeRequest(target, config.GetCookie(), config.GetAgent(), config.GetAccept())

		proxyURL, _ := url.Parse("http://127.0.0.1:7890")
		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}

		utils.Infof("Collector artworks from: %s", target)
		res, err := client.Do(req)
		if err != nil {
			utils.Errorf("Request error: %v", err)
			return nil
		}
		if res.StatusCode != http.StatusOK {
			utils.Errorf("Response StatusCode: %d", res.StatusCode)
			return nil
		}
		defer res.Body.Close()

		if res.Header.Get("X-Userid") == "" {
			utils.Warn("Cookie invalid, please update")
		}

		bytes := decodeZip(res)
		utf8Body := decodeToUTF8(res, bytes)
		return utf8Body
	}

	switch config.GetConfigType() {
	case SEARCH_BY_TAG:
		target := fmt.Sprintf("%s/%s?word=%s&order=%s&mode=%s&p=1&s_mode=s_tag_full&type=%s&lang=zh",
			Global_Search_Tag, config.GetTag(), config.GetTag(), config.GetTag(), config.GetMode(), config.GetMode())
		ret = GetMangaTag(string(getBody(target)), config, true)
	case SEARCH_BY_USER:
		target := fmt.Sprintf("%s/%d/works/latest",
			Global_User, config.GetUser())
		ret = GetMangaUser(string(getBody(target)), config, true)
	case SEARCH_BY_Illust:
		ret = GetMangaIllus(config, true)
	default:

	}

	return ret
}

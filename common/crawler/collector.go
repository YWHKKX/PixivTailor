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

// https://www.pixiv.net/ajax/illust/{userid}/pages
func GetImageUrl(id string, config config) []string {
	var ret []string

	target := fmt.Sprintf("%s/%s/pages", Global_Illust, id)
	req := makeRequest(target, config.cookie, config.agent, config.accept)

	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		utils.Errorf("Request error(%d): %v", err, res.StatusCode)
		return []string{}
	}
	defer res.Body.Close()

	var resp ImageResponse
	err = json.Unmarshal([]byte(decodeToUTF8(res, decodeZip(res))), &resp)
	if err != nil {
		utils.Error(err)
	}

	for _, d := range resp.Body {
		ret = append(ret, d.URLs.Original)
	}

	return ret
}

func GetMangasTag(input string, config config, isDebug bool) (ret []*Manga) {
	var resp TagResponse
	err := json.Unmarshal([]byte(input), &resp)
	if err != nil {
		utils.Error(err)
		return
	}

	for _, d := range resp.Body.IllustManga.Data {
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

func GetMangaUser(input string, config config, isDebug bool) (ret []*Manga) {
	var resp UserResponse
	err := json.Unmarshal([]byte(input), &resp)
	if err != nil {
		utils.Error(err)
		return
	}

	for i, d := range resp.Body.Illusts {
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

func Collector(config config) []*Manga {
	var ret []*Manga

	getBody := func(target string) []byte {
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
			return nil
		}
		if res.Header.Get("X-Userid") == "" {
			utils.Warn("Cookie invalid, please update")
		}
		defer res.Body.Close()

		bytes := decodeZip(res)
		utf8Body := decodeToUTF8(res, bytes)
		return utf8Body
	}

	switch config.configType {
	case SEARCH_BY_TAG:
		target := fmt.Sprintf("%s/%s?word=%s&order=%s&mode=%s&p=1&s_mode=s_tag_full&type=%s&lang=zh",
			Global_Search_Tag, config.tag, config.tag, config.order, config.mode, config.mode)
		ret = GetMangasTag(string(getBody(target)), config, true)
	case SEARCH_BY_USER:
		target := fmt.Sprintf("%s/%d/works/latest",
			Global_User, config.user)
		ret = GetMangaUser(string(getBody(target)), config, true)
	default:

	}

	return ret
}

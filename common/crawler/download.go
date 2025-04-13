package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

// https://i.pximg.net/img-master/img/{time}/{illustID}_p{index}_master1200.jpg
func Download(id, target string, index int, config config, try bool) bool {
	req := makeRequest(target, config.cookie, config.agent, config.accept)

	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	savePath := filepath.Join(config.GetSavePath(), fmt.Sprintf("%s_%d.jpg", id, index))
	if _, err := os.Stat(savePath); err == nil {
		utils.Warnf("file already exists: %s", savePath)
		return true
	}

	if try {
		utils.Infof("Try download artworks: %s", target)
	} else {
		utils.Infof("Start download artworks: %s", target)
	}

	if delay := config.GetDelay(); delay > 0 {
		time.Sleep(delay)
	}

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if try {
			return false
		}
		utils.Errorf("Request error(%d): %v", err, res.StatusCode)
		return false
	}
	defer res.Body.Close()

	file, _ := os.Create(savePath)
	defer file.Close()
	_, err = io.Copy(file, res.Body)

	return true
}

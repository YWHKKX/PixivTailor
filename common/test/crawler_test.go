package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GolangProject/PixivCrawler/common/crawler"
)

func Test_SearchTag(t *testing.T) {
	cookie := ""
	agent := ""
	accept := ""

	currntPath, _ := os.Getwd()
	basePath := filepath.Join(currntPath, "./../..")
	config := crawler.InitTagConfig("Rem", crawler.ORDER_DATE_D, crawler.MODE_SAFE, basePath)
	config.SetCookie(cookie)
	config.SetAgent(agent)
	config.SetAccept(accept)
	config.SetDelay(2 * time.Second)
	config.SetLimit(10)

	crawler := crawler.NewCrawler(config)
	crawler.Run()
}

func Test_SearchUser(t *testing.T) {
	cookie := ""
	agent := ""
	accept := ""

	currntPath, _ := os.Getwd()
	basePath := filepath.Join(currntPath, "./../..")
	config := crawler.InitUserConfig(0, basePath)
	config.SetCookie(cookie)
	config.SetAgent(agent)
	config.SetAccept(accept)
	config.SetDelay(2 * time.Second)
	config.SetLimit(10)

	crawler := crawler.NewCrawler(config)
	crawler.Run()
}

func Test_SearchtIllust(t *testing.T) {
	cookie := ""
	agent := ""
	accept := ""

	currntPath, _ := os.Getwd()
	basePath := filepath.Join(currntPath, "./../..")
	config := crawler.InitIllustConfig(0, basePath)
	config.SetCookie(cookie)
	config.SetAgent(agent)
	config.SetAccept(accept)
	config.SetDelay(2 * time.Second)

	crawler := crawler.NewCrawler(config)
	crawler.Run()
}

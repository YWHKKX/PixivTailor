package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type ProgressWriter struct {
	total   int64
	written int64
}

func NewProgressWriter(total int64) *ProgressWriter {
	return &ProgressWriter{
		total:   total,
		written: 0,
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.written += int64(n)

	percent := float64(pw.written) / float64(pw.total) * 100

	fmt.Printf("\rDownloading: %.2f%% [%s%s] %s/%s\n",
		percent,
		strings.Repeat("=", int(percent/2)),
		strings.Repeat(" ", 50-int(percent/2)),
		formatBytes(pw.written),
		formatBytes(pw.total),
	)

	return n, nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// https://i.pximg.net/img-master/img/{time}/{illustID}_p{index}_master1200.jpg
func Download(id, target string, index int, config CrawlerConfig) bool {
	req := makeRequest(target, config.GetCookie(), config.GetAgent(), config.GetAccept())

	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	getPath := func(id string, index int) (string, string) {
		saveImageFile := fmt.Sprintf("image_%s_%d.jpg", id, index)
		saveImageDir := ""

		if saveName := config.GetSaveName(); saveName != "" {
			saveImageDir = filepath.Join(config.GetSavePath(), saveName)
			return saveImageFile, saveImageDir
		}

		switch config.GetConfigType() {
		case SEARCH_BY_TAG:
			saveImageDir = filepath.Join(config.GetSavePath(), config.GetTag())
			if _, err := os.Stat(saveImageFile); err == nil {
				utils.Warnf("File already exists: %s", saveImageFile)
				return "", ""
			}
		case SEARCH_BY_USER:
			saveImageDir = filepath.Join(config.GetSavePath(), id)
			if _, err := os.Stat(saveImageFile); err == nil {
				utils.Warnf("File already exists: %s", saveImageFile)
				return "", ""
			}
		case SEARCH_BY_Illust:
			saveImageDir = filepath.Join(config.GetSavePath(), id)
			if _, err := os.Stat(saveImageFile); err == nil {
				utils.Warnf("File already exists: %s", saveImageFile)
				return "", ""
			}
		default:

		}
		return saveImageFile, saveImageDir
	}

	saveImageFile, saveImageDir := getPath(id, index)

	utils.Infof("Try to download artworks: %s", target)
	if delay := config.GetDelay(); delay > 0 {
		time.Sleep(delay)
	}

	res, err := client.Do(req)
	if err != nil {
		utils.Errorf("Request error: %v", err)
		return false
	}
	if res.StatusCode != http.StatusOK {
		utils.Errorf("Request StatusCode: %d", res.StatusCode)
		return false
	}
	defer res.Body.Close()

	fileSize := res.ContentLength
	if fileSize <= 0 {
		utils.Errorf("File size <= 0, Possibly a request error")
		return false
	}

	if err := os.MkdirAll(saveImageDir, os.ModePerm); err != nil {
		utils.Errorf("Function os.MkdirAll error: %v", err)
		return false
	}

	file, err := os.Create(filepath.Join(saveImageDir, saveImageFile))
	if err != nil {
		utils.Errorf("Function os.Create error: %v", err)
		return false
	}
	defer file.Close()

	progress := NewProgressWriter(fileSize)
	_, err = io.Copy(file, io.TeeReader(res.Body, progress))
	if err != nil {
		utils.Errorf("Function io.Copy error: %v", err)
		return false
	}

	return true
}

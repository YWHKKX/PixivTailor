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

type progressWriter struct {
	total   int64
	written int64
}

func (pw *progressWriter) Write(p []byte) (int, error) {
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
func Download(id, target string, index int, config config) bool {
	req := makeRequest(target, config.GetCookie(), config.GetAgent(), config.GetAccept())

	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	var savePath, saveImagePath string

	switch config.GetConfigType() {
	case SEARCH_BY_TAG:
		savePath = filepath.Join(config.GetSavePath(), config.GetTag())
		saveImagePath = filepath.Join(savePath, fmt.Sprintf("image_%s_%d.jpg", id, index))
		if _, err := os.Stat(saveImagePath); err == nil {
			utils.Warnf("File already exists: %s", saveImagePath)
			return true
		}
	case SEARCH_BY_USER:
		savePath = filepath.Join(config.GetSavePath(), id)
		saveImagePath = filepath.Join(savePath, fmt.Sprintf("image_%d.jpg", index))
		if _, err := os.Stat(saveImagePath); err == nil {
			utils.Warnf("File already exists: %s", saveImagePath)
			return true
		}
	default:

	}

	utils.Infof("Try download artworks: %s", target)
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
	progress := &progressWriter{
		total:   fileSize,
		written: 0,
	}

	if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
		utils.Errorf("Function os.MkdirAll error: %v", err)
		return false
	}

	file, err := os.Create(saveImagePath)
	if err != nil {
		utils.Errorf("Function os.Create error: %v", err)
		return false
	}
	defer file.Close()

	_, err = io.Copy(file, io.TeeReader(res.Body, progress))
	if err != nil {
		utils.Errorf("Function io.Copy error: %v", err)
		return false
	}

	return true
}

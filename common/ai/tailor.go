package ai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type SDResponse struct {
	Images [][]byte `json:"images"`
}

func Tailor_TXT2IMG(config config) {
	inputPath := config.GetInputPath()
	outputPath := config.GetOutputPath()
	files, _ := filepath.Glob(filepath.Join(inputPath, "*", "*.json"))

	DownloadImage := func(path string, index int, content []byte) {
		file, err := os.Create(fmt.Sprintf("%s_%d.jpg", path, index))
		if err != nil {
			utils.Errorf("Function os.Create error: %v", err)
			return
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		chunkSize := 1024 * 1024

		for i := 0; i < len(content); i += chunkSize {
			end := i + chunkSize
			if end > len(content) {
				end = len(content)
			}
			if _, err = writer.Write(content[i:end]); err != nil {
				utils.Errorf("Function writer.Write error: %v", err)
			}
		}
		writer.Flush()
	}

	for _, file := range files {
		var responses DeepDanbooruResponses
		var SDResponse SDResponse

		utils.Infof("Start read tags: %s", file)
		fileData, err := os.ReadFile(file)
		if err != nil {
			utils.Errorf("Read File Error: %v", err)
			continue
		}

		if err := json.Unmarshal(fileData, &responses); err != nil {
			utils.Error(err)
			continue
		}

		tagString := responses.TagString
		tagNum := responses.TagNum
		if tagString == "" || tagNum == 0 {
			continue
		}

		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		saveDir := filepath.Join(outputPath, config.GetSaveName())
		saveFile := filepath.Join(saveDir, tagName)
		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			utils.Errorf("Function os.MkdirAll error: %v", err)
			return
		}

		utils.Infof("Try to request image for: %s,\tlen(tags) = %d", tagName, tagNum)
		res := makeSDRequest(SD_API_TXT2IMG, tagString, config)
		if res == nil {
			return
		}
		defer res.Body.Close()

		body, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(body, &SDResponse)
		if err != nil {
			utils.Error(err)
		}

		for i, image := range SDResponse.Images {
			utils.Infof("Try to download new image: %s", fmt.Sprintf("%s_%d.jpg", tagName, i))
			DownloadImage(saveFile, i, image)
		}
	}

}

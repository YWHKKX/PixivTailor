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

func CreateImage_TXT2IMG(config ImageConfig) {
	inputPath := config.GetInputPath()
	outputPath := config.GetOutputPath()

	files, _ := filepath.Glob(filepath.Join(inputPath, "*", "*."+string(config.GetSaveType())))

	DownloadImage := func(path string, index int, content []byte) bool {
		savePath := fmt.Sprintf("%s_%d.jpg", path, index)
		if _, err := os.Stat(savePath); err == nil {
			utils.Warnf("File already exists: %s", savePath)
			return false
		}

		file, err := os.Create(savePath)
		if err != nil {
			utils.Errorf("Function os.Create error: %v", err)
			return false
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

		return true
	}

	index := 0
	tmpTag := ""
	for _, file := range files {
		if !config.CheckPathFilter(file) {
			continue
		}

		var responses DeepDanbooruResponses
		var SDResponse SDResponse

		utils.Infof("Start read tags: %s", file)
		fileData, err := os.ReadFile(file)
		if err != nil {
			utils.Errorf("Read File Error: %v", err)
			continue
		}

		tagString := ""
		tagNum := 0
		switch config.GetSaveType() {
		case Save_Json:
			if err := json.Unmarshal(fileData, &responses); err != nil {
				utils.Error(err)
				panic(err)
			}
			tagString = responses.TagString
			tagNum = responses.TagNum
		case Save_Txt:
			tagString = string(fileData)
			tagNum = strings.Count(string(fileData), ",")
		}

		if tagString == "" || tagNum == 0 {
			continue
		}

		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		if tagName != tmpTag {
			index = 0
			tmpTag = tagName
		}
		saveDir := filepath.Join(outputPath, config.GetSavePathName())
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

		ignoreNum := 0
		for _, poseConfig := range config.GetPoseConfigs() {
			ignoreNum += poseConfig.MemberNum/2 - 1
		}
		utils.Infof("The number of bone diagrams: %d", ignoreNum)

		for i, image := range SDResponse.Images {
			if i >= len(SDResponse.Images)-ignoreNum {
				utils.Info("Skip the bone diagram")
				continue
			}
			retry := true
			for retry {
				utils.Infof("Try to download new image: %s", fmt.Sprintf("%s_%d.jpg", tagName, index))
				retry = !DownloadImage(saveFile, index, image)
				index++
			}
		}
	}

}

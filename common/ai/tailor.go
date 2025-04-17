package ai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type SDResponse struct {
	Images [][]byte `json:"images"`
}

func Tailor_TXT2IMG(config imageConfig) {
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

func Tailor_TrainModel(config trainConfig) {
	savePath, ok := buildModelConfig(config)
	if savePath == "" || !ok {
		return
	}

	if ok = buildTrainingSet(config); !ok {
		return
	}

	cmd := exec.Command("python",
		filepath.Join(config.GetBasePath(), "scripts", "train_model.py"),
		filepath.Join(savePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Errorf("Python script execution error: %s", err)
		return
	}
	utils.Infof("Output: %s", string(output))
}

func buildTrainingSet(config trainConfig) bool {
	for _, c := range config.GetTagConfigs() {
		srcDir := c.GetTagSrcPath()
		destDir := filepath.Join(config.GetInputDir(), fmt.Sprintf("%d_%s", c.GetTimes(), c.GetTagName()))

		copyFile := func(src, dst string) error {
			srcFile, err := os.Open(src)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			return err
		}

		index := 0
		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if !config.CheckLimit(index) {
					return nil
				}
				index++

				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".jpg" || ext == ".png" || ext == ".jpeg" || ext == ".gif" {
					relPath, _ := filepath.Rel(srcDir, path)
					destPath := filepath.Join(destDir, relPath)

					if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
						return err
					}

					if err := copyFile(path, destPath); err != nil {
						utils.Errorf("Copy file error from: %s to: %s", path, destPath)
						return err
					}
					utils.Infof("Copy file success from: %s to: %s", path, destPath)
				}
			}
			return nil
		})

		if !config.CheckLimit(index) {
			utils.Infof("Limit reached: %d", config.GetLimit())
		}

		if err != nil {
			utils.Errorf("Failed to traverse the folder: %v\n", err)
			return false
		}
	}

	return true
}

func buildModelConfig(config trainConfig) (string, bool) {
	var trainModelCofig map[string]interface{}

	fileData, err := os.ReadFile(config.examplePath)
	if err != nil {
		utils.Errorf("Read File Error: %v", err)
		return "", false
	}

	if err := json.Unmarshal(fileData, &trainModelCofig); err != nil {
		utils.Error(err)
		return "", false
	}

	utils.Infof("Start build model config, model name: %s", config.GetModelName())

	trainModelCofig["output_name"] = config.GetModelName()
	trainModelCofig["pretrained_model_name_or_path"] = config.GetPretrainedPath()
	trainModelCofig["train_data_dir"] = config.GetInputDir()
	trainModelCofig["output_dir"] = config.GetOutputDir()
	trainModelCofig["logging_dir"] = config.GetLogDir()

	newData, err := json.MarshalIndent(trainModelCofig, "", "  ")
	if err != nil {
		utils.Error(err)
		return "", false
	}

	if err := os.MkdirAll(config.GetInputDir(), os.ModePerm); err != nil {
		utils.Errorf("Function os.MkdirAll error: %v", err)
		return "", false
	}

	savePath := filepath.Join(config.GetInputDir(), "config.json")
	file, err := os.Create(savePath)
	if err != nil {
		utils.Errorf("Function os.Create error: %v", err)
		return "", false
	}
	defer file.Close()

	_, err = file.Write(newData)
	if err != nil {
		utils.Errorf("Function file.Write error: %v", err)
		return "", false
	}

	utils.Infof("Save new model config: %s\n%s", savePath, string(newData))
	return savePath, true
}

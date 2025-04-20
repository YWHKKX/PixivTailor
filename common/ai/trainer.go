package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

func TrainModel(config TrainConfig) {
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

func buildTrainingSet(config TrainConfig) bool {
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

				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".jpg" {
					index++
				}
				if ext == ".jpg" || ext == ".txt" {
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

func buildModelConfig(config TrainConfig) (string, bool) {
	var trainModelCofig map[string]interface{}

	examplePath := filepath.Join(config.GetBasePath(), "scripts", "example.json")
	fileData, err := os.ReadFile(examplePath)
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
	for _, t := range config.GetTagConfigs() {
		trainModelCofig["sample_prompts"] = t.GetTagName()
	}

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

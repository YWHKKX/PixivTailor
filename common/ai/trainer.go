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

func TrainModel(config *TrainConfig) {
	if err := os.RemoveAll(config.GetInputDir()); err != nil {
		utils.Errorf("Function os.RemoveAll error: %v", err)
	}

	if err := os.MkdirAll(config.GetInputDir(), os.ModePerm); err != nil {
		utils.Errorf("Function os.MkdirAll error: %v", err)
		return
	}

	if ok := buildTrainingSet(config); !ok {
		return
	}

	savePath, ok := buildModelConfig(config)
	if savePath == "" || !ok {
		return
	}

	cmd := exec.Command("python",
		filepath.Join(config.GetBasePath(), "scripts", "train_model.py"),
		filepath.Join(savePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Errorf("Python script execution error: %s", err)
		utils.Errorf("Output: %s", string(output))
		return
	}
	utils.Infof("Output: %s", string(output))
}

func buildTrainingSet(config *TrainConfig) bool {
	for _, c := range config.GetTagConfigs() {
		srcDir := c.GetTagSrcPath()
		destDir := filepath.Join(config.GetInputDir(), fmt.Sprintf("%d_%s", c.GetTimes(), c.GetTagName()))

		utils.Infof("Start setting the %s training set", c.GetTagName())

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
		utils.Infof("TagName %s, total of %d images were read", c.GetTagName(), index)
		config.UpTrainImageNum(index)

		if err != nil {
			utils.Errorf("Failed to traverse the folder: %v\n", err)
			return false
		}
	}

	return true
}

func buildModelConfig(config *TrainConfig) (string, bool) {
	var trainModelCofig map[string]interface{}

	examplePath := filepath.Join(config.GetBasePath(), "scripts", "example.json")
	fileData, err := os.ReadFile(examplePath)
	if err != nil {
		utils.Errorf("Read File Error: %v", err)
		return "", false
	}

	if err := json.Unmarshal(fileData, &trainModelCofig); err != nil {
		utils.Errorf("Function Unmarshal Error: %v", err)
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

	switch config.GetTrainType() {
	case TrainSpeedFast:
		trainModelCofig["train_batch_size"] = 4
		trainModelCofig["gradient_accumulation_steps"] = 1
		trainModelCofig["network_alpha"] = 16
		trainModelCofig["network_dim"] = 32
	case TrainSpeedSlow:
		trainModelCofig["train_batch_size"] = 1
		trainModelCofig["gradient_accumulation_steps"] = 6
		trainModelCofig["network_alpha"] = 32
		trainModelCofig["network_dim"] = 64
	case TrainQualityHigh:
		trainModelCofig["train_batch_size"] = 4
		trainModelCofig["gradient_accumulation_steps"] = 2
		trainModelCofig["network_alpha"] = 32
		trainModelCofig["network_dim"] = 64
	case TrainQualityLow:
		trainModelCofig["train_batch_size"] = 1
		trainModelCofig["gradient_accumulation_steps"] = 1
		trainModelCofig["network_alpha"] = 4
		trainModelCofig["network_dim"] = 8
	default:
	}

	newData, err := json.MarshalIndent(trainModelCofig, "", "  ")
	if err != nil {
		utils.Errorf("Function MarshalIndent error: %v", err)
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

	utils.Infof("Save new model config: %s", savePath)
	return savePath, true
}

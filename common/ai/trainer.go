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

	modelPath := filepath.Join(config.GetBasePath(), "scripts", "train_model.py")
	cmd := exec.Command("python", modelPath, filepath.Join(savePath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Errorf("Python script execution error: %s", err)
		utils.Errorf("Output: %s", string(output))
		return
	}
	utils.Infof("Output: %s", string(output))

	utils.KillPortWindows(6006)
	cmd = exec.Command("tensorboard",
		fmt.Sprintf("--logdir=%s", filepath.Join(config.GetBasePath(), "logs")))

	err = cmd.Start()
	if err != nil {
		utils.Errorf("Tensorboard execution error: %s", err)
		return
	}
	utils.Info("Tensorboard started with: http://127.0.0.1:6006/")
}

func buildTrainingSet(config *TrainConfig) bool {
	targetTag := ""
	for _, c := range config.GetTagConfigs() {
		srcDir := c.GetTagSrcPath()
		destDir := ""
		utils.Infof("Start setting the %s training set", c.GetTagName())

		copyImage := func(src, dst string) error {
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
		copyTxt := func(src, dst string) error {
			fileData, err := os.ReadFile(src)
			if err != nil {
				return err
			}

			dstFile, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = dstFile.WriteString(targetTag + "," + string(fileData))
			return err
		}

		index := 0
		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if config.CheckLimit(index) {
					utils.Warnf("Limit reached: %d", config.GetLimit())
					return nil
				}

				ext := strings.ToLower(filepath.Ext(path))
				relPath := filepath.Base(path)
				destPath := filepath.Join(destDir, relPath)
				if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
					utils.Errorf("Function os.MkdirAll error: %v", err)
					return err
				}

				if ext == ".jpg" {
					index++
					if err := copyImage(path, destPath); err != nil {
						utils.Errorf("Copy file error to: %s", destPath)
						return err
					}
					utils.Infof("Copy file success to: %s", destPath)
				}
				if ext == ".txt" {
					if err := copyTxt(path, destPath); err != nil {
						utils.Errorf("Copy file error to: %s", destPath)
						return err
					}
				}
			} else {
				parts := strings.Split(info.Name(), "_")
				if len(parts) > 1 {
					destDir = filepath.Join(config.GetInputDir(), fmt.Sprintf("%d_%s", c.GetTime(parts[1]), parts[1]))
					targetTag = parts[1]
				} else {
					targetTag = ""
				}

			}
			return nil
		})
		utils.Infof("TagName %s, total of %d images were read", c.GetTagName(), index)
		config.UpTrainTotalNum(index)
		config.SetTrainTagNum(c.GetTagName(), index)

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

	if prompts := config.GetPrompts(); prompts != "" {
		trainModelCofig["sample_prompts"] = prompts
	}
	if epoch := config.GetEpoch(); epoch > 0 {
		trainModelCofig["epoch"] = epoch
		trainModelCofig["max_train_epochs"] = epoch
	}

	switch config.GetTrainSpeed() {
	case TrainSpeedFast:
		trainModelCofig["train_batch_size"] = 4
		trainModelCofig["gradient_accumulation_steps"] = 1
	case TrainSpeedMid:
		trainModelCofig["train_batch_size"] = 2
		trainModelCofig["gradient_accumulation_steps"] = 2
	case TrainSpeedSlow:
		trainModelCofig["train_batch_size"] = 1
		trainModelCofig["gradient_accumulation_steps"] = 4
	default:
	}

	switch config.GetTrainQuality() {
	case TrainQualityHigh:
		trainModelCofig["network_alpha"] = 64
		trainModelCofig["network_dim"] = 128
	case TrainQualityMed:
		trainModelCofig["network_alpha"] = 32
		trainModelCofig["network_dim"] = 64
	case TrainQualityLow:
		trainModelCofig["network_alpha"] = 16
		trainModelCofig["network_dim"] = 32
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

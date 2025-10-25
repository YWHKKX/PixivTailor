package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func Test_TrainModel(t *testing.T) {
	currntPath, _ := os.Getwd()
	basePath := filepath.Join(currntPath, "./../..")

	modelName := "re0"
	pretrainedPath := "D:/PythonProject/stable-diffusion-webui/models/Stable-diffusion/chosenMix_bakedVae.safetensors"
	inputDir := filepath.Join(basePath, "images", "Train")

	config := ai.NewTrainConfig(modelName, pretrainedPath, inputDir, basePath)
	config.SetLimit(60)

	config.AddTagConfig(
		"rem", filepath.Join(basePath, "images", "Rem"), map[string]int{
			"rem": 20,
		},
	)
	config.AddTagConfig(
		"ram", filepath.Join(basePath, "images", "Ram"), map[string]int{
			"ram": 20,
		},
	)
	config.AddTagConfig(
		"echidna", filepath.Join(basePath, "images", "Echidna"), map[string]int{
			"echidna": 20,
		},
	)
	config.AddTagConfig(
		"emilia", filepath.Join(basePath, "images", "Emilia"), map[string]int{
			"emilia": 20,
		},
	)
	config.AddTagConfig(
		"priscilla", filepath.Join(basePath, "images", "Priscilla"), map[string]int{
			"priscilla": 20,
		},
	)
	config.AddTagConfig(
		"felt", filepath.Join(basePath, "images", "Felt"), map[string]int{
			"felt": 20,
		},
	)
	config.AddTagConfig(
		"elsa", filepath.Join(basePath, "images", "Elsa"), map[string]int{
			"rem": 20,
		},
	)

	ai.TrainModel(config)
}

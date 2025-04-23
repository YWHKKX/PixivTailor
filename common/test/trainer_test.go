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
		"rem", filepath.Join(basePath, "images", "Rem"), 10,
	)
	config.AddTagConfig(
		"ram", filepath.Join(basePath, "images", "Ram"), 10,
	)
	config.AddTagConfig(
		"echidna", filepath.Join(basePath, "images", "Echidna"), 10,
	)
	config.AddTagConfig(
		"emilia", filepath.Join(basePath, "images", "Emilia"), 10,
	)
	config.AddTagConfig(
		"priscilla", filepath.Join(basePath, "images", "Priscilla"), 10,
	)
	config.AddTagConfig(
		"felt", filepath.Join(basePath, "images", "Felt"), 10,
	)
	config.AddTagConfig(
		"elsa", filepath.Join(basePath, "images", "Elsa"), 10,
	)

	ai.TrainModel(config)
}

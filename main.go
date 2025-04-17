package main

import (
	"os"
	"path/filepath"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func main() {
	basePath, _ := os.Getwd()

	modelName := "re0"
	pretrainedPath := "D:/PythonProject/stable-diffusion-webui/models/Stable-diffusion/chosenMix_bakedVae.safetensors"
	inputDir := filepath.Join(basePath, "images", "Test")

	config := ai.NewTrainConfig(modelName, pretrainedPath, inputDir)
	config.AddTagConfig(
		"re0_Rem", filepath.Join(basePath, "images", "Rem"), 10,
	)
	config.AddTagConfig(
		"re0_Ram", filepath.Join(basePath, "images", "Ram"), 10,
	)
	config.AddTagConfig(
		"re0_Echidna", filepath.Join(basePath, "images", "Echidna"), 10,
	)
	config.AddTagConfig(
		"re0_Emilia", filepath.Join(basePath, "images", "Emilia"), 10,
	)
	config.AddTagConfig(
		"re0_Capella", filepath.Join(basePath, "images", "Capella"), 10,
	)

	ai.Tailor_TrainModel(config)
}

package main

import (
	"os"
	"path/filepath"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func do_trainer() {
	currntPath, _ := os.Getwd()
	basePath := currntPath

	modelName := "re0"
	pretrainedPath := "D:/PythonProject/stable-diffusion-webui/models/Stable-diffusion/chosenMix_bakedVae.safetensors"
	inputDir := filepath.Join(basePath, "images", "Test")

	config := ai.NewTrainConfig(modelName, pretrainedPath, inputDir, basePath)
	config.SetLimit(40)

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

	ai.TrainModel(config)
}

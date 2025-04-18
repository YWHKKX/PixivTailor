package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func Test_Tagger_Webuiwd14tagger(t *testing.T) {
	currntPath, _ := os.Getwd()
	basePath := filepath.Join(currntPath, "./../..")

	config := ai.InitImageConfig("chosenMix_bakedVae.safetensors [52b8ebbd5b]", basePath)
	config.SetShowTags(true)
	config.SetForEach(true)
	config.SetSaveType(ai.Save_Txt)
	config.SetAnalyzeType(ai.Analyze_Webuiwd14tagger)

	config.AddTagConfig(
		"rem", filepath.Join(basePath, "images", "Rem"),
	)
	config.AddTagConfig(
		"ram", filepath.Join(basePath, "images", "Ram"),
	)
	config.AddTagConfig(
		"echidna", filepath.Join(basePath, "images", "Echidna"),
	)
	config.AddTagConfig(
		"emilia", filepath.Join(basePath, "images", "Emilia"),
	)

	ai.SaveTagsFormImage(config)
}

func Test_Tagger_Deepdanbooru(t *testing.T) {
	currntPath, _ := os.Getwd()
	basePath := filepath.Join(currntPath, "./../..")

	config := ai.InitImageConfig("chosenMix_bakedVae.safetensors [52b8ebbd5b]", basePath)
	config.SetShowTags(true)
	config.SetAnalyzeType(ai.Analyze_Deepdanbooru)
	ai.SaveTagsFormImage(config)
}

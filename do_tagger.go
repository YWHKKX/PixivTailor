package main

import (
	"os"
	"path/filepath"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func do_tageer() {
	currntPath, _ := os.Getwd()
	basePath := currntPath

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

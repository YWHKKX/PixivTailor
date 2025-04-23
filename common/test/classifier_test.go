package test

import (
	"path/filepath"
	"testing"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func Test_Classifier(t *testing.T) {
	keyString1 := ""
	keyString2 := ""

	outputName := filepath.Join("scripts", "global_tag.json")

	config := ai.NewCategoryConfig([]string{keyString1, keyString2}, outputName)
	config.AddContainPathName("Input")
	config.SetShowTags(false)
	ai.GetCategory(config)
}

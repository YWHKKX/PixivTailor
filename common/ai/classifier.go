package ai

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type ImageTag struct {
	RawTag     string
	Confidence float64
}

type CategoryType string

const (
	CategoryType_Clothing   CategoryType = "clothing"
	CategoryType_Background CategoryType = "background"
	CategoryType_Character  CategoryType = "character"
	CategoryType_Pose       CategoryType = "pose"
)

type Category struct {
	Name     string       `json:"tag"`
	Kind     CategoryType `json:"kind"`
	Keywords []string     `json:"keywords"`
}

func InitCategories() []*Category {
	return []*Category{
		{
			Name:     "人物特征",
			Kind:     CategoryType_Character,
			Keywords: []string{},
		},
		{
			Name:     "服装特征",
			Kind:     CategoryType_Clothing,
			Keywords: []string{},
		},
		{
			Name:     "动作特征",
			Kind:     CategoryType_Pose,
			Keywords: []string{},
		},
		{
			Name:     "背景特征",
			Kind:     CategoryType_Background,
			Keywords: []string{},
		},
	}
}

func classifyTags(keyString, tagString, basePath string) map[string][]string {
	var result map[string][]string

	cmd := exec.Command("python",
		filepath.Join(basePath, "scripts", "classify_tag.py"),
		tagString, keyString)

	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Errorf("Python script execution error: %s", err)
		utils.Errorf("Output: %s", string(output))
		return result
	}

	if err := json.Unmarshal(output, &result); err != nil {
		utils.Errorf("Function Unmarshal Error: %v", err)
		return result
	}

	return result
}

func saveTagsToFile(tagsMap map[string][]string, outputPath string) {
	categorys := InitCategories()

	if _, err := os.Stat(outputPath); err == nil {
		fileData, err := os.ReadFile(outputPath)
		if err != nil {
			utils.Errorf("Read File Error: %v", err)
			return
		}
		if err := json.Unmarshal(fileData, &categorys); err != nil {
			utils.Errorf("Function Unmarshal Error: %v", err)
			return
		}
	}

	for _, category := range categorys {
		partMap := make(map[string]bool)

		if tags, ok := tagsMap[string(category.Kind)]; ok {
			var ret []string
			for _, k := range category.Keywords {
				partMap[strings.TrimSpace(k)] = true
			}
			for _, t := range tags {
				partMap[strings.TrimSpace(t)] = true
			}
			for p, _ := range partMap {
				ret = append(ret, p)
			}
			category.Keywords = ret
		}
	}

	jsonFile, err := os.Create(outputPath)
	if err != nil {
		utils.Errorf("Failed to create file: %s", err)
		return
	}
	defer jsonFile.Close()

	jsonData, err := json.MarshalIndent(categorys, "", "  ")
	if err != nil {
		utils.Errorf("Function MarshalIndent error: %v", err)
		return
	}

	_, err = jsonFile.Write(jsonData)
	if err != nil {
		utils.Errorf("Function file.Write error: %v", err)
		return
	}
}

func GetCategory(config *CategoryConfig) {
	basePath := config.GetBasePath()

	inputPath := filepath.Join(basePath, "Images")
	outputPath := config.GetOutputPath()

	files, _ := filepath.Glob(filepath.Join(inputPath, "*", "*.txt"))
	for _, file := range files {
		if !config.CheckPathFilter(file) {
			continue
		}
		utils.Infof("Start read tags: %s", file)
		fileData, err := os.ReadFile(file)
		if err != nil {
			utils.Errorf("Read File Error: %v", err)
			continue
		}

		result := classifyTags(config.GetKeyString(), string(fileData), basePath)
		if config.GetShowTags() {
			for k, v := range result {
				utils.Infof("%s: %s", k, strings.Join(v, ","))
			}
		}

		saveTagsToFile(result, outputPath)
	}
}

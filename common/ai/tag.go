package ai

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type DeepDanbooruResponses struct {
	TagName string                 `json:"name"`
	TagPath string                 `json:"path"`
	Results []DeepDanbooruResponse `json:"results"`
}

type DeepDanbooruResponse struct {
	Tags []struct {
		Tag        string  `json:"tag"`
		Confidence float64 `json:"confidence"`
	} `json:"tags"`
}

// deepdanbooru: https://github.com/KichangKim/DeepDanbooru
// deepdanbooru-v3-20211112-sgd-e28: https://github.com/KichangKim/DeepDanbooru/releases/tag/v3-20211112-sgd-e28
func AnalyzeImage(cmdPath, imagePath string) []byte {
	cmd := exec.Command("deepdanbooru", "evaluate", imagePath, "--project-path",
		filepath.Join(cmdPath, "deepdanbooru-v3-20211112-sgd-e28"), "--allow-folder")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error("Error executing command:", err)
		return []byte{}
	}

	return output
}

func ParseSimple(output string) (string, float64) {

	parts := strings.SplitN(output, " ", 2)
	if len(parts) == 2 {
		scoreStr := strings.Trim(parts[0], "()")
		score, _ := strconv.ParseFloat(scoreStr, 64)
		return parts[1], score
	}
	return "", 0.0
}

func GetTags(path string, tag []byte) *DeepDanbooruResponse {
	var response *DeepDanbooruResponse = &DeepDanbooruResponse{}

	lines := strings.Split(string(tag), "\n")
	for _, line := range lines {
		if tag, score := ParseSimple(line); score > 0 {
			response.Tags = append(response.Tags, struct {
				Tag        string  `json:"tag"`
				Confidence float64 `json:"confidence"`
			}{
				Tag:        tag,
				Confidence: score,
			})
		}
	}

	return response
}

func SaveTagsFormImage() {
	var responses *DeepDanbooruResponses = &DeepDanbooruResponses{}

	currentPath, _ := os.Getwd()
	cmdPath := filepath.Join(currentPath, "models")
	imagePath := filepath.Join(currentPath, "images")
	tagPath := ""

	saveResponses := func(path string) {
		file, err := os.Create(path)
		if err != nil {
			utils.Errorf("Function os.Create error: %v", err)
			return
		}
		defer file.Close()

		data, _ := json.Marshal(responses)
		_, err = file.Write(data)
		if err != nil {
			utils.Errorf("Function io.Copy error: %v", err)
			return
		}
		return
	}

	files, _ := filepath.Glob(filepath.Join(imagePath, "*", "*.jpg"))
	for _, file := range files {
		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		tagPath = filepath.Join(imagePath, tagName, tagName+".json")
		if _, err := os.Stat(tagPath); err == nil {
			utils.Warnf("Already exist tag file: %s", tagPath)
			continue
		}

		if responses.TagName == "" {
			responses.TagName = tagName
			responses.TagPath = tagPath
		} else if responses.TagName != tagName {
			saveResponses(responses.TagPath)

			responses = &DeepDanbooruResponses{
				TagName: tagName,
				TagPath: tagPath,
			}
		}

		utils.Infof("Analyze image: %s", file)
		tag := AnalyzeImage(cmdPath, file)
		if len(tag) == 0 {
			continue
		}

		if response := GetTags(tagPath, tag); response != nil {
			responses.Results = append(responses.Results, *response)
		}
	}

	if tagPath != "" {
		saveResponses(tagPath)
	}
}

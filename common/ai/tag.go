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
	TagName   string                 `json:"tag_name"`
	TagPath   string                 `json:"tag_path"`
	Results   []DeepDanbooruResponse `json:"results"`
	TagString string                 `json:"tag_string"`
	TagNum    int                    `json:"tag_num"`
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

func SplitTags(path string, tag []byte) *DeepDanbooruResponse {
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

func TidyTags(responses *DeepDanbooruResponses) (string, int) {
	var tags map[string]struct{} = make(map[string]struct{})
	var ret []string = make([]string, 0)
	for _, result := range responses.Results {
		for _, tag := range result.Tags {
			if tag.Confidence > 0.5 {
				tags[tag.Tag[:len(tag.Tag)-1]] = struct{}{}
			}
		}
	}
	for tag := range tags {
		ret = append(ret, tag)
	}
	return strings.Join(ret, ","), len(ret)
}

func SaveTagsFormImage(config config) {
	var responses *DeepDanbooruResponses = &DeepDanbooruResponses{}

	cmdPath := filepath.Join(config.GetBasePath(), "models")
	imagePath := filepath.Join(config.GetBasePath(), "images")
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
		}
	}

	files, _ := filepath.Glob(filepath.Join(imagePath, "*", "*.jpg"))
	for _, file := range files {
		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		tagPath = filepath.Join(imagePath, tagName, tagName+".json")
		if tagName == config.GetSaveName() {
			utils.Warnf("Skip tag file: %s", tagPath)
			continue
		}

		if responses.TagName == "" || responses.TagPath == "" {
			responses.TagName = tagName
			responses.TagPath = tagPath
		} else if responses.TagName != tagName {
			saveResponses(responses.TagPath)
			responses = &DeepDanbooruResponses{
				TagName: tagName,
				TagPath: tagPath,
			}
		}

		if _, err := os.Stat(tagPath); err == nil {
			utils.Warnf("Already exist tag file: %s", tagPath)
			continue
		}

		utils.Infof("Analyze image: %s", file)
		tag := AnalyzeImage(cmdPath, file)
		if len(tag) == 0 {
			continue
		}

		if response := SplitTags(tagPath, tag); response != nil {
			responses.Results = append(responses.Results, *response)
		}
		tagString, tagNum := TidyTags(responses)
		if config.GetShowTags() && tagNum > 0 {
			utils.Infof("Show tags: %s", tagString)
		}
		responses.TagString = tagString
		responses.TagNum = tagNum
	}

	if responses.TagString != "" {
		saveResponses(responses.TagPath)
	}
}

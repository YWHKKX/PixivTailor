package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

var GlobalTag_Negative []string = []string{
	"low quality", "worst quality", "deformed",
	"distorted", "(extra limbs:1.3)", "(mutated hands:1.4)",
	"fused fingers", "(cloned face:1.2)", "(multiple people:1.5)",
	"overlapping figures", "merged bodies", "bad anatomy",
	"malformed limbs", "extra arms", "out of focus, blurry",
	"(crowd:1.3)", "(congested:1.2)", "no merged poses",
	"distinct postures", "blurry", "lowres", "jpeg artifacts",
	"watermark",
}

var GlobalTag_Clothing []string = []string{
	"clothing", "underwear", "dress", "tunic",
	"shirt", "hat", "beret", "turban",
	"jacket", "glove", "tie", "cravat",
	"heels", "apron", "moccasin", "slippers",
	"bow", "swimsuit", "bathing suit", "bikini",
	"uniform", "stockings", "suspenders", "handkerchief",
	"scrunchie", "slip", "petticoat",
	"bowtie", "stays", "corset", "panties",
	"sleeve", "mantle", "cloak", "bathrobe",
	"socks", "pocket", "cuff", "blouse",
	"ribbon", "vest", "kimono", "trousers",
	"chalkboard", "belt", "underskirt", "briefs",
	"shoes", "brassiere", "bra", "corselet",
	"top", "skirt", "coat", "sweater",
}

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

func AnalyzeImage(config ImageConfig, cmdPath, imagePath string) map[string]float64 {
	if config.GetAnalyzeType() == Analyze_Deepdanbooru {
		return AnalyzeImageByDeepdanbooru(cmdPath, imagePath)
	} else if config.GetAnalyzeType() == Analyze_Webuiwd14tagger {
		return AnalyzeImageByWebuiwd14tagger(imagePath)
	}
	return make(map[string]float64)
}

func AnalyzeImageByWebuiwd14tagger(imagePath string) map[string]float64 {
	url := SD_API_TAGGER

	fileData, err := os.ReadFile(imagePath)
	if err != nil {
		utils.Errorf("Read File Error: %v", err)
		return nil
	}

	requestData := map[string]interface{}{
		"image":     fileData,
		"model":     "wd14-vit-v2-git",
		"threshold": 0.4,
	}

	jsonData, _ := json.Marshal(requestData)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Errorf("Function http.NewRequest error: %v", err)
		return nil
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		utils.Errorf("Request error: %v", err)
		return nil
	}
	if res.StatusCode != http.StatusOK {
		utils.Errorf("Request StatusCode: %d", res.StatusCode)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	var result map[string]map[string]float64

	if err := json.Unmarshal(body, &result); err != nil {
		utils.Error(err)
		return nil
	}

	return result["caption"]
}

// deepdanbooru: https://github.com/KichangKim/DeepDanbooru
// deepdanbooru-v3-20211112-sgd-e28: https://github.com/KichangKim/DeepDanbooru/releases/tag/v3-20211112-sgd-e28
func AnalyzeImageByDeepdanbooru(cmdPath, imagePath string) map[string]float64 {
	cmd := exec.Command("deepdanbooru", "evaluate", imagePath, "--project-path",
		filepath.Join(cmdPath, "deepdanbooru-v3-20211112-sgd-e28"), "--allow-folder")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utils.Error("Error executing command:", err)
		return nil
	}

	return splitTags(output)
}

func parseSimple(output string) (string, float64) {

	parts := strings.SplitN(output, " ", 2)
	if len(parts) == 2 {
		scoreStr := strings.Trim(parts[0], "()")
		score, _ := strconv.ParseFloat(scoreStr, 64)
		tag := strings.TrimRight(parts[1], "\r")
		return tag, score
	}
	return "", 0.0
}

func splitTags(tag []byte) map[string]float64 {
	response := make(map[string]float64)

	lines := strings.Split(string(tag), "\n")
	for _, line := range lines {
		if tag, score := parseSimple(line); score > 0 {
			response[tag] = score
		}
	}

	return response
}

func tidyTags(responses *DeepDanbooruResponses) (string, int) {
	var tags map[string]struct{} = make(map[string]struct{})
	var ret []string = make([]string, 0)
	for _, result := range responses.Results {
		for _, tag := range result.Tags {
			if tag.Confidence > 0.8 {
				tags[tag.Tag] = struct{}{}
			}
		}
	}
	for tag := range tags {
		ret = append(ret, tag)
	}
	return strings.Join(ret, ","), len(ret)
}

func SaveTagsFormImage(config ImageConfig) {
	var responses *DeepDanbooruResponses = &DeepDanbooruResponses{}

	cmdPath := filepath.Join(config.GetBasePath(), "models")
	imagePath := filepath.Join(config.GetBasePath(), "images")
	savePath := ""

	saveResponses := func(path string) {
		file, err := os.Create(path)
		if err != nil {
			utils.Errorf("Function os.Create error: %v", err)
			return
		}
		defer file.Close()

		switch t := config.GetSaveType(); t {
		case Save_Json:
			data, _ := json.Marshal(responses)
			_, err = file.Write(data)
			if err != nil {
				utils.Errorf("Function file.Write error: %v", err)
			}
		case Save_Txt:
			newTag := ""
			for _, e := range config.GetExtendTags("") {
				newTag += fmt.Sprintf("%s,", e)
			}
			data := responses.TagString
			_, err = file.WriteString(fmt.Sprintf("%s%s", newTag, data))
			if err != nil {
				utils.Errorf("Function file.Write error: %v", err)
			}
		default:
			utils.Errorf("SaveType: %s not support", t)
		}

	}

	buildResponse := func(tagPath, file string) (DeepDanbooruResponse, bool) {
		response := DeepDanbooruResponse{}

		if _, err := os.Stat(tagPath); err == nil && config.GetSaveType() == Save_Json {
			utils.Warnf("Already exist tag file: %s", tagPath)
			return response, false
		}

		utils.Infof("Analyze image: %s", file)
		tags := AnalyzeImage(config, cmdPath, file)
		if len(tags) == 0 {
			return response, false
		}

		utils.Infof("Number of labels after analysis: %d", len(tags))
		for tag, score := range tags {
			tag = strings.Replace(tag, "_", " ", -1)
			if config.CheckSkipTags(tag) {
				continue
			}
			response.Tags = append(response.Tags, struct {
				Tag        string  `json:"tag"`
				Confidence float64 `json:"confidence"`
			}{
				Tag:        tag,
				Confidence: score,
			})
		}
		utils.Infof("Number of labels after filtering: %d", len(response.Tags))

		return response, true
	}

	files, _ := filepath.Glob(filepath.Join(imagePath, "*", "*.jpg"))
	for _, file := range files {
		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		if tagName == config.GetSavePathName() {
			utils.Warnf("Skip tag file: %s", savePath)
			continue
		}

		if config.IsForEach() {
			savePath = filepath.Join(imagePath, tagName, filepath.Base(file))
			savePath = strings.TrimSuffix(savePath, ".jpg") + "." + string(config.GetSaveType())

			if !config.CheckPathFilter(file) {
				continue
			}

			responses.TagName = tagName
			responses.TagPath = savePath

			if response, ok := buildResponse(savePath, file); ok {
				responses.Results = append(responses.Results, response)

				tagString, tagNum := tidyTags(responses)
				utils.Infof("Number of labels after tidying: %d", tagNum)
				if config.GetShowTags() && tagNum > 0 {
					utils.Infof("Show tags: %s", tagString)
				}
				responses.TagString = tagString
				responses.TagNum = tagNum

				saveResponses(responses.TagPath)
			}
		} else {
			savePath = filepath.Join(imagePath, tagName, tagName) + "." + string(config.GetSaveType())

			if responses.TagName == "" || responses.TagPath == "" {
				responses.TagName = tagName
				responses.TagPath = savePath
			} else if responses.TagName != tagName && responses.TagString != "" {
				saveResponses(responses.TagPath)
				responses = &DeepDanbooruResponses{
					TagName: tagName,
					TagPath: savePath,
				}
			}

			if response, ok := buildResponse(savePath, file); ok {
				responses.Results = append(responses.Results, response)

				tagString, tagNum := tidyTags(responses)
				if config.GetShowTags() && tagNum > 0 {
					utils.Infof("Show tags: %s", tagString)
				}
				responses.TagString = tagString
				responses.TagNum = tagNum
			}
		}

	}

	if responses.TagString != "" {
		saveResponses(responses.TagPath)
	}
}

func hasClothingTag(text string, skip []string) bool {
	pattern := `(?i)\b(` + strings.Join(skip, "|") + `)\b`
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		utils.Errorf("Regex compilation error: ", err)
		return false
	}
	return matched
}

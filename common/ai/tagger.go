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
	"EasyNegative", "badhandsv5-neg", "Subtitles", "word",
	"((logo))", "watermark", "(3d, photo, hyperrealistic, rough sketch:1.1)",
	"(derpibooru_p_low)", "furry", "source furry", "source comic", "dark skin", "monochrome", "text", "signature",
	"soft focus", "deformed face", "bad proportions", "distorted face",
	"ugly", "mutated face", "poorly drawn face",
	"low quality", "worst quality", "deformed",
	"distorted", "(extra limbs:1.3)", "(mutated hands:1.4)",
	"fused fingers", "(cloned face:1.2)", "(multiple people:1.5)",
	"overlapping figures", "merged bodies", "bad anatomy",
	"malformed limbs", "extra arms", "out of focus", "blurry",
	"(crowd:1.3)", "(congested:1.2)", "no merged poses",
	"distinct postures", "blurry", "lowres", "jpeg artifacts",
	"watermark", "extra legs", "(asymmetrical legs:1.3)",
	"blurry legs", "fused thighs", "(disconnected joints:1.2)",
}

var GlobalTag_Special_NoBackground []string = []string{}

var GlobalTag_Special_Number []string = []string{
	"1girl", "1boy", "2girls", "2boys", "3girls", "3boys",
	"4girls", "4boys", "5girls", "5boys", "6girls", "6boys",
	"7girls", "7boys", "8girls", "8boys", "9girls", "9boys",
}

var GlobalTag_Character []string = []string{}

var GlobalTag_Clothing []string = []string{
	"clothing", "underwear", "dress", "tunic",
	"shirt", "hat", "beret", "turban",
	"jacket", "glove", "tie", "cravat",
	"heels", "apron", "moccasin", "slippers",
	"bow", "swimsuit", "bathing suit", "bikini",
	"uniform", "stockings", "suspenders", "handkerchief",
	"scrunchie", "slip", "petticoat", "sleeves",
	"bowtie", "stays", "corset", "panties",
	"sleeve", "mantle", "cloak", "bathrobe",
	"socks", "pocket", "cuff", "blouse",
	"ribbon", "vest", "kimono", "trousers",
	"chalkboard", "belt", "underskirt", "briefs",
	"shoes", "brassiere", "bra", "corselet",
	"top", "skirt", "coat", "sweater",
	"earrings", "jewelry", "capelet", "hair ornament",
	"miniskirt", "leotard", "pantyhose", "gloves",
	"collar", "leotard", "fishnets", "garter",
}

var GlobalTag_Action []string = []string{}
var GlobalTag_Background []string = []string{}

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

func AnalyzeImage(config *ImageConfig, cmdPath, imagePath string) map[string]float64 {
	var ret map[string]float64

	if config.GetAnalyzeType() == Analyze_Deepdanbooru {
		ret = analyzeImageByDeepdanbooru(cmdPath, imagePath)
	} else if config.GetAnalyzeType() == Analyze_Webuiwd14tagger {
		ret = analyzeImageByWebuiwd14tagger(imagePath)
	}
	return ret
}

func analyzeImageByWebuiwd14tagger(imagePath string) map[string]float64 {
	url := SD_API_TAGGER

	fileData, err := os.ReadFile(imagePath)
	if err != nil {
		utils.Errorf("Read File Error: %v", err)
		return nil
	}

	requestData := map[string]interface{}{
		"image":     fileData,
		"model":     "wd14-vit-v2-git",
		"threshold": 0.35,
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
		utils.Errorf("Response StatusCode: %d", res.StatusCode)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	var result map[string]map[string]float64

	if err := json.Unmarshal(body, &result); err != nil {
		utils.Errorf("Function Unmarshal Error: %v", err)
		return nil
	}

	return result["caption"]
}

// deepdanbooru: https://github.com/KichangKim/DeepDanbooru
// deepdanbooru-v3-20211112-sgd-e28: https://github.com/KichangKim/DeepDanbooru/releases/tag/v3-20211112-sgd-e28
func analyzeImageByDeepdanbooru(cmdPath, imagePath string) map[string]float64 {
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

var Uncategorized map[string]bool = make(map[string]bool)

func tidyTags(responses *DeepDanbooruResponses, extendTags []string, tagOrder TagOrder) (string, int) {
	var tags map[string]bool = make(map[string]bool)
	for _, result := range responses.Results {
		for _, tag := range result.Tags {
			if tag.Confidence > 0.35 {
				tags[tag.Tag] = true
			}
		}
	}
	var CharacterTag []string
	var ClothingTag []string
	var ActionTag []string
	var BackgroundTag []string
	var OtherTag []string

	checkFunc := func(global []string, input string) bool {
		globalTags := strings.Join(global, "|")
		if globalTags == "" {
			return false
		}
		pattern := `(?i)\b(` + globalTags + `)\b`
		matched, err := regexp.MatchString(pattern, input)
		if err != nil {
			utils.Errorf("Regex compilation error: ", err)
			return false
		}
		return matched
	}

	for tag, _ := range tags {
		if checkFunc(GlobalTag_Character, tag) {
			CharacterTag = append(CharacterTag, tag)
		} else if checkFunc(GlobalTag_Clothing, tag) {
			ClothingTag = append(ClothingTag, tag)
		} else if checkFunc(GlobalTag_Action, tag) {
			ActionTag = append(ActionTag, tag)
		} else if checkFunc(GlobalTag_Background, tag) {
			BackgroundTag = append(BackgroundTag, tag)
		} else {
			OtherTag = append(OtherTag, tag)
		}
	}
	if len(OtherTag) > 0 {
		utils.Warnf("Some tags are not classified: %s", strings.Join(OtherTag, ","))
		for _, tag := range OtherTag {
			Uncategorized[tag] = true
		}
	}

	var ret []string = extendTags

	switch tagOrder {
	case TagOrder_Character:
		ret = append(ret, CharacterTag...)
		ret = append(ret, ClothingTag...)
		ret = append(ret, ActionTag...)
		ret = append(ret, BackgroundTag...)
		ret = append(ret, OtherTag...)
	case TagOrder_Action:
		ret = append(ret, ActionTag...)
		ret = append(ret, CharacterTag...)
		ret = append(ret, ClothingTag...)
		ret = append(ret, BackgroundTag...)
		ret = append(ret, OtherTag...)
	default:
		utils.Warnf("Unknown tag order: %s", tagOrder)
	}

	return strings.Join(ret, ","), len(ret)
}

func SaveTagsFormImage(config *ImageConfig) {
	var responses *DeepDanbooruResponses = &DeepDanbooruResponses{}

	if config.GetDeleteTags() {
		DeleteTags(config)
	}

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
			tagName := filepath.Base(strings.TrimSuffix(path, filepath.Base(path)))
			for _, e := range config.GetExtendTags(tagName) {
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
		if len(response.Tags) == 0 {
			utils.Warnf("No tags after filtering")
			return response, false
		}

		return response, true
	}

	files, _ := filepath.Glob(filepath.Join(imagePath, "*", "*.jpg"))
	for _, file := range files {
		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		if tagName == config.GetSavePathName() {
			utils.Warnf("Skip tag file: %s", file)
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
				responses.Results = []DeepDanbooruResponse{response}
				tagString, tagNum := tidyTags(responses, config.GetExtendTags(""), config.GetTagOrder())
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

				tagString, tagNum := tidyTags(responses, config.GetExtendTags(""), config.GetTagOrder())
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

	UncategorizedStr := []string{}
	for u, _ := range Uncategorized {
		UncategorizedStr = append(UncategorizedStr, u)
	}
	if len(UncategorizedStr) > 0 {
		utils.Warnf("Uncategorized Tags: %s", strings.Join(UncategorizedStr, ","))
	}
}

func DeleteTags(config *ImageConfig) {
	imagePath := filepath.Join(config.GetBasePath(), "images")

	files, _ := filepath.Glob(filepath.Join(imagePath, "*", "*.txt"))
	for _, file := range files {
		tagName := filepath.Base(strings.TrimSuffix(file, filepath.Base(file)))
		if tagName == config.GetSavePathName() {
			utils.Warnf("Skip tag file: %s", file)
			continue
		}

		if !config.CheckPathFilter(file) {
			continue
		}

		utils.Infof("Delete tag file: %s", file)
		if err := os.RemoveAll(file); err != nil {
			utils.Errorf("Function os.RemoveAll error: %v", err)
		}
	}
}

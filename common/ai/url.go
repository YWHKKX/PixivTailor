package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

// modle download from: https://civitai.com/

var SD_API_TXT2IMG string = "http://127.0.0.1:7860/sdapi/v1/txt2img"
var SD_API_TAGGER string = "http://127.0.0.1:7860/tagger/v1/interrogate"

/*
{
	(masterpiece:1.2), best quality, PIXIV, blue archive
		......
Write the desired content here
		......
	<lora:blue archive-000018:0.8>,
	Negative prompt: EasyNegative,badhandsv5-neg,Subtitles,word,
	Steps: 32,
	Sampler: DPM++ SDE Karras,
	CFG scale: 7,
	Seed: 3089360094,
	Size: 768x512,
	Model hash: e3edb8a26f,
	Model: ghostmix_v20Bakedvae,
	Denoising strength: 0.5,
	Clip skip: 2,
	ADetailer model: face_yolov8n.pt,
	ADetailer confidence: 0.3,
	ADetailer dilate erode: 4,
	ADetailer mask blur: 4,
	ADetailer denoising strength: 0.4,
	ADetailer inpaint only masked: True,
	ADetailer inpaint padding: 32,
	ADetailer version: 23.11.0,
	Hires upscale: 2,
	Hires steps: 12,
	Hires upscaler:
	R-ESRGAN 4x+ Anime6B,
	Lora hashes: "blue archive-000018: 3fc8d3abb551",
	TI hashes: "EasyNegative: 66a7279a88dd, badhandsv5-neg: aa7651be154c",
	Version: v1.6.0-2-g4afaaf8a
}
*/

func makeSDRequest(tagName, tagString string, config *ImageConfig) *http.Response {
	loraString := ""
	extendTags := ""
	negativeString := strings.Join(GlobalTag_Negative, ",")

	for l, f := range config.GetLoraModel() {
		loraString += fmt.Sprintf("<lora:%s:%f>%s,", l, f, strings.Join(config.GetLoraTags(l), ","))
	}

	if tagName != "" {
		for _, e := range config.GetExtendTags(tagName) {
			extendTags += e + ","
		}
	}
	for _, e := range config.GetExtendTags("") {
		extendTags += e + ","
	}

	args_ADetailer := []map[string]interface{}{}
	args_ControlNet := []map[string]interface{}{}
	args_ComposableLora := []map[string]interface{}{}
	args_LatentCouplet := []map[string]interface{}{}

	if config.GetFixface() {
		args_ADetailer = append(args_ADetailer, map[string]interface{}{
			"ad_model":               "face_yolov8n.pt",
			"ad_confidence":          0.3,
			"ad_dilate_erode":        4,
			"ad_mask_blur":           4,
			"ad_denoising_strength":  0.4,
			"ad_inpaint_only_masked": true,
			"ad_inpaint_padding":     32,
			"ad_version":             "25.3.0",
		})
	}
	if config.GetFixhand() {
		args_ADetailer = append(args_ADetailer, map[string]interface{}{
			"ad_model":              "hand_yolov9c.pt",
			"ad_confidence":         0.2,
			"ad_denoising_strength": 0.5,
			"ad_mask_blur":          8,
			"ad_inpaint_padding":    64,
		})
	}
	if config.GetFixpose() {
		args_ADetailer = append(args_ADetailer, map[string]interface{}{
			"ad_model": "none",
			"controlnet": map[string]interface{}{
				"enabled":        true,
				"preprocessor":   "dw_openpose_full",
				"model":          "control_v11p_sd15_openpose [cab727d4]",
				"weight":         0.8,
				"guidance_start": 0.0,
				"guidance_end":   1.0,
				"pixel_perfect":  true,
			},
		})
	}
	if config.IsManualSetup() {
		args_LatentCouplet = append(args_LatentCouplet, map[string]interface{}{
			"enabled":     true,
			"divisions":   "1:1,1:2,1:2",
			"positions":   "0:0,0:0,0:1",
			"weights":     "0.2,0.8,0.8",
			"end_at_step": 32,
		})
		args_ComposableLora = append(args_ComposableLora, map[string]interface{}{
			"enabled": true,
		})
	}

	for poseName, poseConfig := range config.GetPoseConfigs() {
		poseDir := filepath.Join(config.GetBasePath(), "poses", poseName)
		entries, err := ioutil.ReadDir(poseDir)

		if err != nil {
			utils.Errorf("PosePath dir: %s not find, error: %s", poseDir, err)
			return nil
		}

		setArg := func(imageData []byte) {
			//base64Encoding := base64.StdEncoding.EncodeToString(imageData)
			args_ControlNet = append(args_ControlNet, map[string]interface{}{
				"enabled":        true,
				"preprocessor":   "none",
				"pixel_perfect":  true,
				"allow_preview":  false,
				"low_vram":       false,
				"control_mode":   "Balanced",
				"model":          "control_v11p_sd15_openpose [cab727d4]",
				"weight":         1,
				"threshold_a":    0.5,
				"threshold_b":    0.5,
				"guidance_start": 0,
				"guidance_end":   1,
				"processor_res":  512,
				"image":          imageData,
			})
		}

		var groupNum, memberNum int
		tmpName := ""
		for _, file := range entries {
			filename := file.Name()
			ext := filepath.Ext(filename)
			nameWithoutExt := strings.TrimSuffix(filename, ext)
			parts := strings.Split(nameWithoutExt, "_")
			nameFinal := parts[0]

			if nameFinal != tmpName {
				groupNum++
				memberNum = 0
				tmpName = nameFinal
			}
			memberNum++
		}

		poseConfig.GroupNum = groupNum
		poseConfig.MemberNum = memberNum
		poseKey := poseConfig.PoseKey

		poseFile := filepath.Join(poseDir, poseKey+".png")
		if _, err := os.Stat(poseFile); err == nil {
			utils.Infof("PosePath file: %s to configuration", poseFile)

			for index := 1; index < memberNum+1; index++ {
				poseFile = filepath.Join(poseDir, fmt.Sprintf("%s_%d.png", poseKey, index))
				imageData, err := ioutil.ReadFile(poseFile)
				if err != nil {
					utils.Errorf("PosePath file: %s not find, error: %s", poseFile, err)
					return nil
				}
				setArg(imageData)
			}
		} else {
			randomNum := rand.Intn(groupNum)
			utils.Warnf("PosePath file: %s not find, try using random pose", poseFile)
			utils.Infof("PosePath file: %s to configuration", filepath.Join(poseDir, fmt.Sprintf("%s%d.png", poseKey, randomNum)))

			for index := 1; index < memberNum+1; index++ {
				poseFile = filepath.Join(poseDir, fmt.Sprintf("%s%d_%d.png", poseKey, randomNum, index))
				imageData, err := ioutil.ReadFile(poseFile)
				if err != nil {
					utils.Errorf("PosePath file: %s not find, error: %s", poseFile, err)
					return nil
				}
				setArg(imageData)
			}
		}

	}

	if len(args_ADetailer) > 0 {
		config.AddAlwaysonScripts("ADetailer", map[string]interface{}{
			"args": args_ADetailer,
		})
	}
	if len(args_ControlNet) > 0 {
		config.AddAlwaysonScripts("ControlNet", map[string]interface{}{
			"args": args_ControlNet,
		})
	}
	if len(args_LatentCouplet) > 0 {
		config.AddAlwaysonScripts("Latent Couple extension", map[string]interface{}{
			"args": args_LatentCouplet,
		})
	}
	if len(args_ComposableLora) > 0 {
		config.AddAlwaysonScripts("Composable Lora", map[string]interface{}{
			"args": args_ComposableLora,
		})
	}

	data := map[string]interface{}{
		"prompt":               extendTags + loraString + tagString,
		"negative_prompt":      negativeString,
		"seed":                 -1,
		"sampler_name":         "DPM++ 2M Karras",
		"cfg_scale":            7.5,
		"face_restoration":     "CodeFormer",
		"width":                512,
		"height":               512,
		"batch_size":           config.GetBatchSize(),
		"n_iter":               config.GetNiter(),
		"steps":                32,
		"return_grid":          true,
		"restore_faces":        true,
		"send_images":          true,
		"save_images":          false,
		"do_not_save_samples":  false,
		"do_not_save_grid":     false,
		"enable_hr":            false,
		"denoising_strength":   0.3,
		"firstphase_width":     0,
		"firstphase_height":    0,
		"hr_scale":             2,
		"hr_second_pass_steps": 0,
		"hr_resize_x":          0,
		"hr_resize_y":          0,
		"hr_checkpoint_name":   "",
		"hr_sampler_name":      "",
		"hr_prompt":            "",
		"hr_negative_prompt":   "",
		"override_settings": map[string]interface{}{
			"sd_model_checkpoint": config.GetMainModelName(),
			"sd_vae":              "Automatic",
		},
		"alwayson_scripts":                     config.GetAlwaysonScripts(),
		"override_settings_restore_afterwards": true,
	}

	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", SD_API_TXT2IMG, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Errorf("Function http.NewRequest error: %v", err)
		return nil
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		utils.Errorf("Request error: %v", err)
		return nil
	}
	if res.StatusCode != http.StatusOK {
		utils.Errorf("Response StatusCode: %d", res.StatusCode)

		body, _ := ioutil.ReadAll(res.Body)
		errorMap := make(map[string]interface{})
		err := json.Unmarshal(body, &errorMap)
		if err != nil {
			utils.Errorf("Function Unmarshal Error: %v", err)
			return nil
		}
		utils.Errorf("Response error: %v, %v", errorMap["error"], errorMap["detail"])

		return nil
	}
	return res
}

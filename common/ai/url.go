package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

func makeSDRequest(tagName, tagString string, config *ImageConfig, extendLoras ...string) *http.Response {
	loraString := ""
	extendTags := ""
	negativeString := strings.Join(GlobalTag_Negative, ",")

	if !config.IsManualSetup() {
		for l, c := range config.GetLoraConfig() {
			loraString += fmt.Sprintf("<lora:%s:%f>%s,", l, c.weight, strings.Join(c.tags, ","))
		}
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
	args_ComposableLora := []interface{}{}
	args_LatentCouplet := []interface{}{}
	args_AdditionalNetworks := []interface{}{}

	if config.GetFixface() {
		args_ADetailer = append(args_ADetailer, map[string]interface{}{
			"ad_model":               "face_yolov8n.pt",
			"ad_confidence":          0.3,
			"ad_dilate_erode":        -2,
			"ad_mask_blur":           2,
			"ad_denoising_strength":  0.35,
			"ad_inpaint_only_masked": true,
			"ad_use_separate_width":  false,
			"ad_inpaint_padding":     32,
			"ad_width":               512,
			"ad_height":              512,
		})
	}
	if config.GetFixhand() {
		args_ADetailer = append(args_ADetailer, map[string]interface{}{
			"ad_model":               "hand_yolov9c.pt",
			"ad_confidence":          0.4,
			"ad_dilate_erode":        4,
			"ad_mask_blur":           2,
			"ad_denoising_strength":  0.35,
			"ad_inpaint_only_masked": true,
			"ad_use_separate_width":  false,
			"ad_inpaint_padding":     32,
			"ad_width":               512,
			"ad_height":              512,
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
		if len(extendLoras) == 0 {
			args_LatentCouplet = append(args_LatentCouplet, []interface{}{
				true,
				"1:2,1:2,1:1", // divisions
				"0:1,0:0,0:1", // positions
				"0.2,0.2,0.2", // weights
				50,            // end_at_step
			}...)
			args_ComposableLora = append(args_ComposableLora, []interface{}{
				true, // enabled
				true, // Use Lora in uc text model encoder
				true, // Use Lora in uc diffusion model
			}...)
		}
	}
	if len(extendLoras) == 2 {
		loraString1 := config.GetAndLoraString(extendLoras[0], 1)
		loraString2 := config.GetAndLoraString(extendLoras[1], 2)
		loraString3 := config.GetAndLoraString(extendLoras[1], 3, false)
		tagString = loraString1 + loraString2 + loraString3

		args_AdditionalNetworks = append(args_AdditionalNetworks, true) // AddNet Enabled
		args_AdditionalNetworks = append(args_AdditionalNetworks, true) // Separate UNet/Text Encoder weights
		args_AdditionalNetworks = append(args_AdditionalNetworks,       // AddNet Module 1
			"LoRA",
			config.GetAndLoraConfigs()[extendLoras[0]].loraFullName,
			config.GetAndLoraConfigs()[extendLoras[0]].weight,
			config.GetAndLoraConfigs()[extendLoras[0]].weight,
		)
		args_AdditionalNetworks = append(args_AdditionalNetworks, // AddNet Module 2
			"LoRA",
			config.GetAndLoraConfigs()[extendLoras[1]].loraFullName,
			config.GetAndLoraConfigs()[extendLoras[1]].weight,
			config.GetAndLoraConfigs()[extendLoras[1]].weight,
		)
		args_AdditionalNetworks = append(args_AdditionalNetworks, // AddNet Module 3
			"LoRA",
			config.GetLoraConfigFirst().loraFullName,
			config.GetLoraConfigFirst().weight,
			config.GetLoraConfigFirst().weight,
		)
		args_AdditionalNetworks = append(args_AdditionalNetworks, // AddNet Module 4
			"LoRA", "None", 0, 0,
		)
		args_AdditionalNetworks = append(args_AdditionalNetworks, // AddNet Module 5
			"LoRA", "None", 0, 0,
		)
	}

	width := 512
	height := 512

	for poseName, poseConfig := range config.GetPoseConfigs() {
		poseDir := filepath.Join(config.GetBasePath(), "poses", poseName)
		entries, err := ioutil.ReadDir(poseDir)

		if err != nil {
			utils.Errorf("PosePath dir: %s not find, error: %s", poseDir, err)
			return nil
		}

		setArg := func(imageData string, index int) {
			if config.GetUseHigh() {
				if index == 1 || index == 2 || index == 3 {
					// TODO
					return
				}
			} else {
				if index == 0 || index == 3 {
					// TODO
					return
				}
			}

			ctype := "all"
			cmode := "Balanced"
			model := "None"
			guidance_start := 0.0
			guidance_end := 1.0
			weight := 1.0

			if index == 0 {
				guidance_start = 0
				guidance_end = 1
				weight = 0.6
				ctype = "depth"
				cmode = "Balanced"
				model = "control_v11f1p_sd15_depth [cfd03158]"
			} else if index == 1 {
				guidance_start = 0.0
				guidance_end = 0.40
				if config.GetLoraIndex() == 0 {
					weight = 1.1
				} else {
					weight = 1.0
				}
				ctype = "openPose"
				cmode = "Balanced"
				model = "control_v11p_sd15_openpose [cab727d4]"
			} else if index == 2 {
				guidance_start = 0.35
				guidance_end = 1.0
				if config.GetLoraIndex() == 0 {
					weight = 1.0
				} else {
					weight = 1.1
				}
				ctype = "openPose"
				cmode = "Balanced"
				model = "control_v11p_sd15_openpose [cab727d4]"
			} else if index == 3 {
				guidance_start = 0
				guidance_end = 0.85
				weight = 1.0
				ctype = "openPose"
				cmode = "ControlNet is more important"
				model = "control_v11p_sd15_openpose [cab727d4]"
			}

			args_ControlNet = append(args_ControlNet, map[string]interface{}{
				"enabled":                  true,
				"preprocessor":             "none",
				"pixel_perfect":            true,
				"allow_preview":            false,
				"low_vram":                 false,
				"control_type":             ctype,
				"control_mode":             cmode,
				"inpaint_crop_input_image": true,
				"model":                    model,
				"weight":                   weight,
				"resize_mode":              "Crop and Resize",
				"hr_option":                "Both",
				"threshold_a":              0.5,
				"threshold_b":              0.5,
				"guidance_start":           guidance_start,
				"guidance_end":             guidance_end,
				"processor_res":            512,
				"image":                    imageData,
			})
		}

		var groupNum, memberNum int
		visited := make(map[string]bool)
		for _, file := range entries {
			if ext := filepath.Ext(file.Name()); ext != ".png" {
				continue
			}
			filename := file.Name()
			ext := filepath.Ext(filename)
			nameWithoutExt := strings.TrimSuffix(filename, ext)
			parts := strings.Split(nameWithoutExt, "_")
			nameFinal := parts[0]

			if !visited[nameFinal] {
				groupNum++
				memberNum = 0
				visited[nameFinal] = true
			} else {
				memberNum++
			}
		}

		poseConfig.GroupNum = groupNum
		poseConfig.MemberNum = memberNum
		poseKey := poseConfig.PoseKey

		getImageBase64 := func(imageData []byte) string {
			prefix := "data:image/png;base64,"
			base64Str := base64.StdEncoding.EncodeToString(imageData)
			fullBase64 := prefix + base64Str
			return fullBase64
		}

		poseFile := filepath.Join(poseDir, poseKey+".png")
		if _, err := os.Stat(poseFile); err == nil {
			utils.Infof("PosePath file: %s to configuration", poseFile)

			for index := 0; index < memberNum-1; index++ {
				poseFile = filepath.Join(poseDir, fmt.Sprintf("%s_%d.png", poseKey, index))
				imageData, err := ioutil.ReadFile(poseFile)
				if err != nil {
					utils.Errorf("PosePath file: %s not find, error: %s", poseFile, err)
					return nil
				}
				setArg(getImageBase64(imageData), index)
			}
			poseFile = filepath.Join(poseDir, fmt.Sprintf("%s_%d.png", poseKey, 4))
		} else {
			randomNum := rand.Intn(groupNum)
			utils.Warnf("PosePath file: %s not find, try using random pose", poseFile)
			utils.Infof("PosePath file: %s to configuration", filepath.Join(poseDir, fmt.Sprintf("%s%d.png", poseKey, randomNum)))

			for index := 0; index < memberNum-1; index++ {
				poseFile = filepath.Join(poseDir, fmt.Sprintf("%s%d_%d.png", poseKey, randomNum, index))
				imageData, err := ioutil.ReadFile(poseFile)
				if err != nil {
					utils.Errorf("PosePath file: %s not find, error: %s", poseFile, err)
					return nil
				}
				setArg(getImageBase64(imageData), index)
			}
			poseFile = filepath.Join(poseDir, fmt.Sprintf("%s%d_%d.png", poseKey, randomNum, 4))
		}

		if len(args_AdditionalNetworks) > 0 {
			if _, err := os.Stat(poseFile); err == nil {
				cmd := exec.Command("python",
					filepath.Join(config.GetBasePath(), "scripts", "image_to_numpy.py"), poseFile)
				output, err := cmd.CombinedOutput()
				if err != nil {
					utils.Errorf("Python script execution error: %s", err)
					utils.Errorf("Output: %s", string(output))
					return nil
				}

				imageFile, _ := os.Open(poseFile)
				defer imageFile.Close()
				img, _, err := image.Decode(imageFile)
				if err != nil {
					utils.Errorf("Image decode error: %s", err)
					return nil
				}
				bounds := img.Bounds()

				checkImage := func() [][][]float32 {
					file, err := os.Open("image_data.csv")
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()

					reader := csv.NewReader(file)
					records, err := reader.ReadAll()
					if err != nil {
						log.Fatal(err)
					}

					width, height = bounds.Dx(), bounds.Dy()
					data := make([][][]float32, height)
					for i := range data {
						data[i] = make([][]float32, width)
						for j := range data[i] {
							data[i][j] = make([]float32, 3)
							for k := 0; k < 3; k++ {
								val, _ := strconv.Atoi(records[i*width+j][k])
								data[i][j][k] = float32(val)
							}
						}
					}
					return data
				}

				byteData := checkImage()
				args_AdditionalNetworks = append(args_AdditionalNetworks, byteData)
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
	if len(args_AdditionalNetworks) > 0 {
		config.AddAlwaysonScripts("Additional networks for generating", map[string]interface{}{
			"args": args_AdditionalNetworks,
		})
	}

	seed := config.GetSeed()
	utils.Infof("Random seed: %d", seed)

	data := map[string]interface{}{
		"prompt":                 extendTags + loraString + tagString,
		"negative_prompt":        negativeString,
		"seed":                   seed,
		"sampler_name":           "DPM++ SDE Karras",
		"cfg_scale":              1,
		"clip_skip":              2,
		"width":                  width,
		"height":                 height,
		"batch_size":             config.GetBatchSize(),
		"n_iter":                 config.GetNiter(),
		"steps":                  40,
		"return_grid":            true,
		"restore_faces":          false,
		"face_restoration":       "CodeFormer",
		"face_restoration_model": "null",
		"send_images":            true,
		"save_images":            false,
		"do_not_save_samples":    false,
		"do_not_save_grid":       false,
		"enable_hr":              false,
		"enable-checkbox":        false,
		"switch_at":              0.8,
		"denoising_strength":     0.3,
		"firstphase_width":       0,
		"firstphase_height":      0,
		"hr_scale":               2,
		"hr_second_pass_steps":   0,
		"hr_resize_x":            0,
		"hr_resize_y":            0,
		"hires_steps":            0,
		"hr_checkpoint_name":     "",
		"hr_sampler_name":        "",
		"hr_prompt":              "",
		"hr_negative_prompt":     "",
		"hr-checkbox":            false,
		"hr_upscaler":            "Latent",
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

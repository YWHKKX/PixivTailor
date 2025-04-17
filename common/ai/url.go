package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

// modle download from: https://civitai.com/

var SD_API_TXT2IMG string = "http://127.0.0.1:7860/sdapi/v1/txt2img"

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

func makeSDRequest(url, tag string, config imageConfig) *http.Response {
	loraString := ""
	for l, f := range config.GetLoraModel() {
		loraString += fmt.Sprintf("<lora:%s:%f>,", l, f)
	}

	data := map[string]interface{}{
		"prompt":               loraString + tag,
		"negative_prompt":      "EasyNegative,badhandsv5-neg,Subtitles,word,",
		"seed":                 -1,
		"sampler_name":         "DPM++ SDE",
		"cfg_scale":            7.5,
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
			"sd_model_checkpoint": config.GetModel(),
			"sd_vae":              "Automatic",
		},
		"alwayson_scripts":                     config.GetAlwaysonScripts(),
		"override_settings_restore_afterwards": true,
	}

	jsonData, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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
		utils.Errorf("Request StatusCode: %d", res.StatusCode)
		return nil
	}
	return res
}

package main

import (
	"os"
	"path/filepath"

	"github.com/GolangProject/PixivCrawler/common/ai"
)

func main() {
	config := ai.InitConfig("chosenMix_bakedVae.safetensors [52b8ebbd5b]")
	config.SetBatchSize(1)
	config.SetNiter(1)
	config.SetShowTags(true)

	currntPath, _ := os.Getwd()
	config.SetInputPath(filepath.Join(currntPath, "images"))
	config.SetOutputPath(filepath.Join(currntPath, "images"))

	config.AddAlwaysonScripts("ADetailer", map[string]interface{}{
		"args": []map[string]interface{}{
			{
				"ad_model":               "face_yolov8n.pt",
				"ad_confidence":          0.3,
				"ad_dilate_erode":        4,
				"ad_mask_blur":           4,
				"ad_denoising_strength":  0.4,
				"ad_inpaint_only_masked": true,
				"ad_inpaint_padding":     32,
				"ad_version":             "25.3.0",
			},
		},
	})
	// config.AddAlwaysonScripts("controlnet", map[string]interface{}{
	// 	"args": []map[string]interface{}{
	// 		{
	// 			"enabled":           true,
	// 			"control_mode":      0,
	// 			"model":             "t2i-adapter_diffusers_xl_lineart [bae0efef]",
	// 			"module":            "lineart_standard (from white bg & black line)",
	// 			"weight":            0.45,
	// 			"resize_mode":       "Crop and Resize",
	// 			"threshold_a":       200,
	// 			"threshold_b":       245,
	// 			"guidance_start":    0,
	// 			"guidance_end":      0.7,
	// 			"pixel_perfect":     true,
	// 			"processor_res":     512,
	// 			"save_detected_map": true,
	// 			"input_image":       "",
	// 		},
	// 	},
	// })

	ai.SaveTagsFormImage(config)
	ai.Tailor_TXT2IMG(config)
}

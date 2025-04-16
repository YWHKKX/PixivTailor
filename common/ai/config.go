package ai

import (
	"os"
	"path/filepath"
)

type SDRequest struct {
	modle              string
	batch_size, n_iter int
	alwaysonScripts    map[string]interface{}
}

type SDDownload struct {
	saveName   string
	basePath   string
	inputPath  string
	outputPath string
}

type config struct {
	sdRequestConfig  SDRequest
	sdDownloadConfig SDDownload
	showTags         bool
}

// stable-diffusion-webui: https://github.com/AUTOMATIC1111/stable-diffusion-webui
// modle download from: https://civitai.com/
func InitConfig(modle string) config {
	basePath, _ := os.Getwd()

	inputPath := filepath.Join(basePath, "images")
	outputPath := filepath.Join(basePath, "images")

	return config{
		sdRequestConfig: SDRequest{
			modle:           modle,
			batch_size:      1,
			n_iter:          1,
			alwaysonScripts: make(map[string]interface{}),
		},
		sdDownloadConfig: SDDownload{
			basePath:   basePath,
			inputPath:  inputPath,
			outputPath: outputPath,
			saveName:   "NewImage",
		},
		showTags: false,
	}
}

func (c *config) GetModel() string {
	return c.sdRequestConfig.modle
}

func (c *config) GetBasePath() string {
	return c.sdDownloadConfig.basePath
}

func (c *config) SetBasePath(basePath string) {
	c.sdDownloadConfig.basePath = basePath
}

func (c *config) GetOutputPath() string {
	return c.sdDownloadConfig.outputPath
}

func (c *config) SetOutputPath(outputPath string) {
	c.sdDownloadConfig.outputPath = outputPath
}

func (c *config) GetInputPath() string {
	return c.sdDownloadConfig.inputPath
}

func (c *config) SetInputPath(inputPath string) {
	c.sdDownloadConfig.inputPath = inputPath
}

func (c *config) SetSaveName(saveName string) {
	c.sdDownloadConfig.saveName = saveName
}

func (c *config) GetSaveName() string {
	return c.sdDownloadConfig.saveName
}

func (c *config) SetBatchSize(batch_size int) {
	c.sdRequestConfig.batch_size = batch_size
}

func (c *config) SetNiter(n_iter int) {
	c.sdRequestConfig.n_iter = n_iter
}

func (c *config) GetBatchSize() int {
	return c.sdRequestConfig.batch_size
}

func (c *config) GetNiter() int {
	return c.sdRequestConfig.n_iter
}

func (c *config) GetShowTags() bool {
	return c.showTags
}

func (c *config) SetShowTags(showTags bool) {
	c.showTags = showTags
}

func (c *config) GetAlwaysonScripts() map[string]interface{} {
	return c.sdRequestConfig.alwaysonScripts
}

func (c *config) SetAlwaysonScripts(alwayson_scripts map[string]interface{}) {
	c.sdRequestConfig.alwaysonScripts = alwayson_scripts
}

func (c *config) AddAlwaysonScripts(key string, script interface{}) {
	if c.sdRequestConfig.alwaysonScripts == nil {
		c.sdRequestConfig.alwaysonScripts = make(map[string]interface{})
	}
	c.sdRequestConfig.alwaysonScripts[key] = script
}

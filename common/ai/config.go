package ai

import (
	"os"
	"path/filepath"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type SDRequest struct {
	modle              string
	loraModle          map[string]float64
	batch_size, n_iter int
	alwaysonScripts    map[string]interface{}
}

type SDDownload struct {
	saveName   string
	basePath   string
	inputPath  string
	outputPath string
}

type imageConfig struct {
	sdRequestConfig  SDRequest
	sdDownloadConfig SDDownload
	showTags         bool
}

// stable-diffusion-webui: https://github.com/AUTOMATIC1111/stable-diffusion-webui
// webui.bat --api

func InitImageConfig(modle string, basePaths ...string) imageConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	inputPath := filepath.Join(basePath, "images")
	outputPath := filepath.Join(basePath, "images")

	return imageConfig{
		sdRequestConfig: SDRequest{
			modle:           modle,
			batch_size:      1,
			n_iter:          1,
			alwaysonScripts: make(map[string]interface{}),
			loraModle:       make(map[string]float64),
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

func (c *imageConfig) GetModel() string {
	return c.sdRequestConfig.modle
}

func (c *imageConfig) SetLoraModel(loraModle map[string]float64) {
	c.sdRequestConfig.loraModle = loraModle
}

func (c *imageConfig) GetLoraModel() map[string]float64 {
	return c.sdRequestConfig.loraModle
}

func (c *imageConfig) AddLoraModel(n string, f float64) {
	c.sdRequestConfig.loraModle[n] = f
}

func (c *imageConfig) GetBasePath() string {
	return c.sdDownloadConfig.basePath
}

func (c *imageConfig) GetOutputPath() string {
	return c.sdDownloadConfig.outputPath
}

func (c *imageConfig) SetOutputPath(outputPath string) {
	c.sdDownloadConfig.outputPath = outputPath
}

func (c *imageConfig) GetInputPath() string {
	return c.sdDownloadConfig.inputPath
}

func (c *imageConfig) SetInputPath(inputPath string) {
	c.sdDownloadConfig.inputPath = inputPath
}

func (c *imageConfig) SetSaveName(saveName string) {
	c.sdDownloadConfig.saveName = saveName
}

func (c *imageConfig) GetSaveName() string {
	return c.sdDownloadConfig.saveName
}

func (c *imageConfig) SetBatchSize(batch_size int) {
	c.sdRequestConfig.batch_size = batch_size
}

func (c *imageConfig) SetNiter(n_iter int) {
	c.sdRequestConfig.n_iter = n_iter
}

func (c *imageConfig) GetBatchSize() int {
	return c.sdRequestConfig.batch_size
}

func (c *imageConfig) GetNiter() int {
	return c.sdRequestConfig.n_iter
}

func (c *imageConfig) GetShowTags() bool {
	return c.showTags
}

func (c *imageConfig) SetShowTags(showTags bool) {
	c.showTags = showTags
}

func (c *imageConfig) GetAlwaysonScripts() map[string]interface{} {
	return c.sdRequestConfig.alwaysonScripts
}

func (c *imageConfig) SetAlwaysonScripts(alwayson_scripts map[string]interface{}) {
	c.sdRequestConfig.alwaysonScripts = alwayson_scripts
}

func (c *imageConfig) AddAlwaysonScripts(key string, script interface{}) {
	if c.sdRequestConfig.alwaysonScripts == nil {
		c.sdRequestConfig.alwaysonScripts = make(map[string]interface{})
	}
	c.sdRequestConfig.alwaysonScripts[key] = script
}

type modelConfig struct {
	modelName      string // output_name
	pretrainedPath string // pretrained_model_name_or_path
	inputDir       string // train_data_dir
	outputDir      string // output_dir
	logDir         string // logging_dir
}

type tagConfig struct {
	tagName, tagSrcPath string
	times               int
}

type trainConfig struct {
	basePath, examplePath string
	limit                 int
	modelConfig           modelConfig
	tagConfigs            map[string]tagConfig
}

// kohya-ss: https://github.com/kohya-ss
// gui.bat --listen 127.0.0.1 --server_port 7860 --inbrowser --share

func NewTrainConfig(modelName, pretrainedPath, inputDir string, basePaths ...string) trainConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}
	outputDir := filepath.Join(basePath, "models", modelName)
	logDir := filepath.Join(basePath, "logs", modelName)

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		utils.Errorf("Function os.MkdirAll error: %v", err)
		panic(err)
	}
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		utils.Errorf("Function os.MkdirAll error: %v", err)
		panic(err)
	}

	return trainConfig{
		modelConfig: modelConfig{
			modelName:      modelName,
			pretrainedPath: filepath.Join(pretrainedPath),
			inputDir:       filepath.Join(inputDir),
			outputDir:      outputDir,
			logDir:         logDir,
		},
		tagConfigs:  make(map[string]tagConfig),
		limit:       0,
		basePath:    basePath,
		examplePath: filepath.Join(basePath, "scripts", "example.json"),
	}
}

func (c *trainConfig) GetModelName() string {
	return c.modelConfig.modelName
}

func (c *trainConfig) GetPretrainedPath() string {
	return c.modelConfig.pretrainedPath
}

func (c *trainConfig) GetInputDir() string {
	return c.modelConfig.inputDir
}

func (c *trainConfig) GetOutputDir() string {
	return c.modelConfig.outputDir
}

func (c *trainConfig) GetLogDir() string {
	return c.modelConfig.logDir
}

func (c *trainConfig) GetExamplePath() string {
	return c.examplePath
}

func (c *trainConfig) SetExamplePath(examplePath string) {
	c.examplePath = examplePath
}

func (c *trainConfig) GetBasePath() string {
	return c.basePath
}

func (c *trainConfig) SetBasePath(basePath string) {
	c.basePath = basePath
}

func (c *trainConfig) GetTagConfigs() map[string]tagConfig {
	return c.tagConfigs
}

func (c *trainConfig) SetTagConfigs(tagConfigs map[string]tagConfig) {
	c.tagConfigs = tagConfigs
}

func (c *trainConfig) AddTagConfig(tagName, tagPath string, times int) {
	c.tagConfigs[tagName] = tagConfig{
		tagName:    tagName,
		tagSrcPath: tagPath,
		times:      times,
	}
}

func (c *trainConfig) GetLimit() int {
	return c.limit
}

func (c *trainConfig) SetLimit(limit int) {
	c.limit = limit
}

func (c *trainConfig) CheckLimit(index int) bool {
	if index >= c.GetLimit() && c.GetLimit() != 0 {
		return false
	}
	return true
}

func (t *tagConfig) GetTagName() string {
	return t.tagName
}

func (t *tagConfig) GetTagSrcPath() string {
	return t.tagSrcPath
}

func (t *tagConfig) GetTimes() int {
	return t.times
}

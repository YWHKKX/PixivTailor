package ai

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type SDRequest struct {
	mainModleName      string
	loraModles         map[string]float64
	extendTags         map[string]string
	batch_size, n_iter int
	alwaysonScripts    map[string]interface{}
}

type SDDownload struct {
	saveName   string
	basePath   string
	inputPath  string
	outputPath string
	forEach    bool
	saveType   SaveType
	tagConfigs map[string]tagConfig
}

type AnalyzeType int

const (
	Analyze_Deepdanbooru AnalyzeType = iota
	Analyze_Webuiwd14tagger
)

type SaveType string

const (
	Save_Txt  SaveType = "txt"
	Save_Json SaveType = "json"
)

type ImageConfig struct {
	sdRequestConfig  SDRequest
	sdDownloadConfig SDDownload
	showTags         bool
	analyzeType      AnalyzeType
}

// stable-diffusion-webui: https://github.com/AUTOMATIC1111/stable-diffusion-webui
// webui.bat --api

func InitImageConfig(mainModle string, basePaths ...string) ImageConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	inputPath := filepath.Join(basePath, "images")
	outputPath := filepath.Join(basePath, "images")

	return ImageConfig{
		sdRequestConfig: SDRequest{
			mainModleName:   mainModle,
			batch_size:      1,
			n_iter:          1,
			alwaysonScripts: make(map[string]interface{}),
			loraModles:      make(map[string]float64),
			extendTags:      make(map[string]string),
		},
		sdDownloadConfig: SDDownload{
			basePath:   basePath,
			inputPath:  inputPath,
			outputPath: outputPath,
			saveName:   "NewImage",
			forEach:    false,
			saveType:   Save_Json,
			tagConfigs: make(map[string]tagConfig),
		},
		showTags:    false,
		analyzeType: Analyze_Deepdanbooru,
	}
}

func (c *ImageConfig) GetMainModelName() string {
	return c.sdRequestConfig.mainModleName
}

func (c *ImageConfig) SetLoraModel(loraModle map[string]float64) {
	c.sdRequestConfig.loraModles = loraModle
}

func (c *ImageConfig) GetLoraModel() map[string]float64 {
	return c.sdRequestConfig.loraModles
}

func (c *ImageConfig) AddLoraModel(n string, f float64) {
	c.sdRequestConfig.loraModles[n] = f
}

func (c *ImageConfig) SetExtendTags(extendTags map[string]string) {
	c.sdRequestConfig.extendTags = extendTags
}

func (c *ImageConfig) AddtExtendTag(n string, e string) {
	c.sdRequestConfig.extendTags[n] = e
}

func (c *ImageConfig) GetExtendTags() map[string]string {
	return c.sdRequestConfig.extendTags
}

func (c *ImageConfig) GetBasePath() string {
	return c.sdDownloadConfig.basePath
}

func (c *ImageConfig) GetOutputPath() string {
	return c.sdDownloadConfig.outputPath
}

func (c *ImageConfig) SetOutputPath(outputPath string) {
	c.sdDownloadConfig.outputPath = outputPath
}

func (c *ImageConfig) GetInputPath() string {
	return c.sdDownloadConfig.inputPath
}

func (c *ImageConfig) SetInputPath(inputPath string) {
	c.sdDownloadConfig.inputPath = inputPath
}

func (c *ImageConfig) SetSaveName(saveName string) {
	c.sdDownloadConfig.saveName = saveName
}

func (c *ImageConfig) GetSaveName() string {
	return c.sdDownloadConfig.saveName
}

func (c *ImageConfig) SetBatchSize(batch_size int) {
	c.sdRequestConfig.batch_size = batch_size
}

func (c *ImageConfig) SetNiter(n_iter int) {
	c.sdRequestConfig.n_iter = n_iter
}

func (c *ImageConfig) GetBatchSize() int {
	return c.sdRequestConfig.batch_size
}

func (c *ImageConfig) GetNiter() int {
	return c.sdRequestConfig.n_iter
}

func (c *ImageConfig) GetShowTags() bool {
	return c.showTags
}

func (c *ImageConfig) SetShowTags(showTags bool) {
	c.showTags = showTags
}

func (c *ImageConfig) GetAlwaysonScripts() map[string]interface{} {
	return c.sdRequestConfig.alwaysonScripts
}

func (c *ImageConfig) SetAlwaysonScripts(alwayson_scripts map[string]interface{}) {
	c.sdRequestConfig.alwaysonScripts = alwayson_scripts
}

func (c *ImageConfig) AddAlwaysonScripts(key string, script interface{}) {
	if c.sdRequestConfig.alwaysonScripts == nil {
		c.sdRequestConfig.alwaysonScripts = make(map[string]interface{})
	}
	c.sdRequestConfig.alwaysonScripts[key] = script
}

func (c *ImageConfig) GetAnalyzeType() AnalyzeType {
	return c.analyzeType
}

func (c *ImageConfig) SetAnalyzeType(analyzeType AnalyzeType) {
	c.analyzeType = analyzeType
}

func (c *ImageConfig) GetSaveType() SaveType {
	return c.sdDownloadConfig.saveType
}

func (c *ImageConfig) SetSaveType(saveType SaveType) {
	c.sdDownloadConfig.saveType = saveType
}

func (c *ImageConfig) SetForEach(b bool) {
	c.sdDownloadConfig.forEach = b
}

func (c *ImageConfig) IsForEach() bool {
	return c.sdDownloadConfig.forEach
}

func (c *ImageConfig) GetTagConfigs() map[string]tagConfig {
	return c.sdDownloadConfig.tagConfigs
}

func (c *ImageConfig) SetTagConfigs(tagConfigs map[string]tagConfig) {
	c.sdDownloadConfig.tagConfigs = tagConfigs
}

func (c *ImageConfig) AddTagConfig(tagName, tagPath string) {
	c.sdDownloadConfig.tagConfigs[tagName] = tagConfig{
		tagName:    tagName,
		tagSrcPath: tagPath,
		times:      0,
	}
}

func (c *ImageConfig) CheckTagConfig(path string) string {
	basePath := strings.TrimSuffix(path, filepath.Base(path))
	for _, e := range c.sdDownloadConfig.tagConfigs {
		if filepath.Clean(e.tagSrcPath) == filepath.Clean(basePath) {
			return e.tagName
		}
	}
	return ""
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

type TrainConfig struct {
	basePath, examplePath string
	limit                 int
	modelConfig           modelConfig
	tagConfigs            map[string]tagConfig
}

// kohya-ss: https://github.com/kohya-ss
// gui.bat --listen 127.0.0.1 --server_port 7860 --inbrowser --share

func NewTrainConfig(modelName, pretrainedPath, inputDir string, basePaths ...string) TrainConfig {
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

	return TrainConfig{
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

func (c *TrainConfig) GetModelName() string {
	return c.modelConfig.modelName
}

func (c *TrainConfig) GetPretrainedPath() string {
	return c.modelConfig.pretrainedPath
}

func (c *TrainConfig) GetInputDir() string {
	return c.modelConfig.inputDir
}

func (c *TrainConfig) GetOutputDir() string {
	return c.modelConfig.outputDir
}

func (c *TrainConfig) GetLogDir() string {
	return c.modelConfig.logDir
}

func (c *TrainConfig) GetExamplePath() string {
	return c.examplePath
}

func (c *TrainConfig) SetExamplePath(examplePath string) {
	c.examplePath = examplePath
}

func (c *TrainConfig) GetBasePath() string {
	return c.basePath
}

func (c *TrainConfig) SetBasePath(basePath string) {
	c.basePath = basePath
}

func (c *TrainConfig) GetTagConfigs() map[string]tagConfig {
	return c.tagConfigs
}

func (c *TrainConfig) SetTagConfigs(tagConfigs map[string]tagConfig) {
	c.tagConfigs = tagConfigs
}

func (c *TrainConfig) AddTagConfig(tagName, tagPath string, times int) {
	c.tagConfigs[tagName] = tagConfig{
		tagName:    tagName,
		tagSrcPath: tagPath,
		times:      times,
	}
}

func (c *TrainConfig) GetLimit() int {
	return c.limit
}

func (c *TrainConfig) SetLimit(limit int) {
	c.limit = limit
}

func (c *TrainConfig) CheckLimit(index int) bool {
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

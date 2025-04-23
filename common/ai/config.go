package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GolangProject/PixivCrawler/common/utils"
)

type poseConfig struct {
	PoseKey             string
	GroupNum, MemberNum int
}

type SDRequest struct {
	mainModleName             string
	batch_size, n_iter        int
	fixpose, fixhand, fixface bool
	posePaths                 map[string]*poseConfig
	alwaysonScripts           map[string]interface{}
}

type SDDownload struct {
	savePathName string
	basePath     string
	inputPath    string
	outputPath   string
	forEach      bool
	saveType     SaveType
	tagConfigs   map[string]tagConfig
	loraModles   map[string]float64
	loraTags     map[string][]string
	extendTags   map[string][]string
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
	deleteTags       bool
	containPathName  []string
	ignorePathName   []string
	skipTags         []string
	analyzeType      AnalyzeType
}

// stable-diffusion-webui: https://github.com/AUTOMATIC1111/stable-diffusion-webui
// webui.bat --api

func InitImageConfig(mainModle string, basePaths ...string) *ImageConfig {
	basePath, _ := os.Getwd()
	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}

	tagGlobalPath := filepath.Join(basePath, "scripts", "global_tag.json")
	if _, err := os.Stat(tagGlobalPath); err == nil {
		categorys := InitCategories()
		utils.Warnf("GlobalTag exists: %s, start configuration", tagGlobalPath)
		fileData, err := os.ReadFile(tagGlobalPath)
		if err != nil {
			utils.Errorf("Read File Error: %v", err)
		}
		if err := json.Unmarshal(fileData, &categorys); err != nil {
			utils.Errorf("Function Unmarshal Error: %v", err)
		}
		for _, category := range categorys {
			if category.Kind == CategoryType_Clothing {
				GlobalTag_Clothing = append(GlobalTag_Clothing, category.Keywords...)
			}
			if category.Kind == CategoryType_Character {
				GlobalTag_Character = append(GlobalTag_Character, category.Keywords...)
			}
			if category.Kind == CategoryType_Background {
				GlobalTag_Background = append(GlobalTag_Background, category.Keywords...)
			}
			if category.Kind == CategoryType_Pose {
				GlobalTag_Pose = append(GlobalTag_Pose, category.Keywords...)
			}
		}
	}

	inputPath := filepath.Join(basePath, "images")
	outputPath := filepath.Join(basePath, "images")

	return &ImageConfig{
		sdRequestConfig: SDRequest{
			mainModleName:   mainModle,
			batch_size:      1,
			n_iter:          1,
			alwaysonScripts: make(map[string]interface{}),
			fixhand:         false,
			fixface:         false,
			posePaths:       make(map[string]*poseConfig),
		},
		sdDownloadConfig: SDDownload{
			basePath:     basePath,
			inputPath:    inputPath,
			outputPath:   outputPath,
			savePathName: "NewImage",
			forEach:      false,
			saveType:     Save_Json,
			tagConfigs:   make(map[string]tagConfig),
			loraModles:   make(map[string]float64),
			loraTags:     make(map[string][]string),
			extendTags:   make(map[string][]string),
		},
		showTags:    false,
		deleteTags:  false,
		analyzeType: Analyze_Deepdanbooru,
		skipTags:    []string{},
	}
}

func (c *ImageConfig) GetMainModelName() string {
	return c.sdRequestConfig.mainModleName
}

func (c *ImageConfig) GetLoraModel() map[string]float64 {
	return c.sdDownloadConfig.loraModles
}

func (c *ImageConfig) AddLoraModel(n string, f float64, tags []string) {
	c.sdDownloadConfig.loraModles[n] = f
	c.sdDownloadConfig.loraTags[n] = tags
}

func (c *ImageConfig) GetLoraTags(name string) []string {
	return c.sdDownloadConfig.loraTags[name]
}

func (c *ImageConfig) AddExtendTags(extendTags []string, names ...string) {
	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	c.sdDownloadConfig.extendTags[name] = extendTags
}

func (c *ImageConfig) GetExtendTags(names ...string) []string {
	name := ""
	if len(names) > 0 {
		name = names[0]
	}
	return c.sdDownloadConfig.extendTags[name]
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

func (c *ImageConfig) SetSavePathName(saveName string) {
	c.sdDownloadConfig.savePathName = saveName
}

func (c *ImageConfig) GetSavePathName() string {
	return c.sdDownloadConfig.savePathName
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

func (c *ImageConfig) GetDeleteTags() bool {
	return c.deleteTags
}

func (c *ImageConfig) SetDeleteTags(deleteTags bool) {
	c.deleteTags = deleteTags
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

func (c *ImageConfig) GetFixhand() bool {
	return c.sdRequestConfig.fixhand
}

func (c *ImageConfig) SetFixhand(fixhand bool) {
	if c.GetFixpose() {
		utils.Warn("Fixpose has already been set, it is not recommended to set FixFace or FixHand")
	}
	c.sdRequestConfig.fixhand = fixhand
}

func (c *ImageConfig) GetFixface() bool {
	return c.sdRequestConfig.fixface
}

func (c *ImageConfig) SetFixface(fixface bool) {
	if c.GetFixpose() {
		utils.Warn("Fixpose has already been set, it is not recommended to set FixFace or FixHand")
	}
	c.sdRequestConfig.fixface = fixface
}

func (c *ImageConfig) GetFixpose() bool {
	return c.sdRequestConfig.fixpose
}

func (c *ImageConfig) SetFixpose(fixpose bool) {
	if c.GetFixface() || c.GetFixhand() {
		utils.Warnf("FixFace or FixHand has already been set, it is not recommended to set Fixpose")
	}

	c.sdRequestConfig.fixpose = fixpose
}

func (c *ImageConfig) GetPoseConfigs() map[string]*poseConfig {
	return c.sdRequestConfig.posePaths
}

func (c *ImageConfig) AddPoseConfig(k, v string) {
	if c.GetFixpose() {
		utils.Warn("Fixpose has already been set, adding a new PoseConfig may cause slow image generation")
	}
	if p, ok := c.sdRequestConfig.posePaths[k]; ok {
		p.PoseKey = v
	} else {
		c.sdRequestConfig.posePaths[k] = &poseConfig{
			PoseKey: v,
		}
	}
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

func (c *ImageConfig) AddContainPathName(path string) {
	c.containPathName = append(c.containPathName, path)
}

func (c *ImageConfig) AddIgnorePathName(path string) {
	c.ignorePathName = append(c.ignorePathName, path)
}

func (c *ImageConfig) CheckPathFilter(path string) bool {
	tagPath := strings.TrimSuffix(path, filepath.Base(path))
	tagName := filepath.Base(tagPath)

	for _, i := range c.ignorePathName {
		if i == tagName {
			return false
		}
	}

	if len(c.containPathName) == 0 {
		return true
	}

	for _, i := range c.containPathName {
		if i == tagName {
			return true
		}
	}
	return false
}

func (c *ImageConfig) SetSkipTags(skipTags []string) {
	c.skipTags = skipTags
}

func (c *ImageConfig) AddSkipTags(skipTags []string) {
	c.skipTags = append(c.skipTags, skipTags...)
}

func (c *ImageConfig) CheckSkipTags(input string) bool {
	skipTags := strings.Join(c.skipTags, "|")
	if skipTags == "" {
		return false
	}
	pattern := `(?i)\b(` + skipTags + `)\b`
	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		utils.Errorf("Regex compilation error: ", err)
		return false
	}
	return matched
}

type TrainType int

const (
	TrainSpeedAuto   TrainType = iota
	TrainSpeedSlow             // 100 images ≈ 7h
	TrainSpeedFast             // 100 images ≈ 3h
	TrainQualityLow            // 100 images ≈ 1h
	TrainQualityHigh           // 100 images ≈ 8h
)

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
	basePath      string
	limit         int
	modelConfig   modelConfig
	trainType     TrainType
	trainImageNum int
	tagConfigs    map[string]tagConfig
}

// kohya-ss: https://github.com/kohya-ss
// gui.bat --listen 127.0.0.1 --server_port 7860 --inbrowser --share

func NewTrainConfig(modelName, pretrainedPath, inputDir string, basePaths ...string) *TrainConfig {
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

	return &TrainConfig{
		modelConfig: modelConfig{
			modelName:      modelName,
			pretrainedPath: filepath.Join(pretrainedPath),
			inputDir:       filepath.Join(inputDir),
			outputDir:      outputDir,
			logDir:         logDir,
		},
		tagConfigs:    make(map[string]tagConfig),
		limit:         0,
		basePath:      basePath,
		trainType:     TrainSpeedAuto,
		trainImageNum: 0,
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

func (t *TrainConfig) GetTrainType() TrainType {
	if t.trainType == TrainSpeedAuto {
		if t.trainImageNum < 100 {
			utils.Info("The training set is less than 100 images, use the TrainSpeedFast mode")
			return TrainSpeedFast
		} else {
			utils.Info("The training set is more than 100 images, use the TrainSpeedSlow mode")
			return TrainSpeedSlow
		}
	}
	return t.trainType
}

func (t *TrainConfig) SetTrainType(trainType TrainType) {
	t.trainType = trainType
}

func (t *TrainConfig) UpTrainImageNum(num int) {
	t.trainImageNum += num
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

type CategoryConfig struct {
	keyString                       []string
	keyIndex                        int
	basePath                        string
	outputPath                      string
	containPathName, ignorePathName []string
	showTags                        bool
}

// openai: https://platform.openai.com/settings/organization/api-keys

func NewCategoryConfig(keyString []string, outputPath string, basePaths ...string) *CategoryConfig {
	basePath, _ := os.Getwd()

	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}
	return &CategoryConfig{
		keyString:  keyString,
		keyIndex:   0,
		basePath:   basePath,
		outputPath: filepath.Join(basePath, outputPath),
		showTags:   false,
	}
}

func (c *CategoryConfig) GetKeyString() string {
	c.keyIndex++
	if c.keyIndex >= len(c.keyString) {
		c.keyIndex = 0
	}
	return c.keyString[c.keyIndex]
}

func (c *CategoryConfig) GetBasePath() string {
	return c.basePath
}

func (c *CategoryConfig) GetOutputPath() string {
	return c.outputPath
}

func (c *CategoryConfig) AddContainPathName(path string) {
	c.containPathName = append(c.containPathName, path)
}

func (c *CategoryConfig) AddIgnorePathName(path string) {
	c.ignorePathName = append(c.ignorePathName, path)
}

func (c *CategoryConfig) CheckPathFilter(path string) bool {
	tagPath := strings.TrimSuffix(path, filepath.Base(path))
	tagName := filepath.Base(tagPath)

	for _, i := range c.ignorePathName {
		if i == tagName {
			return false
		}
	}

	if len(c.containPathName) == 0 {
		return true
	}

	for _, i := range c.containPathName {
		if i == tagName {
			return true
		}
	}
	return false
}

func (c *CategoryConfig) SetShowTags(showTags bool) {
	c.showTags = showTags
}

func (c *CategoryConfig) GetShowTags() bool {
	return c.showTags
}

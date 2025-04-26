package ai

import (
	"encoding/json"
	"fmt"
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

type andLoraConfig struct {
	targetPath string
	weight     float64
	extendTags []string
}

type AnalyzeType int

const (
	Analyze_Deepdanbooru AnalyzeType = iota
	Analyze_Webuiwd14tagger
)

type TagOrder string

const (
	TagOrder_Character TagOrder = "training character"
	TagOrder_Action    TagOrder = "training action"
)

type SaveType string

const (
	Save_Txt  SaveType = "txt"
	Save_Json SaveType = "json"
)

type SDRequest struct {
	mainModleName             string
	batch_size, n_iter        int
	fixpose, fixhand, fixface bool
	manualSetup               bool
	posePaths                 map[string]*poseConfig
	alwaysonScripts           map[string]interface{}
}

type SDDownload struct {
	savePathName   string
	basePath       string
	inputPath      string
	outputPath     string
	forEach        bool
	saveType       SaveType
	tagConfigs     map[string]tagConfig
	loraModles     map[string]float64
	loraTags       map[string][]string
	extendTags     map[string][]string
	andLoraConfigs map[string]andLoraConfig
}

type ImageConfig struct {
	sdRequestConfig  SDRequest
	sdDownloadConfig SDDownload
	showTags         bool
	deleteTags       bool
	containPathName  []string
	ignorePathName   []string
	skipTags         []string
	analyzeType      AnalyzeType
	tagOrder         TagOrder
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
			if category.Kind == CategoryType_Action {
				GlobalTag_Action = append(GlobalTag_Action, category.Keywords...)
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
			basePath:       basePath,
			inputPath:      inputPath,
			outputPath:     outputPath,
			savePathName:   "NewImage",
			forEach:        false,
			saveType:       Save_Json,
			tagConfigs:     make(map[string]tagConfig),
			andLoraConfigs: make(map[string]andLoraConfig),
			loraModles:     make(map[string]float64),
			loraTags:       make(map[string][]string),
			extendTags:     make(map[string][]string),
		},
		showTags:    false,
		deleteTags:  false,
		analyzeType: Analyze_Deepdanbooru,
		tagOrder:    TagOrder_Character,
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

// Using extend lora tags for each lora model
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

// Using extend general tags for each directory
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
	if (c.GetFixface() || c.GetFixhand()) && fixpose {
		utils.Warn("FixFace or FixHand has already been set, it is not recommended to set Fixpose")
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

func (c *ImageConfig) GetTagOrder() TagOrder {
	return c.tagOrder
}

func (c *ImageConfig) SetTagOrder(tagOrder TagOrder) {
	c.tagOrder = tagOrder
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

func (c *ImageConfig) GetAndLoraConfigs() map[string]andLoraConfig {
	return c.sdDownloadConfig.andLoraConfigs
}

func (c *ImageConfig) GetAndLoraString(n string) (string, bool) {
	if lore, ok := c.sdDownloadConfig.andLoraConfigs[n]; ok {
		path := filepath.Join(c.sdDownloadConfig.inputPath, c.containPathName[0], lore.targetPath)
		fileData, _ := os.ReadFile(path)
		return fmt.Sprintf("\nAND<lora:%s:%f>%s,%s", n, lore.weight, strings.Join(lore.extendTags, ","), string(fileData)), true
	}
	return "", false
}

func (c *ImageConfig) AddAndLoraConfig(loraName string, weight float64, targetPath string, extendTagsp ...[]string) {
	c.sdRequestConfig.manualSetup = true

	extendTags := []string{}
	if len(extendTagsp) > 0 {
		extendTags = extendTagsp[0]
	}
	c.sdDownloadConfig.andLoraConfigs[loraName] = andLoraConfig{
		targetPath: targetPath,
		extendTags: extendTags,
		weight:     weight,
	}
}

func (c *ImageConfig) IsManualSetup() bool {
	return c.sdRequestConfig.manualSetup
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

type TrainSpeed int

const (
	TrainSpeedAuto TrainSpeed = iota
	TrainSpeedSlow            // 100 images ≈ 7h
	TrainSpeedMid             // 100 images ≈ 5h
	TrainSpeedFast            // 100 images ≈ 3h
)

type TrainQuality int

const (
	TrainQualityHigh TrainQuality = iota
	TrainQualityMed
	TrainQualityLow
)

type modelConfig struct {
	modelName      string // output_name
	pretrainedPath string // pretrained_model_name_or_path
	inputDir       string // train_data_dir
	outputDir      string // output_dir
	logDir         string // logging_dir
	epoch          int    // epoch
	prompts        string // sample_prompts
}

type tagConfig struct {
	tagName, tagSrcPath string
	times               int
	trainTagNum         int
}

type TrainConfig struct {
	basePath      string
	limit         int
	modelConfig   modelConfig
	trainSpeed    TrainSpeed
	trainQuality  TrainQuality
	trainTotalNum int
	tagConfigs    map[string]*tagConfig
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
			epoch:          10,
			prompts:        "",
		},
		tagConfigs:    make(map[string]*tagConfig),
		limit:         0,
		basePath:      basePath,
		trainSpeed:    TrainSpeedAuto,
		trainQuality:  TrainQualityLow,
		trainTotalNum: 0,
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

func (c *TrainConfig) GetTagConfigs() map[string]*tagConfig {
	return c.tagConfigs
}

func (c *TrainConfig) AddTagConfig(tagName, tagPath string, times int) {
	c.tagConfigs[tagName] = &tagConfig{
		tagName:     tagName,
		tagSrcPath:  tagPath,
		times:       times,
		trainTagNum: 0,
	}
}

func (c *TrainConfig) SetTrainTagNum(tagName string, num int) {
	if tag, ok := c.tagConfigs[tagName]; ok {
		totalStep := num * len(c.tagConfigs) * tag.times * c.modelConfig.epoch / 4
		if totalStep > 5000 {
			utils.Warnf("Estimated total number of steps is %d, need to reduce repeat", totalStep)
		} else if totalStep < 2000 {
			utils.Warnf("Estimated total number of steps is %d, need to increase repeat", totalStep)
		}
		tag.trainTagNum = num
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
		return true
	}
	return false
}

func (t *TrainConfig) GetTrainSpeed() TrainSpeed {
	if t.trainSpeed == TrainSpeedAuto {
		if t.trainTotalNum < 100 {
			utils.Info("The training set is less than 100 images, use the TrainSpeedFast mode")
			return TrainSpeedFast
		} else {
			utils.Info("The training set is more than 100 images, use the TrainSpeedSlow mode")
			return TrainSpeedSlow
		}
	}
	return t.trainSpeed
}

func (t *TrainConfig) SetTrainSpeed(trainType TrainSpeed) {
	t.trainSpeed = trainType
}

func (t *TrainConfig) GetTrainQuality() TrainQuality {
	return t.trainQuality
}

func (t *TrainConfig) SetTrainQuality(trainQuality TrainQuality) {
	t.trainQuality = trainQuality
}

func (t *TrainConfig) UpTrainTotalNum(num int) {
	t.trainTotalNum += num
}

func (t *TrainConfig) GetPrompts() string {
	return t.modelConfig.prompts
}

func (t *TrainConfig) SetPrompts(prompts string) {
	t.modelConfig.prompts = prompts
}

func (t *TrainConfig) GetEpoch() int {
	totalStep := 0
	for _, tc := range t.tagConfigs {
		totalStep += tc.times * tc.trainTagNum * t.modelConfig.epoch
	}
	totalStep = totalStep / 4
	utils.Infof("Total step: %d", totalStep)
	if totalStep < 2000 {
		utils.Warn("The training set is less than 2000 steps, it is recommended to increase epoch or repeat")
	} else if totalStep > 5000 {
		utils.Warn("The training set is more than 5000 steps, it is recommended to reduce epoch or repeat")
	}
	return t.modelConfig.epoch
}

func (t *TrainConfig) SetEpoch(epoch int) {
	t.modelConfig.epoch = epoch
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
	limit                           int
	basePath                        string
	outputPath                      string
	containPathName, ignorePathName []string
	showTags                        bool
	directInput                     []string
}

// openai: https://platform.openai.com/settings/organization/api-keys

func NewCategoryConfig(keyString []string, outputPath string, basePaths ...string) *CategoryConfig {
	basePath, _ := os.Getwd()

	if len(basePaths) > 0 {
		basePath = basePaths[0]
	}
	return &CategoryConfig{
		keyString:   keyString,
		keyIndex:    0,
		limit:       0,
		basePath:    basePath,
		outputPath:  filepath.Join(basePath, outputPath),
		showTags:    false,
		directInput: []string{},
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

func (c *CategoryConfig) SetDirectInput(directInput []string) {
	c.directInput = directInput
}

func (c *CategoryConfig) GetDirectInput() []string {
	return c.directInput
}

func (c *CategoryConfig) GetLimit() int {
	return c.limit
}

func (c *CategoryConfig) SetLimit(limit int) {
	c.limit = limit
}

func (c *CategoryConfig) CheckLimit(index int) bool {
	if index >= c.GetLimit() && c.GetLimit() != 0 {
		return true
	}
	return false
}

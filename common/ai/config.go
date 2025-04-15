package ai

type config struct {
	basePath string
}

func InitConfig(basePath string) *config {
	return &config{
		basePath: basePath,
	}
}

package logger

type Config struct {
	Level  string       `yaml:"level"`
	Format string       `yaml:"format"`
	Output OutputConfig `yaml:"output"`
	Hooks  HooksConfig  `yaml:"hooks"`
}

type OutputConfig struct {
	Console bool         `yaml:"console"`
	Files   []FileConfig `yaml:"files"`
}

type FileConfig struct {
	Level      string `yaml:"level"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
}

type HooksConfig struct {
	Wecom WecomConfig `yaml:"wecom"`
}

type WecomConfig struct {
	Enabled    bool     `yaml:"enabled"`
	Levels     []string `yaml:"levels"`
	WebhookURL string   `yaml:"webhook_url"`
}

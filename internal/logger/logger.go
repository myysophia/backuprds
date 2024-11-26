package logger

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

var (
	logger *zap.Logger
	once   sync.Once
	level  zap.AtomicLevel
)

// Field 字段构造函数
type Field = zapcore.Field

func String(key string, value string) Field {
	return zap.String(key, value)
}

func Int(key string, value int) Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

func Float64(key string, value float64) Field {
	return zap.Float64(key, value)
}

func Error(err error) Field {
	return zap.Error(err)
}

func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

func Time(key string, value time.Time) Field {
	return zap.Time(key, value)
}

func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}

// 添加日志级别检查函数
func isLevelEnabled(lvl zapcore.Level) bool {
	return level.Enabled(lvl)
}

// InitFromFile 从配置文件初始化日志系统
func InitFromFile(configPath string) error {
	// 读取配置文件
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// 解析配置
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 确保日志级别正确设置
	level = zap.NewAtomicLevel()
	logLevel := strings.ToLower(cfg.Level)

	// 验证日志级别
	switch logLevel {
	case "debug":
		level.SetLevel(zapcore.DebugLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	case "warn":
		level.SetLevel(zapcore.WarnLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	case "fatal":
		level.SetLevel(zapcore.FatalLevel)
	default:
		return fmt.Errorf("invalid log level: %s", cfg.Level)
	}

	// 创建核心日志组件
	var cores []zapcore.Core

	// 控制台输出
	if cfg.Output.Console {
		consoleEncoder := getEncoder(cfg.Format, encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	for _, fc := range cfg.Output.Files {
		fileEncoder := getEncoder(cfg.Format, encoderConfig)
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   fc.Path,
			MaxSize:    fc.MaxSize,
			MaxBackups: fc.MaxBackups,
			MaxAge:     fc.MaxAge,
			Compress:   true,
		})
		fileCore := zapcore.NewCore(
			fileEncoder,
			writer,
			level, // 使用相同的全局级别
		)
		cores = append(cores, fileCore)
	}

	// 创建日志实例
	core := zapcore.NewTee(cores...)

	// 添加选项
	opts := []zap.Option{
		zap.AddCaller(),                       // 添加这行，启用调用者信息
		zap.AddCallerSkip(1),                  // 跳过一层调用栈以获取正确的调用位置
		zap.AddStacktrace(zapcore.ErrorLevel), // 为 Error 及以上级别添加堆栈信息
	}

	// 添加Hooks
	if cfg.Hooks.Wecom.Enabled {
		hook := NewWecomHook(cfg.Hooks.Wecom.Levels, cfg.Hooks.Wecom.WebhookURL)
		opts = append(opts, zap.Hooks(hook.Fire))
	}

	// 使用选项创建logger
	logger = zap.New(core, opts...)

	return nil
}

// getEncoder 根据格式返回对应的编码器
func getEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}

// 获取调用信息
func getCallerInfo() (string, string, int) {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown", "unknown", 0
	}

	funcName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(funcName, ".")
	funcName = parts[len(parts)-1]

	parts = strings.Split(file, "/")
	file = parts[len(parts)-1]

	return file, funcName, line
}

// 添加调用者信息
func addCallerInfo(msg string, fields []Field) (string, []Field) {
	file, funcName, line := getCallerInfo()
	newFields := make([]Field, 0, len(fields)+3)
	newFields = append(newFields,
		String("file", file),
		String("func", funcName),
		Int("line", line))
	newFields = append(newFields, fields...)
	return msg, newFields
}

// 日志输出函数
func LogDebug(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	msg, newFields := addCallerInfo(msg, fields)
	logger.Debug(msg, newFields...)
}

func LogInfo(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	msg, newFields := addCallerInfo(msg, fields)
	logger.Info(msg, newFields...)
}

func LogWarn(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	msg, newFields := addCallerInfo(msg, fields)
	logger.Warn(msg, newFields...)
}

func LogError(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	msg, newFields := addCallerInfo(msg, fields)
	logger.Error(msg, newFields...)
}

func LogFatal(msg string, fields ...Field) {
	if logger == nil {
		os.Exit(1)
	}
	msg, newFields := addCallerInfo(msg, fields)
	logger.Fatal(msg, newFields...)
}

// 添加获取当前日志级别的方法
func GetLogLevel() string {
	return level.String()
}

// 添加动态修改日志级别的方法
func SetLogLevel(logLevel string) error {
	switch strings.ToLower(logLevel) {
	case "debug":
		level.SetLevel(zapcore.DebugLevel)
	case "info":
		level.SetLevel(zapcore.InfoLevel)
	case "warn":
		level.SetLevel(zapcore.WarnLevel)
	case "error":
		level.SetLevel(zapcore.ErrorLevel)
	case "fatal":
		level.SetLevel(zapcore.FatalLevel)
	default:
		return fmt.Errorf("invalid log level: %s", logLevel)
	}
	return nil
}

package log

/**
 * log  使用工具包
 * Version   : 1.0
 * Create by :liming10
 * Created on: 2022/03/28 15:08
 */
import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
	"wg_goservice/global"

	"wg_goservice/pkg/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// MLogger 日志实例
type MLogger struct {
	writer zapcore.WriteSyncer
	lj     *lumberjack.Logger
	*zap.Logger
}

// 日志实例Map
var insMap sync.Map

// 文件缓存
var cachePath sync.Map

// 返回一个分类&日期的实例
func instance(cat string) MLogger {
	// 获取一个实例
	if v, ok := insMap.Load(cat); ok {
		return v.(MLogger)
	}

	// 编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "MLogger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder, // 大写编码器
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000000"))
		}, // 时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
	}

	// 日志分类  从yaml获取日志文件配置
	logCfg := global.LogsSetting
	logPath := logCfg.LogPath
	if !strings.HasPrefix(logPath, "/") {
		logPath = "/" + logCfg.LogPath
	}
	file := fmt.Sprintf(strings.TrimRight(logPath, "/")+"/%s.log", cat)

	// 日志切割
	lj := &lumberjack.Logger{
		Filename:   file,              // 日志文件路径
		MaxSize:    logCfg.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: logCfg.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     logCfg.MaxAge,     // 文件最多保存多少天
		Compress:   logCfg.Compress,   // 是否压缩
		LocalTime:  true,
	}

	// 日志级别
	var level zapcore.Level
	var writeSync zapcore.WriteSyncer
	if global.LogsSetting.IsWeb {
		level = zap.InfoLevel
		writeSync = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lj))
	} else {
		level = zap.DebugLevel
		writeSync = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig), // 编码器配置
		writeSync,                   // 打印到控制台/文件
		zap.NewAtomicLevelAt(level), // 日志输出级别
	)

	l := MLogger{
		writer: writeSync,
		lj:     lj,
		Logger: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
	}
	insMap.Store(cat, l)

	return l
}

// 关闭文件句柄
func Close() {
	insMap.Range(func(key, value interface{}) bool {
		_ = value.(MLogger).lj.Close()
		return true
	})
}

// Logger 根据文件前缀分类返回日志实例
func Logger() MLogger {
	_, filename, _, _ := runtime.Caller(2)
	RuntimeRoot := path.Dir(path.Dir(path.Dir(filename)))
	filename = strings.TrimPrefix(filename, RuntimeRoot+"/")
	cat, ok := cachePath.Load(filename)
	if !ok {
		//日志名称
		cat = global.LogsSetting.DefaultCategory
		if cat == "" {
			cat = "app"
		}
		cachePath.Store(filename, cat)
	}

	return instance(cat.(string))
}

// 对复杂类型进行JSON格式化
func format(args []interface{}) []interface{} {
	for i, v := range args {
		switch reflect.ValueOf(v).Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			s, _ := json.MarshalToString(v)
			args[i] = s
			continue
		}
	}

	return args
}

// Writer returns a zap writer
func Writer() zapcore.WriteSyncer {
	return Logger().writer
}

// Debug log a message
func Debug(args ...interface{}) {
	Logger().Sugar().Debug(format(args))
}

// Info log a message
func Info(args ...interface{}) {
	Logger().Sugar().Info(format(args)...)
}

// Warn log a message
func Warn(args ...interface{}) {
	Logger().Sugar().Warn(format(args)...)
}

// Error log a message
func Error(args ...interface{}) {
	Logger().Sugar().Error(format(args)...)
}

// Fatal log a message, then calls os.Exit
func Fatal(args ...interface{}) {
	Logger().Sugar().Fatal(format(args)...)
}

// Panic log a message, then panics
func Panic(args ...interface{}) {
	Logger().Sugar().Panic(format(args)...)
}

// Infof log a message
func Infof(tpl string, args ...interface{}) {
	Logger().Sugar().Infof(tpl, format(args)...)
}

// Errorf log a message
func Errorf(tpl string, args ...interface{}) {
	Logger().Sugar().Errorf(tpl, format(args)...)
}

// CPrint only print log in console mode
func CPrint(args ...interface{}) {
	if global.LogsSetting.IsWeb {
		return
	}
	Logger().Sugar().Info(format(args)...)
}

// CPrintf only print log in console mode
func CPrintf(tpl string, args ...interface{}) {
	if global.LogsSetting.IsWeb {
		return
	}
	Logger().Sugar().Infof(tpl, format(args)...)
}

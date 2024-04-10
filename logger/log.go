package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"sync"
	"time"
)

var (
	Zl     *zap.Logger
	Global *zap.SugaredLogger
)

type ZapLog struct {
	Zl        *zap.Logger
	GlobalLog *zap.SugaredLogger
	sync.RWMutex
	fields []string
}

func GetLogger() *ZapLog {
	l := &ZapLog{
		fields: make([]string, 0),
	}
	if Global != nil && Zl != nil {
		l.Zl = Zl
		l.GlobalLog = Global
		return l
	} else {
		IniLogger("info")
	}
	l.Zl = Zl
	l.GlobalLog = Global
	return l
}

func IniLogger(level ...string) *zap.Logger {
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "msg",
		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,
		TimeKey:     "ts",
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		CallerKey:    "file",
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})
	var lv string
	if len(level) != 0 {
		lv = level[0]
	} else {
		lv = "info"
	}
	logLevel := zap.DebugLevel
	switch lv {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel && lvl >= logLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && lvl >= logLevel
	})
	// 获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter := getWriter("./logs/info.logger")
	errorWriter := getWriter("./logs/error.logger")
	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), infoLevel), //打印到控制台
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), errorLevel),
	)
	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.InfoLevel)) //会显示打日志点的文件名和行数
	Global = log.Sugar()
	Zl = log
	return log
}

func getWriter(filename string) io.Writer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return lumberJackLogger
}

func (l *ZapLog) Debug(args ...interface{}) {
	if len(l.fields) != 0 {
		var s string
		for _, v := range l.fields {
			s += v
		}
		args = append(args, " ", s)
	}
	l.GlobalLog.Debug(args)
}

func (l *ZapLog) Debugf(format string, args ...interface{}) {
	var s string
	if len(l.fields) != 0 {
		for _, v := range l.fields {
			s += v
		}
	}
	l.GlobalLog.Debug(s, " ", fmt.Sprintf(format, args...))
}

func (l *ZapLog) Info(args ...interface{}) {
	if len(l.fields) != 0 {
		var s string
		for _, v := range l.fields {
			s += v
		}
		args = append(args, " ", s)
	}
	l.GlobalLog.Info(args...)
}

func (l *ZapLog) Infof(format string, args ...interface{}) {
	var s string
	if len(l.fields) != 0 {
		for _, v := range l.fields {
			s += v
		}
	}
	l.GlobalLog.Info(s, " ", fmt.Sprintf(format, args...))
}

func (l *ZapLog) Error(args ...interface{}) {
	if len(l.fields) != 0 {
		var s string
		for _, v := range l.fields {
			s += v
		}
		args = append(args, " ", s)
	}
	l.GlobalLog.Error(args...)
}

func (l *ZapLog) Errorf(format string, args ...interface{}) {
	var s string
	if len(l.fields) != 0 {
		for _, v := range l.fields {
			s += v
		}
	}
	l.GlobalLog.Error(s, " ", fmt.Sprintf(format, args...))
}

func (l *ZapLog) Panic(args ...interface{}) {
	if len(l.fields) != 0 {
		var s string
		for _, v := range l.fields {
			s += v
		}
		args = append(args, " ", s)
	}
	l.GlobalLog.Panic(args...)
}

func (l *ZapLog) Panicf(format string, args ...interface{}) {
	var s string
	if len(l.fields) != 0 {
		for _, v := range l.fields {
			s += v
		}
	}
	l.GlobalLog.Panicf(s, " ", fmt.Sprintf(format, args...))
}

func (l *ZapLog) WithFileds(args ...string) *ZapLog {
	l.fields = append(l.fields, args...)
	return l
}

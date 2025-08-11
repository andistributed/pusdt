package log

import (
	"io"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"github.com/v03413/bepusdt/app/conf"
)

var logger *logrus.Logger

func Init() error {
	logger = logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     false,
		DisableColors:   true,
		ForceQuote:      false,
		DisableQuote:    true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	logLevelStr := conf.GetLogLevel()
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.DebugLevel
	}
	logger.SetLevel(logLevel)

	// output, err := os.OpenFile(conf.GetOutputLog(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {

	// 	return err
	// }

	cfg := conf.GetConfig().Log
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // MB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 7
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 7 // Days
	}
	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf.GetOutputLog(),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   false,
	}
	logger.SetOutput(lumberJackLogger)

	return nil
}

func Debug(args ...interface{}) {
	logger.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Error(args ...interface{}) {
	logger.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func GetWriter() *io.PipeWriter {
	return logger.Writer()
}

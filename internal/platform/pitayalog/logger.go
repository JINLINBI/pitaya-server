package pitayalog

import (
	"fmt"
	"strings"

	pitaya "github.com/topfreegames/pitaya/v3/pkg"
	plogrus "github.com/topfreegames/pitaya/v3/pkg/logger/logrus"
	"github.com/topfreegames/pitaya/v3/pkg/logger/interfaces"
)

const standaloneRemoteWarn = "registered a remote component but remoteService is not running"

// Init 配置 Pitaya 日志；单体下抑制内置 sys remote 的无害告警。
func Init(suppressStandaloneRemoteWarn bool) {
	base := plogrus.New()
	if suppressStandaloneRemoteWarn {
		pitaya.SetLogger(&filterLogger{inner: base, suppressRemote: true})
		return
	}
	pitaya.SetLogger(base)
}

type filterLogger struct {
	inner          interfaces.Logger
	suppressRemote bool
}

func (f *filterLogger) skipWarn(parts ...string) bool {
	if !f.suppressRemote {
		return false
	}
	for _, p := range parts {
		if strings.Contains(p, standaloneRemoteWarn) {
			return true
		}
	}
	return false
}

func (f *filterLogger) Warn(args ...interface{}) {
	if f.skipWarn(fmt.Sprint(args...)) {
		return
	}
	f.inner.Warn(args...)
}

func (f *filterLogger) Warnf(format string, args ...interface{}) {
	if f.skipWarn(format, fmt.Sprintf(format, args...)) {
		return
	}
	f.inner.Warnf(format, args...)
}

func (f *filterLogger) Warnln(args ...interface{}) {
	if f.skipWarn(fmt.Sprint(args...)) {
		return
	}
	f.inner.Warnln(args...)
}

func (f *filterLogger) Fatal(args ...interface{})                 { f.inner.Fatal(args...) }
func (f *filterLogger) Fatalf(format string, args ...interface{}) { f.inner.Fatalf(format, args...) }
func (f *filterLogger) Fatalln(args ...interface{})               { f.inner.Fatalln(args...) }
func (f *filterLogger) Debug(args ...interface{})                 { f.inner.Debug(args...) }
func (f *filterLogger) Debugf(format string, args ...interface{})   { f.inner.Debugf(format, args...) }
func (f *filterLogger) Debugln(args ...interface{})               { f.inner.Debugln(args...) }
func (f *filterLogger) Error(args ...interface{})                 { f.inner.Error(args...) }
func (f *filterLogger) Errorf(format string, args ...interface{}) { f.inner.Errorf(format, args...) }
func (f *filterLogger) Errorln(args ...interface{})               { f.inner.Errorln(args...) }
func (f *filterLogger) Info(args ...interface{})                  { f.inner.Info(args...) }
func (f *filterLogger) Infof(format string, args ...interface{})  { f.inner.Infof(format, args...) }
func (f *filterLogger) Infoln(args ...interface{})                { f.inner.Infoln(args...) }
func (f *filterLogger) Panic(args ...interface{})                 { f.inner.Panic(args...) }
func (f *filterLogger) Panicf(format string, args ...interface{}) { f.inner.Panicf(format, args...) }
func (f *filterLogger) Panicln(args ...interface{})               { f.inner.Panicln(args...) }
func (f *filterLogger) WithFields(fields map[string]interface{}) interfaces.Logger {
	return &filterLogger{inner: f.inner.WithFields(fields), suppressRemote: f.suppressRemote}
}
func (f *filterLogger) WithField(key string, value interface{}) interfaces.Logger {
	return &filterLogger{inner: f.inner.WithField(key, value), suppressRemote: f.suppressRemote}
}
func (f *filterLogger) WithError(err error) interfaces.Logger {
	return &filterLogger{inner: f.inner.WithError(err), suppressRemote: f.suppressRemote}
}
func (f *filterLogger) GetInternalLogger() any { return f.inner.GetInternalLogger() }
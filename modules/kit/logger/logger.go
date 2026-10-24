package logger

import (
	"bytes"
	"fmt"
	"github.com/silentred/gid"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"sync"
	"time"
)

var (
	logger   *zap.Logger
	goLogger = make(map[int64]*zap.Logger)
	m        sync.Mutex
)

type ChannelAsyncLogger struct {
	logChannel chan *bytes.Buffer
	once       sync.Once
	closeCh    chan struct{}
	done       chan struct{}
	closeFlag  atomic.Bool
	wg         sync.WaitGroup
	WS         interface {
		zapcore.WriteSyncer
		io.Closer
	}
}

// RTMD
// 东西不大，写起来挺麻烦
func (c *ChannelAsyncLogger) Close() error {
	c.closeFlag.Store(true)
	close(c.closeCh)
	c.wg.Wait()
	close(c.logChannel)
	<-c.done
	return c.WS.Close()
}

func (c *ChannelAsyncLogger) loop() {
	c.wg.Add(1)
	defer c.wg.Done()
FORLOOP:
	for {
		select {
		case buf := <-c.logChannel:
			c.WS.Write(buf.Bytes())
		case <-c.closeCh:
			break FORLOOP
		}
	}
	for buf := range c.logChannel {
		_, err := c.WS.Write(buf.Bytes())
		if err != nil {
			// 这里出错了，我也没办法了。。。
			break
		}
	}
}

func (c *ChannelAsyncLogger) Write(p []byte) (n int, err error) {
	c.once.Do(func() {
		go c.loop()
	})
	// must add before
	c.wg.Add(1)
	defer c.wg.Done()
	if c.closeFlag.Load() {
		return 0, fmt.Errorf("file already closed")
	}
	buf := bytes.NewBuffer(make([]byte, 0, len(p)))
	buf.Write(p)
	select {
	case c.logChannel <- buf:
	default:
	}
	return len(p), nil
}

func (c *ChannelAsyncLogger) Sync() error {
	return c.WS.Sync()
}

func InitLogger(fileName string) {
	var w zapcore.WriteSyncer
	if fileName != "" {
		w = zapcore.AddSync(&lumberjack.Logger{
			Filename:   fileName,
			MaxSize:    500, // megabytes
			MaxAge:     0,
			MaxBackups: 0,
			LocalTime:  true,
			Compress:   false,
		})
	} else {
		w = os.Stdout
	}
	ecf := zap.NewProductionEncoderConfig()
	ecf.FunctionKey = "func"
	ecf.EncodeTime = zapcore.ISO8601TimeEncoder
	ecf.ConsoleSeparator = " "
	zap.NewProductionConfig().Build()
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(ecf),
		&zapcore.BufferedWriteSyncer{
			WS:            w,
			Size:          0,
			FlushInterval: time.Millisecond * 10,
			Clock:         nil,
		},
		zap.InfoLevel,
	)
	logger = zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
}

func Logger() *zap.Logger {
	return logger.WithOptions(zap.AddCallerSkip(-1))
}

func innerGoLogger() *zap.Logger {
	id := gid.Get()
	m.Lock()
	l := goLogger[id]
	m.Unlock()
	if l == nil {
		return logger
	}
	return l
}

func GoLogger() *zap.Logger {
	return innerGoLogger().WithOptions(zap.AddCallerSkip(-1))
}

func SetGoLoggerName(name string) {
	id := gid.Get()
	m.Lock()
	defer m.Unlock()
	l := goLogger[id]
	if l != nil {
		goLogger[id] = l.Named(name)
	} else {
		goLogger[id] = logger.Named(name)
	}
}

func ResetGoLogger() *zap.Logger {
	id := gid.Get()
	m.Lock()
	defer m.Unlock()
	l := goLogger[id]
	delete(goLogger, id)
	return l
}

func Debug(msg string, fields ...zap.Field)  { innerGoLogger().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)   { innerGoLogger().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { innerGoLogger().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)  { innerGoLogger().Error(msg, fields...) }
func DPanic(msg string, fields ...zap.Field) { innerGoLogger().DPanic(msg, fields...) }
func Panic(msg string, fields ...zap.Field)  { innerGoLogger().Panic(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { innerGoLogger().Fatal(msg, fields...) }
func Sync()                                  { innerGoLogger().Sync(); logger.Sync() }

func Named(name string) *zap.Logger {
	return innerGoLogger().WithOptions(zap.AddCallerSkip(-1)).Named(name)
}
func Field(field ...zapcore.Field) *zap.Logger {
	return innerGoLogger().WithOptions(zap.AddCallerSkip(-1)).With(field...)
}
func Skip(n int) *zap.Logger { return innerGoLogger().WithOptions(zap.AddCallerSkip(n - 1)) }
func WithOption(opt ...zap.Option) *zap.Logger {
	return innerGoLogger().WithOptions(zap.AddCallerSkip(-1)).WithOptions(opt...)
}

func S() *zap.SugaredLogger { return logger.Sugar() }

func Debugw(msg string, keysAndValues ...interface{}) { S().Debugw(msg, keysAndValues...) }
func Infow(msg string, keysAndValues ...interface{})  { S().Infow(msg, keysAndValues...) }
func Warnw(msg string, keysAndValues ...interface{})  { S().Warnw(msg, keysAndValues...) }
func Errorw(msg string, keysAndValues ...interface{}) { S().Errorw(msg, keysAndValues...) }
func Panicw(msg string, keysAndValues ...interface{}) { S().Panicw(msg, keysAndValues...) }
func Fatalw(msg string, keysAndValues ...interface{}) { S().Fatalw(msg, keysAndValues...) }

func Debugf(msg string, keysAndValues ...interface{}) { S().Debugf(msg, keysAndValues...) }
func Infof(msg string, keysAndValues ...interface{})  { S().Infof(msg, keysAndValues...) }
func Warnf(msg string, keysAndValues ...interface{})  { S().Warnf(msg, keysAndValues...) }
func Errorf(msg string, keysAndValues ...interface{}) { S().Errorf(msg, keysAndValues...) }
func Panicf(msg string, keysAndValues ...interface{}) { S().Panicf(msg, keysAndValues...) }
func Fatalf(msg string, keysAndValues ...interface{}) { S().Fatalf(msg, keysAndValues...) }

# log
go logger with zap

### use default 
The log directory uses the value of the environment variable `LOG_DIR` by default, and if the variable is not set, the current directory is used
```go
package main

import (
	"github.com/syspkg/log"
)

func main()  {
	log.Info("test msg")

}
```
the result is `{"level":"info","timestamp":"2023-01-30T16:56:45.230+0800","caller":"test01/01.go:8","msg":"test msg"}`

### use custom configuration
```go
package main

import (
	"github.com/syspkg/log"
)

func main()  {
	c := log.Config{LogDir: "/data/logs", LogFile: "app.log", JsonEncode: true}
	log.New(&c)
	log.Info("test msg")
}
```

### configuration
Set `FilePerLevel` to `true` if you want each level of log output to a file named after the `level`.log
```go
type Config struct {
	// LogDir dir for logs
	LogDir string `mapstructure:"dir,omitempty" yaml:"dir,omitempty" json:"dir,omitempty"`
	// LogFile file name for logs
	LogFile string `mapstructure:"file,omitempty" yaml:"file,omitempty" json:"file,omitempty"`
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	// It defaults to 100 megabytes.
	MaxSize int `mapstructure:"maxSize,omitempty" yaml:"maxSize,omitempty" json:"maxSize,omitempty"`
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `mapstructure:"maxBackups,omitempty" yaml:"maxBackups,omitempty" json:"maxBackups,omitempty"`
	// MaxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename.
	MaxAge int `mapstructure:"maxAge,omitempty" yaml:"maxAge,omitempty" json:"maxAge,omitempty"`
	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `mapstructure:"compress,omitempty" yaml:"compress,omitempty" json:"compress,omitempty"`

	// LogLevel the level for output log eg debug info warn error
	LogLevel string `mapstructure:"level" yaml:"level" json:"level"`
	// JsonEncode json format logs
	JsonEncode bool `mapstructure:"jsonEncode,omitempty" yaml:"jsonEncode,omitempty" json:"jsonEncode,omitempty"`
	// StacktraceLevel output stack track for this level log eg error
	StacktraceLevel string `mapstructure:"stacktraceLevel,omitempty" yaml:"stacktraceLevel,omitempty" json:"stacktraceLevel,omitempty"`
	// Stdout output log to stdout
	Stdout bool `mapstructure:"stdout,omitempty" yaml:"stdout,omitempty" json:"stdout,omitempty"`
	// FilePerLevel Each level of log output to the corresponding log file. eg debug.log info.log warn.log error.log
	FilePerLevel bool `mapstructure:"filePerLevel,omitempty" yaml:"filePerLevel,omitempty" json:"filePerLevel,omitempty"`
}
```
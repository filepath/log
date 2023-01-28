package log

type Config struct {
	// rotate options
	LogDir     string `json:"logDir,omitempty"`
	LogFile    string `json:"logFile,omitempty"`
	MaxSize    int    `json:"maxSize,omitempty"`
	MaxBackups int    `json:"maxBackups,omitempty"`
	MaxAge     int    `json:"maxAge,omitempty"`
	Compress   bool   `json:"compress,omitempty"`

	// logger options
	LogLevel        string `json:"level"`
	JsonEncode      bool   `json:"jsonEncode,omitempty"`
	StacktraceLevel string `json:"stacktraceLevel,omitempty"`
	Stdout          bool   `json:"stdout,omitempty"`
	FilePerLevel    bool   `json:"filePerLevel"`
}

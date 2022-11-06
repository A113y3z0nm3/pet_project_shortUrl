package myLog

// Config Конфигурация к логгеру
type Config struct {
	Mod    string `json:"mod" xml:"mod"` // development, production
	Level  string `json:"level" xml:"level"`
	Output string `json:"output" xml:"output"`
}

// DefaultConfig Тестовая конфигурация
var DefaultConfig = &Config{
	Mod:    "development",
	Level:  "debug",
	Output: "stdout",
}

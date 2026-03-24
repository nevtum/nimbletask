package cli

// Config represents the configuration structure from config.json
type Config struct {
	Filename        string `json:"filename"`
	DefaultPriority int    `json:"default_priority"`
}

func defaultConfig() Config {
	return Config{
		Filename:        "todos.md",
		DefaultPriority: 3,
	}
}

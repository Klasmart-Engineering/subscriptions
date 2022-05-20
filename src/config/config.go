package config

type Config struct {
	Logger Logger
	Server Server
}

// Server config
type Server struct {
	Port        string
	Development bool
}

// Logger config
type Logger struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}

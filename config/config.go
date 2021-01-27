package config

// Config is for storing confiruables from the env
// The env variable defaults to P4ACCESS_<VAR>, e.g. P4ACCESS_P4PORT
type Config struct {
	P4Port   string
	P4User   string
	P4Client string
}

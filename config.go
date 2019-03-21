package querylog

import "os"

type ConfigRule struct {
	Key     string
	Regex   bool
	Term    string
	Dbs     Actions
	Sites   Actions
	Actions Actions
	Hosts   Actions
	Values  bool
	Error   bool
	Tables  []string
}

type ConfigOutput struct {
	Dest     string
	Perm     os.FileMode
	Truncate bool
	Rules    []ConfigRule
}

type Config struct {
	Outputs []ConfigOutput
}

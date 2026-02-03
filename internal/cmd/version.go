package cmd

import "fmt"

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func VersionString() string {
	if commit == "" && date == "" {
		return version
	}
	if commit == "" {
		return fmt.Sprintf("%s (%s)", version, date)
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return fmt.Sprintf("%s (%s %s)", version, commit, date)
}

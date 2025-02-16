package main

import (
	"strings"

	"github.com/ttl256/euivator/cmd"
)

var version = "unset_version"

func main() {
	cmd.SetVersion(strings.TrimSpace(version))
	cmd.Execute()
}

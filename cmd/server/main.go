package main

import (
	"fmt"
	"github.com/sshirox/isaac/internal/server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	if err := server.Run(); err != nil {
		panic(err)
	}
}

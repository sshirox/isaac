package server

import (
	"flag"
)

var (
	flagRunAddr  string
	flagLogLevel string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.Parse()
}

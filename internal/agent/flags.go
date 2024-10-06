package agent

import "flag"

var (
	flagServerAddr     string
	flagReportInterval int64
	flagPollInterval   int64
	serverAddr         string
	reportInterval     int64
	pollInterval       int64
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server address and port")
	flag.Int64Var(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()
}
package agent

import "flag"

var (
	flagServerAddr     string
	flagGRPCAddr       string
	flagReportInterval int64
	flagPollInterval   int64
	flagEncryptionKey  string
	flagRateLimit      int64
	flagCryptoKeyPath  string
	serverAddr         string
	reportInterval     int64
	pollInterval       int64
	flagConfigPath     string
)

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server address and port")
	flag.StringVar(&flagGRPCAddr, "ga", "", "server grpc address")
	flag.Int64Var(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.StringVar(&flagEncryptionKey, "k", "", "encryption key")
	flag.Int64Var(&flagRateLimit, "l", 10, "rate limit")
	flag.StringVar(&flagCryptoKeyPath, "ck", "", "crypto key path")
	flag.StringVar(&flagConfigPath, "c", "", "config file path")
	flag.Parse()
}

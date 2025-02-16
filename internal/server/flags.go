package server

import (
	"flag"
)

var (
	flagRunAddr         string
	flagGRPCAddr        string
	flagLogLevel        string
	flagStoreInterval   int64
	flagFileStoragePath string
	flagRestoreStr      string
	flagRestore         bool
	flagDatabaseDSN     string
	flagEncryptionKey   string
	flagCryptoKeyPath   string
	flagConfigPath      string
	flagTrustedSubnet   string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagGRPCAddr, "ga", "", "server grpc address")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.Int64Var(&flagStoreInterval, "i", 300, "store interval")
	flag.StringVar(&flagFileStoragePath, "f", "./backups", "file storage path")
	flag.StringVar(&flagRestoreStr, "r", "true", "restore")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database DSN")
	flag.StringVar(&flagEncryptionKey, "k", "", "encryption key")
	flag.StringVar(&flagCryptoKeyPath, "ck", "", "crypto key path")
	flag.StringVar(&flagConfigPath, "c", "", "config file path")
	flag.StringVar(&flagTrustedSubnet, "t", "", "trusted subnet")

	flag.Parse()
}

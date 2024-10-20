package server

import (
	"flag"
)

var (
	flagRunAddr         string
	flagLogLevel        string
	flagStoreInterval   int64
	flagFileStoragePath string
	flagRestoreStr      string
	flagRestore         bool
	flagDatabaseDSN     string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.Int64Var(&flagStoreInterval, "i", 300, "store interval")
	flag.StringVar(&flagFileStoragePath, "f", "./backups", "file storage path")
	flag.StringVar(&flagRestoreStr, "r", "true", "restore")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database DSN")

	flag.Parse()
}

package agent

import (
	"encoding/json"
	"os"
)

type Config struct {
	Address        string `json:"address"`
	HashKey        string `json:"hash_key"`
	CryptoKeyPath  string `json:"crypto_key"`
	ReportInterval int64  `json:"report_interval"`
	PollInterval   int64  `json:"poll_interval"`
	RateLimit      int64  `json:"rate_limit"`
}

func loadConfigs(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var cfg Config

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	if cfg.Address != "" && flagServerAddr == "" {
		flagServerAddr = cfg.Address
	}

	if cfg.HashKey != "" && flagEncryptionKey == "" {
		flagEncryptionKey = cfg.HashKey
	}

	if cfg.CryptoKeyPath != "" && flagCryptoKeyPath == "" {
		flagCryptoKeyPath = cfg.CryptoKeyPath
	}

	if cfg.ReportInterval != 0 && reportInterval == 0 {
		reportInterval = cfg.ReportInterval
	}

	if cfg.PollInterval != 0 && pollInterval == 0 {
		pollInterval = cfg.PollInterval
	}

	if cfg.RateLimit != 0 && flagRateLimit == 0 {
		flagRateLimit = cfg.RateLimit
	}

	return nil
}

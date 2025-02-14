package server

import (
	"encoding/json"
	"os"
)

type Config struct {
	StoreInterval   int64  `json:"store_interval"`
	Address         string `json:"address"`
	Level           string `json:"level"`
	FilePath        string `json:"file_path"`
	RestoreStr      string `json:"restore_str"`
	DatabaseAddress string `json:"database_address"`
	StorePlace      string `json:"store_place"`
	HashKey         string `json:"hash_key"`
	CryptoKeyPath   string `json:"crypto_key"`
	Restore         string `json:"restore"`
	TrustedSubnet   string `json:"trusted_subnet"`
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

	if cfg.StoreInterval != 0 {
		flagStoreInterval = cfg.StoreInterval
	}

	if cfg.Address != "" && flagRunAddr == "" {
		flagRunAddr = cfg.Address
	}

	if cfg.Level != "" && flagLogLevel == "" {
		flagLogLevel = cfg.Level
	}

	if cfg.FilePath != "" && flagFileStoragePath == "" {
		flagFileStoragePath = cfg.FilePath
	}

	if cfg.DatabaseAddress != "" && flagDatabaseDSN == "" {
		flagDatabaseDSN = cfg.DatabaseAddress
	}

	if cfg.CryptoKeyPath != "" && flagCryptoKeyPath == "" {
		flagCryptoKeyPath = cfg.CryptoKeyPath
	}

	if cfg.TrustedSubnet != "" {
		flagTrustedSubnet = cfg.TrustedSubnet
	}

	return nil
}

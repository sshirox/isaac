package server

import (
	"context"
	"crypto/rsa"
	"github.com/pkg/errors"
	grpcHandle "github.com/sshirox/isaac/internal/grpc"
	pb "github.com/sshirox/isaac/internal/proto/metrics/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/sshirox/isaac/internal/backup"
	"github.com/sshirox/isaac/internal/crypto"
	"github.com/sshirox/isaac/internal/handler"
	"github.com/sshirox/isaac/internal/logger"
	"github.com/sshirox/isaac/internal/middleware"
	"github.com/sshirox/isaac/internal/storage"
	"github.com/sshirox/isaac/internal/storage/pg"
)

const (
	dbStorageSource     = "database"
	fileStorageSource   = "file"
	memoryStorageSource = "memory"
	dbDriver            = "postgres"
)

var (
	storageSource string
	privateKey    *rsa.PrivateKey
)

func Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cancel()

	parseFlags()
	err := initConf()

	if err != nil {
		return err
	}

	db, err := pg.Open(dbDriver, flagDatabaseDSN)
	if err != nil {
		slog.Error("open database", "err", err)
	}
	defer db.Close()

	err = pg.Ping(db)
	if err != nil {
		slog.Error("ping database", "err", err)
	} else {
		slog.Info("open database", "addr", flagDatabaseDSN)
	}
	encoder := crypto.NewEncoder(flagEncryptionKey)
	signValidator := middleware.NewSignValidator(encoder).Validate
	cryptoDecoder := middleware.NewCryptoDecoder(privateKey).Decode

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(logger.WithLogging)
	r.Use(middleware.GZipMiddleware)

	trustedSubnetMiddleware, err := middleware.TrustedSubnetMiddleware(flagTrustedSubnet)
	if err != nil {
		slog.Error("Failed to initialize middleware", slog.Any("error", err))
	} else {
		r.Use(trustedSubnetMiddleware)
	}

	s := storage.NewMemStorage()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	r.Get("/", handler.IndexHandler(s))
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handler.UpdateByContentTypeHandler(s))
		r.Post("/{type}/{name}/{value}", handler.UpdateMetricsHandler(s))
	})
	r.Route("/updates", func(r chi.Router) {
		if privateKey != nil {
			r.With(signValidator, cryptoDecoder).Post("/", handler.BulkUpdateHandler(s))
		} else {
			r.With(signValidator).Post("/", handler.BulkUpdateHandler(s))
		}
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handler.ValueByContentTypeHandler(s))
		r.Get("/{type}/{name}", handler.ValueByContentTypeHandler(s))
	})
	r.Get("/ping", handler.PingDBHandler(db))

	switch storageSource {
	case fileStorageSource:
		ms, f, err := backup.RestoreMetrics(s, flagFileStoragePath, flagRestore)
		if err != nil {
			return err
		}
		defer f.Close()

		slog.Info("File is used as storage")

		go backup.RunWorker(ms, flagStoreInterval, f, make(chan struct{}))
	case dbStorageSource:
		err = pg.Bootstrap(db, ctx)
		if err != nil {
			slog.Error("bootstrap database", "err", err)
			return err
		}

		err = pg.ListMetrics(db, s)
		if err != nil {
			return err
		}
		slog.Info("Database is used as a storage")

		go pg.RunSaver(db, s, flagStoreInterval, make(chan struct{}))
	}

	slog.Info("Running server", "address", flagRunAddr)

	if flagGRPCAddr != "" {
		RunGRPCServer(s, flagGRPCAddr)
	}

	err = http.ListenAndServe(flagRunAddr, r)

	if err != nil {
		return err
	}

	return nil
}

func initConf() error {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}

	envStoreInterval := os.Getenv("STORE_INTERVAL")
	if envStoreInterval != "" {
		storeInterval, err := strconv.ParseInt(envStoreInterval, 10, 64)
		if err != nil {
			slog.Error("parse store interval", "err", err)
		}
		flagStoreInterval = storeInterval
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		flagFileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		flagRestoreStr = envRestore
	}

	var err error
	flagRestore, err = strconv.ParseBool(flagRestoreStr)
	if err != nil {
		slog.Error("parse restore", "err", err)
		os.Exit(1)
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		flagDatabaseDSN = envDatabaseDSN
	}

	if envEncryptionKey := os.Getenv("KEY"); envEncryptionKey != "" {
		flagEncryptionKey = envEncryptionKey
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		flagCryptoKeyPath = envCryptoKey
	}

	if flagCryptoKeyPath != "" {
		privateKey, err = crypto.ReadPrivateKey(flagCryptoKeyPath)
		if err != nil {
			return errors.Wrap(err, "[server.initConf] read private key")
		}
	}

	if flagDatabaseDSN != "" {
		storageSource = dbStorageSource
	} else if flagFileStoragePath != "" {
		storageSource = fileStorageSource
	} else {
		storageSource = memoryStorageSource
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		flagTrustedSubnet = envTrustedSubnet
	}

	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		flagConfigPath = envConfigPath
	}

	if flagConfigPath != "" {
		err := loadConfigs(flagConfigPath)
		if err != nil {
			return errors.Wrap(err, "[server.initConf] load config file")
		}
	}

	if err = logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	return nil
}

// RunGRPCServer initializes and starts a gRPC server.
func RunGRPCServer(metricsStorage *storage.MemStorage, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("Failed to start listener", slog.String("address", address), slog.Any("error", err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsServiceServer(grpcServer, grpcHandle.NewServer(metricsStorage))
	reflection.Register(grpcServer)

	slog.Info("Starting gRPC server", slog.String("address", address))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server stopped unexpectedly", slog.Any("error", err))
		}
	}()
}

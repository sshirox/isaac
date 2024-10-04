package server

import (
    "github.com/go-chi/chi/v5"
    chimiddleware "github.com/go-chi/chi/v5/middleware"
    "github.com/sshirox/isaac/internal/handler"
    "github.com/sshirox/isaac/internal/logger"
    "github.com/sshirox/isaac/internal/middleware"
    "github.com/sshirox/isaac/internal/storage"
    "go.uber.org/zap"
    "log/slog"
    "net/http"
    "os"
    "path"
    "strconv"
    "time"
)

func Run() error {
    parseFlags()
    err := initConf()

    if err != nil {
        return err
    }

    r := chi.NewRouter()
    r.Use(chimiddleware.Recoverer)
    r.Use(logger.WithLogging)
    r.Use(middleware.GZipMiddleware)
    s := storage.NewMemStorage()

    r.Get("/", handler.IndexHandler(s))
    r.Route("/update", func(r chi.Router) {
        r.Post("/", handler.UpdateByContentTypeHandler(s))
        r.Post("/{type}/{name}/{value}", handler.UpdateMetricsHandler(s))
    })
    r.Route("/value", func(r chi.Router) {
        r.Post("/", handler.ValueByContentTypeHandler(s))
        r.Get("/{type}/{name}", handler.ValueByContentTypeHandler(s))
    })

    s, f, err := restoreMetrics(s)
    if err != nil {
        return err
    }
    defer f.Close()

    go backupWorker(s, f, make(chan struct{}))

    logger.Log.Info("Running server", zap.String("address", flagRunAddr))

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
            slog.Error("parse store interval", "error", err)
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
        slog.Error("parse restore", "error", err)
        os.Exit(1)
    }

    if err = logger.Initialize(flagLogLevel); err != nil {
        return err
    }

    return nil
}

func restoreMetrics(ms *storage.MemStorage) (*storage.MemStorage, *os.File, error) {
    var err error
    var f *os.File

    if err = os.MkdirAll(flagFileStoragePath, 0755); err != nil {
        slog.Error("create backup dir", "error", err)
        return nil, nil, err
    }

    fp := path.Join(flagFileStoragePath, "metrics.bk")
    f, err = os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        slog.Error("open file", "error", err)
        return nil, nil, err
    }

    if flagRestore {
        err = ms.RestoreMetrics(f)
        if err != nil {
            slog.Error("restore metrics", "error", err)
            return nil, nil, err
        }
    }
    return ms, f, nil
}

func backupWorker(ms *storage.MemStorage, f *os.File, sc chan struct{}) {
    t := time.NewTicker(time.Duration(flagStoreInterval) * time.Second)
    for {
        select {
        case <-t.C:
            if err := ms.BackupMetrics(f); err != nil {
                slog.Error("backup metrics", "error", err)
            }
        case <-sc:
            slog.Info("stop backup worker")
            return
        }
    }
}

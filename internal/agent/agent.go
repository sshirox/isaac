package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sync/errgroup"

	"github.com/sshirox/isaac/internal/compress"
	"github.com/sshirox/isaac/internal/crypto"
	errs "github.com/sshirox/isaac/internal/errors"
	"github.com/sshirox/isaac/internal/metric"
	"github.com/sshirox/isaac/internal/ratelimit"
	"github.com/sshirox/isaac/internal/retries"
)

const (
	proto                 = "http"
	updateMetricsPath     = "update"
	bulkUpdateMetricsPath = "updates"
)

type Monitor struct {
	gauges    map[string]float64
	pollCount int64
	client    *resty.Client
	encoder   *crypto.Encoder
	limiter   *ratelimit.Limiter
}

func (mt *Monitor) pollMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	gauges := map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": m.GCCPUFraction,
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"NumGC":         float64(m.NumGC),
		"OtherSys":      float64(m.OtherSys),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}

	vmem, err := mem.VirtualMemory()
	if err == nil {
		gauges["FreeMemory"] = float64(vmem.Free)
		gauges["TotalMemory"] = float64(vmem.Total)
	}

	counts, err := cpu.Counts(true)
	if err == nil {
		gauges["CPUutilization1"] = float64(counts)
	}

	mt.gauges = gauges
	mt.pollCount++
}

func Run() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		cancel()
	}()

	parseFlags()
	initConf()

	encoder := crypto.NewEncoder(flagEncryptionKey)
	limiter := ratelimit.NewLimiter(flagRateLimit)
	mt := Monitor{
		encoder: encoder,
		client:  resty.New(),
		limiter: limiter,
	}

	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer reportTicker.Stop()

	group, groupCtx := errgroup.WithContext(ctx)

	go func() {
		<-ctx.Done()

		slog.Info("[agent.Run] Graceful shutdown agent")
	}()

	slog.Info("[agent.Run] Agent launched for sending metrics to", "address", serverAddr)

	group.Go(func() error {
		for {
			select {
			case <-groupCtx.Done():
				slog.Info("[agent.Run] Stopped poll metrics")
				return nil
			case <-pollTicker.C:
				slog.Info("[agent.Run] Poll metrics")
				mt.pollMetrics()
			}
		}
	})

	group.Go(func() error {
		for {
			select {
			case <-groupCtx.Done():
				slog.Info("[agent.Run] Stopped bulk sending metrics")
				return nil
			case <-reportTicker.C:
				slog.Info("[agent.Run] Send report")
				err := mt.bulkSendMetrics()
				if err != nil {
					slog.Error("[agent.Run] bulk sending metrics", "error", err)
				}
			}
		}
	})

	if err := group.Wait(); err != nil {
		slog.ErrorContext(ctx, "Run agent", "err", err)
	}
}

func initConf() {
	addrFromEnv := os.Getenv("ADDRESS")
	reportIntervalFromEnv := os.Getenv("REPORT_INTERVAL")
	pollIntervalFromEnv := os.Getenv("POLL_INTERVAL")

	if addrFromEnv != "" {
		serverAddr = addrFromEnv
	} else {
		serverAddr = flagServerAddr
	}

	if reportIntervalFromEnv != "" {
		i, _ := strconv.Atoi(reportIntervalFromEnv)
		reportInterval = int64(i)
	} else {
		reportInterval = flagReportInterval
	}

	if pollIntervalFromEnv != "" {
		i, _ := strconv.Atoi(pollIntervalFromEnv)
		pollInterval = int64(i)
	} else {
		pollInterval = flagPollInterval
	}

	if envEncryptionKey := os.Getenv("KEY"); envEncryptionKey != "" {
		flagEncryptionKey = envEncryptionKey
	}

	if envRateLimitValue := os.Getenv("RATE_LIMIT"); envRateLimitValue != "" {
		limit, err := strconv.Atoi(envRateLimitValue)
		if err != nil {
			slog.Error("rate limit conv", "err", err)
		} else {
			flagRateLimit = int64(limit)
		}
	}
}

func (mt *Monitor) processReport() error {
	for id, value := range mt.gauges {
		m := metric.Metrics{
			ID:    id,
			MType: metric.GaugeMetricType,
			Value: &value,
		}

		err := sendMetric(m)
		if err != nil {
			return err
		}
	}

	pollCount := metric.Metrics{
		ID:    "PollCount",
		MType: metric.CounterMetricType,
		Delta: &mt.pollCount,
	}

	err := sendMetric(pollCount)
	if err != nil {
		return err
	}

	return nil
}

func sendMetric(metric metric.Metrics) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(metric); err != nil {
		return err
	}
	compressed, err := compress.GZipCompress(buf.Bytes())
	if err != nil {
		return err
	}

	client := resty.New()
	err = retries.Retry(func() error {
		resp, respErr := client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(compressed).
			Post(sendMetricAddr())

		if respErr != nil {
			slog.Error("send request", "err", respErr)
			return errs.ErrConnection
		}

		statusCode := resp.StatusCode()

		if statusCode >= http.StatusInternalServerError && statusCode <= http.StatusGatewayTimeout {
			slog.Error("got retry status", "code", statusCode)
			return errs.ErrServer
		}

		if statusCode != http.StatusOK {
			slog.Error("got non retry status", "code", statusCode)
			return errs.ErrNonRetry
		}

		return nil
	})

	if err != nil {
		slog.Error("sending metric", "err", metric)
	}

	return nil
}

func (mt *Monitor) bulkSendMetrics() error {
	slog.Info("[Bulk_Send_Metrics] Start sending metrics")

	var metrics []metric.Metrics
	var err error

	for id, val := range mt.gauges {
		m := metric.Metrics{
			ID:    id,
			MType: metric.GaugeMetricType,
			Value: &val,
		}
		metrics = append(metrics, m)
	}

	pc := metric.Metrics{
		ID:    "PollCount",
		MType: metric.CounterMetricType,
		Delta: &mt.pollCount,
	}
	metrics = append(metrics, pc)

	slog.Info("[Bulk_Send_Metrics] metrics", "set", metrics)

	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(metrics); err != nil {
		slog.Error("[Bulk_Send_Metrics] encode metrics", "err", err)
		return err
	}
	compressedData, err := compress.GZipCompress(buf.Bytes())
	if err != nil {
		slog.Error("[Bulk_Send_Metrics] compress metrics", "err", err)
		return err
	}

	err = retries.Retry(func() error {
		req := mt.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(compressedData)

		if mt.encoder.IsEnabled() {
			req = req.SetHeader(crypto.SignHeader, mt.encoder.Encode(buf.Bytes()))
		}

		mt.limiter.Acquire()
		defer mt.limiter.Release()

		resp, respErr := req.Post(bulkSendMetricsAddr())

		if respErr != nil {
			slog.Error("send request", "err", respErr)
			return errs.ErrConnection
		}

		statusCode := resp.StatusCode()

		if statusCode >= http.StatusInternalServerError && statusCode <= http.StatusGatewayTimeout {
			slog.Error("got retry status", "code", statusCode)
			return errs.ErrServer
		}

		if statusCode != http.StatusOK {
			slog.Error("got non retry status", "code", statusCode)
			return errs.ErrNonRetry
		}

		return nil
	})

	if err != nil {
		slog.Error("[Bulk_Send_Metrics] sending metrics", "err", err)
	}

	return nil
}

func sendMetricAddr() string {
	addr := fmt.Sprintf("%s://%s/%s", proto, serverAddr, updateMetricsPath)

	return addr
}

func bulkSendMetricsAddr() string {
	addr := fmt.Sprintf("%s://%s/%s", proto, serverAddr, bulkUpdateMetricsPath)

	return addr
}

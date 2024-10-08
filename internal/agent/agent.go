package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sshirox/isaac/internal/compress"
	"github.com/sshirox/isaac/internal/metric"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	proto             = "http"
	updateMetricsPath = "update"
)

type Monitor struct {
	gauges    map[string]float64
	pollCount int64
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

	mt.gauges = gauges
	mt.pollCount++
}

func Run() {
	parseFlags()
	initConf()
	mt := Monitor{}

	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)

	slog.Info("Agent launched for sending metrics to", "address", serverAddr)

	for {
		select {
		case <-pollTicker.C:
			slog.Info("Poll metrics")
			mt.pollMetrics()
		case <-reportTicker.C:
			slog.Info("Send report")
			err := mt.processReport()
			if err != nil {
				panic(err)
			}
		}
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
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressed).
		Post(sendMetricAddr())

	if err != nil {
		slog.Error("sending metric", "error", metric)
	} else {
		slog.Info("sending metric", "status code", resp.StatusCode)

		if resp.StatusCode() != http.StatusOK {
			slog.Error("sending metric", "invalid status code", resp.StatusCode)
			slog.Error("sending metric", "body", string(resp.Body()))
		}
	}

	return nil
}

func sendMetricAddr() string {
	addr := fmt.Sprintf("%s://%s/%s", proto, serverAddr, updateMetricsPath)

	return addr
}

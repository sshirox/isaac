package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

var (
	flagServerAddr     string
	flagReportInterval int64
	flagPollInterval   int64
)

const (
	gaugeMetricType   = "gauge"
	counterMetricType = "counter"
	updateMetricsURL  = "update"
)

type GaugeMonitor struct {
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	RandomValue   float64
}

type CounterMonitor struct {
	PollCount int64
}

func pollMetrics(gm *GaugeMonitor, cm *CounterMonitor, memStats *runtime.MemStats) {
	// Read full mem stats
	runtime.ReadMemStats(memStats)

	gm.Alloc = float64(memStats.Alloc)
	gm.BuckHashSys = float64(memStats.BuckHashSys)
	gm.Frees = float64(memStats.Frees)
	gm.GCCPUFraction = float64(memStats.GCCPUFraction)
	gm.GCSys = float64(memStats.GCSys)
	gm.HeapAlloc = float64(memStats.HeapAlloc)
	gm.HeapIdle = float64(memStats.HeapIdle)
	gm.HeapInuse = float64(memStats.HeapInuse)
	gm.HeapObjects = float64(memStats.HeapObjects)
	gm.HeapReleased = float64(memStats.HeapReleased)
	gm.HeapSys = float64(memStats.HeapSys)
	gm.LastGC = float64(memStats.LastGC)
	gm.Lookups = float64(memStats.Lookups)
	gm.MCacheInuse = float64(memStats.MCacheInuse)
	gm.MCacheSys = float64(memStats.MCacheSys)
	gm.MSpanInuse = float64(memStats.MSpanInuse)
	gm.MSpanSys = float64(memStats.MSpanSys)
	gm.Mallocs = float64(memStats.Mallocs)
	gm.NextGC = float64(memStats.NextGC)
	gm.NumForcedGC = float64(memStats.NumForcedGC)
	gm.NumGC = float64(memStats.NumGC)
	gm.OtherSys = float64(memStats.OtherSys)
	gm.PauseTotalNs = float64(memStats.PauseTotalNs)
	gm.StackInuse = float64(memStats.StackInuse)
	gm.StackSys = float64(memStats.StackSys)
	gm.Sys = float64(memStats.Sys)
	gm.TotalAlloc = float64(memStats.TotalAlloc)
	gm.RandomValue = float64(rand.Float64())

	cm.PollCount += 1
}

func main() {
	parseFlags()

	run()
}

func parseFlags() {
	flag.StringVar(&flagServerAddr, "a", "localhost:8080", "server address and port")
	flag.Int64Var(&flagReportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&flagPollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()
}

func run() {
	var gm GaugeMonitor
	var cm CounterMonitor
	var memStats runtime.MemStats

	client := &http.Client{}

	pollTicker := time.NewTicker(time.Duration(flagPollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(flagReportInterval) * time.Second)

	for {
		select {
		case pt := <-pollTicker.C:
			msg := fmt.Sprintf("Poll metrics %v", pt)
			slog.Info(msg)
			pollMetrics(&gm, &cm, &memStats)
		case rt := <-reportTicker.C:
			msg := fmt.Sprintf("Send report %v", rt)
			slog.Info(msg)
			sendReport(client, gm, cm)
		}
	}
}

func sendReport(client *http.Client, gm GaugeMonitor, cm CounterMonitor) {
	gmv := reflect.ValueOf(&gm).Elem()

	for i := 0; i < gmv.NumField(); i++ {
		name := gmv.Type().Field(i).Name
		value := gmv.Field(i).Interface()
		if cnvVal, ok := value.(float64); ok {
			sendGaugeMetric(client, name, fmt.Sprintf("%.2f", cnvVal))
		}
	}

	sendCounterMetric(client, cm)
}

func sendGaugeMetric(client *http.Client, name string, value string) {
	url := fmt.Sprintf("http://%s/%s/%s/%s/%s", flagServerAddr, updateMetricsURL, gaugeMetricType, name, value)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error %s", err)
		return
	}

	msg := fmt.Sprintf("Sent gauge metric %s", url)
	slog.Info(msg)
}

func sendCounterMetric(client *http.Client, cm CounterMonitor) {
	url := fmt.Sprintf("http://%s/%s/%s/PollCount/%d", flagServerAddr, updateMetricsURL, counterMetricType, cm.PollCount)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error %s", err)
		return
	}

	msg := fmt.Sprintf("Sent counter metric %s", url)
	slog.Info(msg)
}

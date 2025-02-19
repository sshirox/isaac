package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/sshirox/isaac/internal/metric"
	pb "github.com/sshirox/isaac/internal/proto/metrics/proto"
	"github.com/sshirox/isaac/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"strings"
	"sync"
)

// Server implements the gRPC server for handling metrics.
type Server struct {
	pb.UnimplementedMetricsServiceServer
	storage *storage.MemStorage
	mu      sync.Mutex
}

// NewServer creates a new instance of the gRPC server.
func NewServer(storage *storage.MemStorage) *Server {
	return &Server{storage: storage}
}

// SendMetrics processes metric submission.
func (s *Server) SendMetrics(ctx context.Context, req *pb.SendMetricsRequest) (*pb.SendMetricsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request canceled: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var updatedMetrics []*pb.Metric
	var errorMessages []string

	for _, m := range req.Metrics {
		switch m.Kind {
		case metric.CounterMetricType:
			if m.Delta == nil {
				slog.Warn("Skipped counter update: delta is nil", slog.String("metric", m.Name))
				errorMessages = append(errorMessages, fmt.Sprintf("delta is nil for metric: %s", m.Name))
				continue
			}
			s.storage.UpdateCounter(m.Name, *m.Delta)

		case metric.GaugeMetricType:
			if m.Value == nil {
				slog.Warn("Skipped gauge update: value is nil", slog.String("metric", m.Name))
				errorMessages = append(errorMessages, fmt.Sprintf("value is nil for metric: %s", m.Name))
				continue
			}
			s.storage.UpdateGauge(m.Name, *m.Value)

		default:
			slog.Warn("Unknown metric type", slog.String("type", m.Kind), slog.String("metric", m.Name))
			errorMessages = append(errorMessages, fmt.Sprintf("unknown metric type: %s", m.Kind))
			continue
		}
		updatedMetrics = append(updatedMetrics, m)
	}

	if len(errorMessages) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Errors while updating metrics:\n%s", errors.New(strings.Join(errorMessages, "\n")))
	}

	slog.Info("Metrics successfully updated", slog.Int("count", len(updatedMetrics)))
	return &pb.SendMetricsResponse{Metrics: updatedMetrics}, nil
}

// GetMetrics returns all stored metrics.
func (s *Server) GetMetrics(ctx context.Context, req *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request canceled: %v", err)
	}

	var metrics []*pb.Metric

	for name, value := range s.storage.ReceiveAllGauges() {
		metrics = append(metrics, &pb.Metric{
			Name:  name,
			Kind:  metric.GaugeMetricType,
			Value: &value,
		})
	}
	for name, value := range s.storage.ReceiveAllCounters() {
		metrics = append(metrics, &pb.Metric{
			Name:  name,
			Kind:  metric.CounterMetricType,
			Delta: &value,
		})
	}

	slog.Info("All metrics requested", slog.Int("count", len(metrics)))
	return &pb.GetMetricsResponse{Metrics: metrics}, nil
}

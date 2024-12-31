package compress

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sshirox/isaac/internal/metric"
)

func TestGZipCompress(t *testing.T) {
	value := 9765.77
	mt := metric.Metrics{
		ID:    "Alloc",
		MType: metric.GaugeMetricType,
		Value: &value,
	}
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(mt)
	require.NoError(t, err)

	_, err = GZipCompress(buf.Bytes())
	require.NoError(t, err)
}

package tatakae

import (
	"context"
	"io"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/require"
)

func TestDatapointGeneration(t *testing.T) {
	metrics := NewMetrics()
	require.NotNil(t, metrics)
	logs := NewLogs()
	require.NotNil(t, logs)
	traces := NewTraces()
	require.NotNil(t, traces)
}

func TestNewExporter(t *testing.T) {
	ms, ls, ts, err := NewOTLPHTTPExporter(context.Background(), log.NewLogfmtLogger(io.Discard), DefaultConfig)
	require.NoError(t, err)
	require.NotNil(t, ms)
	require.NotNil(t, ls)
	require.NotNil(t, ts)
}

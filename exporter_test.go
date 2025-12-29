package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type mockSensor struct {
	humidity    float64
	temperature float64
}

func (m *mockSensor) Get() (float64, float64, error) { return m.humidity, m.temperature, nil }
func (m *mockSensor) Clean()                         {}

func TestMetricsEndpointRespondsWithSensorValues(t *testing.T) {
	sensor := &mockSensor{humidity: 55.5, temperature: 22.25}
	exporter := NewExporter(sensor, 10*time.Second)
	exporter.PollOnce()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(exporter.Registry, promhttp.HandlerOpts{}))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	body := string(bodyBytes)

	if !strings.Contains(body, "sensor_humidity 55.5") {
		t.Fatalf("missing humidity metric, got:\n%s", body)
	}
	if !strings.Contains(body, "sensor_temperature 22.25") {
		t.Fatalf("missing temperature metric, got:\n%s", body)
	}
}


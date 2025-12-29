package main

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Sensor interface {
	Get() (humidity float64, temperature float64, err error)
	Clean()
}

type Exporter struct {
	Sensor   Sensor
	Interval time.Duration
	Registry *prometheus.Registry

	humidityGauge    prometheus.Gauge
	temperatureGauge prometheus.Gauge
}

func NewExporter(sensor Sensor, interval time.Duration) *Exporter {
	reg := prometheus.NewRegistry()
	factory := promauto.With(reg)

	return &Exporter{
		Sensor:   sensor,
		Interval: interval,
		Registry: reg,
		humidityGauge: factory.NewGauge(prometheus.GaugeOpts{
			Name: "sensor_humidity",
			Help: "Humidity of a sensor in %RH.",
		}),
		temperatureGauge: factory.NewGauge(prometheus.GaugeOpts{
			Name: "sensor_temperature",
			Help: "Temperature of a sensor in Celcius temperature.",
		}),
	}
}

func (e *Exporter) PollOnce() {
	hum, tmp, err := e.Sensor.Get()
	if err != nil {
		log.Printf("[error] %v", err)
		return
	}
	e.humidityGauge.Set(hum)
	e.temperatureGauge.Set(tmp)
}

func (e *Exporter) Run(ctx context.Context) {
	e.PollOnce()

	ticker := time.NewTicker(e.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.PollOnce()
		}
	}
}


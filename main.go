package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xruins/dht20-exporter/dht20"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "2112", "the port to listen on")
	flag.Parse()

	sensor, err := dht20.New()
	if err != nil {
		log.Fatal(err)
	}
	defer sensor.Clean()

	exporter := NewExporter(sensor, 30*time.Second)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go exporter.Run(ctx)

	http.Handle("/metrics", promhttp.HandlerFor(exporter.Registry, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

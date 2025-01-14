package main

import (
	"log"
	"net/http"
	"time"

	"github.com/W0n9/iot_bridge_golang/config"
	"github.com/W0n9/iot_bridge_golang/read_sensor"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	logger           *zap.SugaredLogger
	temperatureGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "temperature_metric_celsius",
			Help: "Temperature measured by the WRD Sensor",
		},
		[]string{"node", "campus", "building", "room"},
	)

	humidityGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "humidity_metric_ratio",
			Help: "Humidity percentage measured by the WRD Sensor",
		},
		[]string{"node", "campus", "building", "room"},
	)
)

func init() {
	// 初始化 zap logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	// 创建 SugaredLogger
	logger = zapLogger.Sugar()

	prometheus.MustRegister(temperatureGauge)
	prometheus.MustRegister(humidityGauge)
}

func monitorSensor(s config.Sensor) {
	for {
		reading, err := read_sensor.ReadSensor(s.IP, 80)
		if err != nil {
			logger.Errorw("Failed to read sensor",
				"ip", s.IP,
				"campus", s.Campus,
				"building", s.Building,
				"room", s.Room,
				"error", err,
			)
			temperatureGauge.DeleteLabelValues(s.IP, s.Campus, s.Building, s.Room)
			humidityGauge.DeleteLabelValues(s.IP, s.Campus, s.Building, s.Room)
			time.Sleep(15 * time.Second)
			continue
		}

		temperatureGauge.WithLabelValues(s.IP, s.Campus, s.Building, s.Room).Set(reading.Temperature)
		humidityGauge.WithLabelValues(s.IP, s.Campus, s.Building, s.Room).Set(reading.Humidity)

		time.Sleep(5 * time.Second)
	}
}

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalw("加载配置失败",
			"error", err,
		)
	}

	for _, s := range cfg.Sensors {
		go monitorSensor(s)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9580", nil))
}

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/W0n9/iot_bridge_golang/config"
	"github.com/W0n9/iot_bridge_golang/read_sensor"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	logger           *zap.SugaredLogger
	temperatureGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "temperature_metric_celsius",
			Help: "Temperature measured by the WRD Sensor",
		},
		[]string{"node", "campus", "building", "room"},
	)

	humidityGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "humidity_metric_ratio",
			Help: "Humidity percentage measured by the WRD Sensor",
		},
		[]string{"node", "campus", "building", "room"},
	)
)

func monitorSensor(s config.Sensor) {
	failureCount := 0
	baseDelay := 15 * time.Second
	maxDelay := 24 * time.Hour // 设置最大重试延迟为24小时

	for {
		reading, err := read_sensor.ReadSensor(s.IP, 80)
		if err != nil {
			failureCount++
			logger.Errorw("Failed to read sensor",
				"ip", s.IP,
				"campus", s.Campus,
				"building", s.Building,
				"room", s.Room,
				"error", err,
			)
			temperatureGauge.DeleteLabelValues(s.IP, s.Campus, s.Building, s.Room)
			humidityGauge.DeleteLabelValues(s.IP, s.Campus, s.Building, s.Room)

			// 计算指数退避等待时间，但不超过最大值
			delay := baseDelay * time.Duration(1<<uint(failureCount-1)) // 2^(failureCount-1) * baseDelay
			if delay > maxDelay {
				delay = maxDelay
			}

			// logger.Infow("Retrying after backoff",
			// 	"ip", s.IP,
			// 	"delay", delay.String(),
			// 	"failureCount", failureCount,
			// )
			time.Sleep(delay)
			continue
		}

		// 成功读取，重置失败计数
		if failureCount > 0 {
			logger.Infow("Sensor read successful after failures",
				"ip", s.IP,
				"campus", s.Campus,
				"building", s.Building,
				"room", s.Room,
				"previousFailures", failureCount,
			)
			failureCount = 0
		}

		if reading.Humidity != 0 {
			temperatureGauge.WithLabelValues(s.IP, s.Campus, s.Building, s.Room).Set(reading.Temperature)
			humidityGauge.WithLabelValues(s.IP, s.Campus, s.Building, s.Room).Set(reading.Humidity)
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {
	// 初始化 zap logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// 创建 SugaredLogger
	logger = zapLogger.Sugar()
	defer logger.Sync()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalw("LoadConfig failed",
			"error", err,
		)
	}

	for _, s := range cfg.Sensors {
		go monitorSensor(s)
	}

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9580", nil))
}

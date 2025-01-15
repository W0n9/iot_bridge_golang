package read_sensor

import (
	"bufio"
	"fmt"

	"net"
	"strconv"
	"strings"
	"time"
)

// SensorData 存储传感器返回的数据
type SensorData struct {
	Temperature float64
	Humidity    float64
	RawText     []string
}

// ReadSensor 读取传感器数据
func ReadSensor(serverIP string, serverPort int) (*SensorData, error) {
	// 因为传感器使用的是原始的TCP协议，所以需要使用socket来进行通信
	// 设置超时时间为2秒，避免超时阻塞主线程
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort), 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	// 读取数据
	var lines []string
	reader := bufio.NewReader(conn)
	for i := 0; i < 5; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, strings.TrimSpace(line))
	}

	if len(lines) < 3 {
		return nil, fmt.Errorf("insufficient data received")
	}

	// 解析温度
	var temp float64
	if strings.HasSuffix(lines[1], "C") {
		tempStr := strings.TrimSuffix(strings.Fields(lines[1])[len(strings.Fields(lines[1]))-1], "C")
		temp, err = strconv.ParseFloat(tempStr, 64)
		if err != nil {
			// 处理负数情况
			parts := strings.Split(tempStr, ".")
			if len(parts) == 2 {
				intPart, _ := strconv.Atoi(parts[0])
				decPart, _ := strconv.Atoi(parts[1][1:])
				temp = float64(-1.0) * float64(abs(intPart)+decPart) / 100
			}
		}
	}

	// 解析湿度
	var hum float64
	humStr := ""
	if strings.HasSuffix(lines[2], "%") {
		humStr = strings.TrimSuffix(strings.Fields(lines[2])[len(strings.Fields(lines[2]))-1], "%")
	} else if len(lines) > 3 {
		humStr = strings.TrimSuffix(strings.Fields(lines[3])[len(strings.Fields(lines[3]))-1], "%")
	}

	hum, err = strconv.ParseFloat(humStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse humidity: %v", err)
	}

	return &SensorData{
		Temperature: temp,
		Humidity:    hum,
		RawText:     lines,
	}, nil
}

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

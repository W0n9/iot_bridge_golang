package read_sensor_test

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// Helper function to parse sensor response without network connection
func parseResponse(response string) (float64, float64, error) {
	var err error
	lines := strings.Split(strings.TrimSpace(response), "\n")

	if len(lines) < 3 {
		return 0, 0, fmt.Errorf("insufficient data received")
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
				temp = float64(-1.0) * (float64(abs(intPart)) + float64(decPart)*0.01)
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
		return 0, 0, fmt.Errorf("failed to parse humidity: %v", err)
	}

	return temp, hum, nil
}

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		wantTemp      float64
		wantHum       float64
		wantErr       bool
		errorContains string
	}{
		{
			name: "normal reading",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 22.82C
Humidity = 36.73%`,
			wantTemp: 22.82,
			wantHum:  36.73,
		},
		{
			name: "small negative temperature",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 0.-3C
Humidity = 36.73%`,
			wantTemp: -0.03,
			wantHum:  36.73,
		},
		{
			name: "large negative temperature",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = -1.-38C
Humidity = 36.73%`,
			wantTemp: -1.38,
			wantHum:  36.73,
		},
		{
			name: "very low temperature",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = -10.-5C
Humidity = 36.73%`,
			wantTemp: -10.05,
			wantHum:  36.73,
		},
		{
			name: "zero values",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 0.0C
Humidity = 0.00%`,
			wantTemp: 0.0,
			wantHum:  0.0,
		},
		{
			name: "maximum values",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 100.0C
Humidity = 100.00%`,
			wantTemp: 100.0,
			wantHum:  100.0,
		},
		{
			name: "insufficient data",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 22.82C`,
			wantErr:       true,
			errorContains: "insufficient data",
		},
		{
			name: "invalid temperature format",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = invalidC
Humidity = 36.73%`,
			wantErr:       true,
			errorContains: "failed to parse",
		},
		{
			name: "invalid humidity format",
			response: `Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 22.82C
Humidity = invalid%`,
			wantErr:       true,
			errorContains: "failed to parse humidity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTemp, gotHum, err := parseResponse(tt.response)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if math.Abs(gotTemp-tt.wantTemp) > 0.001 {
				t.Errorf("temperature = %v, want %v", gotTemp, tt.wantTemp)
			}
			if math.Abs(gotHum-tt.wantHum) > 0.001 {
				t.Errorf("humidity = %v, want %v", gotHum, tt.wantHum)
			}
		})
	}
}

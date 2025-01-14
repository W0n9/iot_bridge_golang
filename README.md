# IoT_Bridge_Golang

## Use Prometheus to monitor Temperature and Humidity

### How to use

1. Create a config file, for example:

```yaml
sensors:
  - ip: 1.2.3.4
    campus: Main
    building: Tower-1
    room: 34F
  - ip: 5.6.7.8
    campus: 昌平校区
    building: 第二实验楼文科楼
    room: 一层1-5汇聚机房
```

2. Run the following command:

```docker
docker run -itd -p 9580:9580 -v /path/to/config.yaml:/app/config.yaml w0n9/iot_bridge_golang
```

### Example of Response from Sensors

```
Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 22.82C
Humidity = 36.73%
```

or a Bug when the temperature is lower than 0C

```
Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = 0.-3C # Eq. to -0.03C
Humidity = 36.73%
```

or a Bug when the temperature is lower than -1C

```
Copyright (c) 2010 WRD Tech. Co., Ltd. All rights reserved.
Temperature = -1.-38 # Eq. to -1.38C
Humidity = 36.73%
```

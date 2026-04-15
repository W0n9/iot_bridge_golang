# IoT Bridge 优化设计方案

> 日期：2026-04-15
> 状态：待评审

---

## 背景

本项目是一个基于 Go 的 IoT 传感器数据采集工具，通过 TCP 协议读取温湿度传感器数据，并暴露 Prometheus 指标供监控系统采集。

在代码审查中发现存在以下问题：
1. 核心功能存在 bug 影响数据正确性
2. Docker 镜像体积未优化
3. 日志时区为 UTC，运维排查时需要时间转换

---

## 目标

通过渐进式修复方案，在低风险的前提下：
1. 修复影响核心功能的 bug
2. 减小 Docker 镜像体积
3. 提升运维友好性

---

## 方案：渐进式修复

### 第一阶段：核心 bug 修复

#### 1.1 温度解析 Bug

**文件：** `read_sensor.go:58`

**问题描述：**

传感器在负温度时的输出格式存在异常：
- `-0.03°C` 显示为 `0.-3C`
- `-1.38°C` 显示为 `-1.-38C`

当前解析逻辑错误：
```go
// 错误实现
intPart, _ := strconv.Atoi(parts[0])      // -1
decPart, _ := strconv.Atoi(parts[1][1:])   // 38
temp = float64(-1.0) * float64(abs(intPart)+decPart) / 100
// 计算结果：-(1+38)/100 = -0.39 ❌ (期望 -1.38)
```

**修复方案：**
```go
// 正确实现（与测试文件 read_sensor_test.go:31 对齐）
intPart, _ := strconv.Atoi(parts[0])
decPart, _ := strconv.Atoi(parts[1][1:])
temp = float64(-1.0) * (float64(abs(intPart)) + float64(decPart)*0.01)
// 计算结果：-(1 + 0.38) = -1.38 ✅
```

---

#### 1.2 指数退避 Bug

**文件：** `main.go:56`

**问题描述：**

当 `failureCount = 0` 时，表达式 `1<<uint(failureCount-1)` 会产生未定义行为：
- `failureCount - 1 = -1`
- `uint(-1)` 在 64 位系统上为 `18446744073709551615`
- `1 << 18446744073709551615` 行为未定义

**修复方案：**
```go
// 修改前
delay := baseDelay * time.Duration(1<<uint(failureCount-1))

// 修改后
delay := baseDelay * time.Duration(1<<uint(failureCount))
```

**验证：**
| failureCount | 修改前 | 修改后 |
|-------------|-------|-------|
| 0 | 未定义 | 15s |
| 1 | 15s | 30s |
| 2 | 30s | 60s |
| 3 | 60s | 120s |

---

### 第二阶段：日志降噪（可选）

**文件：** `main.go:45-51`

**问题描述：**

当前每次读取失败都记录 Error 级别日志。在指数退避期间（最长 24 小时），会产生大量重复日志。

**优化方案：**
```go
// 仅在首次失败时记录 Error
if failureCount == 0 {
    logger.Errorw("Sensor network unreachable", "ip", s.IP, ...)
}
// 后续失败静默，直到恢复

// 恢复时记录 Info
if failureCount > 0 && err == nil {
    logger.Infow("Sensor network recovered", "previousFailures", failureCount)
}
```

**收益：** 日志量减少约 90%

---

### 第三阶段：Docker 优化

#### 3.1 二进制体积优化

**文件：** `Dockerfile:16`

**修改：**
```dockerfile
# 修改前
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 修改后（移除调试符号）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .
```

**参数说明：**
- `-s`：移除符号表
- `-w`：移除 DWARF 调试信息

**预期收益：** 二进制体积减少约 50%（从 ~10MB 降至 ~5MB）

---

#### 3.2 时区设置

**文件：** `Dockerfile`

**修改：**
```dockerfile
# 在 FROM alpine:3.21 后添加
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
```

**收益：** 日志时间戳为本地时间（Asia/Shanghai），无需转换

---

## 实施计划

| 阶段 | 文件 | 修改行数 | 风险等级 | 预计工时 |
|-----|-----|---------|---------|---------|
| 第一阶段 | `read_sensor.go` | ~1 | 低 | 10 分钟 |
| 第一阶段 | `main.go` | ~1 | 低 | 10 分钟 |
| 第二阶段 | `main.go` | ~5 | 低 | 15 分钟 |
| 第三阶段 | `Dockerfile` | ~3 | 低 | 10 分钟 |

**总计：** 约 45 分钟

---

## 测试方案

### 第一阶段测试

**温度解析测试：**
```bash
go test -v ./read_sensor/... -run TestParseResponse
```

预期所有测试用例通过，包括：
- `small negative temperature`：`0.-3C` → `-0.03`
- `large negative temperature`：`-1.-38C` → `-1.38`
- `very low temperature`：`-10.-5C` → `-10.05`

### 第三阶段测试

**Docker 镜像构建：**
```bash
docker build -t iot_bridge:test .
docker images | grep iot_bridge
```

验证镜像大小减少，并检查日志时区：
```bash
docker run --rm iot_bridge:test
# 日志时间应为 Asia/Shanghai
```

---

## 风险评估

| 风险项 | 概率 | 影响 | 缓解措施 |
|-------|-----|-----|---------|
| 温度解析逻辑遗漏边界情况 | 低 | 中 | 现有测试覆盖完整 |
| 指数退避修改影响现有行为 | 低 | 低 | 修复 bug 不改变设计 |
| Docker 构建失败 | 低 | 低 | 本地验证后再提交 |

---

## 待确认事项

无。

---

## 参考文献

- Go 官方文档：[go build -ldflags](https://pkg.go.dev/cmd/go#hdr-Adding_ldflags_during_the_build)
- Prometheus 最佳实践：[Metric Naming](https://prometheus.io/docs/practices/naming/)

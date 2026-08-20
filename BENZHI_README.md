# et0-fao56

et0-fao56 是 FAO-56 参考作物腾发核算工具：输入日气象与作物系数，按彭曼–蒙蒂斯日尺度公式计算 ET0（mm/d）与 ETc = Kc·ET0，
固定系数 0.408 / 900 / 0.34，Δ 由 es(T) 同源求导、γ 随气压与潜热变化，辐射项与空气动力项共用同一套温度单位。
纯标准库，无网络依赖，无 cgo。

## 构建 / 运行 / 测试

```text
go build ./...             # 编译
go run . -et0 example/arid-day.json   # CLI：ET0 分项（-calm 1 附无风对照）
go run . -etc example/arid-day.json   # CLI：ET0 + ETc
go run . -http :8080       # Web 控制台
go test ./...              # 单元测试（meteo / penman / crop）
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

两种架构都要构建并进容器验证：

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```

# et0-fao56：Go FAO-56 参考作物腾发 Web 服务（彭曼–蒙蒂斯 ET0/ETc + 前端控制台）

et0-fao56 按 FAO-56 彭曼–蒙蒂斯日尺度公式计算参考作物腾发 ET0（mm/d）与 ETc = Kc·ET0。输入日气象与作物系数，固定系数 0.408 / 900 / 0.34；提供 `/api/et0`、`/api/etc` 与嵌入 Web 控制台。

## 构建 / 运行 / 测试

```text
go build ./...
./et0-fao56 -http :8080
curl -s http://127.0.0.1:8080/api/example
curl -s -X POST http://127.0.0.1:8080/api/et0 -d @example/arid-day.json
go test ./...
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -d -P --name et0-b14 <image-name>:latest
curl -s http://127.0.0.1:$(docker port et0-b14 8080 | cut -d: -f2)/api/example
docker rm -f et0-b14
```

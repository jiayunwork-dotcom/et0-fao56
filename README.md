# et0-fao56

et0-fao56 是一个 FAO-56 参考作物腾发（reference evapotranspiration）核算工具：给定日气象（净辐射 Rn、土壤热通量 G、气温 T、2 m 风速 u2、相对湿度 RH 或直接的水汽压 ea、气压或海拔），按彭曼–蒙蒂斯公式计算 ET0（mm/d），并按作物系数算出 ETc = Kc·ET0（支持单值 Kc 或 ini/mid/end 三段生长曲线，可选土壤水分胁迫系数 Ks∈(0,1]）。

公式（FAO-56 日尺度，本仓钉死）：

```
ET0 = (0.408·Δ·(Rn−G) + γ·(900/(T+273))·u2·(es−ea)) / (Δ + γ·(1 + 0.34·u2))
```

固定系数 0.408（MJ/m²·d → mm/d）、900（日尺度空气动力项分子）、0.34（日尺度分母风速系数）均为 FAO-56 日尺度取值；若改用小时公式必须整表换系数并全程一致。Δ 为饱和水汽压对温度的斜率（kPa/°C），由同一 Tetens 曲线解析求导；γ 为湿度计常数（kPa/°C），γ = cp·P/(ε·λ)，随气压与潜热变化。es 用气温计算，ea 由 RH 或直接给定；当多个湿度通道同时给出时必须互相一致。

温度单位纪律：辐射项与空气动力项共用同一套 Δ、γ、λ 与同一温度读数——Δ 由 es(T) 在摄氏下求导，空气动力项分母用 T+273（FAO-56 日尺度取 273，气象学精确值为 273.15，本仓对两者都做了区分与审计），es 与 Δ 不允许一项用摄氏、一项用开尔文。`penman.Result` 的审计输出会给出 Δ 与 es 数值导数的相对偏差，以及 ET0·λ 与两项分子按能量单位还原的残差。

- 输入：JSON 文档（weather + 可选 crop），示例见 `example/arid-day.json`
- 输出：ET0、辐射项、空气动力项、Δ、γ、es/ea/饱和差、ETc 与分段 Kc 表
- 边界：仅日/小时气象驱动，不做灌溉制度排班；风速为负、RH 越界、λ 或气压非正、Rn−G 与 ET0 长期矛盾却声称「已蒸腾」等均以 error 返回，不会 panic
- 交叉校验：u2=0 时空气动力项消失、分母退化为 Δ+γ（辐射项主导）；其他条件不变时提高饱和差 es−ea，ET0 不下降；ET0·λ 与辐射/动力两项分子在能量单位下还原一致；Kc 从 ini 提到 mid，ETc 按比例增加

## 构建 / 运行 / 测试

```text
go build ./...          # 编译（纯标准库，无第三方依赖）
go test ./...           # 单元测试（meteo / penman / crop 三个包）
go run . -et0 example/arid-day.json            # CLI：ET0 分项
go run . -et0 example/arid-day.json -calm 1    # CLI：并排无风对照
go run . -etc example/arid-day.json            # CLI：ET0 + ETc
go run . -http :8080    # Web 控制台：http://localhost:8080
```

## Web 界面

`/` 提供静态页面：自动加载 `example/arid-day.json`，列出 ET0 分项（辐射项、空气动力项、Δ、γ、分母、es/ea、饱和差）与 ETc（Kc、阶段、Ks、ETc 及其无胁迫值），并给出风速扫描表。修改任一气象输入（如风速）后自动重新调用后端接口，数字全部来自 API。

## 命令行

```text
go run . -et0 <file.json> [-calm <任意值>] [-wind <m/s>]
go run . -etc <file.json> [-wind <m/s>]
go run . -http :8080
```

## API

- `POST /api/et0`：日气象 → ET0、辐射项、空气动力项、Δ、γ（可选风速扫描 `windSweep`）
- `POST /api/etc`：ET0 或气象 + Kc（+ 可选 Ks）→ ETc 与分段表
- 非法输入返回 JSON 错误体（如 `{"error": "..."}`），HTTP 400

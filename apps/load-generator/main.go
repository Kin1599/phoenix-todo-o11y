package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AttackRequest struct {
	URL         string `json:"url"`
	Method      string `json:"method"`
	Payload     string `json:"payload"`
	Headers     string `json:"headers"`
	RPS         int    `json:"rps"`
	Duration    int    `json:"duration"`
	LoadProfile string `json:"load_profile"`
}

type MetricsMessage struct {
	Total       uint64  `json:"total"`
	Success     uint64  `json:"success"`
	Fail        uint64  `json:"fail"`
	RPS         int     `json:"rps"`
	SuccessRate float64 `json:"success_rate"`
	Done        bool    `json:"done"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	currentRPS = atomic.Uint64{}
	rpsGauge   = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "loadgen_rps_current",
		Help: "Current RPS measured during test run",
	})
	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadgen_requests_total", Help: "Total requests sent",
	})
	requestsSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadgen_requests_success_total", Help: "Successful requests",
	})
	requestsFail = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "loadgen_requests_fail_total", Help: "Failed requests",
	})
	latencyHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "loadgen_response_latency_seconds",
		Help:    "Histogram of response latencies",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
	})
)

func main() {
	prometheus.MustRegister(requestsTotal, requestsSuccess, requestsFail, latencyHistogram, rpsGauge)

	r := gin.Default()
	r.GET("/", servePage)
	r.GET("/ws", handleWebSocket)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	fmt.Println("Listening on http://localhost:8081")
	r.Run(":8081")
}

func servePage(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, htmlPage)
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	var req AttackRequest
	if err := conn.ReadJSON(&req); err != nil {
		fmt.Println("Invalid JSON:", err)
		return
	}

	headerMap := http.Header{}
	for _, line := range strings.Split(req.Headers, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headerMap.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	target := vegeta.Target{
		Method: req.Method,
		URL:    req.URL,
		Header: headerMap,
	}
	if req.Method == "POST" || req.Method == "PUT" {
		target.Body = []byte(req.Payload)
	}

	attacker := vegeta.NewAttacker()
	targeter := vegeta.NewStaticTargeter(target)
	duration := time.Duration(req.Duration) * time.Second
	resCh := profileAttack(attacker, targeter, req.LoadProfile, req.RPS, duration)

	var total, success, fail uint64
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	done := make(chan struct{})

	go func() {
		var lastTotal uint64
		for {
			select {
			case <-ticker.C:
				newTotal := total
				delta := newTotal - lastTotal
				lastTotal = newTotal
				currentRPS.Store(delta)
				rpsGauge.Set(float64(delta))

				successRate := 0.0
				if total > 0 {
					successRate = (float64(success) / float64(total)) * 100
				}
				conn.WriteJSON(MetricsMessage{
					Total:       total,
					Success:     success,
					Fail:        fail,
					RPS:         int(delta),
					SuccessRate: successRate,
					Done:        false,
				})
			case <-done:
				return
			}
		}
	}()

	for res := range resCh {
		total++
		requestsTotal.Inc()
		latencyHistogram.Observe(res.Latency.Seconds())
		if res.Code >= 200 && res.Code < 300 {
			success++
			requestsSuccess.Inc()
		} else {
			fail++
			requestsFail.Inc()
		}
	}
	close(done)

	successRate := 0.0
	if total > 0 {
		successRate = (float64(success) / float64(total)) * 100
	}
	conn.WriteJSON(MetricsMessage{
		Total:       total,
		Success:     success,
		Fail:        fail,
		RPS:         int(currentRPS.Load()),
		SuccessRate: successRate,
		Done:        true,
	})
}

func profileAttack(attacker *vegeta.Attacker, targeter vegeta.Targeter, profile string, rps int, duration time.Duration) <-chan *vegeta.Result {
	switch strings.ToLower(profile) {
	case "хаотичная":
		return chaoticAttack(attacker, targeter, rps, duration)
	case "спайковая":
		return spikyAttack(attacker, targeter, rps, duration)
	case "умеренная":
		return attacker.Attack(targeter, vegeta.Rate{Freq: rps / 2, Per: time.Second}, duration, "moderate")
	case "волнообразная":
		return waveAttack(attacker, targeter, rps, duration)
	case "нагрев":
		return warmupAttack(attacker, targeter, rps, duration)
	case "ночной режим":
		return attacker.Attack(targeter, vegeta.Rate{Freq: rps / 10, Per: time.Second}, duration, "night")
	default:
		return attacker.Attack(targeter, vegeta.Rate{Freq: rps, Per: time.Second}, duration, "constant")
	}
}

func chaoticAttack(attacker *vegeta.Attacker, targeter vegeta.Targeter, rps int, duration time.Duration) <-chan *vegeta.Result {
	out := make(chan *vegeta.Result)
	go func() {
		defer close(out)
		end := time.Now().Add(duration)
		for time.Now().Before(end) {
			res := attacker.Attack(targeter, vegeta.Rate{Freq: rps, Per: time.Second}, 1*time.Second, "chaotic")
			for r := range res {
				out <- r
			}
			time.Sleep(time.Duration(100+time.Now().UnixNano()%300) * time.Millisecond)
		}
	}()
	return out
}

func spikyAttack(attacker *vegeta.Attacker, targeter vegeta.Targeter, rps int, duration time.Duration) <-chan *vegeta.Result {
	out := make(chan *vegeta.Result)
	go func() {
		defer close(out)
		end := time.Now().Add(duration)
		for time.Now().Before(end) {
			isSpike := time.Now().Unix()%5 < 2
			rate := rps
			if isSpike {
				rate *= 2
			}
			res := attacker.Attack(targeter, vegeta.Rate{Freq: rate, Per: time.Second}, 1*time.Second, "spiky")
			for r := range res {
				out <- r
			}
		}
	}()
	return out
}

func waveAttack(attacker *vegeta.Attacker, targeter vegeta.Targeter, rps int, duration time.Duration) <-chan *vegeta.Result {
	out := make(chan *vegeta.Result)
	go func() {
		defer close(out)
		end := time.Now().Add(duration)
		for time.Now().Before(end) {
			mod := time.Now().Unix() % 4
			rate := rps
			if mod >= 2 {
				rate /= 2
			}
			res := attacker.Attack(targeter, vegeta.Rate{Freq: rate, Per: time.Second}, 1*time.Second, "wave")
			for r := range res {
				out <- r
			}
		}
	}()
	return out
}

func warmupAttack(attacker *vegeta.Attacker, targeter vegeta.Targeter, rps int, duration time.Duration) <-chan *vegeta.Result {
	out := make(chan *vegeta.Result)
	go func() {
		defer close(out)
		end := time.Now().Add(duration)
		start := time.Now()
		for time.Now().Before(end) {
			t := time.Since(start).Seconds()
			currentRPS := int(float64(rps) * (0.5 + 0.5*t/float64(duration.Seconds())))
			res := attacker.Attack(targeter, vegeta.Rate{Freq: currentRPS, Per: time.Second}, 1*time.Second, "warmup")
			for r := range res {
				out <- r
			}
		}
	}()
	return out
}

const htmlPage = `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Vegeta WebSocket Dashboard</title>
	<style>
		body { font-family: sans-serif; background: #f5f6fa; margin: 0; padding: 20px; }
		.container { display: flex; flex-wrap: wrap; gap: 20px; max-width: 1200px; margin: auto; }
		.card {
			background: white;
			padding: 20px;
			border-radius: 10px;
			box-shadow: 0 4px 8px rgba(0,0,0,0.1);
			flex: 1 1 400px;
		}
		input, textarea, select {
			width: 100%;
			margin-top: 8px;
			padding: 10px;
			border: 1px solid #ccc;
			border-radius: 5px;
		}
		label { margin-top: 12px; display: block; font-weight: bold; }
		button {
			margin-top: 20px;
			padding: 12px;
			width: 100%;
			background: linear-gradient(to right, #4facfe, #00f2fe);
			border: none;
			color: white;
			font-size: 16px;
			border-radius: 6px;
			cursor: pointer;
		}
		h2, h3 { margin-bottom: 10px; }
		.metrics p {
			font-size: 18px;
			margin: 6px 0;
		}
		.metrics span {
			font-weight: bold;
			color: #2c3e50;
		}
	</style>
</head>
<body>
	<h1 style="text-align:center;">🔥 WebSocket UI для Vegeta</h1>
	<div class="container">
		<div class="card">
			<h2>Настройки теста</h2>
			<label>URL:<input id="url" value="http://app:4000/api/ping" /></label>
			<label>Метод:
				<select id="method">
					<option>GET</option>
					<option>POST</option>
				</select>
			</label>
			<label>Payload:<textarea id="payload">{}</textarea></label>
			<label>Заголовки:
			<textarea id="headers">Content-Type: application/json
			Authorization: Bearer token</textarea>
			</label>
			<label>RPS:<input type="number" id="rps" value="10" /></label>
			<label>Длительность (сек):<input type="number" id="duration" value="5" /></label>
			<label>Профиль нагрузки:
				<select id="loadProfile">
					<option>Постоянная</option>
					<option>Умеренная</option>
					<option>Хаотичная</option>
					<option>Спайковая</option>
					<option>Волнообразная</option>
					<option>Нагрев</option>
					<option>Ночной режим</option>
				</select>
			</label>
			<button onclick="start()">🚀 Запустить</button>
		</div>

		<div class="card metrics" id="metrics" style="display:none;">
			<h3>📊 Метрики:</h3>
			<p>Всего запросов: <span id="total">0</span></p>
			<p>Успешных: <span id="success">0</span></p>
			<p>Ошибок: <span id="fail">0</span></p>
			<p>Успешность: <span id="successRate">0</span>%</p>
			<p>RPS: <span id="rpsShow">0</span></p>
		</div>
	</div>

<script>
function start() {
	const socket = new WebSocket("ws://" + location.host + "/ws");
	socket.onopen = () => {
		const payload = {
			url: document.getElementById("url").value,
			method: document.getElementById("method").value,
			payload: document.getElementById("payload").value,
			headers: document.getElementById("headers").value,
			rps: parseInt(document.getElementById("rps").value),
			duration: parseInt(document.getElementById("duration").value),
			load_profile: document.getElementById("loadProfile").value
		};
		socket.send(JSON.stringify(payload));
		document.getElementById("metrics").style.display = "block";
	};
	socket.onmessage = (msg) => {
		const data = JSON.parse(msg.data);
		document.getElementById("total").textContent = data.total;
		document.getElementById("success").textContent = data.success;
		document.getElementById("fail").textContent = data.fail;
		document.getElementById("successRate").textContent = data.success_rate.toFixed(1);
		document.getElementById("rpsShow").textContent = data.rps;
		if (data.done) {
			console.log("Тест завершён");
		}
	};
	socket.onerror = err => console.error("WebSocket Error", err);
}
</script>
</body>
</html>
`

package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
	"fortio.org/log"
)

var formTemplate = template.Must(template.New("form").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Fortio Load Test</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            margin: 20px;
            background-color: #f4f7fa;
            color: #333;
        }
        form {
            max-width: 500px;
            padding: 20px;
            border: 1px solid #ccc;
            border-radius: 10px;
            background-color: #fff;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
        }
        input, select, textarea {
            margin: 10px 0;
            padding: 10px;
            width: 100%;
            border: 1px solid #ddd;
            border-radius: 5px;
            box-sizing: border-box;
        }
        input[type="submit"] {
            background: #4CAF50;
            color: white;
            border: none;
            padding: 12px;
            border-radius: 5px;
            cursor: pointer;
            transition: background 0.3s;
        }
        input[type="submit"]:hover {
            background: #45a049;
        }
        .result-card {
            max-width: 600px;
            padding: 20px;
            margin-top: 20px;
            border-radius: 10px;
            background: linear-gradient(135deg, #ffffff, #e0f7fa);
            box-shadow: 0 6px 12px rgba(0, 0, 0, 0.1);
            animation: fadeIn 0.5s ease-in;
        }
        .result-card h2 {
            color: #2c3e50;
            font-size: 24px;
            margin-bottom: 15px;
            text-align: center;
        }
        .result-card p {
            margin: 10px 0;
            font-size: 16px;
            color: #34495e;
        }
        .metric-success {
            color: #27ae60;
            font-weight: bold;
        }
        .metric-error {
            color: #e74c3c;
            font-weight: bold;
        }
        .result-card a {
            display: inline-block;
            margin-top: 15px;
            padding: 10px 20px;
            background-color: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: background 0.3s;
        }
        .result-card a:hover {
            background-color: #2980b9;
        }
        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }
    </style>
</head>
<body>
    <h1>Fortio Load Test UI</h1>
    <form method="POST" action="/run">
        <label>URL: <input type="text" name="url" value="http://app:4000/api/tasks"></label><br>
        <label>Method: <select name="method">
            <option value="GET">GET</option>
            <option value="POST">POST</option>
        </select></label><br>
        <label>Payload (for POST): <textarea name="payload" placeholder='{"test": "load"}'></textarea></label><br>
        <label>Headers (Key:Value, one per line): <textarea name="headers" placeholder="Authorization: Bearer <token>"></textarea></label><br>
        <label>QPS: <input type="number" name="qps" value="10" min="1"></label><br>
        <label>Duration (s): <input type="number" name="duration" value="10" min="1"></label><br>
        <label>Threads: <input type="number" name="threads" value="2" min="1"></label><br>
        <input type="submit" value="Start Test">
    </form>
</body>
</html>
`))

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		formTemplate.Execute(w, nil)
	})

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		url := r.FormValue("url")
		if url == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		method := r.FormValue("method")
		if method != "GET" && method != "POST" {
			http.Error(w, "Invalid method: only GET or POST allowed", http.StatusBadRequest)
			return
		}

		qps, err := strconv.ParseFloat(r.FormValue("qps"), 64)
		if err != nil || qps <= 0 {
			http.Error(w, "Invalid QPS value", http.StatusBadRequest)
			return
		}

		duration, err := strconv.Atoi(r.FormValue("duration"))
		if err != nil || duration <= 0 {
			http.Error(w, "Invalid duration value", http.StatusBadRequest)
			return
		}

		threads, err := strconv.Atoi(r.FormValue("threads"))
		if err != nil || threads <= 0 {
			http.Error(w, "Invalid threads value", http.StatusBadRequest)
			return
		}

		// Parse headers
		headers := strings.Split(r.FormValue("headers"), "\n")
		for i, h := range headers {
			headers[i] = strings.TrimSpace(h)
		}

		// Parse payload for POST
		payload := strings.TrimSpace(r.FormValue("payload"))
		if method == "POST" && payload == "" {
			payload = `{"test": "load"}`
		}

		log.Infof("Running test: %s %s QPS=%v Duration=%v Threads=%v Headers=%v Payload=%s",
			method, url, qps, duration, threads, headers, payload)

		opts := fhttp.HTTPRunnerOptions{
			HTTPOptions: fhttp.HTTPOptions{
				URL: url,
			},
			RunnerOptions: periodic.RunnerOptions{
				QPS:         qps,
				Duration:    time.Duration(duration) * time.Second,
				NumThreads:  threads,
				Percentiles: []float64{50.0, 90.0},
			},
		}

		// Add headers
		for _, header := range headers {
			if header == "" {
				continue
			}
			if !strings.Contains(header, ":") {
				http.Error(w, fmt.Sprintf("Invalid header format: %s (use Key:Value)", header), http.StatusBadRequest)
				return
			}
			if err := opts.HTTPOptions.AddAndValidateExtraHeader(header); err != nil {
				http.Error(w, fmt.Sprintf("Invalid header: %v", err), http.StatusBadRequest)
				return
			}
		}

		// For POST, add payload
		if method == "POST" {
			opts.HTTPOptions.Payload = []byte(payload)
		}

		res, err := fhttp.RunHTTPTest(&opts)
		if err != nil {
			log.Errf("Test failed: %v", err)
			http.Error(w, fmt.Sprintf("Test failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Extract metrics from DurationHistogram and RetCodes
		totalRequests := res.DurationHistogram.Count
		durationSeconds := float64(duration)                // Use input duration in seconds
		avgResponseTime := res.DurationHistogram.Avg * 1000 // Convert to milliseconds
		p50 := res.DurationHistogram.Percentiles[0].Value * 1000
		p90 := res.DurationHistogram.Percentiles[1].Value * 1000
		var errors int64
		for code, count := range res.RetCodes {
			if code != http.StatusOK {
				errors += count
			}
		}

		// Output results with styled HTML
		fmt.Fprintf(w, `<div class="result-card">
            <h2>Test Result</h2>
            <p>Total Requests: <span class="metric-success">%d</span></p>
            <p>Duration: <span class="metric-success">%.2f seconds</span></p>
            <p>Average Response Time: <span class="metric-success">%.2f ms</span></p>
            <p>P50 Latency: <span class="metric-success">%.2f ms</span></p>
            <p>P90 Latency: <span class="metric-success">%.2f ms</span></p>
            <p>Errors: <span class="metric-error">%d</span></p>
            <a href="/">Run Another Test</a>
        </div>`,
			totalRequests, durationSeconds, avgResponseTime, p50, p90, errors)
	})

	// Start server
	srv := &http.Server{Addr: ":8081"}
	go func() {
		log.Infof("Listening on http://localhost:8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	log.Infof("Shutting down server...")
	srv.Shutdown(context.Background())
}

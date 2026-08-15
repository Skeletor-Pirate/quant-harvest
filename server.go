// Copyright 2026 Skeletor-Pirate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	store   *TaskStore
	metrics *Metrics
	config  Config
}

func NewServer(store *TaskStore, metrics *Metrics, config Config) *Server {
	return &Server{store: store, metrics: metrics, config: config}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleConsole)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	return withHeaders(mux)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasks, err := s.store.Pending(r.Context(), 100)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read task queue"})
			return
		}
		pending, running, completed, failed, err := s.store.Counts(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read task counts"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks, "pending": pending, "running": running, "completed": completed, "failed": failed})
	case http.MethodPost:
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "content type must be application/json"})
			return
		}
		task, err := readTask(r, s.config.MaxBodyBytes)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		stored, duplicate, err := s.store.Enqueue(r.Context(), task, r.Header.Get("Idempotency-Key"))
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "task already exists or could not be queued"})
			return
		}
		if duplicate {
			writeJSON(w, http.StatusOK, stored)
			return
		}
		writeJSON(w, http.StatusAccepted, stored)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	pending, running, completed, failed, err := s.store.Counts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read status"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"service": "q-harvest", "state": "operational", "carbon_intensity": s.metrics.CarbonIntensity(), "carbon_threshold": s.config.CarbonThreshold, "queue_depth": pending, "running": running, "completed": completed, "failed": failed, "pqc": "integration boundary only", "updated_at": time.Now().UTC()})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "store unavailable"})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	pending, running, completed, failed, err := s.store.Counts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "metrics unavailable"})
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintf(w, "# HELP qharvest_task_queue_size Tasks waiting for execution.\n# TYPE qharvest_task_queue_size gauge\nqharvest_task_queue_size %d\n", pending)
	fmt.Fprintf(w, "qharvest_tasks_running %d\nqharvest_tasks_completed_total %d\nqharvest_tasks_failed %d\nqharvest_carbon_intensity %d\n", running, completed, failed, s.metrics.CarbonIntensity())
}

func (s *Server) handleConsole(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(consoleHTML))
}

func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; script-src 'self' 'unsafe-inline'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

const consoleHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>⚡ Q-HARVEST / Control Plane</title>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=Outfit:wght@600;700;800&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #070709;
      --card: rgba(18, 18, 24, 0.7);
      --border: rgba(255, 255, 255, 0.08);
      --text: #F3F4F6;
      --text-soft: #9CA3AF;
      --violet: #7C3AED;
      --cyan: #06B6D4;
      --green: #10B981;
      --red: #EF4444;
      --amber: #F59E0B;
    }
    
    * { box-sizing: border-box; margin: 0; padding: 0; }
    
    body {
      font-family: 'Inter', system-ui, sans-serif;
      background: var(--bg);
      color: var(--text);
      max-width: 1100px;
      margin: 0 auto;
      padding: 40px 20px;
      line-height: 1.5;
      background-image: 
        radial-gradient(circle at 10% 20%, rgba(124, 58, 237, 0.06) 0%, transparent 40%),
        radial-gradient(circle at 90% 80%, rgba(6, 182, 212, 0.06) 0%, transparent 40%);
      background-attachment: fixed;
    }
    
    header {
      margin-bottom: 32px;
    }
    
    .logo-area {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 12px;
    }
    
    .logo {
      width: 40px;
      height: 40px;
      border-radius: 12px;
      background: linear-gradient(135deg, var(--violet), var(--cyan));
      display: grid;
      place-items: center;
      font-family: 'Outfit', sans-serif;
      font-weight: 800;
      font-size: 20px;
      color: white;
      box-shadow: 0 0 20px rgba(6, 182, 212, 0.2);
    }
    
    .badge {
      font-size: 10px;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      color: var(--cyan);
      background: rgba(6, 182, 212, 0.1);
      border: 1px solid rgba(6, 182, 212, 0.2);
      padding: 4px 10px;
      border-radius: 20px;
    }
    
    h1 {
      font-family: 'Outfit', sans-serif;
      font-size: clamp(32px, 5vw, 48px);
      font-weight: 800;
      letter-spacing: -0.04em;
      line-height: 1.1;
      margin-bottom: 8px;
      background: linear-gradient(120deg, #FFFFFF 30%, var(--text-soft) 100%);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }
    
    .desc {
      color: var(--text-soft);
      max-width: 680px;
      font-size: 15px;
      margin-bottom: 24px;
    }
    
    .grid {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;
      margin-bottom: 24px;
    }
    
    .card {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 18px;
      padding: 20px;
      transition: all 0.3s ease;
      backdrop-filter: blur(10px);
    }
    
    .card:hover {
      border-color: rgba(6, 182, 212, 0.3);
      box-shadow: 0 8px 30px rgba(0,0,0,0.4);
    }
    
    .card-label {
      font-size: 10px;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      color: var(--text-soft);
      font-weight: 600;
    }
    
    .card-value {
      font-family: 'Outfit', sans-serif;
      font-size: 28px;
      font-weight: 700;
      margin-top: 14px;
      display: flex;
      align-items: baseline;
      gap: 4px;
    }
    
    .layout-main {
      display: grid;
      grid-template-columns: 1fr 340px;
      gap: 20px;
      margin-top: 24px;
    }
    
    .panel {
      background: var(--card);
      border: 1px solid var(--border);
      border-radius: 20px;
      padding: 24px;
      backdrop-filter: blur(10px);
    }
    
    .panel-title {
      font-family: 'Outfit', sans-serif;
      font-size: 18px;
      font-weight: 700;
      margin-bottom: 16px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    
    /* Live Chart styling */
    .chart-container {
      height: 140px;
      margin-bottom: 24px;
      position: relative;
    }
    
    .chart-svg {
      width: 100%;
      height: 100%;
      overflow: visible;
    }
    
    .threshold-line {
      stroke: var(--red);
      stroke-dasharray: 4 4;
      stroke-width: 1.5;
    }
    
    .chart-line {
      fill: none;
      stroke: var(--cyan);
      stroke-width: 3;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
    
    .chart-area {
      fill: url(#gradient-fill);
      opacity: 0.15;
    }
    
    /* Logger Panel */
    .logger-panel {
      display: flex;
      flex-direction: column;
      height: 100%;
    }
    
    .console-win {
      flex: 1;
      background: #09090D;
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 14px;
      font-family: 'Courier New', Courier, monospace;
      font-size: 12px;
      color: var(--text-soft);
      overflow-y: auto;
      max-height: 250px;
      min-height: 180px;
    }
    
    .log-line {
      margin-bottom: 6px;
      display: flex;
      gap: 8px;
    }
    
    .log-time { color: var(--violet); }
    .log-text-info { color: var(--text); }
    .log-text-warning { color: var(--amber); }
    .log-text-success { color: var(--green); }
    .log-text-system { color: var(--cyan); }
    
    /* Task Queue Row styling */
    .queue-container {
      margin-top: 16px;
      border: 1px solid var(--border);
      border-radius: 14px;
      overflow: hidden;
      background: #09090D;
    }
    
    .queue-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 14px 18px;
      border-bottom: 1px solid var(--border);
      transition: background 0.2s;
    }
    
    .queue-row:last-child { border-bottom: none; }
    .queue-row:hover { background: rgba(255,255,255,0.02); }
    
    .status-badge {
      font-size: 11px;
      font-weight: 600;
      padding: 3px 8px;
      border-radius: 6px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    
    .status-queued { background: rgba(245, 158, 11, 0.15); color: var(--amber); }
    .status-running { background: rgba(6, 182, 212, 0.15); color: var(--cyan); }
    .status-completed { background: rgba(16, 185, 129, 0.15); color: var(--green); }
    
    .empty-state {
      padding: 40px;
      text-align: center;
      color: var(--text-soft);
      font-size: 14px;
    }
    
    /* Flowchart / Explanation Cards */
    .flow-steps {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;
      margin-top: 32px;
      border-top: 1px solid var(--border);
      padding-top: 24px;
    }
    
    .step-card {
      background: rgba(255,255,255,0.01);
      border: 1px dashed var(--border);
      border-radius: 14px;
      padding: 16px;
    }
    
    .step-num {
      width: 22px;
      height: 22px;
      border-radius: 50%;
      background: var(--violet);
      display: grid;
      place-items: center;
      font-size: 11px;
      font-weight: 700;
      margin-bottom: 10px;
      color: white;
    }
    
    .step-title {
      font-size: 13px;
      font-weight: 600;
      margin-bottom: 4px;
      color: white;
    }
    
    .step-desc {
      font-size: 12px;
      color: var(--text-soft);
    }
    
    /* Form */
    form {
      display: flex;
      gap: 12px;
      margin-top: 16px;
    }
    
    input {
      flex: 1;
      background: #09090D;
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 14px;
      color: white;
      font-family: inherit;
      font-size: 14px;
      transition: all 0.3s;
    }
    
    input:focus {
      outline: none;
      border-color: var(--cyan);
      box-shadow: 0 0 10px rgba(6, 182, 212, 0.15);
    }
    
    button {
      background: linear-gradient(135deg, var(--violet), var(--cyan));
      border: none;
      color: white;
      font-weight: 700;
      font-size: 14px;
      padding: 0 24px;
      border-radius: 12px;
      cursor: pointer;
      transition: opacity 0.2s;
    }
    
    button:hover { opacity: 0.9; }
    
    @media(max-width: 900px) {
      .layout-main { grid-template-columns: 1fr; }
      .grid { grid-template-columns: repeat(2, 1fr); }
      .flow-steps { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <header>
    <div class="logo-area">
      <div class="logo">Q</div>
      <div class="badge">Operational Control Plane</div>
    </div>
    <h1>Compute when the grid is clean.</h1>
    <p class="desc">A carbon-aware workload queue that defers computationally heavy post-quantum cryptographic (PQC) operations when carbon intensity exceeds local thresholds.</p>
  </header>

  <section class="grid">
    <div class="card">
      <div class="card-label">Grid Carbon Intensity</div>
      <div class="card-value" style="color: var(--cyan)" id="carbon">—</div>
    </div>
    <div class="card">
      <div class="card-label">Execution Cutoff</div>
      <div class="card-value" style="color: var(--red)" id="threshold">—</div>
    </div>
    <div class="card">
      <div class="card-label">Queue Depth</div>
      <div class="card-value" style="color: var(--amber)" id="queue">—</div>
    </div>
    <div class="card">
      <div class="card-label">Completed Jobs</div>
      <div class="card-value" style="color: var(--green)" id="completed">—</div>
    </div>
  </section>

  <div class="layout-main">
    <div class="panel">
      <div class="panel-title">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
        Grid Carbon Intensity History
      </div>
      <div class="chart-container">
        <svg class="chart-svg" id="chart" viewBox="0 0 500 120" preserveAspectRatio="none">
          <defs>
            <linearGradient id="gradient-fill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="var(--cyan)" />
              <stop offset="100%" stop-color="transparent" />
            </linearGradient>
          </defs>
          <!-- Threshold line -->
          <line id="thresh-line" class="threshold-line" x1="0" y1="60" x2="500" y2="60" />
          <!-- Area under curve -->
          <path id="chart-area-path" class="chart-area" d="" />
          <!-- Carbon line -->
          <path id="chart-line-path" class="chart-line" d="" />
        </svg>
      </div>

      <div class="panel-title">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1zM4 22v-7"/></svg>
        Workload Dispatch Queue
      </div>
      
      <div class="queue-container" id="tasks">
        <div class="empty-state">Loading workloads...</div>
      </div>

      <form onsubmit="enqueue(event)">
        <input id="name" maxlength="120" placeholder="Workload name (e.g., pqc-key-rotation-job)" required>
        <button>Queue Job</button>
      </form>
    </div>

    <div class="panel logger-panel">
      <div class="panel-title">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 17L10 11L4 5M12 19H20"/></svg>
        System Logs
      </div>
      <div class="console-win" id="console">
        <div class="log-line">
          <span class="log-time">[System]</span>
          <span class="log-text-system">Initializing Quant Harvest operator console...</span>
        </div>
      </div>
    </div>
  </div>

  <section class="flow-steps">
    <div class="step-card">
      <div class="step-num">1</div>
      <div class="step-title">Queue PQC Job</div>
      <div class="step-desc">Enter a workload to schedule computationally heavy signing or key exchanges.</div>
    </div>
    <div class="step-card">
      <div class="step-num">2</div>
      <div class="step-title">Real-Time Grid Check</div>
      <div class="step-desc">Scheduler continuously queries local carbon intensity metrics of the energy grid.</div>
    </div>
    <div class="step-card">
      <div class="step-num">3</div>
      <div class="step-title">Carbon Threshold Guard</div>
      <div class="step-desc">If the grid is dirty (exceeds threshold), execution is dynamically deferred.</div>
    </div>
    <div class="step-card">
      <div class="step-num">4</div>
      <div class="step-title">Execute on Clean Energy</div>
      <div class="step-desc">When carbon drops, scheduler claims tasks and processes them safely in SQLite WAL.</div>
    </div>
  </section>

  <script>
    let carbonHistory = [];
    const maxHistoryPoints = 25;
    let lastCarbonVal = null;
    let lastQueueVal = null;

    function addLog(text, type = 'info') {
      const win = document.getElementById('console');
      const timeStr = new Date().toLocaleTimeString();
      const line = document.createElement('div');
      line.className = 'log-line';
      line.innerHTML = '<span class="log-time">['+timeStr+']</span> <span class="log-text-'+type+'">'+text+'</span>';
      win.appendChild(line);
      win.scrollTop = win.scrollHeight;
    }

    function updateChart(history, threshold) {
      if (history.length < 2) return;
      const width = 500;
      const height = 120;
      
      // Calculate min/max for scale (or fit to standard grid intensity 100-300)
      const minVal = 80;
      const maxVal = 300;
      const range = maxVal - minVal;

      // Map threshold Y coordinate
      const threshY = height - ((threshold - minVal) / range) * height;
      document.getElementById('thresh-line').setAttribute('y1', threshY);
      document.getElementById('thresh-line').setAttribute('y2', threshY);

      // Construct points
      let points = [];
      const step = width / (maxHistoryPoints - 1);
      
      history.forEach((pt, i) => {
        // Center the line mapping if history isn't full yet
        const offset = maxHistoryPoints - history.length;
        const x = (i + offset) * step;
        const y = height - ((pt.value - minVal) / range) * height;
        points.push({x, y});
      });

      // Line Path
      let pathD = 'M ' + points[0].x + ' ' + points[0].y;
      for (let i = 1; i < points.length; i++) {
        pathD += ' L ' + points[i].x + ' ' + points[i].y;
      }
      document.getElementById('chart-line-path').setAttribute('d', pathD);

      // Area Path
      let areaD = pathD + ' L ' + points[points.length - 1].x + ' ' + height + ' L ' + points[0].x + ' ' + height + ' Z';
      document.getElementById('chart-area-path').setAttribute('d', areaD);
    }

    async function refresh() {
      try {
        const [s, t] = await Promise.all([
          fetch('/api/status').then(r => r.json()),
          fetch('/api/tasks').then(r => r.json())
        ]);

        carbon.textContent = s.carbon_intensity + ' g';
        threshold.textContent = s.carbon_threshold + ' g';
        queue.textContent = s.queue_depth;
        completed.textContent = s.completed;

        // Log grid checks
        if (s.carbon_intensity !== lastCarbonVal) {
          const stateStr = s.carbon_intensity > s.carbon_threshold ? 'DIRTY (Deferred)' : 'CLEAN (Ready)';
          const logType = s.carbon_intensity > s.carbon_threshold ? 'warning' : 'success';
          addLog('Grid check: ' + s.carbon_intensity + ' gCO2e/kWh - Status: ' + stateStr, logType);
          
          carbonHistory.push({ value: s.carbon_intensity });
          if (carbonHistory.length > maxHistoryPoints) carbonHistory.shift();
          updateChart(carbonHistory, s.carbon_threshold);
          lastCarbonVal = s.carbon_intensity;
        }

        if (s.queue_depth !== lastQueueVal && lastQueueVal !== null) {
          if (s.queue_depth > lastQueueVal) {
            addLog('New task received in scheduler queue. Queue depth: ' + s.queue_depth, 'system');
          } else {
            addLog('Workload processed and claimed. Queue depth: ' + s.queue_depth, 'success');
          }
          lastQueueVal = s.queue_depth;
        } else if (lastQueueVal === null) {
          lastQueueVal = s.queue_depth;
        }

        // Render Queue
        const tasksContainer = document.getElementById('tasks');
        if (t.tasks && t.tasks.length) {
          tasksContainer.innerHTML = t.tasks.map(x => {
            const statusClass = x.status === 'queued' ? 'status-queued' : (x.status === 'running' ? 'status-running' : 'status-completed');
            return '<div class="queue-row"><strong>' + esc(x.name) + '</strong><span class="status-badge ' + statusClass + '">' + x.status + '</span></div>';
          }).join('');
        } else {
          tasksContainer.innerHTML = '<div class="empty-state">No workloads waiting in the queue.</div>';
        }

      } catch (err) {
        addLog('Error reading control plane status: ' + err.message, 'warning');
      }
    }

    function esc(x) {
      return x.replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
    }

    async function enqueue(e) {
      e.preventDefault();
      const input = document.getElementById('name');
      const taskName = input.value.trim();
      if (!taskName) return;

      try {
        addLog('Submitting workload "' + taskName + '" with idempotency token...', 'system');
        const res = await fetch('/api/tasks', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Idempotency-Key': crypto.randomUUID()
          },
          body: JSON.stringify({ name: taskName })
        });
        
        if (res.ok) {
          addLog('Workload "' + taskName + '" enqueued successfully.', 'info');
        } else {
          const data = await res.json();
          addLog('Failed to queue: ' + (data.error || 'Server error'), 'warning');
        }
        input.value = '';
        refresh();
      } catch (err) {
        addLog('Network error enqueuing task: ' + err.message, 'warning');
      }
    }

    // Seed initial historical chart data
    for (let i = 0; i < maxHistoryPoints; i++) {
      carbonHistory.push({ value: 150 + Math.floor(Math.random() * 80) });
    }

    refresh();
    setInterval(refresh, 5000);
  </script>
</body>
</html>`

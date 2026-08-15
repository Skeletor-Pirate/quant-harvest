# ⚡ Quant Harvest

[![Go Report Card](https://goreportcard.com/badge/github.com/Skeletor-Pirate/quant-harvest)](https://goreportcard.com/report/github.com/Skeletor-Pirate/quant-harvest)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

**Quant Harvest** is a cloud-native, carbon-aware control-plane prototype designed for scheduling and executing computationally heavy post-quantum cryptography (PQC) workloads when the energy grid is cleanest.

It implements a bounded, deferred-execution queue that checks real-time carbon intensity against a user-defined threshold, dynamically pausing and resuming queue processing.

---

## 🚀 Key Features

*   **Carbon-Aware Scheduling**: Defers queue processing automatically when the grid's carbon intensity (gCO2e/kWh) exceeds the configured threshold.
*   **Idempotency & Reliability**: SQLite-backed task store running in WAL (Write-Ahead Logging) mode with robust transaction handling and automatic retries.
*   **Built-in Operator Console**: A responsive, self-contained dashboard served at the root path (`/`) to monitor grid status and manage workloads in real-time.
*   **Observability-Ready**: Native Prometheus `/metrics` endpoint, along with standard `/healthz` (liveness) and `/readyz` (readiness) check endpoints.
*   **Explicit Cryptographic Boundary**: Includes a structured boundary (`pqc.go`) ready for ML-KEM/PQC library integration.

---

## 🛠️ Tech Stack & Architecture

*   **Language**: Go 1.20+ (Standard library focused)
*   **Database**: Embedded SQLite (using the modern pure-Go `modernc.org/sqlite` driver)
*   **Observability**: Prometheus instrumentation + HTTP middleware hardening (CSP, nosniff, cache-control headers).

---

## 📋 API Reference

| Endpoint | Method | Description | Headers / Payload |
| :--- | :--- | :--- | :--- |
| `/` | `GET` | Serves the Operator Console dashboard. | — |
| `/healthz` | `GET` | Liveness check. Returns `200 OK`. | — |
| `/readyz` | `GET` | Readiness check. Verifies SQLite connection. | — |
| `/metrics` | `GET` | Exposes Prometheus metrics (`qharvest_task_queue_size`, etc.). | — |
| `/api/status` | `GET` | Fetches scheduler status and current grid intensity. | — |
| `/api/tasks` | `GET` | Lists all pending/queued tasks. | — |
| `/api/tasks` | `POST` | Enqueues a new PQC workload. | JSON body: `{"name": "...", "payload": "..."}`<br>Supports optional `Idempotency-Key` header. |

### Enqueue Workload Example

**Request:**
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-uuid-12345" \
  -d '{"name": "pqc-batch-signing", "payload": "ml-kem-768-request-block"}'
```

**Response:**
```json
{
  "id": "qh-1786772642674",
  "name": "pqc-batch-signing",
  "payload": "ml-kem-768-request-block",
  "status": "queued",
  "created_at": "2026-08-15T06:30:00Z"
}
```

---

## ⚙️ Configuration Options

Configuration is managed via environment variables or command-line flags (flags take precedence).

| Environment Variable | CLI Flag | Default | Description |
| :--- | :--- | :--- | :--- |
| `QHARVEST_ADDR` | `-addr` | `:8080` | Bind address for the HTTP service. |
| `QHARVEST_CARBON_THRESHOLD` | `-carbon-threshold` | `200` | Maximum carbon intensity (gCO2e/kWh) before execution is paused. |
| `QHARVEST_TICK` | `-tick` | `15s` | Scheduler tick interval for grid polling. |
| `QHARVEST_MAX_CONCURRENCY` | `-max-concurrency` | `2` | Maximum concurrent workers running tasks. |
| `QHARVEST_STORE` | `-store` | `data/qharvest.db` | Location of the SQLite storage file. |

---

## 🏃 Getting Started

### Local Setup

Ensure you have [Go](https://go.dev/doc/install) installed.

1. **Run the project locally**:
   ```bash
   go run .
   ```
2. Open [http://localhost:8080](http://localhost:8080) in your browser to view the operator console.
3. **Run unit tests**:
   ```bash
   go test -v ./...
   ```

### Docker Container

Build and launch the lightweight multi-stage Docker container locally:

```bash
# Build the image
docker build -t qharvest:latest .

# Run the container
docker run --rm -p 8080:8080 -v qharvest-data:/app/data qharvest:latest
```

### Kubernetes Deployment

Kubernetes manifests are located in the `deploy/` folder:

```bash
kubectl apply -f deploy/config.yaml
kubectl apply -f deploy/kubernetes.yaml
```

---

## 🔒 Security & Post-Quantum Boundary

> [!WARNING]
> The current cryptographic boundary defined in `pqc.go` is an envelope representation for scheduling testing. Vetted implementations of ML-KEM/FIPS 203 standards must be imported into `pqc.go` before using this system in production cryptographic environments.

---

## 📄 License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.


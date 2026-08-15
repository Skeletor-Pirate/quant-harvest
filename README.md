# Quant Harvest

Quant Harvest is a small, cloud-native control-plane prototype for carbon-aware execution of post-quantum cryptography workloads.

## Project layout

- `*.go` — application source and tests
- `deploy/` — Kubernetes and runtime configuration
- `data/` — local database volume for development

## Run locally

```powershell
go run .
```

Open `http://localhost:8080` for the operator console.

## API

- `GET /healthz` — liveness
- `GET /readyz` — readiness
- `GET /api/status` — scheduler and queue state
- `GET /api/tasks` — queued workloads
- `POST /api/tasks` — enqueue `{"name":"batch-signing"}`
- `GET /metrics` — Prometheus-compatible metrics

## Configuration

- `QHARVEST_ADDR` or `-addr` — bind address, default `:8080`
- `QHARVEST_CARBON_THRESHOLD` or `-carbon-threshold` — execution threshold
- `QHARVEST_TICK` or `-tick` — scheduler interval, default `15s`

The PQC boundary is intentionally explicit: `pqc.go` contains a non-cryptographic envelope identifier only. Integrate a vetted ML-KEM implementation before processing secrets or representing this prototype as production cryptography.

## Container

```powershell
docker build -t qharvest:dev .
docker run --rm -p 8080:8080 qharvest:dev
```

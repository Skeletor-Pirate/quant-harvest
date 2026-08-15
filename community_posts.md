# Quant Harvest Community Promotion Kit

This file contains copy-pasteable titles and descriptions to promote **Quant Harvest** on Hacker News and Reddit.

---

## 📰 Hacker News (Show HN)

**Suggested Title:**
`Show HN: Quant Harvest – A carbon-aware scheduler for post-quantum workloads`

**Suggested Text:**
```markdown
Hi HN,

I built Quant Harvest, a Go-based control-plane prototype for carbon-aware execution of post-quantum cryptography (PQC) workloads. 

As we migrate to post-quantum standards (like ML-KEM), our cryptographic operations are becoming significantly more CPU-intensive and energy-expensive. Quant Harvest schedules these delayable tasks (like bulk document signing or backups) to run only when the local energy grid is cleanest.

Features:
- Scheduler that polls grid carbon intensity and defers queue execution when carbon is high.
- Pure Go implementation using SQLite in WAL mode for task storage.
- Idempotency key tracking to avoid double-signing/double-execution.
- Exposes Prometheus metrics and standard health endpoints.
- Lightweight built-in HTML operator dashboard.

It's fully open source under the Apache 2.0 license. I'd love to hear your thoughts on the concept and architecture!

GitHub: https://github.com/Skeletor-Pirate/quant-harvest
```

---

## 👽 Reddit

### Subreddit: `r/golang`
**Title:**
`Show Go: Quant Harvest – SQLite-backed carbon-aware queue for PQC workloads`

**Post Text:**
```markdown
Hi all,

I wanted to share my new Go project: **Quant Harvest**. It is a cloud-native control plane prototype that schedules CPU-heavy post-quantum cryptography (PQC) workloads to run only when the local grid's carbon intensity is below a set threshold.

I kept the stack minimal and relied strictly on Go standard libraries + SQLite for storage:
- SQLite runs in WAL mode with active busy timeouts to prevent locking under high-concurrency claims.
- The scheduler runs on a standard Go ticker, claiming and locking tasks concurrently.
- Enforce idempotency on enqueues via custom SQLite unique constraints.
- Native `/metrics` endpoint serves Prometheus-compatible counters.

Code is available under Apache 2.0. Any feedback on the scheduler architecture or SQL schemas would be greatly appreciated!

GitHub: https://github.com/Skeletor-Pirate/quant-harvest
```

### Subreddit: `r/devops` or `r/GreenComputing`
**Title:**
`Quant Harvest: Carbon-aware deferred-execution queue for PQC workloads`

**Post Text:**
```markdown
Hey everyone,

With the ongoing migration to Post-Quantum Cryptography (PQC), compute overhead for signing/encrypting bulk operations is about to increase dramatically. 

I've put together a Go-based prototype called **Quant Harvest** to address this. It implements a carbon-aware queue that monitors the carbon intensity of the local grid (gCO2e/kWh) and automatically defers or executes scheduled workloads (like batch PQC signing) depending on how clean the grid is.

It's containerized, comes with Kubernetes manifests, and exposes standard Prometheus metrics + health/readiness endpoints for clustering.

Check it out here: https://github.com/Skeletor-Pirate/quant-harvest

Feedback on carbon-aware scheduling heuristics is welcome!
```

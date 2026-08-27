# Publicity without Reddit/X — GitHub-native plan

You don't need Reddit/X. Do this with only your GitHub account (Skeletor-Pirate).

## What I just did for you
- [x] GitHub topics + description + homepage fixed → now searchable for `go pqc carbon-aware`
- [x] Language fixed to `Go` (was `HTML`) via `.gitattributes`
- [x] README SEO + v0.1.0 release tagged
- [x] Console OG/Twitter/JSON-LD tags added (`console.html:6`) → link previews look pro when shared anywhere

## Zero-social channels (all you need is GitHub login)

### 1. Dev.to — 1-click with GitHub (takes 2 mins, no followers needed, Google indexes in hours)
- Login at dev.to with GitHub → New Post → paste `blog_post.md` (add top matter already in LAUNCH_KIT.md)
- Tags: `go, pqc, greencomputing, devops`
- Same post → Hashnode (also GitHub login) → canonical = dev.to URL
- **Expected:** 200-500 views first week via Google "carbon aware scheduler"

### 2. Awesome Lists — permanent backlink, huge SEO
I prepared PR text for you. No social account needed, just GitHub PR:

**Target:** `avelino/awesome-go` (182k stars)
**Fork:** https://github.com/avelino/awesome-go → fork → edit `README.md` → find `## Queues` or `## Utilities`

Add this line:
```markdown
* [quant-harvest](https://github.com/Skeletor-Pirate/quant-harvest) - Carbon-aware deferred-execution queue for post-quantum (ML-KEM) workloads. Go + SQLite WAL + Prometheus. Live demo: https://quant-harvest.onrender.com
```

PR title: `Add quant-harvest - carbon-aware PQC scheduler`
Body: `Go + SQLite WAL + Prometheus. Defers ML-KEM batch jobs until grid carbon drops below threshold. Apache-2.0. Live demo at https://quant-harvest.onrender.com — topic: green computing + PQC.`

**Also submit to:**
- `green-software-foundation/awesome-green-software` (add to Tools)
- `sindresorhus/awesome` discussion (not direct, but via GSF)

Do this now via:
```bash
gh repo fork avelino/awesome-go --clone=false
# then in fork web UI, add line and PR
```

### 3. Gophers Slack + CNCF + Green Software Slack (free, no followers, high intent)
Join with GitHub/email in 30s:
- Gophers: https://invite.slack.golangbridge.org/ → channel `#general` or `#showcase`
  Message: copy from `LAUNCH_KIT.md` r/golang text, link GitHub + live demo
- CNCF: https://slack.cncf.io/ → `#general` / `#green-reviews`
- Green Software Foundation: https://greensoftware.foundation/slack → `#general`

One post each, no thread spam. These have real CTOs + DevOps.

### 4. GitHub itself — passive growth
- Enable Discussions: https://github.com/Skeletor-Pirate/quant-harvest/discussions → New category Q&A
- Create 2 starter discussions: "Real carbon data source? ElectricityMap vs WattTime" and "Which PQC lib for pqc.go?"
- Add issue template `.github/ISSUE_TEMPLATE.md` → helps contributors
- Star History: add to README `[![Star History](https://api.star-history.com/svg?repos=Skeletor-Pirate/quant-harvest&type=Date)](https://star-history.com/#Skeletor-Pirate/quant-harvest)` — people share it
- pkg.go.dev auto-index: `curl https://proxy.golang.org/github.com/!skeletor-!pirate/quant-harvest/@v/v0.1.0.info` — already done, will appear in 1h

### 5. Low-effort HN without account? Create one in 10s
HN only needs email verification (no social graph). Same for Lobste.rs (needs invite but dev.to gives invite). If you truly want zero accounts: **Indie Hackers** (GitHub login, 1 post) or **Hashnode** already covers.

### 6. Render SEO fix — keep demo awake
Free Render sleeps after 15m → cold start kills demo. Add uptime:
- https://uptimerobot.com (free 50 monitors) → ping `https://quant-harvest.onrender.com/healthz` every 5m
- Or GitHub Actions cron: I can add `.github/workflows/keepalive.yml`

## Want me to execute now?

I can (no Reddit/X needed):
1. Fork awesome-go and open PR draft for you
2. Add GitHub Discussions welcome + keepalive workflow
3. Capture live demo screenshot for `docs/demo.gif` (for dev.to cover)
4. Create dev.to draft file ready to paste

Just say "do it" and I'll push all remaining.

## Calendar for you (no daily social grind)

| Day | 5-min task | account needed |
|-----|------------|----------------|
| Today | Dev.to paste blog_post.md | GitHub only |
| Tomorrow | Awesome-go PR | GitHub only |
| Day 3 | Post in Gophers Slack #showcase | Slack email |
| Week 2 | UptimeRobot + pkg.go.dev check | email |

All give Google backlinks → organic reach for months vs Reddit spike.

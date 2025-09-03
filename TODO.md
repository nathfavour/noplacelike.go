# NoPlaceLike 2.0 — TODO and Roadmap

A living plan tracking what’s complete and what remains to make NoPlaceLike a robust, production-grade distributed platform.

Legend
- Priority: P0 (Critical), P1 (High), P2 (Medium), P3 (Low)
- Status: [x] Done, [ ] Pending, [~] In progress


## 1) What’s done

Core Platform
- [x] v2.0 Platform scaffolding with service lifecycle management
- [x] In-memory implementations of core managers to provide working APIs:
  - [x] EventBus (publish, subscribe, wildcard subscribe)
  - [x] MetricsCollector (counters, gauges, histograms, timers; export json/text)
  - [x] SecurityManager (accept any non-empty token; allow-all authorization)
  - [x] NetworkManager (in-memory peers; discover/connect/list; messaging stubs)
  - [x] ResourceManager (in-memory registry; list/get/create/delete/stream stub)
  - [x] ServiceManager (register/start/stop/health)
- [x] HTTP service started by platform (not manually) with:
  - [x] Platform routes: GET /health, GET /info
  - [x] API routes under /api:
    - [x] /api/platform: health, info, metrics
    - [x] /api/plugins: list/get/start/stop/health
    - [x] /api/network: list peers, discover peers
    - [x] /api/resources: list/get/create/delete/stream
    - [x] /api/events: SSE stream and publish
  - [x] Plugin routes auto-registered under /plugins/{plugin}
- [x] Preload built-in plugins before platform Start; auto-start during Start:
  - [x] File Manager plugin (list/upload/download/delete)
  - [x] Clipboard plugin (get/set/history/clear)
  - [x] System Info plugin (info/health)
- [x] Sample in-memory resource registered on boot to validate resource APIs
- [x] Graceful shutdown stopping the platform (and all services/plugins)
- [x] Configuration file auto-create (~/.noplacelike.json) with sensible defaults
- [x] Structured logging via Zap
- [x] Network access QR code and URL display on boot (server package)

Legacy Surface (present, not default)
- [x] Legacy API & UI code path exists (clipboard, files, filesystem, shell, media, ollama proxy)
- [x] Can be enabled or integrated later; currently not started by v2 main


## 2) Gaps to production (by area)

P0 — Critical
- Security
  - [ ] Implement real token management (JWT) and validation; rotation; clock skew; revocation
  - [ ] RBAC with permissions/roles and route-level enforcement
  - [ ] Optional OAuth2/OIDC provider integration
  - [ ] TLS enablement (certs/keys, auto-redirect HTTP→HTTPS, HSTS tuning)
  - [ ] CSRF protection for browser flows; session handling for UI
  - [ ] Secure headers hardening and CORS allowlist model
- Networking
  - [ ] Actual peer discovery (mDNS/Bonjour) and/or gossip; configurable
  - [ ] Peer connection management and keepalive (QUIC/TCP/WebSocket)
  - [ ] NAT traversal strategy (UPnP/PCP/ICE/STUN/TURN) if required
  - [ ] Peer health checks, backoff, and failure detection
- Metrics/Observability
  - [ ] Prometheus exporter with consistent naming/labels
  - [ ] pprof endpoints (protected) and performance profiling guides
  - [ ] OpenTelemetry traces (HTTP handlers, plugin handlers, network IO)
  - [ ] Structured logs with request IDs, correlation IDs, peer IDs
  - [ ] Metrics for API latencies, error rates, plugin health, network events
- Resources
  - [ ] Streaming: chunked transfer, resumable downloads/uploads, backpressure
  - [ ] File resources: Range requests (partial content), MIME detection, ETag/If-None-Match
  - [ ] Persistence: storage drivers (filesystem, object store), GC/retention
  - [ ] Quotas, limits, and policies
  - [ ] Strong error model and status codes
- EventBus
  - [ ] Async dispatch/buffering; fan-out; backpressure; topics; priorities
  - [ ] Durable subscribers and replay (optional)
- HTTP Service
  - [ ] Unify duplicated HTTP implementations (internal/services vs internal/core). Decide on one and remove the other.
  - [ ] Middlewares: gzip compression, real rate limiting (token bucket/leaky-bucket)
  - [ ] Max request size with streaming (not only Content-Length pre-check)
  - [ ] OpenAPI/Swagger docs generation and UI
  - [ ] API versioning strategy across v1/v2 (deprecations, compatibility)
- Plugin System
  - [ ] Plugin discovery from directories (with signature checks)
  - [ ] Plugin sandboxing/isolation strategy (process or WASM) and resource caps
  - [ ] Plugin configuration schema/validation and hot-reload
  - [ ] Robust plugin lifecycle hooks and failure handling (circuit breaker)
- Testing & Reliability
  - [ ] Unit tests for managers, services, handlers (>=80% coverage target)
  - [ ] Integration tests for API flows; e2e test harness
  - [ ] Concurrency/race tests; fuzz testing on parsers and inputs
  - [ ] Load/stress benchmarks on critical paths (file transfer, events)
- CI/CD & Supply Chain
  - [ ] CI pipelines: lint (golangci-lint), test, coverage, build
  - [ ] Release automation (goreleaser), checksums, SBOM, signatures
  - [ ] Container builds (non-root, distroless), CVE scanning

P1 — High
- Legacy API/UI
  - [ ] Strategy: integrate legacy routes into v2 server or deprecate
  - [ ] If integrate: mount UI under /ui and legacy JSON under /api/v1; guard with feature flags
- Configuration
  - [ ] Standardize env/flags/config file precedence; support live reload
  - [ ] Secret management (env, file mounts, vault)
  - [ ] Validations and helpful error messages
- API UX
  - [ ] Pagination, filtering, and sorting conventions
  - [ ] Consistent error schema; error codes and docs
  - [ ] WebSockets for events and streams (in addition to SSE)
- Persistence
  - [ ] Local data dir structure; SQLite/BoltDB for metadata/state
  - [ ] Migrations and backups
- Deployment
  - [ ] Dockerfile hardening (small base, reproducible builds)
  - [ ] Helm chart, K8s manifests, Kustomize overlays
  - [ ] Readiness/liveness probes; pod security context
- Documentation
  - [ ] Architecture docs and diagrams
  - [ ] Security model and deployment hardening guide
  - [ ] Operator guide (backups, upgrades, scaling)
  - [ ] Contributor guide with local dev scripts and style guide

P2 — Medium
- UX & Clients
  - [ ] Simple web UI for admin/monitoring
  - [ ] CLI client & SDKs for common languages
  - [ ] Desktop tray app (optional)
- Features
  - [ ] Peer-to-peer file sync; conflict resolution
  - [ ] QUIC/HTTP3 support for low latency
  - [ ] Feature flags toggles at runtime
- Internationalization & Accessibility
  - [ ] i18n scaffolding for UI
  - [ ] Accessibility guidelines

P3 — Low
- Marketplace
  - [ ] Plugin registry metadata, semantic version constraints, auto-updates
- Community
  - [ ] Code of conduct, governance doc, roadmap page


## 3) Detailed tasks and acceptance criteria

Security (P0)
- [ ] JWT auth
  - [ ] HS256/RS256 support, key rotation
  - [ ] Issuer, audience, expiry, nbf validation
  - [ ] Middleware enforcing permissions on protected routes
  - Acceptance: Endpoints reject invalid/missing tokens; integration tests pass
- [ ] RBAC
  - [ ] Role-to-permissions mapping
  - [ ] Route-level required permissions
  - Acceptance: Matrix test verifying role access rules
- [ ] TLS
  - [ ] Configurable certs/keys; TLS 1.3; modern ciphers
  - [ ] HSTS; HTTP→HTTPS redirect
  - Acceptance: SSL Labs-like checks; curl tests succeed

Networking (P0)
- [ ] Peer discovery
  - [ ] mDNS-based discovery with configurable domain
  - [ ] Manual peer join
  - Acceptance: Peers find each other automatically in LAN
- [ ] Health and messaging
  - [ ] Keepalive pings; peer eviction; reconnects
  - [ ] Basic messaging API with backpressure
  - Acceptance: Chaos tests demonstrate resilience

Metrics & Observability (P0)
- [ ] Prometheus metrics
  - [ ] /metrics endpoint with standard Go, HTTP, custom metrics
  - [ ] Labels: service, route, status_code, peer_id
  - Acceptance: Prometheus scrape works; sample Grafana boards load
- [ ] Tracing
  - [ ] OpenTelemetry with OTLP exporter; spans for HTTP and plugins
  - Acceptance: Traces visible in Jaeger/Tempo

Resources (P0)
- [ ] Streaming & persistence
  - [ ] Range requests; resumable uploads (tus/HTTP chunking)
  - [ ] Storage drivers; retention policies
  - Acceptance: Transfer large files reliably under load
- [ ] API model
  - [ ] Consistent JSON schema; error codes
  - Acceptance: Contract tests across create/get/delete/stream

EventBus (P0)
- [ ] Async/evented core
  - [ ] Buffered pub/sub; topics; wildcard; backpressure
  - Acceptance: Load tests show no drops under expected rates

HTTP Service (P0)
- [ ] Consolidation
  - [ ] Remove duplicate HTTP implementations; single authoritative server
  - Acceptance: Grep shows one HTTP service; tests green
- [ ] OpenAPI docs
  - [ ] Generate spec; serve Swagger UI
  - Acceptance: Docs accessible at /api/docs; matches handlers

Plugin System (P0)
- [ ] Discovery and sandbox
  - [ ] Load from directories; signature checks
  - [ ] Isolation (process/WASM); resource quotas
  - Acceptance: Malicious plugin cannot crash platform; tests pass

Testing & CI (P0)
- [ ] Test coverage
  - [ ] Unit tests for managers/services/plugins (>=80%)
  - [ ] Integration/e2e with docker-compose
  - Acceptance: CI reports coverage and passes
- [ ] Pipelines
  - [ ] Lint (golangci-lint), vet, race detector
  - [ ] Release builds for linux/amd64, arm64, darwin, windows
  - Acceptance: Tagged releases upload assets with checksums/signatures


## 4) API status matrix (selected)

Platform
- [x] GET /health
- [x] GET /info
- [x] GET /api/platform/health
- [x] GET /api/platform/info
- [~] GET /api/platform/metrics (json/text) — needs Prometheus exporter

Plugins
- [x] GET /api/plugins
- [x] GET /api/plugins/:name
- [x] POST /api/plugins/:name/start
- [x] POST /api/plugins/:name/stop
- [x] GET /api/plugins/:name/health
- [x] Dynamic plugin routes under /plugins/{plugin}

Network
- [x] GET /api/network/peers
- [~] POST /api/network/peers/discover — returns current; needs real discovery

Resources
- [x] GET /api/resources
- [x] GET /api/resources/:id
- [x] POST /api/resources — creates memory resource (json)
- [x] DELETE /api/resources/:id
- [~] GET /api/resources/:id/stream — minimal stub; needs real streaming

Events
- [x] GET /api/events/stream (SSE)
- [x] POST /api/events/publish

Built-in Plugin Routes
- [x] File Manager: list/upload/download/delete
- [x] Clipboard: get/set/history/clear
- [x] System Info: info/health


## 5) Architectural cleanups and risks

- [ ] Duplicate HTTP service implementations (internal/services/http.go vs internal/core/http.go)
  - Risk: Divergent behavior, maintenance burden
  - Action: Choose internal/services/http.go as canonical and delete/redirect the other
- [ ] Mixed “platform” concepts (internal/platform vs internal/core Platform types)
  - Risk: Confusion and conflicting APIs
  - Action: Consolidate to one Platform type; migrate or remove unused one
- [ ] In-memory managers suitable for demo, not for production
  - Action: Implement production-grade managers as listed above
- [ ] Plugin isolation/sandboxing
  - Risk: Plugin can crash/compromise host
  - Action: Externalize plugin process or WASM runtime with strict caps
- [ ] Context propagation
  - Action: Ensure request contexts pass through all subsystems (timeouts/cancellation everywhere)
- [ ] Error handling and logging consistency
  - Action: Define error schema; propagate and log with context and correlation IDs


## 6) Milestones & timeline (suggested)

M1: Production HTTP & Security (P0)
- Auth (JWT), RBAC, TLS, unified HTTP service, OpenAPI docs
- Prometheus metrics, basic tracing
- ETA: 2–3 weeks

M2: Networking & Eventing (P0)
- mDNS discovery, peer lifecycle, async EventBus
- ETA: 2 weeks

M3: Resources & Persistence (P0)
- Streaming, resumable transfers, storage driver, retention
- ETA: 2–3 weeks

M4: Observability & Reliability (P0)
- pprof, tracing expansion, load tests, chaos tests
- ETA: 1–2 weeks

M5: CI/CD & Packaging (P0/P1)
- Lint, test, build, release; container hardening; Helm chart
- ETA: 1–2 weeks

M6: Legacy/UI Strategy (P1)
- Decide integrate/deprecate; initial admin UI
- ETA: 2 weeks


## 7) Immediate next steps (next PRs)

- [ ] Consolidate HTTP service (remove duplicate; keep internal/services/http.go)
- [ ] Implement Prometheus exporter and expose /metrics
- [ ] Introduce JWT-based auth middleware and enforce on privileged endpoints
- [ ] Add OpenAPI spec and Swagger UI at /api/docs
- [ ] Add unit tests for managers and HTTP handlers; wire CI with lint/test
- [ ] Implement mDNS peer discovery (opt-in via config)
- [ ] Implement proper resource streaming and Range support
- [ ] Add plugin config schema and route auth requirements


## 8) References & conventions

- API design: JSON, snake_case fields? Decide and standardize (document in API style guide)
- Logging fields: time, level, service, route, peer_id, request_id, err
- Versioning: SemVer; document breaking changes policy
- Config precedence: flags > env > file; document supported keys

This document will evolve as we implement and learn. Contributions welcome—please keep tasks scoped and reference this TODO in PR descriptions.
## 🚀 Project Roadmap & Architecture Evolution

This project follows an evolutionary architecture approach, starting from a foundational CRUD application and progressively scaling into a robust, production-ready system capable of handling high concurrency, asynchronous tasks, and AI integrations.

### Phase 1: Project Initialization & Core Infrastructure
- [x] **Monorepo Setup:** Scaffold Next.js frontend and Go standard library backend.
- [x] **Development Environment:** Containerize database infrastructure using Docker Compose.
- [x] **Data Access Layer:** Design relational PostgreSQL schema and implement type-safe queries using `sqlc`.
- [x] **Identity Management:** Implement secure, stateless Authentication and Authorization using JWT and HTTP-only cookies.

### Phase 2: Core Domain & Data Lifecycle
- [x] **Forum Foundation:** Implement end-to-end CRUD capabilities for Public Policy Posts and nested Comments.
- [x] **Data Lifecycle Management:** Implement soft-delete logic and edit histories to manage typos and content moderation securely.
- [x] **Data Bounding (Pagination):** Implement limit/offset sorting to prevent unbounded arrays from overwhelming the client and server memory.

### Phase 3: Discoverability & Engagement
- [x] **Discovery (Search & Filtering):** Implement robust search capabilities (transitioning from standard SQL `ILIKE` to Postgres Full-Text Search) to ensure forum discoverability.
- [x] **Engagement Base:** Introduce relational tables for user upvotes and downvotes.
- [ ] **Concurrency & Data Integrity:** Resolve "Double Vote" race conditions by implementing atomic increment queries (`SET votes = votes + 1`), database-level `UNIQUE` constraints, and Go-level transaction rollbacks.

### Phase 4: Payload Integrity & Standardization
- [ ] **Implement Request Validation:** Replace manual backend nil-checks with declarative struct tags using `go-playground/validator`, returning standardized error JSON for Next.js form handling.
- [ ] **Migrate to Structured Logging:** Strip out standard `log.Printf` and implement `log/slog` for JSON-formatted, key-value logging to enable machine-readable log aggregation.

### Phase 5: High-Performance Read Paths
- [ ] **Implement Cache-Aside Pattern (Redis):** Integrate `go-redis` to cache the "Trending/Popular Posts" hot path. 
- [ ] **Dynamic Cache Invalidation:** Implement targeted cache deletion (explicit `DEL`) triggered by new high-value votes to prevent stale data delivery.
- [ ] **Infrastructure Optimization:** Configure Dockerized Redis for local development with a deployment strategy targeting Upstash Serverless Redis to minimize cloud overhead.

### Phase 6: AI Integration & Asynchronous Workflows
- [x] **Synchronous AI Integration:** Implement initial LLM categorization for forum posts using structured prompt engineering.
- [ ] **Asynchronous AI Summarization Worker:** Decouple long-running LLM network calls by building an asynchronous Go background worker.
- [ ] **Stateful Job Queue:** Implement transactional job state tracking (`PENDING`, `PROCESSING`, `COMPLETED`) in PostgreSQL utilizing `FOR UPDATE SKIP LOCKED` to prevent worker collisions.
- [ ] **Uni-directional Frontend Streaming:** Replace inefficient client-side polling with Server-Sent Events (SSE) to stream job completion status directly to the Next.js UI.
- [ ] **Feature-Flagged Document Generation:** Implement the Strategy Pattern via a Go interface for Executive Policy PDF generation.
  - *Development:* Route to a Gotenberg container for headless-Chrome enterprise PDF rendering.
  - *Production:* Fallback to a `NoOpGenerator` paired with a meticulously styled `@media print` CSS solution on the frontend to optimize cloud costs while retaining professional output.

### Phase 7: Distributed Observability
- [ ] **Instrument OpenTelemetry (OTel):** Inject trace IDs at the Next.js edge and propagate context through the Go HTTP router, Redis cache, and PostgreSQL queries.
- [ ] **Latency Visualization:** Output distributed traces to a local Jaeger container to generate waterfall charts for bottleneck identification.

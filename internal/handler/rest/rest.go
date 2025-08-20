package rest

// GET /ping — super fast “is the process alive?” check.

// GET /health — full check that dependencies (DB, cache, etc.) are working.

// GET /live (liveness probe) — only checks that the app process hasn’t crashed/hung.

// GET /ready (readiness probe) — checks that the app is ready to handle requests (e.g., DB
// connected).

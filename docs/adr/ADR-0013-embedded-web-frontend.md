# ADR-0013: Embedded Web Frontend

**Status:** Accepted

**Date:** 2026-07-07

## Context

The platform exposes 18 REST API endpoints serving telemetry, PLC configuration, alarms, and system status. A web-based dashboard is required to visualize this data and provide clickable controls for pause/resume.

The frontend must be:

* Deployable as a single binary with no separate build or deployment step.
* Zero-dependency for development — no npm install, no build tooling, no React CLI.
* Easy for an operator to open by navigating to the API root.

## Decision

Embed a single-page application (SPA) directly in the Go binary using `//go:embed` and serve it at the API root (`/`).

### Architecture

```
internal/web/
    embed.go         # embeds static/ via //go:embed
    static/
        index.html   # single HTML file with inline CSS and JS
```

The API router (chi) registers a catch-all `/*` handler that serves the embedded file system. API routes take precedence because they are registered first.

### Frontend Stack

* **Zero frameworks** — plain HTML, CSS, and vanilla JavaScript.
* **SPA routing** — a simple `navigate()` function manages page state and renders content into a container div.
* **Data fetching** — `fetch()` calls to the API with JSON parsing.
* **DOM manipulation** — a small `$el()` helper creates elements declaratively.
* **Pages** — Dashboard, PLCs, Telemetry (latest/history/aggregate), Alarms, Controls.

### Page Design

| Page | Endpoints Used |
|---|---|
| Dashboard | `/system/status`, `/telemetry/latest` |
| PLCs | `/plcs`, `/plcs/{id}`, `/plcs/{id}/tags`, `/plcs/{id}/status`, `/telemetry/latest/{id}` |
| Telemetry | `/telemetry/latest`, `/telemetry/history`, `/telemetry/aggregate` |
| Alarms | `/alarms`, `/alarms/active` |
| Controls | `/collector/status`, `POST /collector/pause`, `POST /collector/resume` |

## Alternatives Considered

### React SPA in a Separate Repository

Pros

* Rich component ecosystem
* Familiar to most frontend developers
* Hot reload during development

Cons

* Requires Node.js and npm
* Separate build and deployment pipeline
* CORS configuration needed for cross-origin requests
* Increases cognitive load for operators who primarily know Go

---

### Go Templates (server-side rendering)

Pros

* No client-side JavaScript complexity
* Works without JS enabled

Cons

* Full page reload on every navigation
* Mixes presentation logic with Go code
* Harder to build interactive controls like pause/resume

---

### Embedded Vanilla SPA (Selected)

Pros

* Single binary deployment — no external files or servers
* No Node.js, npm, or build step
* Hot-reload by refreshing the browser during development
* API and frontend are always in sync (same binary, same port)
* No CORS configuration
* Minimal surface area — easy to understand and modify

Cons

* No component library — all UI is hand-crafted
* No hot module replacement — must refresh the browser
* Limited to what vanilla JS can do efficiently
* Not suitable for highly interactive or complex UIs

## Rationale

For an industrial monitoring dashboard with primarily read-only data display and a handful of control actions, a vanilla SPA is the most pragmatic choice. The primary goal is to visualize telemetry data — not to build a highly interactive consumer application.

Embedding the frontend in the binary eliminates an entire class of deployment concerns: version mismatches between API and frontend, CORS configuration, static file servers, and build pipeline maintenance.

The catch-all route pattern ensures that navigating to `http://localhost:8081/` in a browser immediately serves the dashboard, while API clients using the same port receive JSON.

## Consequences

### Positive

* Zero frontend build tooling.
* Single binary includes everything.
* API and frontend are always version-matched.
* Frontend development is just editing one HTML file.
* Dashboard works immediately after `go run`.

### Negative

* No component reuse — each page duplicates HTML/CSS patterns.
* Large single file (index.html) as the application grows.
* No TypeScript, linting, or testing for the frontend.

## Future Considerations

If the dashboard grows beyond a handful of pages with complex interactivity, a dedicated frontend application (React, Vue, or similar) should be built in a separate repository. At that point, the embedded SPA can remain as a lightweight operational overview while the dedicated frontend handles advanced features.

CORS support should be added to the API server before developing a separate frontend.

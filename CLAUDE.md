# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Layout

Two independent projects, no shared build:

- `api/` — Go module `github.com/omjogani/shape-hill` (go 1.26.5). Currently a `main.go` hello-world, no dependencies.
- `web/` — Next.js 16 + React 19 + Tailwind v4 app (App Router, TypeScript), managed with **pnpm**. Currently the default starter.

Both are scaffolds; there is no wiring between them yet. Expect to establish conventions rather than follow existing ones.

## Commands

```bash
# api/
go run .
go build ./...
go test ./...            # single test: go test -run TestName ./...

# web/  (pnpm, not npm)
pnpm dev                 # next dev
pnpm build
pnpm lint                # eslint
```

No test setup exists in `web/` yet.

## Next.js version caveat

From `web/AGENTS.md`: this Next.js version has breaking changes vs. what's likely in training data. Read the relevant guide in `web/node_modules/next/dist/docs/` before writing Next-specific code, and heed deprecation notices.

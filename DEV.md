# Development

- `SPECIFICATION.md`: the normative spec.
- `tik-go/`: Go parser and TIK to ICU translator.
- `cheatsheet/`: site generator (Go + templ) for the landing page and the cheatsheet.
- `site/`: GitHub Pages root with the shared theme and the generated (uncommitted) pages.

Requires Go 1.27.2 (as in CI) and [golangci-lint](https://golangci-lint.run).

## tik-go

```sh
cd tik-go
go test ./...
go test -run '^$' -bench . # benchmarks
go test -fuzz FuzzTokenize # fuzzing
golangci-lint run
```

## Site

- `cheatsheet/landing.templ`: the landing page. Its spec link uses the latest `v*` tag.
- `cheatsheet/main.go`: cheatsheet entries. Markup and styles are in `cheatsheet/template.templ`.
- `site/theme.css`: color tokens for both pages. `site/theme.js` sets `data-theme` from the system color scheme.

```sh
cd cheatsheet
make dev # opens a live-reloading preview at http://127.0.0.1:7331
make all # tidy, run templ, write the pages into site/
```

Run `make all` before committing; CI fails if generated Go code is stale. Edits to `site/theme.*` need a manual refresh.

Pushes to `main` and spec release tags rebuild the site and deploy it to GitHub Pages.

## Spec changes

Update `SPECIFICATION.md`, `tik-go/` and `cheatsheet/main.go` together.
Keep the tables of contents in `README.md` and `SPECIFICATION.md` in sync with the headings.

## Releasing

- Spec: bump `Version` and `Last updated` in `SPECIFICATION.md`, then tag `vX.Y.Z`. Pushing the tag redeploys the site with the new spec link.
- tik-go: tag `tik-go/vX.Y.Z`.

## Commits

[Conventional Commits](https://www.conventionalcommits.org) with a capitalized subject, e.g. `fix: Incorrect ICU in examples`.
Mark breaking changes with `!`, e.g. `feat!: Blank cardinal plural`.

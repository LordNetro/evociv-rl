---
name: Pull Request
about: Enviar cambios para revisión
title: "type(scope): description"
labels: []
---

## Linked Issue

Closes #

<!-- REQUIRED: El issue debe tener label status:approved -->

## PR Type

Selecciona EXACTAMENTE UNO:

- [ ] 🐛 Bug fix (`type:bug`)
- [ ] ✨ New feature (`type:feature`)
- [ ] 📝 Documentation only (`type:docs`)
- [ ] ♻️ Code refactoring (`type:refactor`)
- [ ] 🔧 Maintenance/tooling (`type:chore`)
- [ ] 💥 Breaking change (`type:breaking-change`)

## Summary

<!-- 1-3 bullet points describiendo los cambios -->

-

## Changes

| File | Change |
|------|--------|
| `path/to/file` |  |

## Test Plan

- [ ] `go build ./cmd/evociv` compila sin errores
- [ ] `go test ./... -v` todos los tests pasan
- [ ] `go vet ./...` sin advertencias
- [ ] Probado manualmente (describir cómo)

## Contributor Checklist

- [ ] Linked an approved issue with `status:approved`
- [ ] Added exactly one `type:*` label
- [ ] `go build` y `go test ./...` pasan
- [ ] `go vet ./...` limpio
- [ ] Conventional commit format usado en los mensajes
- [ ] Sin `Co-Authored-By` trailers

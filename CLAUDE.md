# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go client library for the [Azure Naming Tool](https://github.com/mspnp/AzureNamingTool) API (v5+). It is consumed by [terraform-provider-proactnaming](https://github.com/proact-global/terraform-provider-proactnaming) and can also be imported directly into Go programs. The module path is `github.com/proact-global/azurenamingtool-client-go`, package name `azurenamingtool`.

## Commands

```bash
go build ./...   # verify the module builds
go vet ./...     # static checks
go test ./...    # run tests (none currently exist)
```

There is no Makefile or lint config beyond `go vet`. CI (`.github/workflows/release.yml`) runs `go build ./...` and `go test ./...` on every `v*` tag push, then creates a GitHub Release — this is a library with no binaries, so the release step is just the tag + generated notes.

## Architecture

Each API operation lives in its own top-level file, all in package `azurenamingtool`:

- `client.go` — `Client` struct, `NewClient(host, apiKey, adminPassword *string)`, and `doRequest`, the single low-level HTTP helper every operation funnels through. It sets the `APIKey`/`AdminPassword`/`Content-Type` headers, and on non-200 responses tries to unmarshal the V2 `{error: {code, message}, metadata: {correlationId}}` envelope into a formatted error before falling back to a raw status/body error.
- `models.go` — all request/response structs, including the generic `ApiResponse[T]` V2 envelope (`{success, data, error, metadata}`).
- `generate_name.go`, `delete_name.go`, `resource_types.go` — one file per operation (`GenerateName`, `GetName`, `DeleteName`, `GetResourceTypes`).

Two API surfaces are mixed in this client, and new operations must match the right one:

- **V2 API** (`/api/v2.0/...`): responses are wrapped in `ApiResponse[T]`; check `resp.Success` and unwrap `resp.Data`. Used by `GenerateName` and `GetResourceTypes`.
- **Admin V1 API** (`/api/Admin/...`, no version prefix): requires `AdminPassword` to be set on the client, returns raw objects (not the `ApiResponse` envelope), and always answers HTTP 200 — failures are signaled in the body text (e.g. containing `"FAILURE"` or `"not found"`) rather than via status code. `GetName` and `DeleteName` follow this pattern; both string-match on the error/body text to map "not found" to `ErrNotFound` (`GetName`) or to a success no-op (`DeleteName`, since the target is already gone).

`Client` has a `sync.Mutex` (`c.mu`) held around `doRequest` calls in the request functions to serialize access — follow this pattern (`c.mu.Lock()` / `doRequest` / `c.mu.Unlock()`) when adding new operations that call `doRequest`. Note `resource_types.go`'s `GetResourceTypes` does not currently take the lock; be aware of this inconsistency rather than assuming it's the template to copy.

Use `errors.Is(err, azurenamingtool.ErrNotFound)` to detect a missing generated name from `GetName`.

## Leftover scaffold files

`auth.go.old`, `coffees.go.old`, and `orders.go.old` are dead template code (a coffee-shop/orders example, entirely commented out or unused) left over from the original HashiCorp terraform-provider scaffold this repo was bootstrapped from. The `.old` extension excludes them from the Go build — they are not part of the client's actual API surface. Don't extend them or treat them as reference for current patterns; look at `generate_name.go` / `delete_name.go` / `resource_types.go` instead.

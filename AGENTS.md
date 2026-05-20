# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go Terraform provider for vpsAdmin. The provider entry point is `main.go`; provider schemas, resources, data sources, and API wiring live in `vpsadmin/`. Generated Terraform Registry documentation is stored in `docs/`, with templates in `templates/` and runnable Terraform examples in `examples/`. The `get-token/` directory is a separate Go module for the helper CLI used to obtain API tokens. Tool dependencies used by `go generate` are pinned through `tools/tools.go`.

## Build, Test, and Development Commands

- `make build`: compiles `terraform-provider-vpsadmin` in the repository root.
- `make install`: copies the built provider to `~/.terraform.d/plugins/terraform.vpsfree.cz/vpsfreecz/vpsadmin/<version>/<platform>/`.
- `make fmt`: runs `go fmt` for the root package and `vpsadmin/`.
- `make docs`: runs `go generate`, which invokes `tfplugindocs` from `main.go`.
- `go test ./...`: runs tests for the main provider module.
- `(cd get-token && go test ./...)`: runs tests for the nested helper module.

Use `nix-shell` when you want the repository-provided development environment. The examples use OpenTofu/Terraform style commands: `init`, `plan`, and `apply`.

## Coding Style & Naming Conventions

Follow `gofmt` formatting. Go files use tabs with width 4; Terraform files use two-space indentation, as defined in `.editorconfig`. Keep provider names and Terraform identifiers in the existing `vpsadmin_<name>` pattern, for example `vpsadmin_vps` or `vpsadmin_ssh_key`. Resource and data source files follow `resource_<name>.go` and `datasource_<name>.go`.

## Testing Guidelines

Add focused Go tests next to the code they cover using standard `*_test.go` naming. Prefer unit tests for schema behavior, helper logic, and API request shaping. Run `go test ./...` before submitting provider changes, and run the `get-token` test command when changing that module.

## Commit & Pull Request Guidelines

Recent commits use short, imperative messages, often scoped with a prefix such as `examples:` or `get-token:`. Keep the first line concise, for example `examples: use opentofu` or `get-token: update dependencies`. Pull requests should describe the behavior change, note any generated docs updates, list test commands run, and link related issues when available.

Every commit must be created from a temporary message file, not with `git commit -m`. Each commit message must explain what changed and why, including the problem and the solution. No commit message line may exceed 80 characters; validate the temporary file line lengths before committing with `git commit -F "$msgfile"`.

## Security & Configuration Tips

Do not commit API tokens, `.tfvars` files, Terraform state, crash logs, or local `.terraform/` directories. Prefer `VPSADMIN_API_TOKEN` and `VPSADMIN_API_URL` for local testing. When running example applies against real infrastructure, review plans carefully and use `-parallelism=1` when managing multiple VPS instances.

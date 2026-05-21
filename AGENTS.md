# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go Terraform provider for vpsAdmin. The provider entry point is `main.go`; provider schemas, resources, data sources, and API wiring live in `vpsadmin/`. Generated Terraform Registry documentation is stored in `docs/`, with templates in `templates/` and runnable Terraform examples in `examples/`. The `get-token/` directory is a separate Go module for the helper CLI used to obtain API tokens. Tool dependencies used by `go generate` are pinned through `tools/tools.go`.

## Build, Test, and Development Commands

- `make build`: compiles `terraform-provider-vpsadmin` in the repository root.
- `make install`: copies the built provider to `~/.terraform.d/plugins/terraform.vpsfree.cz/vpsfreecz/vpsadmin/<version>/<platform>/`.
- `make fmt`: runs `go fmt` for the root package and `vpsadmin/`.
- `make docs`: runs `go generate`, which invokes `tfplugindocs` from `main.go`.
- `make test`: runs Go tests for the main provider module.
- `make test-get-token`: runs Go tests for the nested helper module.
- `make test-integration`: runs ci-tagged integration tests through `test-runner.sh`.

Use `nix develop` when you want the repository-provided development environment. The examples use OpenTofu/Terraform style commands: `init`, `plan`, and `apply`.

## Coding Style & Naming Conventions

Follow `gofmt` formatting. Go files use tabs with width 4; Terraform files use two-space indentation, as defined in `.editorconfig`. Keep provider names and Terraform identifiers in the existing `vpsadmin_<name>` pattern, for example `vpsadmin_vps` or `vpsadmin_ssh_key`. Resource and data source files follow `resource_<name>.go` and `datasource_<name>.go`.

## Testing Guidelines

Add focused Go tests next to the code they cover using standard `*_test.go` naming. Prefer unit tests for schema behavior, helper logic, and API request shaping. Run both `make test` and `make test-get-token` before submitting Go changes. CI runs both commands against the latest patch releases of the supported Go major versions; update the workflow matrix when Go's supported release window changes. The `go` directive in `go.mod` is the minimum supported toolchain version, not a patch-level lock. For integration-test or provider workflow changes, run `make test-integration` or the focused `./test-runner.sh test ...` command.

## Commit & Pull Request Guidelines

- Use short imperative subjects, often scoped (`examples: use opentofu`,
  `get-token: update dependencies`); keep one logical change per commit.
- Every commit message must explain what the change does and why it is
  needed; use the subject for the action and the body for the rationale
  when needed.
- Wrap every commit message line at 80 characters or fewer.
- Always write the commit message to a temporary file and commit with
  `git commit -F <tmpfile>` instead of passing the message inline.

- Flake input updates flow:
  1. Read current rev: `nix flake metadata --json . | jq -r '.locks.nodes.<input>.locked.rev'`.
  2. Update input: `nix flake update <input>` (or `nix flake lock --update-input <input>`).
  3. Verify only `flake.lock` changed for this update commit.
  4. Commit with subject format: `flake: <input> <old9> -> <new9>` (example: `flake: vpsadmin f9db2d4ff -> 123456789`).

Pull requests should describe the behavior change, note any generated docs updates, list test commands run, and link related issues when available.

## Security & Configuration Tips

Do not commit API tokens, `.tfvars` files, Terraform state, crash logs, or local `.terraform/` directories. Prefer `VPSADMIN_API_TOKEN` and `VPSADMIN_API_URL` for local testing. When running example applies against real infrastructure, review plans carefully and use `-parallelism=1` when managing multiple VPS instances.

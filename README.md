Terraform Provider for vpsAdmin
===============================

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 1.x
- [Go](https://golang.org/doc/install) 1.26.6+ (to build the provider plugin)

With Nix, enter the provider development environment with:

```sh
$ nix develop
```

For the examples environment, including OpenTofu and the `get-token` helper,
run:

```sh
$ nix develop .#examples
```

Building The Provider
---------------------

Clone the repository anywhere. Enter the directory and build the provider with:

```sh
$ make build
```

Terraform can then be called from the provider directory. For the provider to
be available outside of the provider directory, it has to be installed
to `~/.terraform.d/plugins`:

```sh
$ make install
```

Using the provider
------------------

See the [examples](./examples).

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org)
installed on your machine (version 1.26.6 or newer is required).
To compile the provider, run:

```sh
$ make build
```

Run local checks with:

```sh
$ make test
$ make test-get-token
$ make test-integration
```

## Automated flake updates

The `update-vpsadmin-flake.yml` workflow uses a separately pinned confctl CI
package to require commit ancestry preservation across the full changed lock graph.
Rejected candidates and unavailable evidence stop publication. The workflow
checks the final lock against its saved baseline after hooks and checks, then
uses an ordinary non-force push. If the target branch advances, rerun from its
new baseline instead of rebasing the old lock update.

For a local update, set `CONFCTL_FLAKE` to the immutable package reference in
`.github/workflows/update-vpsadmin-flake.yml`, then run
`tools/update-vpsadmin-flake-input.sh` and review its generated commit.

The workflow saves confctl's Git mirrors in `~/.cache/confctl` even when an
update is rejected. Every update checks ancestry, whether the cache was
restored or freshly cloned.

See the [shared guard contract](https://github.com/vpsfreecz/confctl/blob/822c0e35dbe2549111912f539116b4225dfd3035/docs/flake-inputs.md#guarded-automated-updates) for supported sources and failure behavior.

Terraform Provider for vpsAdmin
===============================

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 1.x
- [Go](https://golang.org/doc/install) 1.16+ (to build the provider plugin)

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
installed on your machine (version 1.16+ is *required*).
To compile the provider, run:

```sh
$ make build
```

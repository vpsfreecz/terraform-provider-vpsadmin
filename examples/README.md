# vpsAdmin VPS configuration example
The example configuration creates one VPS with `remote-exec` provisioner
and deploys a public key.

## Obtaining API authentication token
The provider needs an authentication token to the vpsAdmin API. The token can
be obtain using any of [HaveAPI clients](https://github.com/vpsfreecz/haveapi),
but the provider also comes with a simple CLI utility [get-token](../get-token).

For this example, the token should be put in an arbitrary tfvars file, e.g.
`token.auto.tfvars`:

```
vpsadmin_token = "your token"
```

## Setup
Edit `main.tf` and set up your public key for deployment and private key
for provisioner.

## Run it
```sh
$ terraform init
$ terraform plan
$ terraform apply
```

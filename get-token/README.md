# get-token
`get-token` is a CLI utility that can be used to obtain vpsAdmin authentication
token for use with the terraform provider.

## Build

```sh
$ go build -o get-token
```

## Usage
```sh
$ get-token --help
Usage:
  ./get-token [options] [api url]

Options:
  -interval int
        How long should the token be valid, in seconds (default 3600)
  -lifetime string
        Token lifetime (default "renewable_auto")
  -tfvars string
        Write the token to a dedicated .tfvars file
  -user string
        User name
```

The API URL defaults to <https://api.vpsfree.cz>.

Without any parameters, the program will return a token which will be valid
for one hour, or an hour since it was last used for authentication.

Possible token lifetimes are:

 - fixed
 - renewable\_manual
 - renewable\_auto
 - permanent

`get-token` can also write the token to a tfvars file for use with terraform:

```sh
$ get-token -tfvars ../examples/token.auto.tfvars
```

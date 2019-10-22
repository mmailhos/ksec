# Ksec
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/MathieuMailhos/ksec)](https://goreportcard.com/report/github.com/MathieuMailhos/ksec) [![Build Status](https://travis-ci.org/MathieuMailhos/ksec.svg?branch=master)](https://travis-ci.org/MathieuMailhos/ksec)


ksec is an easy-to-use tool to find, sort and decode your Kubernetes secrets. It supports Helm versioning.
No crazy combination of grep/yq/base64 anymore.

## Installation

Set up your `$GOPATH`, then run:
```
go get github.com/mathieumailhos/ksec
```

## Usage

```
ksec [OPTIONS] <secret> [KEY]
```

This example retrieves the latest deployed secret that contains `mysql` and find all data keys with `pass`:

```
$ ksec mysql pass
DATA_DB_MYSQL_PASSWORD: r4nd0mP4ss
USER_DB_MYSQL_PASSWORD: pa$$w0rd123
```

## Options

```
--namespace <namespace name>
--label <kubernetes label>
--selector <kubernetes selector>
--type <secret type (ex: Opaque)>
--color
--out [env,yaml,json]
--metadata
```

## Logic

Taken an input, here is the ordered list of rules:
  * Return if ksec finds an exact match (`my-unique-object`)
  * Return the latest version if the input is versioned (`my-object-5`)
  * Else: return the list of matchs to allow the user to rerun the command

## Roadmap

Potential incoming features:
  * Fix --metadata for all outputs
  * Fix --color for all outputs
  * Multiple outputs: yaml, json, bash env...
  * Improve argopt (ex: print Usage...)
  * Add an option to also print the configmap related to the found secret
  * Get all secrets found in a deployment
  * Any other idea? Please open an issue

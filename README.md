# Ksec

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/MathieuMailhos/ksec)](https://goreportcard.com/report/github.com/MathieuMailhos/ksec) [![Build Status](https://travis-ci.org/MathieuMailhos/ksec.png?branch=master)](https://travis-ci.org/MathieuMailhos/ksec)


ksec is an easy-to-use tool to decode your Kubernetes secrets. It supports Helm versioning.
No more crazy combination of grep/yq/base64.

## Install

```
go get github.com/MathieuMailhos/ksec
```

## Use

```
ksec [OPTIONS] <secret> [KEY]
```

Example:

```
ksec --color --namespace prod apache password
```

## Options

```
--namespace <namespace name>
--label <kubernetes label>
--selector <kubernetes selector>
--type <secret type (ex: Opaque)>
--color
--metadata
```

## Logic

Taken an input, here is the ordered list of rules:
  * Return if ksec finds an exact match (`my-unique-object`)
  * Return the latest version if the input is versioned (`my-object-5`)
  * Else: return the list of matchs to allow the user to rerun the command

## Roadmap

Potential incoming features:
  * Improve argopt and error handling (ex: print Usage, no panic...)
  * Add an option to also print the configmap related to the found secret
  * Get all secrets found in a deployment
  * Multiple outputs: yaml, json, bash env...
  * Any other idea? Please open an issue

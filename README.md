# Ksec

![goreportcard](https://goreportcard.com/badge/github.com/MathieuMailhos/ksec)

ksec is an easy-to-use tool to decode your Kubernetes secrets. It supports Helm versioning.
No more crazy combination of grep/yq/base64.

## Install

```
go get github.com/MathieuMailhos/ksec
```

## Use

```
ksec [OPTIONS] <secret>
```

## Logic

Taking an input, here is the ordered list of rules:
  * Return if ksec finds an exact match (`my-unique-object`)
  * Return the latest version if the input is versioned (`my-object-5`)
  * Else: return the list of matchs to allow the user to rerun the command

## Parameters

```
--namespace <namespace name>
--label <kubernetes label>
--selector <kubernetes selector>
--type <secret type (ex: Opaque)>
--no-color
```

## Roadmap

Potential incoming features:
  * Travis CI
  * Improve argopt and error handling (ex: print Usage, no panic...)
  * Add an option to also print the configmap related to the found secret
  * Get all secrets found in a deployment
  * Any other idea? Please open an issue

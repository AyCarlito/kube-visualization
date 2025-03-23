# kube-visualization

kube-visualization visualizes namespaces within a Kubernetes cluster. Resources are represented heirarchically,
in a graphviz directed graph.

## Prerequisites

Install the following:

1. [Go](https://go.dev/dl/)

## Usage

```shell
$ go run . --help
Allows resources in a given namespace in a Kubernetes cluster to be visualised

Usage:
  kube-visualization [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  visualize   List resources in a namespace and generate a heirarchical graph of them

Flags:
      --config string           Path to configuration file. (default "config/config.json")
  -h, --help                    help for kube-visualization
      --label-selector string   Filter resources by label. Comma separated key-value pairs.
      --namespace string        Namespace of resources. (default "default")
      --output string           Path to output file. (default "assets/output.dot")

Use "kube-visualization [command] --help" for more information about a command.
```

## Example

- Visualize the resources in the "default" namespace.

```shell
make run FLAGS="--namespace default"
```

- Generate a png image from the output graph.

```shell
make generate
```

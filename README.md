[![Go Report Card](https://goreportcard.com/badge/github.com/scaleway/scaleway-operator)](https://goreportcard.com/report/github.com/scaleway/scaleway-operator)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/scaleway/scaleway-operator)
![GitHub](https://img.shields.io/github/license/scaleway/scaleway-operator?style=flat)

# Scaleway Operator for Kubernetes

⚠️  this project is not maintained anymore.

The **Scaleway Operator** is a Kubernetes controller that lets you create Scaleway Resources directly from Kubernetes via [Kubernetes Custom Resources Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).

**WARNING**: this project is under active development and should be considered alpha.

## Features

Currently, **Scaleway Operator** only supports RDB instances, databases and users. Other resources will be implemented, and [contributions](./CONTRIBUTING.md) are more than welcome!

If you want to see a specific Scaleway product, please [open an issue](https://github.com/scaleway/scaleway-operator/issues/new) describing which product you'd like to see.

## Getting Started

First step is to install [Cert Manager](https://cert-manager.io/docs/installation/kubernetes/) in order to handle the webhooks certificates.

```bash
$ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.0.3/cert-manager.yaml
```

Once Cert Manager is up and running, you have to install the CRDs. First clone the repo, and then making sure your `KUBECONFIG` environment variable is set on the right cluster, run:
```bash
$ make install
```

Then, run:
```bash
kubectl create -f deploy/scaleway-operator-secrets.yml --edit --namespace=scaleway-operator-system
```

and replace the values.

Finally, in order to deploy the Scaleway Operator, run:
```bash
kubectl apply -k config/default
```

## Development

If you are looking for a way to contribute please read [CONTRIBUTING](./CONTRIBUTING.md).

### Code of conduct

Participation in the Kubernetes community is governed by the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Reach us

We love feedback. Feel free to reach us on [Scaleway Slack community](https://slack.scaleway.com), we are waiting for you on #k8s.

You can also join the official Kubernetes slack on #scaleway-k8s channel

You can also [raise an issue](https://github.com/scaleway/scaleway-operator/issues/new) if you think you've found a bug.

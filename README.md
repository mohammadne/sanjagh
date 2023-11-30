# Sanjagh 📌

> `Sanjagh` is a Persian word meaning Pin

Sanjagh is a simple k8s operator to handle deployments and aims to demonstrate best practices for how using operators, I tried to keep things as simple as possible and demonstrate how to deploy and structure your operator projects.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

```sh
# initialize base project
operator-sdk init --domain=mohammadne.me --project-name=sanjagh --repo=github.com/mohammadne/sanjagh

# https://book.kubebuilder.io/migration/multi-group.html
operator-sdk edit --multigroup=false

# resource -> generates the api directory
# controller -> generates the controller directory
operator-sdk create api --group=apps --version=v1alpha1 --kind=Executer --controller --resource

operator-sdk create webhook --group apps --version v1alpha1 --kind Executer --programmatic-validation
```

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Deployment

1. Push your operator image to the remote registry

**NOTE:** Make sure to update the operator image in the deployment as well.

**NOTE:** If you are using ghcr registry, you have to get a personal access token from the `Developer Settings` tab of the settings of the github and proceed with [this document](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

```sh
# tag your HEAD
git tag v0.1.0-rc1

# trigger Github CI action automatically and skip the following
git push origin v0.1.0-rc1

# see your version to be deployed to the registry
make version

# build and push your image to the location specified by `IMAGE`
# skip this command if you have already triggered the Github CI.
make docker-build docker-push
```

2. Install CRDs and other related k8s manifests

```sh
# install the CRDs into the cluster
make install

# install sanjagh dependency charts
helm repo add mohammadne https://mohammadne.me/charts/
helm dependency build deployments/sanjagh

# deploy sanjagh to the cluster via the helmsman
make deploy
```

3. Undeploy controller

```sh
# delete the controller from the cluster
make undeploy

# delete the CRDs from the cluster
make uninstall
```

## Development

> **_NOTE:_** if you don't have the webhook, you don't need to do the [Deployments](##Deployments), also you have to skip steps 0, 1, 2 and some part of step 4.

> **_NOTE:_** you have to build the correct image for webhook and manager images and the webhook deployment should have at least one available replicas in order for telepresence to inject its sidecar container to intercept network traffic.

0. Deactivate the manager from the cluster

```sh
kubectl scale deployment sanjagh-manager -n operators --replicas=0
```

1. Clone `tls` secrets and put them in appropriate directory.

```sh
tls_secret=$(kubectl get secrets sanjagh-webhook-tls -n operators -ojson)
echo $tls_secret | jq '.data."tls.key"' -r | base64 -d > secrets/tls/key.pem
echo $tls_secret | jq '.data."tls.crt"' -r | base64 -d > secrets/tls/crt.pem
```

2. Ensure `telepresence` is installed in your cluster with the same version with your client, you can run `telepresence helm upgrade` to sync the versions

```sh
# make sure consistency between client and server versions
telepresence helm upgrade

# start telepresence agent
telepresence connect

# intercept api-server traffic into your localhost via injecting side-car container
telepresence intercept sanjagh-webhook --port 9443 -n operators --service sanjagh-webhook
```

3. Install the CRDs into the (local) cluster

```sh
make install
```

4. Run webhook and manager with appropriate configuration

```bash
# start webhook server
go run main.go webhook

# start controllers (controller manager)
go run main.go manager
```

5. Install sample-executer in your (local) cluster

```sh
helm upgrade --install sample-executer ./deployments/sample-executer
```

6. Cleaning up

```sh
helm uninstall sample-executer

# delete the CRDs from the cluster
make uninstall
```

# Sanjagh 📌

> `Sanjagh` is a Persian word meaning pin which I choose this word because it hooks things together.

Sanjagh is a simple k8s operator to handle deployments and aims to demonstrate best practices for how using operators, I tried to keep things as simple as possible and demonstrate how to deploy and structure your operator projects.

## How it works

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

## Infrastructure Provisioning

I have developed an ansible playbook in order to provision the required infrastructure required for deploying and testing the Sanjagh.

As we are under sanctions, normally in Iran we have to use proxy servers, the playbook contains 4 roles and in order to run the `proxy` role you have to define your proxy server credentials, also the `docker` role contains proxy server to set which enables us to use it for pulling and pushing images to appropriate registry, and you have to set that variable too.
also in the `kind` role which installs the kind binary and sets up the k8s cluster we have to specify server_ip_address variable to be used when provisioning the kubernetes cluster.

## Deployment

1. Install required tools for working with the application, you can refer to [devenv](https://github.com/mohammadne/devenv) repository to install all the mentioned tools as below:

   - Go (+1.19)
   - kubectl
   - helm (helm-secret, helm-diff)
   - helmsman
   - telepresence

2. Push your operator image to the remote registry

    **NOTE 1:** Make sure to update the operator image in the deployment as well.

    **NOTE 2:** If you are using ghcr registry, you have to get a personal access token from the `Developer Settings` tab of the settings of the github and proceed with [this document](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

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

3. Put appropriate sops credentials into `~/.config/sops/age/keys.txt`

4. Install CRDs and other related k8s manifests

    ```sh
    # install the CRDs into the cluster
    make install

    # install sanjagh dependency charts
    helm repo add mohammadne https://mohammadne.me/charts/
    helm dependency build deployments/sanjagh

    # deploy sanjagh to the cluster via the helmsman
    make deploy
    ```

5. Undeploy controller

    ```sh
    # delete the controller from the cluster
    make undeploy

    # delete the CRDs from the cluster
    make uninstall
    ```

## Development

> **NOTE 1:** if you don't have webhook, you don't need to do the `Deployments` section, also you have to skip steps 1, 2, 3 and some part of step 5.
>
> **NOTE 2:** you have to build the correct image for webhook and manager images and the webhook deployment should have at least one available replicas in order for telepresence to inject its sidecar container to intercept network traffic.

1. Deactivate the manager from the cluster

    ```sh
    kubectl scale deployment sanjagh-manager -n operators --replicas=0
    ```

2. Clone `tls` secrets and put them in appropriate directory.

    ```sh
    tls_secret=$(kubectl get secrets sanjagh-webhook-tls -n operators -ojson)
    echo $tls_secret | jq '.data."tls.key"' -r | base64 -d > secrets/tls/key.pem
    echo $tls_secret | jq '.data."tls.crt"' -r | base64 -d > secrets/tls/crt.pem
    ```

3. Ensure `telepresence` is installed in your cluster with the same version with your client

    ```sh
    # make sure consistency (sync) between client and server versions
    telepresence helm upgrade

    # start telepresence agent
    telepresence connect

    # intercept api-server traffic into your localhost via injecting side-car container
    telepresence intercept sanjagh-webhook -n operators --service sanjagh-webhook --port 8443:master
    ```

4. Install the CRDs into the (local) cluster

    ```sh
    make install
    ```

5. Run webhook and manager with appropriate configuration

    ```bash
    # start webhook server
    go run main.go webhook

    # start controllers (controller manager)
    go run main.go manager
    ```

6. Install sample-executer in your (local) cluster

    ```sh
    helm upgrade --install sample-executer ./deployments/sample-executer
    ```

7. Cleaning up

    ```sh
    helm uninstall sample-executer

    # delete the CRDs from the cluster
    make uninstall
    ```

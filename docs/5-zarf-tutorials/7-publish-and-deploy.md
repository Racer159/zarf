# Store & Deploy Packages with OCI

## Introduction

In this tutorial, we are going to run through how to publish a Zarf package to an [OCI](https://github.com/opencontainers/image-spec) compliant registry, allowing end users to pull and deploy packages without needing to build locally, or transfer the package to their environment.

## System Requirements

- Internet access to download resources or upload packages
- Access to a registry (this tutorial uses Docker Hub)

## Prerequisites

For following along locally, please ensure the following prerequisites are met:

- Zarf binary installed on your `$PATH`: ([Installing Zarf](../1-getting-started/index.md#installing-zarf))
- Access to a [Registry supporting the OCI Distribution Spec](https://oras.land/implementors/#registries-supporting-oci-artifacts), this tutorial will be using Docker Hub
- Access to a cluster that has been [initialized with zarf](./1-initializing-a-k8s-cluster.md).

## Youtube Tutorial
[![Using OCI to Store & Deploy Zarf Packages Video on YouTube](../.images/tutorials/publish_and_deploy_thumbnail.png)](https://www.youtube.com/watch?v=QKxgJnC_37Y "Using OCI to Store & Deploy Zarf Packages")

## Working with OCI Packages

This tutorial will require a registry to be configured (see [prerequisites](#prerequisites) for more information).  The below sets up some variables for us to use when logging into the registry:

<iframe src="/docs/tutorials/publish_and_deploy_setup.html" height="130px" width="100%"></iframe>

With those set, you can tell Zarf to login to your registry with the following:

<iframe src="/docs/tutorials/publish_and_deploy_login.html" width="100%" height="80px"></iframe>

:::note

If you do not have the Docker CLI installed, you may need to create a Docker compliant auth config file manually:

<iframe src="/docs/tutorials/publish_and_deploy_docker_config.html" width="100%" height="225px"></iframe>

:::

### Publish Package

First, create a valid Zarf package definition (`zarf.yaml`), with the `metadata.version` key set.

<iframe src="/docs/tutorials/publish_and_deploy_manifest.html" width="100%" height="400px"></iframe>


Create the package locally:

[CLI Reference](../2-the-zarf-cli/100-cli-commands/zarf_package_create.md)

<iframe src="/docs/tutorials/publish_and_deploy_create.html" width="100%" height="600px"></iframe>

Then publish the package to the registry:

:::note

Your package tarball may be named differently based on your machine's architecture.  For example, if you are running on an AMD64 machine, the tarball will be named `zarf-package-helm-oci-chart-amd64-0.0.1.tar.zst`.

:::

[CLI Reference](../2-the-zarf-cli/100-cli-commands/zarf_package_publish.md)

<iframe src="/docs/tutorials/publish_and_deploy_publish.html" width="100%" height="700px"></iframe>

:::note

The name and reference of this OCI artifact is derived from the package metadata, e.g.: `helm-oci-chart:0.0.1-arm64`

To modify, edit `zarf.yaml` and re-run `zarf package create .`

:::

### Inspect Package

[CLI Reference](../2-the-zarf-cli/100-cli-commands/zarf_package_inspect.md)

Inspecting a Zarf package stored in an OCI registry is the same as inspecting a local package and has the same flags:

<iframe src="/docs/tutorials/publish_and_deploy_inspect.html" width="100%" height="520px"></iframe>


### Deploy Package

[CLI Reference](../2-the-zarf-cli/100-cli-commands/zarf_package_deploy.md)

Deploying a package stored in an OCI registry is nearly the same experience as deploying a local package:

<iframe src="/docs/tutorials/publish_and_deploy_deploy.html" width="100%" height="800px"></iframe>

### Pull Package

[CLI Reference](../2-the-zarf-cli/100-cli-commands/zarf_package_pull.md)

Packages can be saved to the local disk in order to deploy a package multiple times without needing to fetch it every time.

<iframe src="/docs/tutorials/publish_and_deploy_pull.html" width="100%" height="450px"></iframe>


## Removal

1. Use the `zarf package list` command to get a list of the installed packages.  This will give you the name of the games package to remove it.

<iframe src="/docs/tutorials/publish_and_deploy_list.html" width="100%"></iframe>

2. Use the `zarf package remove` command to remove the `helm-oci-chart` package.  Don't forget the `--confirm` flag.  Otherwise you'll receive an error.

<iframe src="/docs/tutorials/publish_and_deploy_remove.html" width="100%" height="115px"></iframe>

## Troubleshooting

### Failed to publish package: version is required for publishing

<iframe src="/docs/tutorials/troubleshoot_version_required_publish.html" width="100%" height="130px"></iframe>

:::info Remediation

You attempted to publish a package with no version metadata.

<iframe src="/docs/tutorials/troubleshoot_version_required_no_version.html" width="100%" height="300px"></iframe>

1. Open the zarf.yaml file.
2. Add a version attribute to the [package metadata](https://docs.zarf.dev/docs/user-guide/zarf-schema#metadata)
3. Recreate the package with the `zarf package create` command.
4. Publish the package.  The filename will now have the version as part of it.

:::

### Failed to publish, http: server gave HTTP response to HTTPS client

<iframe src="/docs/tutorials/troubleshoot_insecure_registry.html" width="100%" height="375px"></iframe>

:::info Remediation

You attempted to publish a package to an insecure registry, using http instead of https.

1. Use the `--insecure` flag.  Note that this is not suitable for production workloads.

:::

### Unable to connect to the Kubernetes cluster.

<iframe src="/docs/tutorials/troubleshoot_unreachable.html" width="100%" height="200px"></iframe>

:::info Remediation

If you receive this error, either you don't have a Kubernetes cluster, your cluster is down, or your cluster is unreachable.

1. Check your kubectl configuration, then try again.  For more information about kubectl configuration see [Configure Access to Multiple Clusters](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/) from the Kubernetes documentation.

If you need to setup a cluster, you can perform the following.

1. Deploy a Kubernetes cluster with the [Creating a K8s Cluster with Zarf](./5-creating-a-k8s-cluster-with-zarf.md) tutorial.
2. Perform the [Initialize a cluster](./1-initializing-a-k8s-cluster.md) tutorial.

After that you can try deploying the package again.

:::

### Secrets "zarf-state" not found.

<iframe src="/docs/tutorials/troubleshoot_uninitialized.html" width="100%" height="250px"></iframe>

:::info Remediation

If you receive this error when zarf is attempting to deploy any component, this means you have not initialized the kubernetes cluster.  This is one of the prerequisites for this tutorial.  Perform the [Initialize a cluster](./1-initializing-a-k8s-cluster.md) tutorial, then try again.

:::

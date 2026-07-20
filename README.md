[[_TOC_]]

# Cassandra Operator

## Installation Guide

This guide covers the full installation sequence: Cassandra node configuration charts, the Cassandra Operator, and the Cassandra Services (supplementary services). Install them in the order shown below.

### Prerequisites

- Kubernetes or OpenShift
- Helm
- `kubectl` configured against the target cluster
- A namespace created for the deployment
- Container image access to `ghcr.io/netcracker` (ensure image pull credentials are configured if the cluster cannot reach ghcr.io anonymously)

---

### Step 1 — Install the Cassandra Configuration Chart (cassandra_4.x.x or cassandra_5.x.x)

These charts produce the Kubernetes ConfigMaps that the operator mounts into Cassandra pods. Install the chart matching target Cassandra version before installing the operator.

**Cassandra 4.x.x**

```bash
helm install cassandra cassandra/cassandra-image/deployments/charts/cassandra_4.x.x
```

**Cassandra 5.x.x**

```bash
helm install cassandra cassandra/cassandra-image/deployments/charts/cassandra_5.x.x 
```

**Common overrides**

Update the `values.yaml` file according to installation requirements. Modify the required parameters to match deployment environment before installing the Helm chart.

**Verify**

```bash
kubectl get configmaps -n <namespace> | grep cassandra
```

Expected ConfigMaps: `cassandra-configuration`, `cassandra-env`, `cassandra-jvm`, `cassandra-logback`, `cassandra-audit`, `cassandra-major-version`.

---

### Step 2 — Install the Cassandra Operator Chart

The operator chart deploys the `cassandra-operator` Deployment, installs the `CassandraDeployment` CRD, and creates the `CassandraDeployment` CR that drives the Cassandra `StatefulSet` provisioning.

```bash
helm install cassandra-operator operator/charts/helm/cassandra-operator
```

**Verify the operator is running**

```bash
kubectl get pods -n <namespace> -l app=cassandra-operator
kubectl get cassandras -n <namespace>
```

**Verify Cassandra StatefulSets and pods are ready**

```bash
kubectl get statefulsets -n <namespace>
kubectl get pods -n <namespace> -l app=cassandra
```

All pods should reach `Running` status. Wait for all replicas to be ready before proceeding to the next step.

---

### Step 3 — Install the Cassandra Services Chart

The services chart deploys the supplementary services: backup-daemon, dbaas-adapter, monitoring agent, and optionally Cassandra Reaper and robot framework tests. This chart requires a running Cassandra cluster from Step 2.

```bash
helm install cassandra-services services/service/charts/helm/cassandra-services
```

**Verify all services are running**

```bash
kubectl get pods -n <namespace>
kubectl get deployments -n <namespace>
```

Expected deployments: `cassandra-services` (operator), `dbaas-adapter`, `backup-daemon`, `robot-test`.

---

### Upgrading

To upgrade any chart after modifying values, use `helm upgrade`:

```bash
# Upgrade the operator chart
helm upgrade cassandra-operator operator/charts/helm/cassandra-operator

# Upgrade the services chart
helm upgrade cassandra-services services/service/charts/helm/cassandra-services
```

---

### Uninstalling

```bash
helm uninstall cassandra-services --namespace <namespace>
helm uninstall cassandra-operator  --namespace <namespace>
helm uninstall cassandra  --namespace <namespace>
```

> **Note:** PersistentVolumeClaims are not deleted automatically. To remove them:
> ```bash
> kubectl delete pvc -n <namespace> --all
> ```

---

## Repository structure

* `./.github` - CI/CD workflow definitions, build configuration, and automation scripts for GitHub Actions.
* `./cassandra` - Cassandra-related Docker images and configuration.
  * `./cassandra/cassandra-image` - Dockerfile and Helm charts for Cassandra versions 4.1.x and 5.0.x.
  * `./cassandra/reaper` - Dockerfile and startup script for Cassandra Reaper (repair service).
* `./operator` - The main Cassandra Operator (Go).
  * `./operator/api/v1alpha1` - API type definitions for the `CassandraDeployment` custom resource.
  * `./operator/bin` - `controller-gen` binary used by the Makefile to generate CRD and DeepCopy methods.
  * `./operator/build` - Entrypoint scripts for the Cassandra Operator Docker image.
  * `./operator/charts` - Helm chart for deploying the Cassandra Operator.
  * `./operator/config` - Kubernetes manifests for CRD, RBAC roles, and sample resources.
  * `./operator/controllers` - Operator reconciliation controller.
  * `./operator/hack` - License boilerplate file.
  * `./operator/migration-artifacts` - Scripts for migration artifact handling.
  * `./operator/pkg` - Core operator source code (reconciliation logic, utilities).
  * `./operator/tests` - Unit tests for the operator.
* `./services` - Supporting services deployed alongside the operator.
  * `./services/backup-daemon` - Python-based backup and restore daemon using Ansible playbooks.
  * `./services/dbaas-adapter` - Go service implementing the DBaaS adapter interface for Cassandra.
  * `./services/service` - Cassandra supplementary operator managing `CassandraService` custom resources.
  * `./services/test` - Robot Framework integration tests.

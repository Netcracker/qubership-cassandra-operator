[[_TOC_]]

# Cassandra Operator

## Installation Guide

This guide covers the full installation sequence: Cassandra node configuration charts, the Cassandra Operator, and the Cassandra Services (supplementary services). Install them in the order shown below.

### Prerequisites

- Kubernetes 1.24+ or OpenShift 4.10+
- Helm 3.10+
- `kubectl` configured against the target cluster
- A namespace created for the deployment
- Container image access to `ghcr.io/netcracker` (ensure image pull credentials are configured if the cluster cannot reach ghcr.io anonymously)

---

### Step 1 — Install the Cassandra Configuration Chart (cassandra_4.x.x or cassandra_5.x.x)

These charts produce the Kubernetes ConfigMaps that the operator mounts into Cassandra pods. Install the chart matching your target Cassandra version **before** installing the operator.

**Cassandra 4.x.x**

```bash
helm install cassandra-config cassandra/cassandra-image/deployments/charts/cassandra_4.x.x
```

**Cassandra 5.x.x**

```bash
helm install cassandra-config cassandra/cassandra-image/deployments/charts/cassandra_5.x.x 
```

**Common overrides**

Update the `values.yaml` file according to your installation requirements. Modify the required parameters to match deployment environment before installing the Helm chart.

**Verify**

```bash
kubectl get configmaps -n <your-namespace> | grep cassandra
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
kubectl get pods -n <your-namespace> -l app=cassandra-operator
kubectl get cassandras -n <your-namespace>
```

**Verify Cassandra StatefulSets and pods are ready**

```bash
kubectl get statefulsets -n <your-namespace>
kubectl get pods -n <your-namespace> -l app=cassandra
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
kubectl get pods -n <your-namespace>
kubectl get deployments -n <your-namespace>
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
helm uninstall cassandra-services --namespace <your-namespace>
helm uninstall cassandra-operator  --namespace <your-namespace>
helm uninstall cassandra-config    --namespace <your-namespace>
```

> **Note:** PersistentVolumeClaims are not deleted automatically. To remove them:
> ```bash
> kubectl delete pvc -n <your-namespace> --all
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

---

## .github

* `assets-config.yml` - Configures documentation archive assets uploaded as release artifacts.
* `build-config.cfg` - JSON listing all Docker image components with Dockerfile paths, build contexts, and changeset paths for selective CI builds.
* `dependabot.yml` - Configures weekly Dependabot updates for Go modules and Docker base images across the monorepo.
* `docker-build-config.json` - Subset of `build-config.cfg` used exclusively by the release pipeline for operator images.
* `helm-charts-release-config.yaml` - Defines the operator Helm chart location and the images whose tags are stamped into chart values at release time.
* `release-drafter-config.yml` - Configures automated GitHub Release notes with categories and version label resolution.
* `scripts/` - Helper scripts for CI workflows.
  * `matrix.sh` - Filters build components by comparing changeset paths against changed files; returns all components if nothing matches.
* `workflows/` - GitHub Actions workflow definitions.
  * `build.yaml` - Main CI workflow; determines changed components and builds/pushes multi-arch Docker images to `ghcr.io/netcracker`.
  * `clean.yaml` - Triggered on branch deletion; removes container image versions tagged with the deleted branch name.
  * `helm-charts-release.yaml` - Manually triggered release workflow; updates image tags, builds release images, runs chart-releaser and release-drafter.
  * `license.yaml` - Runs `google/addlicense` over all `.go`/`.sh`/`.py` files on push to `main` and opens a PR if headers are missing.

---

## cassandra

### cassandra/cassandra-image

* `Dockerfile` - Builds the Cassandra node image from `eclipse-temurin:11-jre`; installs Python3, `cassandra-driver`, `cqlsh`, the Cassandra binary, and the Cassandra Exporter agent JAR.
* `pip.conf` - PyPI configuration copied into the image at build time; sets the index URL and disables the pip cache.
* `.gitignore` - Git ignore rules scoped to the `cassandra-image` subdirectory.
* `CODE-OF-CONDUCT.md` - Community code of conduct for contributors.
* `CONTRIBUTING.md` - Contribution guidelines for the cassandra-image component.
* `README.md` - Component README describing the image structure and Evergreen maintenance strategy.
* `files/` - Runtime files copied into the container.
  * `run.sh` - Container startup script; applies operator-injected config, resolves seeds, generates TLS keystores, starts `sshd` on port 2222, then starts `cassandra -f`.
  * `sshd_config` - SSH daemon config listening on port 2222; used by the backup-daemon's Ansible playbooks for remote backup operations.
* `deployments/charts/cassandra_4.x.x/` - Helm chart producing Kubernetes ConfigMaps for Cassandra 4.1.x configuration.
  * `Chart.yaml` - Chart metadata (name: `cassandra`, version: `0.1.0`).
  * `values.yaml` - Default values: TLS disabled, audit disabled, commitlog archiving disabled, empty override blocks.
  * `templates/audit.yaml` - Produces the `cassandra-audit` ConfigMap with ecAudit config when `auditLogEnabled` is `true`.
  * `templates/cassandra_default_configuration.yaml` - Produces the `cassandra-configuration` ConfigMap with the base `cassandra.yaml`.
  * `templates/cassandra_env.yaml` - Produces the `cassandra-env` ConfigMap with `cassandra-env.sh` (heap, GC, JMX, Exporter agent).
  * `templates/cassandra_version.yaml` - Produces the `cassandra-major-version` ConfigMap with `majorVersion: "4"` read by the backup-daemon.
  * `templates/commitlog_archiving.yaml` - Produces the commitlog archiving ConfigMap when `commitlogArchiving.enabled` is `true`.
  * `templates/jvm_options.yaml` - Produces the `cassandra-jvm` ConfigMap with `jvm-server.options` content.
  * `templates/logback.yaml` - Produces the `cassandra-logback` ConfigMap with Logback XML; adds an audit appender when `auditLogEnabled` is `true`.
* `deployments/charts/cassandra_5.x.x/` - Helm chart with identical template structure as `cassandra_4.x.x` but for Cassandra 5.0.x nodes.
* `deployments/charts/values.schema.json` - JSON Schema (Draft-07) validating `values.yaml` for both Cassandra charts.
* `docker-transfer/` - Artifact transfer images for shipping Helm charts.
  * `cassandra4/Dockerfile` - `FROM scratch` image copying the `cassandra_4.x.x` chart for use in the operator image build.
  * `cassandra5/Dockerfile` - `FROM scratch` image copying the `cassandra_5.x.x` chart for use in the operator image build.

### cassandra/reaper

* `Dockerfile` - Extends `thelastpickle/cassandra-reaper:4.2.3` with `openssl`; sets entrypoint to `run.sh`; runs as UID 999.
* `run.sh` - Container entrypoint; generates TLS keystores from mounted secrets when SSL is enabled; copies Reaper YAML config to the config volume; delegates to the upstream Reaper entrypoint.
* `README.md` - Component README describing the Reaper image and Evergreen maintenance strategy.

---

## operator

* `main.go` - Entrypoint; initializes the controller-runtime Manager, registers `CassandraReconciler`, and binds health/readiness probes and metrics endpoints.
* `Dockerfile` - Multi-stage build producing a minimal `alpine:3.24` runtime image with the operator binary and Helm chart.
* `Makefile` - Targets for generating CRD/DeepCopy code, running tests, building/pushing the Docker image, and installing CRDs.
* `go.mod` - Go module definition; key deps: `controller-runtime`, `qubership-nosqldb-operator-core`, `qubership-cql-driver`, `gocql`, `gofiber`.
* `go.sum` - Go module checksum file; auto-generated.
* `LICENSE.txt` - Apache 2.0 license for the operator component.
* `CODE-OF-CONDUCT.md` - Community code of conduct for operator contributors.
* `CONTRIBUTING.md` - Contribution guidelines for the operator component.
* `sf-class2-root.crt` - Starfield Class 2 CA root certificate used for TLS trust when connecting to AWS services.
* `api/v1alpha1/` - CRD type definitions.
  * `cassandra_types.go` - Defines `CassandraDeployment`, `CassandraSpec`, `DataCenter`, `Reaper`, `TLS`, and related types.
  * `groupversion_info.go` - Registers API group `netcracker.com` / version `v1alpha1`.
  * `zz_generated.deepcopy.go` - Auto-generated `DeepCopy` methods; regenerated by `make generate`.
* `bin/` - Build tooling binaries.
  * `controller-gen` - Binary used by the Makefile to generate CRD YAML and DeepCopy methods from Go annotations.
* `build/bin/` - Container entrypoint scripts.
  * `entrypoint` - Executes the operator binary with all passed arguments.
  * `user_setup` - Sets home directory permissions for arbitrary UID / OpenShift compatibility.
* `charts/helm/cassandra-operator/` - Helm chart for deploying the operator and its dependencies.
  * `Chart.yaml` - Chart metadata.
  * `values.yaml` - Default deployment values covering operator, Cassandra, Reaper, backup daemon, dbaas, TLS, and monitoring.
  * `values.schema.json` - JSON Schema validating chart values.
  * `crds/crd.yaml` - Generated CRD manifest for `CassandraDeployment`; installed by Helm pre-install hook.
  * `templates/` - Helm templates for operator Deployment, CR, RBAC, Secrets, TLS certificates, hooks, and test config.
  * `tests/` - Helm unit test files and value fixtures.
* `config/` - Raw Kubernetes manifests for `kustomize`-based installation.
  * `crd/` - CRD kustomization and webhook/CA-injection patches.
  * `rbac/` - `ClusterRole` definitions for manager, editor, and viewer access on `CassandraDeployment` and `CassandraService`.
  * `samples/` - Sample CR manifests for `CassandraDeployment` and `CassandraService`.
* `controllers/` - Controller reconciliation logic.
  * `cassandra_controller.go` - `CassandraReconciler` wrapping `qubership-nosqldb-operator-core`; registers and wires the controller for `CassandraDeployment`.
* `docker-transfer/` - Artifact transfer image.
  * `Dockerfile` - `FROM scratch` image copying the operator Helm chart for use in CI builds.
* `hack/` - Code generation tooling.
  * `boilerplate.go.txt` - Apache 2.0 license header template for `controller-gen` generated files.
* `migration-artifacts/` - Upgrade helpers.
  * `migration-artifacts.sh` - Removes `ownerReference` links from existing resources to allow safe Helm release replacement during upgrades.
* `pkg/` - Core operator implementation.
  * `impl/cassandra/` - Reconciliation steps: StatefulSets, Services, ConfigMaps, Reaper, scaling, node rebuild, credential rotation, user management.
  * `impl/common/` - Shared steps: initial validation, seed resolution, service user creation, internal Fiber server startup.
  * `impl/fiber/` - Internal HTTP server (GoFiber) for credential update callbacks and health endpoints.
  * `impl/utils/` - Utilities: CQL client setup, SSH helpers, metrics, constants, and mocks.
* `tests/` - Operator unit and integration tests.
  * `module_test.go` - Tests covering multi-DC deployment, node removal, ConfigMap updates, Service labels, seed lists, and credentials.
  * `objects_generator.go` - Builds Kubernetes Secret, PersistentVolume, and ConfigMap fixtures for tests.
  * `spec_builder.go` - Builds a default `CassandraDeployment` with typical test settings.

---

## services

### services/backup-daemon

* `Dockerfile` - Base: `qubership-backup-daemon-go-debian`; installs Python3, Ansible, `boto3`, `cassandra-driver`; copies JRE; downloads Cassandra 4.1.9 binary.
* `main.py` - CLI entrypoint; dispatches `backup`, `restore`, `aws-restore`, or `list-dbs` subcommands using credentials from Kubernetes secret paths.
* `backup-daemon.conf` - HOCON config for the Go backup-daemon sidecar; defines command templates for backup/restore/list.
* `README.md` - Component README for the backup-daemon.
* `.gitignore` - Git ignore rules for the backup-daemon directory.
* `config/` - SSH client configuration.
  * `ssh_config` - SSH client config for Ansible connections to Cassandra pods on port 2222.
* `files/` - Runtime files added to the container.
  * `run.sh` - Container startup script; handles TLS setup and executes the Go backup-daemon binary.
  * `ansible.cfg` - Ansible configuration tuned for containerized operation (`remote_tmp`, `transfer_method: piped`).
  * `hosts_template` - Ansible inventory template with a `[cassandra]` group and SSH connection variables.
  * `playbooks/backup.yaml` - Ansible playbook that performs Cassandra snapshot backup on remote pods over SSH.
* `src/` - Core Python backup/restore logic.
  * `aws_restore.py` - Downloads backup files from S3 and restores them to Cassandra pods.
  * `backup_and_restore.py` - Core backup and restore logic; runs Ansible playbooks and manages local backup storage.
  * `backup_remote.py` - Coordinates snapshot creation and file transfer from Cassandra pods.
  * `cassandra_client.py` - Python Cassandra client wrapper for pre/post-backup CQL operations.
  * `os_utils.py` - OS utility functions: file system operations and subprocess execution helpers.
* `tests/` - Unit tests for the backup-daemon Python code.
  * `backups_generator.py` - Generates backup fixture data for tests.
  * `cassandra_client_test.py` - Unit tests for `cassandra_client.py`.
  * `mocks.py` - Mock objects and patches shared across tests.
  * `restore_test.py` - Unit tests for the restore logic.

---

### services/dbaas-adapter

* `Dockerfile` - Multi-stage build producing a minimal `alpine:3.24` runtime image with the `dbaas-cassandra-adapter` binary.
* `main.go` - HTTP server entrypoint (GoFiber); connects to Cassandra; registers with DBaaS aggregator; exposes the adapter REST API.
* `go.mod` - Go module definition; key deps: `qubership-dbaas-adapter-core`, `gocql`, `gofiber`, `zap`.
* `go.sum` - Go module checksum file; auto-generated.
* `README.md` - Component README for the dbaas-adapter.
* `.gitignore` - Git ignore rules for the dbaas-adapter directory.
* `renovate.json` - Renovate bot configuration for automated dependency updates.
* `build/bin/` - Container entrypoint scripts (same pattern as operator).
  * `entrypoint` - Executes the `dbaas-cassandra-adapter` binary.
  * `user_setup` - Sets home directory permissions for arbitrary UID / OpenShift compatibility.
* `impl/` - Adapter implementation.
  * `db_admin.go` - Implements `CreateDatabase`, `CreateUser`, `DropResources`, `GetDatabases`, `GetMetadata`, and the settings update HTTP handler.
  * `db_admin_test.go` - Unit tests for database and user lifecycle operations.
  * `cassandra/` - Cassandra session and CQL helpers, type definitions, and testability mocks.
* `module_test.go` - Top-level integration tests for the dbaas-adapter module.
* `utils/` - Shared utilities.
  * `const.go` - Constants: `MetadataKey`, feature flags, resource kinds, `DefaultPort`.
  * `utils.go` - String helpers, error formatting, and env var reading utilities.

---

### services/service

* `Dockerfile` - Multi-stage build producing an `alpine:3.24` runtime image with the `cassandra-services` binary and Helm charts.
* `main.go` - Entrypoint for the CassandraSupplementary operator; registers `CassandraSupplServiceReconciler`.
* `go.mod` - Go module definition for the cassandra-services operator.
* `go.sum` - Go module checksum file; auto-generated.
* `Makefile` - Targets for generating CRDs, running tests, and building/pushing the Docker image.
* `PROJECT` - Kubebuilder project metadata describing the project layout and API groups.
* `README.md` - Component README for the cassandra-services operator.
* `.gitignore` - Git ignore rules for the service directory.
* `api/v1alpha1/` - CRD type definitions for `CassandraService`.
  * `cassandraservices_types.go` - Defines `CassandraSupplService` spec fields for backup daemon, dbaas adapter, monitoring agent, reaper, and robot tests.
  * `groupversion_info.go` - Registers the API group and version for the `CassandraService` CRD.
  * `zz_generated.deepcopy.go` - Auto-generated `DeepCopy` methods; do not edit manually.
* `bin/` - Build tooling binaries.
  * `controller-gen` - Binary for generating CRD YAML and DeepCopy methods.
* `build/bin/` - Container entrypoint scripts.
  * `entrypoint` - Executes the `cassandra-services` binary.
  * `user_setup` - Sets home directory permissions for arbitrary UID / OpenShift compatibility.
* `charts/helm/cassandra-services/` - Helm chart for deploying all supplementary services.
  * `Chart.yaml` - Chart metadata.
  * `values.yaml` - Default values for backup daemon, dbaas adapter, monitoring agent, reaper, robot tests, TLS, and Consul.
  * `values.schema.json` - JSON Schema validating chart values.
  * `monitoring/` - Grafana dashboard JSON for Cassandra metrics visualization.
  * `templates/` - Helm templates for CR, operator Deployment, RBAC, Services, Secrets, TLS certificates, PodDisruptionBudgets, Prometheus alerts, ServiceMonitor, and Grafana dashboard ConfigMap.
  * `tests/` - Helm unit test files and value fixtures.
* `config/` - Raw Kubernetes manifests for `kustomize`-based installation.
  * `crd/` - CRD kustomization and webhook/CA-injection patches.
  * `rbac/` - `ClusterRole` definitions for manager, editor, viewer, leader election, and auth proxy access.
  * `samples/` - Sample `CassandraService` CR manifest.
  * `default/`, `manager/`, `manifests/`, `prometheus/`, `scorecard/` - Kustomize overlays for default deployment, OLM bundle, Prometheus monitoring, and operator scorecard testing.
* `controllers/` - Controller reconciliation logic.
  * `cassandraservices_controller.go` - Reconciles backup-daemon, dbaas-adapter, monitoring agent, reaper, and robot test deployments.
* `docker-transfer/` - Artifact transfer image.
  * `Dockerfile` - `FROM scratch` image copying the cassandra-services Helm chart for use in CI builds.
* `hack/` - Code generation tooling.
  * `boilerplate.go.txt` - Apache 2.0 license header template for `controller-gen` generated files.
* `module_test.go` - Integration tests for the cassandra-services operator module.
* `pkg/` - Core operator implementation.
  * `service.go` - Top-level builder wiring backup, dbaas, robotTests, and common sub-packages.
  * `backup/` - Reconciliation steps for backup-daemon Deployment, Service, ConfigMaps, and SSH key generation.
  * `common/` - Shared step for creating Cassandra users for supplementary services.
  * `dbaas/` - Reconciliation steps for dbaas-adapter Deployment and Service.
  * `robotTests/` - Reconciliation steps for the robot framework test pod.
  * `utils/` - Constants, credential manager, SSH key generation, and general helpers.

---

### services/test

* `Dockerfile` - Base: `qubership-docker-integration-tests-debian`; installs `python3-dev`, `libffi-dev`; creates `cassandra` user (UID 999); installs `robotframework-requests` and `cassandra-driver`.
* `requirements.txt` - Python dependencies: `robotframework-requests`, `cassandra-driver`.
* `README.md` - Component README for the robot framework integration tests.
* `.gitignore` - Git ignore rules for the test directory.
* `alpine-repositories` - Alpine package repository list used during the image build.
* `robot/tests/` - Robot Framework test suites.
  * `alerts/alerts.robot` - Tests validating Prometheus alert rules fire correctly under failure conditions.
  * `backup/backup.robot` - Tests for backup and restore operations via the backup-daemon.
  * `crud/crud.robot` - Basic Cassandra CRUD operation tests.
  * `dbaas/dbaas.robot` - Tests for the dbaas-adapter API (database creation, user management, metadata).
  * `dbaas/dbaas_shared.robot` - Shared keywords and variables reused across dbaas test suites.
  * `ha/ha.robot` - High availability scenario tests (node failure, rolling restart, quorum).
  * `image_tests/image_tests.robot` - Validates the Cassandra node image (version, ports, user setup).
  * `tls/tls.robot` - TLS connectivity tests (encrypted CQL, certificate validation).
  * `shared/keywords.robot` - Shared Robot Framework keywords reused across all test suites.
  * `lib/CassandraLibrary.py` - Custom Robot Framework library implementing Cassandra-specific keywords using `cassandra-driver`.

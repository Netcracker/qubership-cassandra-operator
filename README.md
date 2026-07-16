[[_TOC_]]

# Cassandra Operator

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

| File | Description |
|---|---|
| `assets-config.yml` | Configures documentation archive assets for the release workflow; archives the `docs` directory as a tar file to be uploaded as a release asset. |
| `build-config.cfg` | JSON listing all Docker image components (operator, backup-daemon, dbaas-adapter, tests, Cassandra images, Reaper) with their Dockerfile paths, build contexts, build args, and changeset paths for selective builds. |
| `dependabot.yml` | Configures weekly Dependabot updates for Go modules and Docker base images across operator, dbaas-adapter, and service directories, grouped under `monorepo-dependencies`. |
| `docker-build-config.json` | Subset of `build-config.cfg` listing only the two operator images used exclusively by the release pipeline. |
| `helm-charts-release-config.yaml` | Defines the operator Helm chart location and the list of images whose tags are stamped into chart values at release time. |
| `release-drafter-config.yml` | Configures automated GitHub Release notes with categories (Breaking Changes, Features, Bug Fixes, etc.) and version label resolution for major/minor/patch. |
| `scripts/matrix.sh` | Bash script that filters build components by comparing changeset paths against changed files; returns all components if nothing matches; used by `build.yaml` for smart selective builds. |
| `workflows/build.yaml` | Main CI workflow; runs on push/PR/release; determines changed components via `matrix.sh` then builds and pushes multi-arch Docker images to `ghcr.io/netcracker`. |
| `workflows/clean.yaml` | Triggered on branch deletion; removes all container image versions from `ghcr.io` tagged with the deleted branch name. |
| `workflows/helm-charts-release.yaml` | Manually triggered release workflow; validates config, updates image tags in chart values, builds and pushes release images, runs chart-releaser, runs release-drafter, uploads chart assets. |
| `workflows/license.yaml` | Triggered on push to `main`; runs `google/addlicense` over all `.go`/`.sh`/`.py` files and opens a PR if Apache 2.0 license headers are missing. |

---

## cassandra

### cassandra/cassandra-image

| File | Description |
|---|---|
| `Dockerfile` | Builds the Cassandra node image from `eclipse-temurin:11-jre`; installs Python3, `cassandra-driver`, `cqlsh`, downloads the Cassandra binary tarball, SJK and Cassandra Exporter agent JARs, and configures the `cassandra` user (UID 999). |
| `pip.conf` | PyPI configuration copied into the image at build time; sets `index-url` to `pypi.org` and disables the pip cache. |
| `.gitignore` | Git ignore rules scoped to the `cassandra-image` subdirectory. |
| `CODE-OF-CONDUCT.md` | Community code of conduct for contributors to the cassandra-image component. |
| `CONTRIBUTING.md` | Contribution guidelines covering branching strategy and review process for cassandra-image. |
| `README.md` | Component README for cassandra-image; describes structure and Evergreen maintenance strategy. |
| `files/run.sh` | Container startup script; applies operator-injected config files from `CASSANDRA_INIT_CONFIG_DIR`, resolves seeds, configures `GossipingPropertyFileSnitch`, generates TLS keystores via `openssl pkcs12`, starts `sshd` on port 2222, then starts `cassandra -f`. |
| `files/sshd_config` | SSH daemon config for the Cassandra container; listens on port 2222; used by the backup-daemon's Ansible playbooks for remote backup and restore operations. |
| `files/.gitkeep` | Empty placeholder to keep the `files/` directory tracked in git. |

#### cassandra/cassandra-image/deployments/charts/cassandra_4.1.11

| File | Description |
|---|---|
| `Chart.yaml` | Helm chart metadata for the Cassandra 4.1.x configuration chart (name: `cassandra`, version: `0.1.0`). |
| `values.yaml` | Default values: TLS disabled, audit disabled, commitlog archiving disabled, empty `configuration`/`jvm_options`/`cassandra_env` override blocks, `PART_OF` and `MANAGED_BY` labels. |
| `templates/audit.yaml` | Produces the `cassandra-audit` ConfigMap; contains full ecAudit YAML config when `auditLogEnabled` is `true`, empty ConfigMap otherwise. |
| `templates/cassandra_default_configuration.yaml` | Produces the `cassandra-configuration` ConfigMap with a near-complete `cassandra.yaml`; sets cluster name, auth, partitioner, data directories, and TLS encryption options. |
| `templates/cassandra_env.yaml` | Produces the `cassandra-env` ConfigMap containing `cassandra-env.sh`; configures heap, GC, JMX on port 7199, and attaches the Cassandra Exporter agent JAR. |
| `templates/cassandra_version.yaml` | Produces the `cassandra-major-version` ConfigMap with `majorVersion: "4"`; read by the backup-daemon to select the correct Cassandra binary directory. |
| `templates/commitlog_archiving.yaml` | Produces the `cassandra-commitlog-archiving` ConfigMap with archive/restore commands; only rendered when `commitlogArchiving.enabled` is `true`. |
| `templates/jvm_options.yaml` | Produces the `cassandra-jvm` ConfigMap containing `jvm-server.options`; sets stack size, NUMA/TLAB tuning, OOM heap dump, and IPv4 stack preference. |
| `templates/logback.yaml` | Produces the `cassandra-logback` ConfigMap with Logback XML; configures `SYSTEMLOG`, `DEBUGLOG`, and `STDOUT` appenders; adds a dedicated `AUDIT` rolling appender when `auditLogEnabled` is `true`. |

#### cassandra/cassandra-image/deployments/charts/cassandra_5.0.8

Same template structure as `cassandra_4.1.11`. All seven templates serve the same purpose for Cassandra 5.x nodes.

| File | Description |
|---|---|
| `Chart.yaml` | Helm chart metadata for the Cassandra 5.0.x configuration chart. |
| `values.yaml` | Default values for the Cassandra 5.0.8 chart; identical structure to `cassandra_4.1.11/values.yaml`. |
| `templates/audit.yaml` | Produces the `cassandra-audit` ConfigMap for Cassandra 5.x nodes. |
| `templates/cassandra_default_configuration.yaml` | Produces the `cassandra-configuration` ConfigMap with Cassandra 5.x-appropriate `cassandra.yaml` settings. |
| `templates/cassandra_env.yaml` | Produces the `cassandra-env` ConfigMap for Cassandra 5.x. |
| `templates/cassandra_version.yaml` | Produces the `cassandra-major-version` ConfigMap; read by the backup-daemon to select the Cassandra 5.x binary directory. |
| `templates/commitlog_archiving.yaml` | Produces the commitlog archiving ConfigMap for Cassandra 5.x. |
| `templates/jvm_options.yaml` | Produces the `cassandra-jvm` ConfigMap for Cassandra 5.x. |
| `templates/logback.yaml` | Produces the `cassandra-logback` ConfigMap with Logback XML for Cassandra 5.x. |

#### cassandra/cassandra-image/deployments/charts

| File | Description |
|---|---|
| `values.schema.json` | JSON Schema (Draft-07) validating `values.yaml` inputs for both Cassandra charts; covers TLS, cassandra config, backup daemon, dbaas, monitoring agent, reaper, and robot tests sections. |

#### cassandra/cassandra-image/docker-transfer

| File | Description |
|---|---|
| `cassandra4/Dockerfile` | `FROM scratch` image that copies the `cassandra_4.1.11` chart and `values.schema.json` into `/deployments/charts`; artifact container for shipping chart files to the operator image build. |
| `cassandra5/Dockerfile` | `FROM scratch` image that copies the `cassandra_5.0.8` chart into `/deployments/charts`; artifact container for shipping chart files to the operator image build. |

---

### cassandra/reaper

| File | Description |
|---|---|
| `Dockerfile` | Extends `thelastpickle/cassandra-reaper:4.2.3` with `openssl`; fixes file ownership to UID 999; sets entrypoint to `run.sh`; runs the Cassandra Reaper repair service. |
| `run.sh` | Container entrypoint; generates JKS truststore and PKCS12 keystore from mounted TLS secrets when SSL is enabled; copies Reaper YAML config to the config volume; delegates to the upstream Reaper entrypoint. |
| `README.md` | Component README for the Reaper image; describes the Dockerfile and `run.sh` and the Evergreen maintenance strategy. |

---

## operator

| File | Description |
|---|---|
| `main.go` | Entrypoint of the Cassandra Operator; initializes the controller-runtime Manager, registers `CassandraReconciler` for the `CassandraDeployment` CRD, and binds health/readiness probes and metrics endpoints. |
| `Dockerfile` | Multi-stage build: compiles the operator binary with `golang:1.26.5-alpine`, produces a minimal `alpine:3.24` runtime image with the binary, entrypoint scripts, and Helm chart. |
| `Makefile` | Build targets: generate CRD YAML and DeepCopy methods via `controller-gen`, run tests, build/push Docker image, install/uninstall CRDs via kustomize, deploy/undeploy operator. |
| `go.mod` | Go module definition; module `github.com/Netcracker/qubership-cassandra-operator`; key deps: `controller-runtime`, `qubership-nosqldb-operator-core`, `qubership-cql-driver`, `qubership-credential-manager`, `gocql`, `gofiber`. |
| `go.sum` | Go module checksum file; auto-generated, do not edit manually. |
| `LICENSE.txt` | Apache 2.0 license text for the operator component. |
| `CODE-OF-CONDUCT.md` | Community code of conduct for operator contributors. |
| `CONTRIBUTING.md` | Contribution guidelines for the operator component. |
| `sf-class2-root.crt` | Starfield Class 2 CA root certificate (PEM); used for TLS trust when connecting to AWS services from the operator. |

### operator/api/v1alpha1

| File | Description |
|---|---|
| `cassandra_types.go` | Defines the `CassandraDeployment` CRD types: `CassandraDeployment`, `CassandraSpec`, `DataCenter`, `DeploymentSchema`, `Reaper`, `TLS`, `Policies`, `CommitlogArchiving`. |
| `groupversion_info.go` | Registers the API group `netcracker.com` and version `v1alpha1`; exports `GroupVersion`, `SchemeBuilder`, `AddToScheme`. |
| `zz_generated.deepcopy.go` | Auto-generated `DeepCopy` methods for all CRD types; regenerated by `make generate`, do not edit manually. |

### operator/bin

| File | Description |
|---|---|
| `controller-gen` | The `controller-gen` binary downloaded by the Makefile; used to generate CRD YAML and DeepCopy methods from Go type annotations. |

### operator/build/bin

| File | Description |
|---|---|
| `entrypoint` | Shell script that `exec`s the operator binary with all passed arguments; serves as the container `ENTRYPOINT`. |
| `user_setup` | Creates the home directory and sets permissions for arbitrary UID compatibility (OpenShift); deletes itself after running. |

### operator/charts/helm/cassandra-operator

| File | Description |
|---|---|
| `Chart.yaml` | Helm chart metadata for the `cassandra-operator` deployment chart. |
| `values.yaml` | Default values for the operator Helm chart; covers cassandra, reaper, backup daemon, dbaas adapter, TLS, monitoring, and robot tests configuration. |
| `values.schema.json` | JSON Schema validating operator chart `values.yaml` inputs. |
| `deployment-configuration.json` | Deployment configuration metadata describing chart capabilities for the deployment system. |
| `crds/crd.yaml` | Generated CRD manifest for `CassandraDeployment`; produced by `make manifests` via `controller-gen`; installed by Helm pre-install hook. |
| `templates/_helpers.tpl` | Helm template helper functions (name, labels, selectors) shared across all templates in this chart. |
| `templates/cr_and_operator.yaml` | Renders the `CassandraDeployment` CR and the operator `Deployment` resource. |
| `templates/cassandra-tests-config.yaml` | `ConfigMap` for robot framework test configuration; rendered when `robotTests.install` is `true`. |
| `templates/role.yaml` | `ClusterRole` granting the operator permissions to manage `CassandraDeployment` resources. |
| `templates/role_binding.yaml` | `ClusterRoleBinding` associating the operator `ServiceAccount` with the manager `ClusterRole`. |
| `templates/service_account.yaml` | `ServiceAccount` for the cassandra-operator pod. |
| `templates/configmaps/cassandra-reaper.yaml` | `ConfigMap` containing `cassandra-reaper.yml` configuration; mounted into the Reaper container. |
| `templates/hooks/creds-hook.yaml` | Helm pre-install/pre-upgrade hook `Job` that handles credential initialization before the main deployment. |
| `templates/hooks/role.yaml` | `Role` granting the hook Job the permissions needed during install/upgrade. |
| `templates/hooks/role_binding.yaml` | `RoleBinding` for the hook Job's `ServiceAccount`. |
| `templates/hooks/serviceaccount.yaml` | `ServiceAccount` used by the hook Job. |
| `templates/secrets/cassandra_secret.yaml` | `Secret` containing Cassandra admin credentials (username/password). |
| `templates/secrets/reaper-truststore_secret.yaml` | `Secret` containing the Reaper JKS truststore; rendered when TLS is enabled. |
| `templates/secrets/reaper_webui_credentials.yaml` | `Secret` containing Reaper web UI username and password. |
| `templates/tls/cassandra-tls-certificate.yaml` | cert-manager `Certificate` resource for Cassandra node TLS; rendered when `tls.generateCerts.enabled` is `true`. |
| `templates/tls/root-ca.yaml` | `Secret` or `Certificate` resource for the TLS root CA; rendered when `tls.generateCerts.enabled` is `true`. |
| `tests/cloud-passport_test.yaml` | Helm unit test for cloud passport / deployment metadata rendering. |
| `tests/envtest.yaml` | Helm unit test validating environment variable rendering in the operator `Deployment`. |
| `tests/ha_scheme_test.yaml` | Helm unit test verifying multi-DC HA scheme values are rendered correctly. |
| `tests/jvmtest.yaml` | Helm unit test for JVM options `ConfigMap` rendering. |
| `tests/multi_kuber_test.yaml` | Helm unit test for multi-Kubernetes cluster deployment rendering. |
| `tests/multi_storage_test.yaml` | Helm unit test for multiple storage volume configurations. |
| `tests/non_default_test.yaml` | Helm unit test exercising non-default chart values. |
| `tests/values/ha_values.yaml` | Test values fixture for HA scheme tests. |
| `tests/values/jvm_values.yaml` | Test values fixture for JVM options tests. |
| `tests/values/jvm_values_ipv6.yaml` | Test values fixture for JVM options tests with IPv6 enabled. |
| `tests/values/multi_kuber_dc2_values.yaml` | Test values fixture for the second DC in multi-Kubernetes cluster tests. |
| `tests/values/multi_kuber_values.yaml` | Test values fixture for multi-Kubernetes cluster tests. |
| `tests/values/non_default_values.yaml` | Test values fixture with non-default settings. |
| `tests/values/separate_commitlog_values.yaml` | Test values fixture for separate commit log storage configuration. |
| `tests/values/values.yaml` | Default test values fixture used across multiple Helm unit tests. |

### operator/config/crd

| File | Description |
|---|---|
| `kustomization.yaml` | Kustomize config listing CRD base resources; includes commented-out webhook and cert-manager patches. |
| `kustomizeconfig.yaml` | Kustomize field replacement config for CRD webhook and CA injection annotations. |
| `patches/cainjection_in_netcracker.com_cassandras.yaml` | Kustomize patch to inject CA bundle annotation for cert-manager on the `CassandraDeployment` CRD. |
| `patches/cainjection_in_netcracker.com_cassandraservices.yaml` | Kustomize patch to inject CA bundle annotation for cert-manager on the `CassandraService` CRD. |
| `patches/webhook_in_netcracker.com_cassandras.yaml` | Kustomize patch to add webhook conversion configuration to the `CassandraDeployment` CRD. |
| `patches/webhook_in_netcracker.com_cassandraservices.yaml` | Kustomize patch to add webhook conversion configuration to the `CassandraService` CRD. |

### operator/config/rbac

| File | Description |
|---|---|
| `role.yaml` | `ClusterRole manager-role`; grants get/list/watch/create/update/patch/delete on `cassandras` resources and status/finalizers subresources. |
| `netcracker.com_cassandra_editor_role.yaml` | `ClusterRole` granting full edit access to `CassandraDeployment` resources. |
| `netcracker.com_cassandra_viewer_role.yaml` | `ClusterRole` granting read-only access to `CassandraDeployment` resources. |
| `netcracker.com_cassandraservice_editor_role.yaml` | `ClusterRole` granting full edit access to `CassandraService` resources. |
| `netcracker.com_cassandraservice_viewer_role.yaml` | `ClusterRole` granting read-only access to `CassandraService` resources. |

### operator/config/samples

| File | Description |
|---|---|
| `kustomization.yaml` | Kustomize manifest listing the sample CR files. |
| `netcracker.com_v1alpha1_cassandra.yaml` | Sample `CassandraDeployment` CR manifest with standard kubebuilder labels and an empty spec. |
| `netcracker.com_v1_cassandraservice.yaml` | Sample `CassandraService` CR manifest. |

### operator/controllers

| File | Description |
|---|---|
| `cassandra_controller.go` | `CassandraReconciler` wrapping `qubership-nosqldb-operator-core`'s `ReconcileCommonService`; registers the controller for `CassandraDeployment`; wires `CassandraBuilder`, `PreDeployCassandraBuilder`, and `CassandraInstanceReconciler`. |

### operator/docker-transfer

| File | Description |
|---|---|
| `Dockerfile` | `FROM scratch` image that copies the operator Helm chart into `/charts`; used as an artifact container for shipping chart files in CI builds. |

### operator/hack

| File | Description |
|---|---|
| `boilerplate.go.txt` | Apache 2.0 license header template used by `controller-gen` when generating new Go files. |

### operator/migration-artifacts

| File | Description |
|---|---|
| `migration-artifacts.sh` | Migration script for upgrading from older operator versions; removes `ownerReference` links from `Deployments`/`Services`/`StatefulSets` so they survive Helm release replacement. |

### operator/pkg/impl

| File | Description |
|---|---|
| `cassandra.go` | Top-level entry point for the operator's implementation package; wires together the cassandra builder and reconciler sub-packages. |
| `cassandra/cassandra.go` | Core `CassandraBuilder`; orchestrates all reconciliation steps for a `CassandraDeployment`. |
| `cassandra/configmaps_step.go` | Reconciliation step that creates or updates Cassandra configuration ConfigMaps in the target namespace. |
| `cassandra/nodetool_rebuild.go` | Reconciliation step that triggers `nodetool rebuild` on Cassandra nodes when datacenter topology changes. |
| `cassandra/reaper.go` | Reconciliation step that deploys or removes the Cassandra Reaper `Deployment` and `Service`. |
| `cassandra/scaling.go` | Handles scale-up and scale-down logic for Cassandra `StatefulSets`, including node decommissioning. |
| `cassandra/serivce_lb_step.go` | Reconciliation step that manages `LoadBalancer` Services for external Cassandra access. |
| `cassandra/service_step.go` | Reconciliation step that creates or updates headless and client `Services` for Cassandra `StatefulSets`. |
| `cassandra/statefulsets_step.go` | Reconciliation step that creates or updates Cassandra `StatefulSets` for each data center. |
| `cassandra/system_keyspaces.go` | Updates replication factor on Cassandra system keyspaces (`system_auth`, `system_distributed`, `system_traces`) after topology changes. |
| `cassandra/templates.go` | Go templates for rendering Kubernetes resource manifests used in cassandra reconciliation steps. |
| `cassandra/updateCreds.go` | Handles Cassandra credential rotation; updates the superuser password in Cassandra and the associated Kubernetes `Secret`. |
| `cassandra/user_steps.go` | Reconciliation steps that create or update Cassandra users defined in the `CassandraDeployment` spec. |
| `common/add_services_users_step.go` | Shared reconciliation step that adds service-level Cassandra users used across operator and service controllers. |
| `common/initial_validations_step.go` | Validates the `CassandraDeployment` spec before reconciliation begins; returns errors for invalid configurations. |
| `common/run_fiber_server.go` | Starts the internal Fiber HTTP server used for operator-to-operator communication and hook callbacks. |
| `common/seeds_step.go` | Resolves and builds the Cassandra seed list from the `CassandraDeployment` spec and existing pod DNS names. |
| `fiber/handlers.go` | HTTP route handlers for the internal Fiber server; handles credential update callbacks and health endpoints. |
| `fiber/server.go` | Initializes and starts the Fiber HTTP server with configured routes and middleware. |
| `fiber/service.go` | Service layer for the Fiber server; implements business logic called by the HTTP handlers. |
| `utils/cassandra_metric.go` | Utilities for collecting and exposing Cassandra-specific metrics via the Prometheus endpoint. |
| `utils/const.go` | Package-level constants for the operator implementation (environment variable names, default values, label keys). |
| `utils/mocks/CassandraUtils.go` | Mock implementation of the `CassandraUtils` interface for use in unit tests. |
| `utils/ssh.go` | SSH client utilities for connecting to Cassandra pods over port 2222 (used for `nodetool` operations). |
| `utils/stream.go` | Utilities for streaming command output from SSH sessions on Cassandra pods. |
| `utils/utils.go` | General utility functions: Cassandra CQL client setup, secret reading, label building, and shared helpers. |
| `utils/utils_test.go` | Unit tests for the `utils` package. |

### operator/tests

| File | Description |
|---|---|
| `module_test.go` | Integration-style unit tests using a fake Kubernetes client; covers two-DC deployment, node removal, `ConfigMap` updates, `Service` labels, seed lists, `hostNetwork`, and credential reading. |
| `objects_generator.go` | Helper builder for constructing Kubernetes `Secret`, `PersistentVolume`, and `ConfigMap` objects in tests. |
| `spec_builder.go` | `GenerateDefaultCassandra` helper that builds a `CassandraDeployment` with typical test defaults for use across test cases. |

---

## services

### services/backup-daemon

| File | Description |
|---|---|
| `Dockerfile` | Base: `qubership-backup-daemon-go-debian`; installs Python3, Ansible, `boto3`, `cassandra-driver`; copies JRE from `eclipse-temurin:11-jre`; downloads Cassandra 4.1.9 binary; sets `CMD` to `run.sh`. |
| `main.py` | Python CLI entrypoint; reads Cassandra and AWS credentials from Kubernetes secret file paths; dispatches `backup`, `restore`, `aws-restore`, or `list-dbs` subcommands. |
| `backup-daemon.conf` | HOCON config for the Go backup-daemon sidecar; defines command templates for backup/restore/list using `main.py`; references TLS environment variables. |
| `README.md` | Component README for the backup-daemon. |
| `.gitignore` | Git ignore rules for the backup-daemon directory. |
| `config/ssh_config` | SSH client config for Ansible connections to Cassandra pods: `ControlMaster auto`, port 2222, user `cassandra`, identity file from a mounted Kubernetes secret. |
| `files/.gitkeep` | Empty placeholder to keep the `files/` directory tracked in git. |
| `files/ansible.cfg` | Ansible configuration; sets `remote_tmp`, `local_tmp`, `control_path_dir` to `/tmp` and `transfer_method` to `piped` for containerized operation. |
| `files/hosts_template` | Ansible inventory template with a `[cassandra]` group header and SSH connection variables for connecting to Cassandra pods. |
| `files/run.sh` | Container startup script; optionally imports TLS root cert into a JKS truststore; copies the correct Cassandra version directory; optionally sets remote debug params; executes the Go backup-daemon binary. |
| `files/playbooks/backup.yaml` | Ansible playbook that performs Cassandra snapshot backup on remote pods over SSH. |
| `src/__init__.py` | Python package marker for the `src` module. |
| `src/aws_restore.py` | Implements AWS S3-based Cassandra restore logic; downloads backup files from S3 and restores them to Cassandra pods. |
| `src/backup_and_restore.py` | Core backup and restore logic; runs Ansible playbooks and manages local backup storage. |
| `src/backup_remote.py` | Handles the remote side of backup operations; coordinates snapshot creation and file transfer from Cassandra pods. |
| `src/cassandra_client.py` | Python Cassandra client wrapper using `cassandra-driver`; used for pre/post-backup CQL operations. |
| `src/os_utils.py` | OS utility functions: file system operations, path handling, and subprocess execution helpers. |
| `tests/__init__.py` | Python package marker for the `tests` module. |
| `tests/backups_generator.py` | Test data generator that creates backup fixture data for unit tests. |
| `tests/cassandra_client_test.py` | Unit tests for `cassandra_client.py`. |
| `tests/mocks.py` | Mock objects and patches used across backup-daemon unit tests. |
| `tests/restore_test.py` | Unit tests for the restore logic in `backup_and_restore.py`. |

---

### services/dbaas-adapter

| File | Description |
|---|---|
| `Dockerfile` | Multi-stage build: compiles `dbaas-cassandra-adapter` binary with `golang:1.26.5-alpine`; produces minimal `alpine:3.24` runtime image with `curl`. |
| `main.go` | HTTP server entrypoint using GoFiber; reads config from env vars and Kubernetes secret paths; connects to Cassandra via gocql; registers with the DBaaS aggregator; exposes the adapter REST API. |
| `go.mod` | Go module definition; key deps: `qubership-dbaas-adapter-core`, `gocql`, `gofiber`, `zap`. |
| `go.sum` | Go module checksum file; auto-generated. |
| `README.md` | Component README for the dbaas-adapter. |
| `.gitignore` | Git ignore rules for the dbaas-adapter directory. |
| `renovate.json` | Renovate bot configuration for automated dependency updates in the dbaas-adapter module. |
| `build/bin/entrypoint` | Shell script that `exec`s the `dbaas-cassandra-adapter` binary; serves as the container `ENTRYPOINT`. |
| `build/bin/user_setup` | Creates home directory and sets permissions for arbitrary UID / OpenShift compatibility; deletes itself after running. |
| `go/bin/dlv` | Delve debugger binary used for remote debugging of the dbaas-adapter in development. |
| `impl/db_admin.go` | `CassandraDbAdministration` implementation; handles `CreateDatabase`, `CreateUser`, `DropResources`, `GetDatabases`, `GetMetadata`, `UpdateMetadata`, and the settings update HTTP handler. |
| `impl/db_admin_test.go` | Unit tests for `db_admin.go` covering database and user lifecycle operations. |
| `impl/cassandra/cassandraService.go` | `CassandraService` interface definition; abstracts Cassandra session operations for testability. |
| `impl/cassandra/configuration.go` | Builds gocql cluster configuration from environment variables including TLS settings. |
| `impl/cassandra/cql.go` | CQL query helpers: keyspace creation, role management, permission grants, metadata table operations. |
| `impl/cassandra/session_service.go` | Implements `CassandraService` using a real gocql session; manages connection lifecycle. |
| `impl/cassandra/types.go` | Type definitions for Cassandra-specific data structures used in the dbaas-adapter. |
| `impl/cassandra/mocks/cassandraService_mock.go` | Mock for the `CassandraService` interface used in unit tests. |
| `impl/cassandra/mocks/cluster_mock.go` | Mock for the gocql `Cluster` for testing configuration logic. |
| `impl/cassandra/mocks/configuration_mock.go` | Mock for the configuration interface used in unit tests. |
| `impl/cassandra/mocks/query_mock.go` | Mock for gocql `Query` objects used in CQL operation tests. |
| `impl/cassandra/mocks/session_mock.go` | Mock for gocql `Session` objects used in unit tests. |
| `module_test.go` | Top-level integration tests for the dbaas-adapter module. |
| `utils/const.go` | Constants: `MetadataKey`, `FeatureMultiUsers`, `FeatureTLS`, `UserResourceKind`, `DbResourceKind`, `DefaultPort`. |
| `utils/utils.go` | Shared utility functions for the dbaas-adapter: string helpers, error formatting, env var reading. |

---

### services/service

| File | Description |
|---|---|
| `Dockerfile` | Multi-stage build: compiles `cassandra-services` binary with `golang:1.26.5-alpine`; produces `alpine:3.24` runtime image with `openssl` and `curl`; copies Helm charts into the image. |
| `main.go` | Entrypoint for the CassandraSupplementary operator; initializes the controller-runtime Manager and registers `CassandraSupplServiceReconciler`. |
| `go.mod` | Go module definition for the cassandra-services operator. |
| `go.sum` | Go module checksum file; auto-generated. |
| `Makefile` | Build targets: generate CRDs, run tests, build/push Docker image, install CRDs via kustomize. |
| `PROJECT` | Kubebuilder project metadata file describing the project layout and API group versions. |
| `README.md` | Component README for the cassandra-services operator. |
| `.gitignore` | Git ignore rules for the service directory. |
| `.dockerignore` | Docker ignore rules to exclude unnecessary files from the build context. |
| `build/Dockerfile` | Alternative Dockerfile for the services operator build (legacy/alternative build path). |
| `build/bin/entrypoint` | Shell script that `exec`s the `cassandra-services` binary; serves as the container `ENTRYPOINT`. |
| `build/bin/user_setup` | Creates home directory and sets permissions for arbitrary UID / OpenShift compatibility. |
| `bin/controller-gen` | The `controller-gen` binary used by the Makefile to generate CRD YAML and DeepCopy methods. |
| `hack/boilerplate.go.txt` | Apache 2.0 license header template for `controller-gen` generated files. |
| `api/v1alpha1/cassandraservices_types.go` | Defines the `CassandraSupplService` CRD types: spec fields for backup daemon, dbaas adapter, monitoring agent, reaper, and robot tests. |
| `api/v1alpha1/groupversion_info.go` | Registers the API group and version for the `CassandraService` CRD. |
| `api/v1alpha1/zz_generated.deepcopy.go` | Auto-generated `DeepCopy` methods for `CassandraService` types; do not edit manually. |
| `controllers/cassandraservices_controller.go` | `CassandraSupplServiceReconciler`; reconciles backup-daemon, dbaas-adapter, monitoring agent, reaper, and robot test deployments. |
| `module_test.go` | Integration tests for the cassandra-services operator module. |
| `docker-transfer/Dockerfile` | `FROM scratch` image that copies the cassandra-services Helm chart into `/charts` for artifact transfer. |

#### services/service/charts/helm/cassandra-services

| File | Description |
|---|---|
| `Chart.yaml` | Helm chart metadata for the `cassandra-services` deployment chart. |
| `values.yaml` | Default values for the `cassandra-services` chart; covers backup daemon, dbaas adapter, monitoring agent, reaper, robot tests, TLS, and Consul settings. |
| `values.schema.json` | JSON Schema validating `cassandra-services` chart `values.yaml` inputs. |
| `deployment-configuration.json` | Deployment configuration metadata describing chart capabilities for the deployment system. |
| `monitoring/grafana-dashboard.json` | Grafana dashboard definition for Cassandra metrics visualization. |
| `monitoring/grafana-dashboard.json.gz` | Gzip-compressed Grafana dashboard for `ConfigMap` size efficiency. |
| `templates/_helpers.tpl` | Helm template helper functions shared across `cassandra-services` templates. |
| `templates/consul-acls.yaml` | Renders Consul ACL token resources when Consul service registration is enabled. |
| `templates/cr_and_operator.yaml` | Renders the `CassandraService` CR and the `cassandra-services` operator `Deployment`. |
| `templates/dbaas-ingress-deployment.yaml` | Renders an `Ingress` resource for the dbaas-adapter when ingress is configured. |
| `templates/dbaas-physical-databases-labels.yaml` | Renders labels/annotations to register Cassandra as a physical DBaaS database with the aggregator. |
| `templates/grafana_dashboard.yaml` | `ConfigMap` wrapping the Grafana dashboard JSON for automatic Grafana discovery. |
| `templates/operator_service.yaml` | `Service` resource exposing the cassandra-services operator's internal HTTP port. |
| `templates/poddisruptionbudgets.yaml` | `PodDisruptionBudgets` for Cassandra `StatefulSet` pods to limit simultaneous disruptions during maintenance. |
| `templates/prometheus_alerts.yaml` | `PrometheusRule` defining alerting rules for Cassandra (node down, high latency, disk usage, etc.). |
| `templates/role.yaml` | `ClusterRole` granting the cassandra-services operator permissions to manage its resources. |
| `templates/role_binding.yaml` | `ClusterRoleBinding` associating the cassandra-services `ServiceAccount` with the manager role. |
| `templates/service_account.yaml` | `ServiceAccount` for the cassandra-services operator pod. |
| `templates/service_export.yaml` | `ServiceExport` resource for multi-cluster service mesh connectivity. |
| `templates/service_monitor.yaml` | Prometheus `ServiceMonitor` for scraping Cassandra metrics from the Exporter agent port (8778). |
| `templates/supplementary-test-config.yaml` | `ConfigMap` with test configuration for the robot framework tests pod. |
| `templates/secrets/aws_credentials.yaml` | `Secret` containing AWS access key and secret for S3 backup operations. |
| `templates/secrets/backup-daemon-s3-tls-secret.yaml` | `Secret` containing TLS certificates for S3-over-TLS connections from the backup-daemon. |
| `templates/secrets/backup_api_credentials.yaml` | `Secret` containing backup API username and password for the backup-daemon REST API. |
| `templates/secrets/dbaas-adapter-credentials.yaml` | `Secret` containing dbaas-adapter HTTP basic auth credentials. |
| `templates/secrets/dbaas-aggregator-credentials.yaml` | `Secret` containing DBaaS aggregator registration credentials for the dbaas-adapter. |
| `templates/secrets/dbaas_streaming_role.yaml` | `Secret` containing the streaming role credentials for DBaaS multi-user mode. |
| `templates/secrets/s3_credentials.yaml` | `Secret` containing S3 bucket name and endpoint URL for backup storage. |
| `templates/tls/backup-daemon-cetificate-secret.yaml` | `Secret` holding the TLS certificate and key for the backup-daemon. |
| `templates/tls/backup-daemon-tls-certificate.yaml` | cert-manager `Certificate` resource for the backup-daemon TLS certificate. |
| `templates/tls/dbaas-adapter-certificate-secret.yaml` | `Secret` holding the TLS certificate and key for the dbaas-adapter. |
| `templates/tls/dbaas-tls-certificate.yaml` | cert-manager `Certificate` resource for the dbaas-adapter TLS certificate. |
| `templates/tls/tls_metrics.yaml` | TLS configuration for securing Prometheus metrics endpoints. |
| `tests/alerts_test.yaml` | Helm unit test verifying Prometheus alert rules are rendered correctly. |
| `tests/ha_scheme_test.yaml` | Helm unit test for HA scheme rendering in `cassandra-services`. |
| `tests/prometheus_metricRelabelings_test.yaml` | Helm unit test for Prometheus metric relabeling configuration. |
| `tests/values/ha_values.yaml` | Test values fixture for HA scheme tests. |
| `tests/values/prometheus_values.yaml` | Test values fixture for Prometheus configuration tests. |

#### services/service/config

| File | Description |
|---|---|
| `crd/kustomization.yaml` | Kustomize config listing `CassandraService` CRD base resources. |
| `crd/kustomizeconfig.yaml` | Kustomize field replacement config for CRD webhook and CA injection annotations. |
| `crd/patches/cainjection_in_cassandraservices.yaml` | Kustomize patch to inject CA bundle annotation for cert-manager on the `CassandraService` CRD. |
| `crd/patches/webhook_in_cassandraservices.yaml` | Kustomize patch to add webhook conversion configuration to the `CassandraService` CRD. |
| `default/kustomization.yaml` | Default Kustomize overlay composing manager, RBAC, CRD, and metrics resources. |
| `default/manager_auth_proxy_patch.yaml` | Kustomize patch to inject the `kube-rbac-proxy` sidecar container for metrics auth. |
| `default/manager_config_patch.yaml` | Kustomize patch to set controller manager configuration (leader election, metrics, health probes). |
| `manager/kustomization.yaml` | Kustomize base for the controller manager `Deployment`. |
| `manager/manager.yaml` | `Deployment` manifest template for the cassandra-services controller manager. |
| `manifests/kustomization.yaml` | Kustomize config for generating OLM bundle manifests. |
| `prometheus/kustomization.yaml` | Kustomize config for Prometheus monitoring resources. |
| `prometheus/monitor.yaml` | `ServiceMonitor` for Prometheus scraping of the controller manager metrics endpoint. |
| `rbac/auth_proxy_client_clusterrole.yaml` | `ClusterRole` allowing the auth proxy to request token reviews. |
| `rbac/auth_proxy_role.yaml` | `Role` for the `kube-rbac-proxy` sidecar. |
| `rbac/auth_proxy_role_binding.yaml` | `RoleBinding` for the `kube-rbac-proxy` sidecar. |
| `rbac/auth_proxy_service.yaml` | `Service` exposing the metrics endpoint through the auth proxy. |
| `rbac/cassandraservices_editor_role.yaml` | `ClusterRole` granting full edit access to `CassandraService` resources. |
| `rbac/cassandraservices_viewer_role.yaml` | `ClusterRole` granting read-only access to `CassandraService` resources. |
| `rbac/kustomization.yaml` | Kustomize config listing all RBAC resource files. |
| `rbac/leader_election_role.yaml` | `Role` granting the controller manager permission to use leader election (ConfigMap/Lease locks). |
| `rbac/leader_election_role_binding.yaml` | `RoleBinding` for the leader election role. |
| `rbac/role_binding.yaml` | `ClusterRoleBinding` associating the cassandra-services `ServiceAccount` with the manager `ClusterRole`. |
| `rbac/service_account.yaml` | `ServiceAccount` for the cassandra-services controller manager. |
| `samples/kustomization.yaml` | Kustomize manifest listing the `CassandraService` sample CR file. |
| `samples/qubership.org_v1alpha1_cassandraservices.yaml` | Sample `CassandraService` CR manifest for testing and reference. |
| `scorecard/bases/config.yaml` | Operator SDK scorecard base configuration for conformance and OLM tests. |
| `scorecard/kustomization.yaml` | Kustomize config for scorecard resources. |
| `scorecard/patches/basic.config.yaml` | Scorecard patch for basic conformance test configuration. |
| `scorecard/patches/olm.config.yaml` | Scorecard patch for OLM-specific test configuration. |

#### services/service/pkg

| File | Description |
|---|---|
| `service.go` | Top-level service builder; wires together backup, dbaas, robotTests, and common reconciliation sub-packages. |
| `backup/backup.go` | `BackupBuilder` orchestrating the backup-daemon deployment reconciliation steps. |
| `backup/backup_service_step.go` | Reconciliation step that creates or updates the backup-daemon `Service` resource. |
| `backup/configmaps_step.go` | Reconciliation step that creates or updates backup-daemon `ConfigMaps` (Ansible config, hosts template, etc.). |
| `backup/deployment.go` | Reconciliation step that creates or updates the backup-daemon `Deployment`. |
| `backup/ssh_key_step.go` | Reconciliation step that generates an SSH key pair and stores it in a Kubernetes `Secret` for backup-daemon-to-Cassandra connectivity. |
| `backup/templates.go` | Go templates for rendering backup-daemon Kubernetes resource manifests. |
| `common/add_services_users_step.go` | Shared reconciliation step that creates Cassandra users for supplementary services. |
| `dbaas/dbaas.go` | `DBaaSBuilder` orchestrating the dbaas-adapter deployment reconciliation steps. |
| `dbaas/dbaas_service_step.go` | Reconciliation step that creates or updates the dbaas-adapter `Service`. |
| `dbaas/deployment_step.go` | Reconciliation step that creates or updates the dbaas-adapter `Deployment`. |
| `dbaas/templates.go` | Go templates for rendering dbaas-adapter Kubernetes resource manifests. |
| `robotTests/deployment_step.go` | Reconciliation step that creates or updates the robot tests `Deployment`. |
| `robotTests/robot_tests.go` | `RobotTestsBuilder` orchestrating robot framework test pod reconciliation. |
| `robotTests/templates.go` | Go templates for rendering robot tests Kubernetes resource manifests. |
| `utils/const.go` | Package-level constants for the services operator (label keys, annotation names, default values). |
| `utils/credsManager.go` | Credential manager utilities for reading and rotating Cassandra credentials stored in Kubernetes `Secrets`. |
| `utils/ssh.go` | SSH key generation utilities used by `ssh_key_step` to produce the backup-daemon keypair. |
| `utils/stream.go` | Utilities for streaming command output from operator operations. |
| `utils/utils.go` | General utility functions: label building, resource name helpers, and env var construction. |

---

### services/test

| File | Description |
|---|---|
| `Dockerfile` | Base: `qubership-docker-integration-tests-debian`; installs `gcc`, `python3-dev`, `libffi-dev`; creates `cassandra` user (UID 999); installs `robotframework-requests` and `cassandra-driver`; copies robot tests; exposes port 8080. |
| `requirements.txt` | Python dependencies for the test image: `robotframework-requests`, `cassandra-driver`. |
| `README.md` | Component README for the robot framework integration tests. |
| `.gitignore` | Git ignore rules for the test directory. |
| `alpine-repositories` | Alpine package repository configuration used during the image build. |
| `robot/tests/alerts/alerts.robot` | Robot Framework test suite validating Prometheus alert rules fire correctly under failure conditions. |
| `robot/tests/alerts/tags_exclusion.py` | Python helper for dynamically excluding Robot Framework test tags based on environment conditions. |
| `robot/tests/backup/backup.robot` | Robot Framework test suite for backup and restore operations. |
| `robot/tests/crud/crud.robot` | Robot Framework test suite for basic Cassandra CRUD operations. |
| `robot/tests/dbaas/dbaas.robot` | Robot Framework test suite for the dbaas-adapter API (database creation, user management, metadata). |
| `robot/tests/dbaas/dbaas_shared.robot` | Shared keywords and variables reused across dbaas test suites. |
| `robot/tests/ha/ha.robot` | Robot Framework test suite for high availability scenarios (node failure, rolling restart, quorum loss). |
| `robot/tests/image_tests/image_tests.robot` | Robot Framework test suite for validating the Cassandra node image (version checks, port availability, user setup). |
| `robot/tests/tls/tls.robot` | Robot Framework test suite for TLS connectivity (encrypted CQL, certificate validation, keystore tests). |
| `robot/tests/shared/keywords.robot` | Shared Robot Framework keywords (CQL helpers, wait conditions, assertion macros) reused across all test suites. |
| `robot/tests/lib/CassandraLibrary.py` | Custom Robot Framework library implementing Cassandra-specific keywords using the Python `cassandra-driver`. |

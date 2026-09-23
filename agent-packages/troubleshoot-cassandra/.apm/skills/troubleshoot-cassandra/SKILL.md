---
name: troubleshoot-cassandra

description: Diagnose and resolve failures in operator-managed CASSANDRA clusters using the documented Cassandra troubleshooting reference. Use whenever the user reports a Cassandra pod startup failure, CrashLoopBackOff, gossip/connectivity issue, commit-log corruption, direct-buffer-memory problem, logs stuck during startup, invalid GrafanaDashboard, Prometheus alert, high CPU or memory usage, high disk or inode usage, Cassandra nodes not running, fewer pods than expected, Unavailable exceptions, backup-daemon disk/inode pressure, a failed backup, or Cassandra backup failures caused by GC/heap pressure or excessive schema metadata. Match the reported symptom to the documented Cassandra troubleshooting section first. If no documented section is a plausible match, fall back to a general Cassandra/operator checklist covering pod state, operator logs, Cassandra logs, events, networking/DNS, storage, resources, configuration, nodetool status, and backup-daemon logs.

---

## Reading the reference file

1. Grep issue headers with line numbers:
   `grep -n "^\*\*## " references/troubleshooting.md`
   This identifies the documented Cassandra issue sections.

2. If nothing in Step 1's output looks like a plausible match, also inspect headings that may have been formatted differently:
   `grep -n "^## " references/troubleshooting.md`
   Do not conclude that an issue is undocumented until both heading forms have been checked.

3. Match the reported symptom, exact error text, alert, or affected component against the jump table below to select the most relevant issue.

4. Read only the matched section, from its issue header through the next issue header. Do not load the whole troubleshooting file for a single lookup.

## Symptom → reference section

| Symptom | Header in references/troubleshooting.md |
|---|---|
| Cassandra operator needs to be restarted | If operator needs to be restarted |
| App Job fails because a rendered resource already exists / invalid Helm ownership metadata | App Job has failed with Error: rendered manifests contain a resource that already exists |
| Cassandra pod fails to start with `Could not read commit log` / `CommitLogReadException` | Cassandra Pod Fails to Start. Could not read commit log. |
| Cassandra pod fails to start with `Unable to gossip with any peers` | Cassandra Pod Fails to Start. Unable to gossip with any peers |
| Cassandra pod fails to start with direct buffer memory error | Cassandra Pod Fails to Start. Direct buffer memory. |
| Cassandra startup logs are stuck at `Initializing system.IndexInfo` | Cassandra Pod Fails to Start. Logs stuck. |
| GrafanaDashboard is invalid because annotations exceed the size limit | The GrafanaDashboard is invalid |
| Cassandra pod is in CrashLoopBackOff with `Saved cluster name ... != configured name cassandra_cluster` | Cassandra Pod goes CrashLoopBack. |
| Prometheus CPU usage alert | Prometheus Alerts Troubleshooting. CPU Usage. |
| Prometheus memory usage alert | Prometheus Alerts Troubleshooting. Memory Usage. |
| Cassandra disk-space usage is higher than the configured threshold | Prometheus Alerts Troubleshooting. Cassandra Disk Space Usage is Higher Than N Percent. |
| Cassandra pod is not running / unable to gossip with peers | Prometheus Alerts Troubleshooting. Cassandra Pod is Not Running. |
| Cassandra pod count is lower than expected | Prometheus Alerts Troubleshooting. Cassandra Pods Count is Lower Than Expected |
| Cassandra JVM heap usage is above the documented threshold | Prometheus Alerts Troubleshooting. JVM Memory Heap Usage |
| Unavailable exceptions are reported over 5 minutes | Prometheus Alerts Troubleshooting. Unavailable Exceptions Count Over 5 Minutes |
| Backup daemon disk-space usage is higher than the configured threshold | Prometheus Alerts Troubleshooting. Backup Daemon Disk Space Usage is Higher Than N Percent |
| Backup daemon inode usage is higher than the configured threshold | Prometheus Alerts Troubleshooting. Backup Daemon Uses More Than N Percent of Available Inodes |
| Last backup failed | Prometheus Alerts Troubleshooting. Last Backup Failed |
| Cassandra backup fails with exit code 2, especially with long GC pauses or heap pressure | Cassandra Backup Failure (Exit Code 2) |

Start every diagnosis by identifying the exact error text, alert name, startup log message, or affected component. The reference sections generally match specific error strings or symptoms.

Check the conversation attachments and previously provided logs, operator output, events, or `nodetool status` output before asking the user to provide them again.

If the symptom plausibly matches more than one section, use the exact error text and surrounding context to distinguish them. For example, `Unable to gossip with any peers` appears in both the startup and Prometheus alert sections; determine whether the user is reporting a startup failure or an alert before selecting the section.

## Guardrails

The Cassandra operator is operator-managed. Prefer the documented configuration/deployment path rather than making undocumented live changes that can be overwritten by reconciliation.

- For operator restart, follow the documented procedure: remove the `summary-spec` value from the `last-applied-configuration-info` ConfigMap and restart the operator pod.
- For Helm ownership/resource conflicts, follow the documented cleanup and deployment procedure. Do not assume that every existing resource should be deleted without first confirming that the resource matches the documented failure.
- For commit-log corruption, deleting a single affected commit-log file is documented. Clearing the entire commit-log directory can cause partial data loss and must be treated as a higher-risk recovery action.
- The documented `cassandra.commitlog.ignorereplayerrors=true` JVM option is a temporary recovery measure: apply it through the documented ConfigMap path, restart, and revert the change afterward.
- For resource alerts, follow the documented resource/storage scaling path rather than making unrelated Cassandra configuration changes.
- For backup failures, investigate the Cassandra/backup-daemon logs and the documented backup path before changing Cassandra data or schema.
- Do not claim that a workaround is safe when the reference explicitly identifies a data-loss or recovery risk.

## Configuration conventions

Use the documented configuration path for the specific issue:

- `DEPLOY_W_HELM: "true"` is the documented setting for the invalid GrafanaDashboard case and is also relevant to the Helm resource-ownership deployment scenario.
- `operator.podName` should be removed when the documented App Job failure shows an existing `cassandra-operator` Deployment with conflicting Helm ownership metadata.
- `file_cache_size_in_mb` in the `cassandra-configuration` ConfigMap is the documented parameter to check for the direct-buffer-memory startup issue.
- `cassandra-env` is the documented ConfigMap for the temporary `JVM_OPTS="$JVM_OPTS -Dcassandra.commitlog.ignorereplayerrors=true"` recovery option.
- For a saved cluster-name mismatch, configure `cassandra.configuration` / `cluster_name` to match the saved cluster name reported by Cassandra.
- Do not invent configuration keys or replace the documented configuration path with a live `nodetool` or JVM command unless the reference explicitly calls for it.

## Cluster and diagnostic conventions

- Substitute the actual Cassandra namespace for `<namespace>`.
- Check Cassandra pod state before diagnosing a startup or availability issue.
- Use `nodetool status` from a Cassandra pod to identify nodes that are `DOWN` when investigating Unavailable exceptions.
- For gossip failures, verify that the Cassandra headless service has `clusterIP: None`.
- Check pod/service DNS resolution and network connectivity when gossip is failing.
- For startup logs stuck at `Initializing system.IndexInfo`, check both PV free space and inode availability:
  `df -h -i /var/lib/cassandra/data`
- For backup-daemon disk/inode alerts, check available disk space and remove unnecessary backups according to the documented procedure.
- For a failed backup, use the backup ID from the alert and inspect:
  `cat /backup-storage/{backup_id}/.console`
- For Cassandra backup exit code 2, check Cassandra logs for long GC pauses and heap pressure, then review schema/keyspace structure and the number of tables.

## Fallback checklist

If no documented section is a plausible match:

1. Capture the exact Cassandra error, alert, or startup message.
2. Identify the affected pod and namespace.
3. Check pod status and Kubernetes events.
4. Check Cassandra logs around the failure.
5. Check operator logs if the issue occurs during deployment or reconciliation.
6. Run `nodetool status` when the Cassandra service is available.
7. Check DNS/network connectivity for gossip or peer communication issues.
8. Check PV capacity and inode usage for storage/startup issues.
9. Check CPU, memory, and JVM heap pressure for resource-related symptoms.
10. For backup issues, inspect backup-daemon logs and the backup ID-specific `.console` file.
11. Re-check the reference file for a more specific documented symptom before proposing a workaround.

When a documented fix exists, follow the reference section rather than replacing it with a generic troubleshooting path.
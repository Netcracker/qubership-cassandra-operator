---
name: troubleshoot-cassandra

description: Diagnose and resolve failures in operator-managed Cassandra clusters using the documented troubleshooting reference. Use whenever the user reports a Cassandra pod startup failure, CrashLoopBackOff, gossip/connectivity issue, commit-log corruption, direct-buffer-memory problem, logs stuck during startup, invalid GrafanaDashboard, Prometheus alert, high CPU/memory/JVM heap/disk/inode usage, nodes DOWN or not running, fewer pods than expected, Unavailable exceptions, backup-daemon disk/inode pressure, failed backups (GC/heap pressure, excessive schema metadata), an operator not reacting to a restart, or an App Job/Helm deploy error. Match the reported symptom to the documented Cassandra troubleshooting section first. If no documented section is a plausible match, fall back to a general Cassandra/operator checklist covering pod state, operator logs, Cassandra logs, events, networking/DNS, storage, resources, configuration, nodetool status, and backup-daemon logs.

---

## Workflow

Follow these steps in order. Do not skip step 1 or step 5.

1. **Collect evidence.** Identify the exact error text, alert name, startup log line, affected pod, and namespace. Check the conversation, attachments, and earlier tool output first; ask the user only for what is still missing (`kubectl get pods -n <namespace>`, the failing pod's logs, `kubectl describe pod`, recent events, `nodetool status`).
2. **Classify the context.** Decide which one applies, because the same error string can appear in several sections:
   - *Startup failure*: pod is `CrashLoopBackOff`/`Error`/not Ready and Cassandra logs show a startup exception.
   - *Prometheus alert*: user quotes an alert name or threshold.
   - *Deployment/operator*: App Job, Helm, or operator reconciliation error.
   - *Backup*: backup-daemon alert, backup ID, or backup playbook failure.
3. **Select the section** with the lookup below and the jump table.
4. **Read only that section** and follow its "How to solve" steps. Verify each documented cause against the evidence before proposing it (see "Confirm before fixing").
5. **Report** using the output format at the end. State which section was used, which evidence supports it, and what has not been verified.

## Reading the reference file

All paths are relative to this skill directory. The file uses CRLF line endings, so strip `\r` when extracting.

1. List issue headers with line numbers (headers are plain `## ` lines):
   `grep -n "^## " references/troubleshooting.md`
   If this unexpectedly returns nothing, also try `grep -n "^\*\*## " references/troubleshooting.md` in case the heading format changed. Do not conclude that an issue is undocumented until both forms have been checked.

2. Search by distinctive error text to catch matches the jump table misses (the stack traces and descriptions are in the file):
   `grep -n -i "<key phrase>" references/troubleshooting.md`

3. Extract just the matched section, from its header to the next `## ` header:
   `tr -d '\r' < references/troubleshooting.md | awk -v h="<exact header text>" '/^## /{p = (index($0, h) > 0)} p'`
   Do not load the whole file for a single lookup. If two sections are candidates, read both.

## Symptom → reference section

| Symptom | Header in references/troubleshooting.md |
|---|---|
| Operator restarted but nothing is reconciled / operator needs a forced re-run with an unchanged CR | If operator needs to be restarted |
| App Job fails: `rendered manifests contain a resource that already exists` / `invalid ownership metadata` (ConfigMap, Ingress, ServiceAccount, Role, etc.) | App Job has failed with Error: rendered manifests contain a resource that already exists |
| Same App Job error, but the conflicting object is the `cassandra-operator` **Deployment** (`meta.helm.sh/release-name` must equal `cassandra-services`) | Same section, second stack trace: remove `operator.podName` |
| Pod fails to start with `Could not read commit log` / `CommitLogReadException` | Cassandra Pod Fails to Start. Could not read commit log. |
| Pod **fails to start** with `Unable to gossip with any peers` (startup exception in the pod log) | Cassandra Pod Fails to Start. Unable to gossip with any peers |
| Pod fails to start with a direct buffer memory error | Cassandra Pod Fails to Start. Direct buffer memory. |
| Startup logs stuck at `Initializing system.IndexInfo` | Cassandra Pod Fails to Start. Logs stuck. |
| `GrafanaDashboard "cassandra-grafana-dashboard" is invalid ... annotations: Too long` | The GrafanaDashboard is invalid |
| CrashLoopBackOff with `Saved cluster name ... != configured name cassandra_cluster` | Cassandra Pod goes CrashLoopBack. |
| CPU usage alert (>95% for some pods) | Prometheus Alerts Troubleshooting. CPU Usage. |
| Memory usage alert (>95% for some pods) | Prometheus Alerts Troubleshooting. Memory Usage. |
| Cassandra disk-space usage above threshold | Prometheus Alerts Troubleshooting. Cassandra Disk Space Usage is Higher Than N Percent. |
| **Alert** that a Cassandra pod is not running / node cannot gossip | Prometheus Alerts Troubleshooting. Cassandra Pod is Not Running. |
| Cassandra pod count lower than expected | Prometheus Alerts Troubleshooting. Cassandra Pods Count is Lower Than Expected |
| JVM heap usage above 80% | Prometheus Alerts Troubleshooting. JVM Memory Heap Usage |
| Unavailable exceptions over 5 minutes | Prometheus Alerts Troubleshooting. Unavailable Exceptions Count Over 5 Minutes |
| Backup daemon disk usage above threshold | Prometheus Alerts Troubleshooting. Backup Daemon Disk Space Usage is Higher Than N Percent |
| Backup daemon inode usage above threshold | Prometheus Alerts Troubleshooting. Backup Daemon Uses More Than N Percent of Available Inodes |
| Last backup failed (alert with a backup `id`) | Prometheus Alerts Troubleshooting. Last Backup Failed |
| Backup fails with exit code 2, all nodes `UN`, long GC pauses or heap pressure | Cassandra Backup Failure (Exit Code 2) |

### Disambiguation rules

- **`Unable to gossip with any peers`** has two sections. Pick *Fails to Start* when the pod is crashing at startup and the exception is in its log; pick *Pod is Not Running* when the user quotes the alert. If unclear, read both: the alert section adds the headless-service (`clusterIP: None`) and inter-node network checks, the startup section adds the `cassandra0-0` reachability and DNS/Calico checks. Use the union of checks that the evidence supports.
- **Two `Stack trace(s)` blocks in the App Job section** are two different causes. Match the object named in the error: ConfigMap/other objects → missing `DEPLOY_W_HELM: "true"` and the documented cleanup; `Deployment "cassandra-operator"` → remove `operator.podName`.
- **High heap vs. backup exit code 2 vs. Unavailable exceptions** can be the same underlying GC/heap problem. Start with the section that matches the reported alert, and mention the others if logs show long GC pauses or a `DOWN` node.
- **Disk alerts**: Cassandra disk (data PV) and Backup Daemon disk/inodes are different volumes and different sections.
- **Pod startup stuck with no error**: use *Logs stuck* only if the last log line is `Initializing system.IndexInfo`; otherwise fall back to the checklist.
- Match on the exact string, not a loose keyword (for example, "Commit log got corrupted" is also the stated description under the gossip startup section, but the evidence decides, not the description).

## Confirm before fixing

Each documented fix assumes a specific cause. Verify it first with a read-only check:

| Section | Read-only confirmation |
|---|---|
| Could not read commit log | The path in the exception is the file to act on; the stack trace names `CommitLogReadException` |
| Gossip (startup or alert) | `kubectl get pods -n <namespace>` (is `cassandra0-0` Running?); `kubectl get svc cassandra -n <namespace> -o jsonpath='{.spec.clusterIP}'` returns `None`; `nslookup cassandra0-0.cassandra.<namespace>.svc.cluster.local` from another pod |
| Direct buffer memory | `file_cache_size_in_mb` is actually set in the `cassandra-configuration` ConfigMap |
| Logs stuck | `df -h -i /var/lib/cassandra/data` shows low free space or inodes |
| Saved cluster name | The error message shows the saved name to copy (usually the namespace name) |
| Unavailable exceptions | `nodetool status` shows at least one `DN` node; pod logs show why |
| Pods count lower than expected | `kubectl get statefulset,deployment,pods -n <namespace>` and `kubectl get events -n <namespace>` to see whether pods were scaled down or cannot be created (Pending, PVC, scheduling, quota) |
| Resource alerts | Current usage vs. limits (`kubectl top pod -n <namespace>`, `kubectl describe pod`) and whether the usage is sustained rather than a spike |
| Last backup failed | The backup `id` from the alert; read `/backup-storage/{backup_id}/.console` |
| Backup exit code 2 | `nodetool status` shows all `UN` **and** Cassandra logs show long GC pauses; otherwise it is the generic "Last Backup Failed" investigation |

If the confirmation does not match, say so and move to the next candidate section or the fallback checklist rather than applying the fix anyway.

## Guardrails

The Cassandra operator is operator-managed. Prefer the documented configuration/deployment path rather than making undocumented live changes that can be overwritten by reconciliation.

- **Ask for confirmation before any destructive or hard-to-reverse step**: deleting files on a PV, deleting Kubernetes objects, recreating services, restarting DNS/Calico pods, removing backups, scaling. Present the exact command and its impact, and let the user run it or approve it.
- For operator restart, follow the documented procedure: remove the `summary-spec` value from the `last-applied-configuration-info` ConfigMap and restart the operator pod.
- For Helm ownership/resource conflicts, the documented cleanup is `kubectl delete configmaps,ingresses,serviceaccounts,roles,rolebindings,grafanadashboard,prometheusrule,servicemonitor --all --namespace=<namespace>`. This deletes **every** object of those kinds in the namespace, including ones not owned by Cassandra. Confirm the namespace is dedicated to this deployment and that the error matches the documented failure, list what would be deleted first (`kubectl get ... -n <namespace>`), and only then run it. Then re-run the job.
- For commit-log corruption, prefer the least invasive option in order: (1) delete only the single file named in the stack trace and restart; (2) if that is insufficient, use the temporary `ignorereplayerrors` option; (3) clearing the entire `/var/lib/cassandra/data/commitlog/` directory can cause partial data loss and is a last resort that needs explicit user approval.
- The `cassandra.commitlog.ignorereplayerrors=true` JVM option is temporary: apply it through the `cassandra-env` ConfigMap (add the line before the last line), restart, and revert the ConfigMap afterward. Remind the user to revert.
- For resource alerts, follow the documented resource/storage scaling path rather than making unrelated Cassandra configuration changes.
- For backup failures, investigate the Cassandra/backup-daemon logs and the documented backup path before changing Cassandra data or schema. Dropping tables or restructuring keyspaces is a data-affecting change that requires user approval and a check that the tables are truly unused.
- Do not claim that a workaround is safe when the reference explicitly identifies a data-loss or recovery risk.
- Do not invent configuration keys, thresholds, or commands that are not in the reference. If you add a generic Kubernetes check, label it as supplementary to the documented procedure.

## Configuration conventions

- `DEPLOY_W_HELM: "true"` is the documented deploy parameter for the invalid GrafanaDashboard case and the cause of the first App Job ownership error (previous deploy was done without it).
- `operator.podName` should be removed when the App Job error shows the existing `cassandra-operator` Deployment with conflicting Helm ownership (`meta.helm.sh/release-name` must equal `cassandra-services`).
- `file_cache_size_in_mb` in the `cassandra-configuration` ConfigMap: comment out or delete it for the direct-buffer-memory startup issue.
- `cassandra-env` is the ConfigMap for the temporary `JVM_OPTS="$JVM_OPTS -Dcassandra.commitlog.ignorereplayerrors=true"` recovery option.
- Saved cluster-name mismatch: set the deploy parameter to either form below, using the saved name from the error (usually the namespace name), then restart the job:
  ```
  cassandra.configuration: "cluster_name: <saved cluster name>"
  ```
  (or the equivalent multi-line `cassandra: configuration: |- cluster_name: <saved cluster name>`).
- Do not replace the documented configuration path with a live `nodetool` or JVM command unless the reference explicitly calls for it.

## Cluster and diagnostic conventions

- Substitute the actual Cassandra namespace for `<namespace>`; never run commands with the literal placeholder.
- Check Cassandra pod state before diagnosing a startup or availability issue.
- Use `nodetool status` from a Cassandra pod: `UN` = up/normal, `DN` = down. Look for `DN` nodes when investigating Unavailable exceptions, and read that node's pod logs for the cause.
- The reference's Unavailable-exceptions section links to a "Cassandra Node Down" section that does not exist in this file. If a node is `DOWN`, use its pod logs plus the *Pod is Not Running* / gossip sections instead.
- For gossip failures, verify the Cassandra headless service has `clusterIP: None`, check DNS resolution of pod/service names, and check connectivity between Kubernetes nodes. Escalate suspected cluster network/DNS problems to the cluster administrators.
- For startup logs stuck at `Initializing system.IndexInfo`, check both PV free space and inode availability: `df -h -i /var/lib/cassandra/data`.
- For backup-daemon disk/inode alerts, check available space and inodes in the backup daemon pod (`df -h -i /backup-storage`) and remove unnecessary backups, or increase the volume.
- For a failed backup, use the backup ID from the alert: `cat /backup-storage/{backup_id}/.console` in the Backup Daemon pod.
- For Cassandra backup exit code 2, check Cassandra logs for long GC pauses and heap pressure, then review schema/keyspace structure and the number of tables per keyspace.
- Read-only commands are safe to run directly; anything that changes state follows the Guardrails.

## Fallback checklist

Use only when no documented section is a plausible match after the header grep, the phrase grep, and the disambiguation rules.

1. Capture the exact Cassandra error, alert, or startup message.
2. Identify the affected pod and namespace.
3. Check pod status and Kubernetes events (`kubectl get pods`, `kubectl describe pod`, `kubectl get events --sort-by=.lastTimestamp`).
4. Check Cassandra logs around the failure (`kubectl logs <pod>` and `--previous` for crash loops).
5. Check operator logs if the issue occurs during deployment or reconciliation.
6. Run `nodetool status` when the Cassandra service is available.
7. Check DNS/network connectivity for gossip or peer communication issues.
8. Check PV capacity and inode usage for storage/startup issues.
9. Check CPU, memory, and JVM heap pressure for resource-related symptoms.
10. For backup issues, inspect backup-daemon logs and the backup-ID-specific `.console` file.
11. Re-check the reference file for a more specific documented symptom before proposing a workaround.

State clearly that the issue is not documented, present findings as hypotheses with the evidence for each, and propose read-only next steps before any change.

When a documented fix exists, follow the reference section rather than replacing it with a generic path.

## Output format

Keep the answer short and structured:

1. **Matched section**: the exact reference header, or "no documented match (fallback)".
2. **Evidence**: the error text, alert, or command output that supports the match, and any conflicting evidence.
3. **Cause**: the documented cause, marked *confirmed* or *unverified*.
4. **Steps**: the documented fix, in order, with real namespace/pod names. Flag every destructive step and its risk.
5. **Verification**: how to confirm recovery (pod Ready, `nodetool status` all `UN`, alert cleared, job succeeds, backup re-run succeeds).
6. **Follow-ups**: revert temporary changes (for example `ignorereplayerrors`), and any recommendations from the section.

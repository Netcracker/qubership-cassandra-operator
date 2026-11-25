This section provides information on how to resolve the commonly encountered Cassandra issues.
<!-- #GFCFilterMarkerStart# -->

[[_TOC_]]

<!-- #GFCFilterMarkerEnd# -->

## If operator needs to be restarted

This section provides information about the Cassandra operator needs to be restarted.

### Description

The cassandra operator is designed in a way that restart will not execute any steps, if Cassandra service CR has no changes. 

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Edit the Config Map:** 
   - Edit the Config Map "last-applied-configuration-info" - delete value of key `summary-spec` and save.
   - Restart the operator pod.

### Recommendations

"Not applicable"

## App Job has failed with Error: rendered manifests contain a resource that already exists

This section provides information about the App Job has failed with Error.

### Description

Previous deploy was done without mandatory parameter DEPLOY_W_HELM: true.

### Alerts

"Not applicable"

### Stack trace(s)

```text
Error: rendered manifests contain a resource that already exists. Unable to continue with install: ConfigMap "cassandra-major-version" in namespace "diag-cassandra-22" exists and cannot be imported into the current release: invalid ownership metadata; label validation error: missing key "app.kubernetes.io/managed-by": must be set to "Helm"; annotation validation error: missing key "meta.helm.sh/release-name": must be set to "cassandra"; annotation validation error: missing key "meta.helm.sh/release-namespace": must be set to "diag-cassandra-22"
```

### How to solve

1. **delete the following objects:**
    - Use the following command to delete: ```kubectl delete configmaps,ingresses,serviceaccounts,roles,rolebindings,grafanadashboard,prometheusrule,servicemonitor --all --namespace=<namespace_name>``` 
    - re-run the job.

### Stack trace(s)

```text
Error: rendered manifests contain a resource that already exists. Unable to continue with install: Deployment "cassandra-operator" in namespace "ocs-cicd-dev-2-qcm-cassandra" exists and cannot be imported into the current release: invalid ownership metadata; annotation validation error: key "meta.helm.sh/release-name" must equal "cassandra-services": current value is "cassandra-operator" 
```

The main difference here is that in this case **Deployment object** is already exists.

### How to solve

You have to remove next deploy parameter if it set:

```
operator.podName
```

### Recommendations

"Not applicable"

## Cassandra Pod Fails to Start. Could not read commit log.

This section provides information about the Cassandra Pod Fails to Start because one could not read commit log issue.

### Description

Commit log got corrupted.

### Alerts

"Not applicable"

### Stack trace(s)

In the cassandra logs stacktrace looks like the following:
```text
[2023-08-10T07:34:57,224][ERROR][method=inspectCommitLogError]Exiting due to error while processing commit log during initialization.
org.apache.cassandra.db.commitlog.CommitLogReadHandler$CommitLogReadException: Could not read commit log descriptor in file /var/lib/cassandra/data/commitlog/CommitLog-7-1691592418508.log
	at org.apache.cassandra.db.commitlog.CommitLogReader.readCommitLogSegment(CommitLogReader.java:195)
	at org.apache.cassandra.db.commitlog.CommitLogReader.readCommitLogSegment(CommitLogReader.java:146)
	at org.apache.cassandra.db.commitlog.CommitLogReplayer.replayFiles(CommitLogReplayer.java:157)
	at org.apache.cassandra.db.commitlog.CommitLog.recoverFiles(CommitLog.java:193)
	at org.apache.cassandra.db.commitlog.CommitLog.recoverSegmentsOnDisk(CommitLog.java:174)
	at org.apache.cassandra.service.CassandraDaemon.setup(CassandraDaemon.java:360)
	at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:765)
	at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:889)
[2023-08-10T07:34:57,262][INFO][method=delete]Unfinished transaction log, deleting 
```

### How to solve

1. **delete the log file:** 
    - Delete the file the stacktrace is complaining about (in this example /var/lib/cassandra/data/commitlog/CommitLog.log).
    - Restart the pod.
2. **clear the log directory:** 
    - Clear the whole directory **/var/lib/cassandra/data/commitlog/**, but in some cases that can lead to partial data loss.
3. **add option in Config map:** 
    - Add `JVM_OPTS="$JVM_OPTS -Dcassandra.commitlog.ignorereplayerrors=true"` line to Config Map `cassandra-env` before the last line and restart the pod. After that revert changes in Config Map.

### Recommendations

"Not applicable"

## Cassandra Pod Fails to Start. Unable to gossip with any peers

This section provides information about the Cassandra Pod Fails to Start because unable to gossip with any peers issue.

### Description

Commit log got corrupted.

### Alerts

"Not applicable"

### Stack trace(s)

```text
[2023-08-11T07:10:10,572][ERROR][method=exitOrFail]Exception encountered during startup
java.lang.RuntimeException: Unable to gossip with any peers
	at org.apache.cassandra.gms.Gossiper.doShadowRound(Gossiper.java:1603)
	at org.apache.cassandra.service.StorageService.checkForEndpointCollision(StorageService.java:628)
	at org.apache.cassandra.service.StorageService.prepareToJoin(StorageService.java:888)
	at org.apache.cassandra.service.StorageService.initServer(StorageService.java:745)
	at org.apache.cassandra.service.StorageService.initServer(StorageService.java:694)
	at org.apache.cassandra.service.CassandraDaemon.setup(CassandraDaemon.java:395)
	at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:633)
	at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:786)
java.lang.RuntimeException: Unable to gossip with any peers
	at org.apache.cassandra.gms.Gossiper.doShadowRound(Gossiper.java:1603)
	at org.apache.cassandra.service.StorageService.checkForEndpointCollision(StorageService.java:628)
	at org.apache.cassandra.service.StorageService.prepareToJoin(StorageService.java:888)
	at org.apache.cassandra.service.StorageService.initServer(StorageService.java:745)
	at org.apache.cassandra.service.StorageService.initServer(StorageService.java:694)
	at org.apache.cassandra.service.CassandraDaemon.setup(CassandraDaemon.java:395)
	at org.apache.cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:633)
	at org.apache.cassandra.service.CassandraDaemon.main(CassandraDaemon.java:786)
[2023-08-11T07:10:10,600][INFO][method=pauseDispatch]Paused hints dispatch
```

### How to solve

1. **The pod cassandra0-0 is not running or unreachable:**
   - check cassandra0-0 pod state and make it running if it's not.
   - check connectivity to cassandra0-0 pod via `nslookup cassandra0-0.cassandra.<namespace>.svc.cluster.local or via ping and resolve the problem.

2. **An issues with DNS in the kubernetes:**
   - restart all DNS and Calico pods.

### Recommendations

"Not applicable"

## Cassandra Pod Fails to Start. Direct buffer memory.

This section provides information about the Cassandra Pod Fails to Start because direct buffer memory issue.

### Description

Direct Buffer overflows JVM limit.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve
1. **Direct Buffer overflows JVM limit**
    - check if `file_cache_size_in_mb` parameter in `cassandra-configuration` config map is defined and comment it or delete.

### Recommendations

"Not applicable"

## Cassandra Pod Fails to Start. Logs stuck.

This section provides information about the Cassandra Pod Fails to Start because logs stuck issue.

### Description

Cassandra Pod Fails to Start because logs stuck on `[method=<init>]Initializing system.IndexInfo`

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **logs stuck on `[method=<init>]Initializing system.IndexInfo`**
   - Check the free space on PV and if needed clean it.
   - Check inodes (df -h -i /var/lib/cassandra/data).

### Recommendations

"Not applicable"

## The GrafanaDashboard is invalid

This section provides information about the invalid GrafanaDashboard issue.

### Description

The GrafanaDashboard "cassandra-grafana-dashboard" is invalid: metadata.annotations: Too long: must have at most 262144 bytes.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Add `DEPLOY_W_HELM: "true"` to deploy parameters.**

### Recommendations

"Not applicable"

## Cassandra Pod goes CrashLoopBack.

This section provides information about the Cassandra Pod goes CrashLoopBack issue.

### Description

Cassandra Pod goes CrashLoopBack with Error **Saved cluster name <namespace-name> != configured name cassandra_cluster**.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve
1. **Change CMDB parameters:**
   - add following:
        ```text
        cassandra:
        configuration:  |-
            cluster_name: <saved cluster name> 
        ```

        or

        ```text
        cassandra.configuration: "cluster_name: <saved cluster name>"
        ```
        Where <saved cluster name> is the saved cluster name from error message (usually namespace name)
   - Restart the job.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. CPU Usage.

This section provides information about the Prometheus Alerts Troubleshooting with CPU Usage issue.

### Description

For some pods the CPU load is higher than 95 percent.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Insufficient resources.**
    - Increase the CPU for pods.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Memory Usage.

This section provides information about the Prometheus Alerts Troubleshooting with Memory Usage.

### Description

For some pods the memory usage is higher than 95 percent.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Insufficient resources.**
    - Increase RAM for pods.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Cassandra Disk Space Usage is Higher Than N Percent.

This section provides information about the Prometheus Alerts Troubleshooting with Cassandra Disk Space.

### Description

Disk space usage is higher than N percent of available space.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Free disk space is low.**
    - Increase the available disk space or add new node into the Cassandra cluster or clean up DB for obsolete tables if any.

### Recommendations

For more information, refer to the _Scaling Up_ section in the _[Cassandra Installation Procedure](https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/blob/master/docs/installation_guide.md#scaling-up)_.

## Prometheus Alerts Troubleshooting. Cassandra Pod is Not Running.

This section provides information about the Prometheus Alerts Troubleshooting with Cassandra Pod is Not Running issue.

### Description

Failed node unable to gossip with any peers.

### Alerts

"Not applicable"

### Stack trace(s)

```text
java. lang-RuntimeException: Unable to gossip with any peers
    at org-apache.cassandra.gms.Gossiper.doShadowRound(Gossiper.java:1443)
    at org-apache. cassandra-service.StorageService. CheckForEndpointCollision(StorageService.java:547)
    at org-apache.cassandra.service.StorageService-prepareToJoin(StorageService-java:804)
    at org-apache. cassandra. service.StorageService-initServer(StorageService.java:664) at org-apache.cassandra.service.StorageService.initServer(StorageService.java:613)
    at org-apache.cassandra.service.CassandraDaemon.setup(CassandraDaemon.java:379)
    at org-apache. cassandra.service.CassandraDaemon.activate(CassandraDaemon.java:602)
    at org-apache.cassandra.service.CassandraDaemon.main(CassandraDaemon-java:691)
[2020-07-1608:26:32,429] [ERROR] [method=exitOrFail]Exception encountered during startup
```

### How to solve

1. **Cassandra headless service has been changed manually.**
   - It is required to check cluster IP of `cassandra` service. It should be `None`.
   - ![Cluster IP of Cassandra service](/docs/public/images/cassandra-headless-check.png)
   - Otherwise, it is required to re-create `cassandra` service with `spec.clusterIP` value `None`.
2. **Network issues between Kubernetes nodes.**
   - To make sure there are no network issues, try to ping pods/services presented on the Cassandra's failed node from another nodes.
     If there is a network issue, ask the responsible IT Engineers for the cluster to check the environment for network issues.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Cassandra Pods Count is Lower Than Expected

This section provides information about the Prometheus Alerts Troubleshooting with Cassandra Pods Count is Lower Than Expected.

### Description

Number of currently running Cassandra pods is lower than expected.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Pods are manually scaled down or cannot be created.**
   - create pods

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. JVM Memory Heap Usage

This section provides information about the Prometheus Alerts Troubleshooting with JVM Memory Heap Usage.

### Description

Cassandra heap usage is currently greater than 80% of available capacity.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Insufficient resources.**
   - Increase the RAM for Cassandra pods.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Unavailable Exceptions Count Over 5 Minutes

This section provides information about the Prometheus Alerts Troubleshooting with Unavailable Exceptions Count Over 5 Minutes.

### Description

Unavailable exceptions count over 5 minutes. Cassandra node does not behave as expected.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **one of Cassandra node is down**
   - see [Cassandra Node Down](#cassandra-node-down) section.
2. **all pods are working** you need to find out which node is down:
  - Navigate to any Cassandra pod terminal.
  - Type `nodetool status`.
  - Find replica(s) with `DOWN` status(es).
  - In the pod logs look for exception messages that tell about the possible reasons for the error.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Backup Daemon Disk Space Usage is Higher Than N Percent

This section provides information about the Prometheus Alerts Troubleshooting with Backup Daemon Disk Space Usage is Higher Than N Percent.

### Description

Disk space usage is higher than N percent of available space.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Free disk space is low**
   - Increase available disk space or remove unnecessary backups.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Backup Daemon Uses More Than N Percent of Available Inodes

This section provides information about the Prometheus Alerts Troubleshooting with Backup Daemon Uses More Than N Percent of Available Inodes.

### Description

Amount of used inodes is higher than N percent of available.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Low amount of available inodes**
   - Increase available disk space or remove unnecessary backups.

### Recommendations

"Not applicable"

## Prometheus Alerts Troubleshooting. Last Backup Failed

This section provides information about the Prometheus Alerts Troubleshooting with Last Backup Failed.

### Description

The last performed backup failed.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Various reasons**
   - Find the `id` of the backup in the Prometheus alert.
   - Navigate to the Backup Daemon pod terminal and run the following command to get backup logs for investigating the failure: `cat /backup-storage/{backup_id}/.console`.

### Recommendations

"Not applicable"

## Cassandra Backup Failure (Exit Code 2)

This section provides information about the Cassandra backup failure issue caused by long GC pauses and heap pressure.

### Description

Cassandra backups fail with exit code 2 during execution of the backup playbook. The issue is observed even though all Cassandra nodes report UN (Up/Normal) status via nodetool status.

### Alerts

"Not applicable"

### Stack trace(s)

"Not applicable"

### How to solve

1. **Check Cassandra logs for long GC pauses and heap pressure.**

   Long GC pauses can temporarily make Cassandra pods unavailable, causing backup operations to fail.

2. **Review the Cassandra schema design and keyspace structure.**

   Excessive numbers of tables within a single keyspace can create high schema metadata overhead in the JVM heap.

3. **Identify keyspaces containing a very large number of tables.**

   A large number of tables or excessive schema metadata within a keyspace can significantly increase JVM heap utilization and trigger frequent GC activity.

4. **Clean up or optimize the keyspaces.**

   Recommended actions:

   - Remove unused tables.
   - Split excessively large keyspaces into smaller logical groups.
   - Reduce schema complexity where possible.

5. **Re-run the backup after schema optimization.**

   Once heap pressure and GC activity are reduced, backup operations should complete successfully.

### Recommendations

1. **Avoid maintaining an excessively large number of tables within a single Cassandra keyspace; this is an [anti-pattern](https://docs.datastax.com/en/planning/oss/anti-patterns.html#too-many-tables).**
2. **Regularly monitor:**
   - JVM heap usage
   - GC pause duration
   - Cassandra schema growth
3. **Perform periodic cleanup of unused tables and obsolete schema objects to minimize JVM metadata overhead.**
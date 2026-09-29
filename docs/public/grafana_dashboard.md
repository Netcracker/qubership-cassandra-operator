This section describes the Cassandra dashboards, metrics and their significance.

# Dashboard for Prometheus Metrics

An overview of the Cassandra dashboard is shown in the image below.

![Dashboard Overview](/docs/public/images/grafana-prometheus-overview.png)

## Metrics

This section describes the metrics and their meanings.

### Overview

* `Number of replicas` - Displays the number of working nodes.
* `Cassandra Pods Status` - Displays the status of the pods.

### All Pods: Disk

* `Disk I/O Utilization` - Displays the data read/write in bytes.
* `Disk I/O Utilization By Pods` - Displays the data read/write in bytes by pods.
* `IOps` - Displays the count of writes/reads completed.
* `IOps By Pods` - Displays the count of writes/reads completed by pods.

### All Pods: Network

* `Receive/Transmit Bandwidth` - Displays the overall incoming and outgoing network traffic in bytes per second.
* `Receive/Transmit Bandwidth Usage By Pods` - Displays the overall incoming and outgoing network traffic in bytes per second by pods.
* `Rate of Received/Transmitted Packets` - Displays the overall incoming and outgoing packets per second for the namespace.
* `Rate of Received/Transmitted Packets By Pods` - Displays the overall incoming and outgoing packets per second by pods for the namespace.

### Request Errors

* `Unavailable exceptions/_Interval_ Instance` - Displays the unavailable exceptions per _Interval_ Instance.
* `Timeout exceptions/_Interval_ Instance` - Displays the timeout exceptions per _Interval_ Instance.
* `Storage exceptions/_Interval_ Instance` - Displays the storage exceptions per _Interval_ Instance.

### CPU/RAM Usage

* `CPU usage` - Displays the CPU usage of each pod in the Namespace. The red lines show limits of CPUs usage.
  The values are indicated in Millicores. It is a special Kubernetes metric where CPU core split into 1000 units.

* `Memory usage` - Displays the RAM usage of each pod in the Namespace. The red lines show limits of RAM usage.

### Network

* `Network TX/RX` - Displays the Network TX/RX Instance.
* `Network errors TX/RX` - Displays the Network errors TX/RX Instance.

### Client Read\Write Latency

* `Client request latency (99%)` - Displays the 99th percentile client request latency.
* `Client request rate _Interval_` - Displays the client request latency rate.
* `Client request latency (avg)` - Displays the average client request latency for an instance.

### Local Read\Write Latency

* `Read rate per table per _Interval_` - Displays the read per table latency rate.
* `Read rate per keyspace per _Interval_` - Displays the read per keyspace latency rate.
* `Read latency per table` - Displays the read latency per table.
* `Read latency per keyspace` - Displays the read latency per keyspace.
* `Read latency per table per _Interval_` - Displays the the read per table latency rate.
* `Read latency per keyspace (Max)` - Displays the maximum latency per keyspace.
* `Read latency per keyspace (50thPercentile)` - Displays the 50th percentile read latency per keyspace.
* `Read latency per keyspace (95thPercentile)` - Displays the 95th percentile read latency per keyspace.
* `Read latency per keyspace (99thPercentile)` - Displays the 99th percentile read latency per keyspace.
* `Write rate per table per _Interval_` - Displays the write per table latency rate.
* `Write rate per keyspace per _Interval_` - Displays the write per keyspace latency rate.
* `Write latency per table` - Displays the write latency per table.
* `Write latency per keyspace` - Displays the write latency per keyspace.
* `Write latency per table per _Interval_` - Displays the write per table latency rate. 
* `Write latency per keyspace (Max)` - Displays the maximum write latency per keyspace.
* `Write latency per keyspace (50thPercentile)` - Displays the 50th percentile write latency per keyspace.
* `Write latency per keyspace (95thPercentile)` - Displays the 95th percentile write latency per keyspace.
* `Write latency per keyspace (99thPercentile)` - Displays the 99th percentile write latency per keyspace.

### Coordinator Read Latency

* `Coordinator read rate per table per _Interval_` - Displays read rate of Coordinator (node that recieves read/write request).
* `Coordinator read latency per table (avg)` - Displays the average read latencey per table

### Table Metrics (debug: LiveScanned, SSTablesPerRead)

* `Histogram of live cells scanned in queries on this table` - Displays the number of live cells scanned by queries on a table.
* `Histogram of live cells scanned in queries on this table (Max)` - Displays the maximum number of live cells scanned by queries on a table.
* `Histogram of live cells scanned in queries on this table (99thPercentile)` - Displays the 99th percentile of live cells scanned by queries on a table.
* `Number of sstable data files accessed per single partition read (Mean) per keyspace - SSTablesPerRead` - Displays the mean number of data files accessed per partition per keyspace.
* `Number of sstable data files accessed per single partition read (Max) - SSTablesPerRead` - Displays the maximum number of data files accessed per partition.
* `Number of sstable data files accessed per single partition read (95th Percentile) - SSTablesPerRead` - Displays the 95th percentile of data files accessed per partitiion.

### Thread Pools

* `Completed tasks per _Interval_` - Displays the completed tasks per second for an instance.
* `Total blocked tasks per _Interval_` - Displays the total blocked tasks per second for an instance.
* `Active tasks Instance` - Displays the active tasks for an instance.
* `Currently blocked tasks` - Displays the currently blocked tasks.

### JVM

* `JVM Memory Pools (used)` - Displays used memory JVM pools(spaces).
* `Last GC duration` - Displays the duration of the last GC.
* `JVM Threads` - Displays the number of user and daemon threads.
* `JVM Memory Pools (max)` - Displays current max pool size.

### Compaction

* `Completed tasks per _Interval_` - Displays the number of completed compaction task per _Interval_.
* `Pending tasks per _Interval_` - Displays the number of pending compaction task per interval.
* `Written compaction bytes per _Interval_` - Displays the written compaction bytes per _Interval_.
* `Compacted bytes per _Interval_` - Displays the compacted bytes per _Interval_.

### Commit Log Metrics

* `Commit Log Pending Tasks` - Displays the pending tasks for a commit log.
* `Commit Log Waiting time (Mean/Max)` - Displays the mean and maximum waiting time for commit log.
* `Commit Log Completed Tasks per $inter` - Displays the completed tasks for a commit log.
* `Commit Log Waiting time (Rate per $inter)` - Displays the waiting time for a commit log.

### SSTable Files and Hints

* `Total number of live SSTables` - Displays the total number of live SSTables.
* `Top 10 of live SSTables for key tables` - Displays the top 10 live SSTables for key tables.
* `Total hints per _Interval_` - Displays the total number of hints per instance.
* `Total hints in progress per _Interval_` - Displays the number of hint in progress per instance.

### Speculative Retry Metrics

The `Speculative retries per _Interval_` displays the number of speculative retries per instance.

### Disk Usage

* `Top 10 live disk space used` - Displays the disk space used by Keyspaces.
* `Cassandra tables disk space usage` - Displays the disk usage by Cassandra tables. Snopshots disk usage not included.
* `All mem tables live data size` - Displays the size of data in memtables

### Chunk Cache

* `Miss latency per _Interval_` - Displays the miss latency per Interval for an instance.
* `Miss latency 99th Percentile` - Displays the miss latency 99th percentile for an instance.
* `Size` - Displays the chunk cache size in bytes.

### Key Cache

* `Hit rate` - Displays the hit rate for an instance.
* `Size` - Displays the key cache size in bytes.

### Authentication

The `Requests per _Interval_` displays the number of authentication requests per _Interval_.

### Connected Clients

The `Client count` displays the number of connected clients.

### CQL

* `CQL regular statements rate per $inter` - Displays the number of regular statements executed per _Interval_.
* `CQL prepared statements rate per $inter` - Displays the number of prepared statements executed per _Interval_.

### Snapshots

The `Snapshots Size`  displays the snapshots size for an instance.

### Backup Daemon

* `Status` - Displays the Backup Daemon pod status.
* `CPU usage` - Displays the CPU usage of Backup Daemon pod in the Namespace. The red lines show limits of CPUs usage.
  The values are indicated in Millicores. It is a special Kubernetes metric where CPU core is split into 1000 units.
* `Memory usage` - Displays the RAM usage of Backup Daemon pod in the Namespace. The red lines show limits of RAM usage.
* `Last Backup Status` - Displays the last backup status.
* `Last Backup Size` - Displays the last backup size.
* `Dumps count` - Displays the amount of backups done.
* `Storage Space Usage` - Displays the space used by backups against the total space.
* `Storage Inodes Usage` - Displays the amount of used inodes against the total inodes.

### DBaaS Adapter

* `Status` - Displays the DBaaS Adapter pod status.
* `CPU usage` - Displays the CPU usage of DBaaS Adapter pod in the Namespace. The red lines show limits of CPUs usage.
  The values are indicated in Millicores. It is a special Kubernetes metric where CPU core is split into 1000 units.
* `Memory usage` - Displays the RAM usage of DBaaS Adapter pod in the Namespace. The red lines show limits of RAM usage.
* `GC Duration` - Displays the duration of the GC.
* `Threads` - Displays the number of user and daemon threads.
* `Requests Count` - Displays the total number of requests DBaaS Adapter received per response status.
* `DBaaS API Requests Duration` - Displays the requests duration per each DBaaS Adapter endpoint.

This section provides information about changing passwords for the users in Cassandra.

The following topics are covered in this section:

- [Change Password of Cassandra Superuser](#change-password-of-cassandra-superuser)
  - [Change Cassandra Superuser Password in DB](#change-cassandra-superuser-password-in-db)
  - [Change Cassandra Superuser Password in Services](#change-cassandra-superuser-password-in-services)
- [Change Password of Backup REST API](#change-password-of-backup-rest-api)
- [Change Password of DBaaS Adapter REST API](#change-password-of-dbaas-adapter-rest-api)
- [Change Password of DBaaS Aggregator REST API](#change-password-of-dbaas-aggregator-rest-api)
- [Change JVM Settings of Cassandra](#change-jvm-settings-of-cassandra)
- [Point-in-time Recovery for Amazon Keyspaces](#point-in-time-recovery-for-amazon-keyspaces)
  - [Limitations](#limitations)
  - [Enable PITR](#enable-pitr)
  - [Restoring an Amazon Keyspaces Table to a Point-in-time](#restoring-an-amazon-keyspaces-table-to-a-point-in-time)
  - [Repair after Loss of Connection between Clusters](#repair-after-loss-of-connection-between-clusters)
- [Audit Logs](#audit-logs)
  - [What does Audit Logging Captures](#what-does-audit-logging-captures)
  - [How to Configure](#how-to-configure)
    - [cassandra.yaml Configurations for AuditLog](#cassandrayaml-configurations-for-auditlog)
  - [Performance Comparison with Audit logs Enabled and Disabled](#performance-comparison-with-audit-logs-enabled-and-disabled)  

# Change Password of Cassandra Superuser

You can change the password of the Superuser in the database as well as in the services.

The following sections describe in detail the procedure to change the password for the Superuser.

## Change Cassandra Superuser Password in DB

To change the password in database:

1. For OpenShift, open OpenShift user interface console and select the **Cassandra** project. For Kubernetes, open Kubernetes UI Console and select the **Cassandra** namespace.
1. Login to the terminal of any Cassandra pod using the cqlsh tool and the current credentials. You have to only change the password on one pod and it broadcasts to all Cassandra pods in the ring. Execute the following command:
   
   ```
   cqlsh -u <current_user> -p <current_password>
   ```

   The default user is **admin**. The default password is **admin**. If you changed the password previously, use the current password.
1. Run the following command at the cqlsh prompt to update the password:
   
   ```
   ALTER USER <current_user> WITH PASSWORD 'NEW_PASSWORD';
   ```
   
1. Close the terminal.

## Change Cassandra Superuser Password in Services

To change the password of Superuser in services, follow the below sequence:

**Update Secret**

For OpenShift:

1. Open OpenShift UI Console.
1. Select the **Cassandra** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** list, select **cassandra-secret**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open Kubernetes UI Console.
1. Select the **Cassandra** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **cassandra-secret**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart Backup Daemon**

For OpenShift:

1. Navigate to **Applications > Pods**.
1. From the pod drop-down list, select **cassandra-backup-daemon-0**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**.

For Kubernetes:

1. Navigate to **Workloads > Pods**.
1. From the pod drop-down list, select **cassandra-backup-daemon-0**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

# Change Password of Backup REST API

To change the password of the backup REST API, follow the below sequence:

**Update Secret**

For OpenShift:

1. Open OpenShift UI Console.
1. Select **Cassandra** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **cassandra-backup-api-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open Kubernetes UI Console.
1. Select **Cassandra** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **cassandra-backup-api-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart Backup Daemon**

For OpenShift:

1. Navigate to **Applications > Pods**.
1. From the pod drop-down list, select **cassandra-backup-daemon-0**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**.

For Kubernetes:

1. Navigate to **Workloads > Pods**.
1. From the pod drop-down list, select **cassandra-backup-daemon-0**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

# Change Password of DBaaS Adapter REST API

To change the password of the DBaaS adapter REST API, follow the below sequence:

**Update Secret**

For OpenShift:

1. Open OpenShift UI Console.
1. Select the **Cassandra** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-adapter-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open Kubernetes UI Console.
1. Select the **Cassandra** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-adapter-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

# Change Password of DBaaS Aggregator REST API

To change the password of the DBaaS aggregator REST API, follow the below sequence:

**Update Secret**

For OpenShift:

1. Open OpenShift UI Console.
1. Select the **Cassandra** project.
1. Navigate to **Resources > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-aggregator-credentials**.
1. From the **Actions** drop-down list, select **Edit YAML**.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Save**.

For Kubernetes:

1. Open Kubernetes UI Console.
1. Select the **Cassandra** namespace.
1. Navigate to **Config and Storage > Secrets**.
1. From the **Secrets** drop-down list, select **dbaas-aggregator-credentials**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Replace the value of **data.password** with the new password encoded to base64.
1. Click **Update**.

**Restart DBaaS Adapter**

For OpenShift:

1. Navigate to **Applications > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. From the **Actions** drop-down list, select **Delete**.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

For Kubernetes:

1. Navigate to **Workloads > Deployments**.
1. Click **dbaas-cassandra-adapter**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**. Wait until the redeployment is completed.

# Change JVM Settings of Cassandra

To change the JVM settings of Cassandra, follow the below sequence:

**Update Config Map**

For OpenShift:

1. Open OpenShift UI Console.
1. Select the **Cassandra** project.
1. Navigate to **Resources > Config Maps**.
1. From the **Config Maps** drop-down list, select **cassandra-jvm**.
1. From the **Actions** drop-down list, select **Edit**.
1. Update the required JVM options of the **config**.
1. Click **Save**.

For Kubernetes:

1. Open Kubernetes UI Console.
1. Select the **Cassandra** namespace.
1. Navigate to **Config and Storage > Config Maps**.
1. From the **Config Maps** drop-down list, select **cassandra-jvm**.
1. In the upper right corner, click the pencil icon to edit.
1. Click the **YAML** tab.
1. Update the required JVM options of the **config**.
1. Click **Update**.

**Restart Cassandra**

There are several ways to restart Cassandra as described below.

**Restart Cassandra replicas manually**

For OpenShift:

1. Open OpenShift UI Console.
1. Select the **Cassandra** project.
1. Navigate to **Applications > Pods**.
1. From the pod drop-down list, select **cassandra%i**.
1. From the **Actions** drop-down list, select **Delete**.
1. Wait until the **cassandra%i** pod is in the `Running` state and `1/1` containers are ready.
1. Repeat the above steps for each Cassandra **Pod**.

For Kubernetes:

1. Open Kubernetes UI Console.
1. Select the **Cassandra** namespace.
1. Navigate to **Workloads > Pods**.
1. From the pod drop-down list, select **cassandra%i**.
1. In the upper right corner, click the trash icon to delete.
1. In the pop-up window, click **Delete**.
1. Wait until the **cassandra%i** pod is in the `Running` state and `1/1` containers are ready.
1. Repeat the above steps for each Cassandra **Pod**.

# Point-in-time Recovery for Amazon Keyspaces

Point-in-time recovery (PITR) helps protect your Amazon Keyspaces tables from accidental write or delete operations by providing you continuous backups of your table data.
With point-in-time recovery, you can restore a table's data to any second in time since PITR was enabled within the last 35 days. 

For more information on PITR, refer to [https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery.html](https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery.html).

## Limitations

* PITR does not overwrite existing tables. You can only restore data to a new table.
* Backup/restore works with tables and not keyspaces.
* If you delete a table with point-in-time recovery enabled, you can query for the deleted table's data for 35 day, and restore it to the state it was in just before the point of deletion.

## Enable PITR

You can enable PITR by using AWS Management Console, or you can enable it programmatically. For more information, refer to [https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery_HowItWorks.html#howitworks_enabling](https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery_HowItWorks.html#howitworks_enabling).

## Restoring an Amazon Keyspaces Table to a Point-in-time

To restore an Amazon Keyspaces table to a point in time, refer to the official documentation at [https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery_Tutorial.html](https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery_Tutorial.html).

## Repair after Loss of Connection between Clusters

If Cassandra is deployed in a Multi-DC schema and one of the datacenters is down for some time, the data can be out of sync when the DC returns.
To fix the data replication, the Cassandra Reaper tool can be used. To deploy Cassandra with Reaper, see [Cassandra Reaper Installation](installation_guide.md#cassandra-reaper-installation) in the _Cassandra Operator Installation Procedure_ guide.

Manual or a scheduled repair processes can be configured for specific keyspaces. For more information, refer to Scheduling a Cluster Repair at [http://cassandra-reaper.io/docs/usage/schedule/](http://cassandra-reaper.io/docs/usage/schedule/) and Running a Cluster Repair at [http://cassandra-reaper.io/docs/usage/single/](http://cassandra-reaper.io/docs/usage/single/).

# Audit Logs

Audit logging in Cassandra logs every incoming CQL command request, as well as authentication (successful/unsuccessful login) to a Cassandra node.

## What does Audit Logging Captures

Audit logging captures the following events:
- Successful as well as unsuccessful login attempts
- All database commands executed via native CQL protocol attempted or successfully executed

## How to Configure

By default, the audit logs are disabled. If you just need to enable it without any additional options, use the following parameters:

```
cassandra:
  auditLogEnabled: true
```

### cassandra.yaml Configurations for AuditLog

The following options are supported:

<!-- * `logger`: Class name of the logger/ custom logger.
* `audit_logs_dir`: Auditlogs directory location, if not set, default to
[.title-ref]#cassandra.logdir.audit# or [.title-ref]#cassandra.logdir# +
/audit/ -->
* `enabled`: This option enables/ disables audit log
* `included_keyspaces`: Comma separated list of keyspaces to be included in audit log, default - includes all keyspaces
* `excluded_keyspaces`: Comma separated list of keyspaces to be excluded
from audit log, default - excludes no keyspace except [.title-ref]#system#, [.title-ref]#system_schema# and [.title-ref]#system_virtual_schema#
* `included_categories`: Comma separated list of Audit Log Categories to be included in audit log, default - includes all categories
* `excluded_categories`: Comma separated list of Audit Log Categories to be excluded from audit log, default - excludes no category
* `included_users`: Comma separated list of users to be included in audit log, default - includes all users
* `excluded_users`: Comma separated list of users to be excluded from audit log, default - excludes no user

To set up audit logs with extra options, provide the required parameters to the `cassandra.configuration` field in the **values.yaml** file. **It is important to enable audit logs here and set logger class_name: FileAuditLogger.**<br>
For example, to configure `excluded_keyspaces`:

```yaml
cassandra:
  configuration: |-
    audit_logging_options:
		  enabled: true
		  excluded_keyspaces: test_ks, test_ks-1
		  logger:
		    - class_name: FileAuditLogger
```

The logfile is saved at **/var/lib/cassandra/data/logs/audit/audit.log**.

## Performance Comparison with Audit logs Enabled and Disabled

On average, cassandra performance with audit logs enabled decreases by 12%.
This is a rough estimate based on running [Cassandra Benchmarks](https://git.netcracker.com/PROD.Platform.Databases/benchmarks/cassandra-benchmarks) with the following parameters:

``` json
curl -XPOST -H "Application/json" cassandra-benchmarks:8080/benchmarks -d '{"parameters": {"command": "write", "count": 1000000, "rate": "threads=3", "mode": "native cql3"}}'
```

|          | 1000 | 10000 | 100000 | 1000000 | 10000000 |
|----------|------|-------|--------|---------|----------|
| Disabled | 4    | 20    | 98     | 974     | 9248     |
| Enabled  | 6    | 24    | 111    | 1187    | 11490    |

The following graph shows performance comparison with audit logs enabled and disabled:

![Performance Comparison with Audit logs Enabled and Disabled](/docs/public/images/performance_comparison_with_audit_logs_enabled_and_disabled.png)

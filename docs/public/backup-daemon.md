For POST operations, you must specify the user name and password using `BACKUP_DAEMON_API_CREDENTIALS_USERNAME` and `BACKUP_DAEMON_API_CREDENTIALS_PASSWORD` env parameters so that you can use REST API to run the backup tasks:

# Run Full Manual Backup

This step returns the backup folder (vault) as a plain-text response. You can later use this vault name to get a backup status (7):

```
curl -XPOST  -u  username:password localhost:8080/backup
```

## Run Manual Backup For Some Subset of DBs (granular backup)

This section provide details about manual backup. 

## Run Manual Backup, Passing DBs and Tables

This returns the backup folder (vault) as a plain-text response. You can later use this vault name to get a backup status (7).

For databases (DBs), use the following command:  

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"dbs":["db_name1","db_name2"]}' localhost:8080/backup
```

If you want to run manual backup only for specific tables in the database, use the following command:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"dbs":["db_name1",{"db_name2":{"tables":["table_1","table_2"]}}]}' localhost:8080/backup
```

<!-- For DBs with collections and queries use the following command:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"dbs":["db_name1",{"db_name2":{"collections":["first",{"second":{"test1":"1"}}]}]}' localhost:8080/backup
``` -->

## Run Manual Backup That Will Not Be Deleted Ever

If you do not want your backup to be evicted, add `allow_eviction":"False"` in your request. It works both for full and granular backups: 

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d '{"allow_eviction":"False","dbs":["arg1","arg2"]}' localhost:8080/backup
```

## Run Manual Eviction

For manual eviction, use the following command:

```
curl -XPOST -u username:password localhost:8080/evict
```

## Remove Specific Backup by ID

To remove the specific manual eviction, use the following command:

```
curl -XPOST -u username:password localhost:8080/evict/backupid
```
It returns `200` for OK.

## Get Health

The Get Health method returns JSON with the following information:

```
"status": status of backup daemon   
"backup_queue_size": backup daemon queue size (if > 0 then there are 1 or tasks waiting for execution)
 "storage": storage info:
  "total_space": total storage space in bytes
  "dump_count": number of backup
  "free_space": free space left in bytes
  "size": used space in bytes
  "total_inodes": total number of inodes on storage
  "free_inodes": free number of inodes on storage
  "used_inodes": used number of inodes on storage
  "last": last backup metrics
    "metrics['exit_code']": exit code of script 
    "metrics['exception']": python exception if backup failed
    "metrics['spent_time']": spent time
    "metrics['size']": backup size in bytes
    "failed": is failed or not
    "locked": is locked or not
    "id": vault name of backup
    "ts": timestamp of backup  
  "lastSuccessful": last succesfull backup metrics
    "metrics['exit_code']": exit code of script 
    "metrics['spent_time']": spent time
    "metrics['size']": backup size in bytes
    "failed": is failed or not
    "locked": is locked or not
    "id": vault name of backup
    "ts": timestamp of backup\
```

Use the command below to get JSON health:

```
curl -XGET localhost:8080/health
```

## Run Recovery

You must specify JSON with vault. You must specify databases in the databases list. Also, you might need to specify the tables for each keyspace in format: `"dbs":["<db_name>", {"<db_name_1>":{"tables":["<table_name>"]}}]`
You cannot run a recovery without database specified.   

To run the full recovery, you need to use database-specific methods. You can not run full recovery from API.

You will recieve `task_id` as a response. Use it in (7) to get the status of recovery:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d  '{"vault":"20170913T1114", "dbs":["db1","db2"]}' localhost:8080/restore
```

If you need to copy a database, you can use the `changeDbName` arg in JSON.   

An example is given below:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d  '{"vault":"20170913T1114", "dbs":["db1","db2","db3""], "changeDbNames":{"db1":"new_db1_name","db2":"new_db2_name"}}' localhost:8080/restore
```

An example of running restore for specific tables:

```
curl -XPOST -u username:password -v -H "Content-Type: application/json" -d  '{"vault":"20230901T072354", "dbs":["db1", {"db2":{"tables":["tb1"]}}]}' localhost:8080/restore
```

This saves `db1` and `db2` as they are on a DB server, and restore `db1` and `db2` into new (or existing) databases called `new_db1_name` and `new_db2_name`. Database `db3` will be rewritten, because it is not in the `changeDbNames` list.

## Run Point In Time Recovery in AWS Keyspaces

For more detailed information on PITR, follow the link: https://docs.aws.amazon.com/keyspaces/latest/devguide/PointInTimeRecovery.html

Request:

```
curl -u <username>:<passwoed> -XPOST localhost:8080/external/restore -d '{"restore_timestamp":"<timestamp in ISO 8601 format>", "table":"<table name>", "ks_name":"<keyspace name>", "restored_table_name": "<name of restored table>"}' -H "Content-Type: application/json"
```

Body parameters:

- `restore_timestamp` (_optional_)—A timestamp in ISO 8601 format the table must be restored to. If not set, the current timestamp is used.
- `ks_name` (_required_)—A keypace that holds the table
- `table` (_required_)—A table to restore
- `restored_table_name` (_required_)—A new table name

Response:

`<restore_id>`

Use [Get Backup/Recovery Status](#get-backup-recovery-status) to track restore status.

An example is given below:  

```
curl -u backup:backup localhost:8080/external/restore -d '{"restore_timestamp":"2023-08-15T11:22:31.000Z", "table":"test", "ks_name":"clpl_tst", "restored_table_name": "new_test_table"}' -H "Content-Type: application/json"
```

## Get Backup/Recovery Status

You recieve the HTTP responses: `200` for `OK`, `206` for `Still in process` and `500` for `NOT OK`. Use the following command to get recovery status:

```
curl -XGET localhost:8080/jobstatus/<task_id>
```

or 

```
curl -XGET localhost:8080/jobstatus/<vault_name>
```

Also, you recieve a JSON string as plain-text with the following information:

* `status`—Successful/Queued/Processing/Failed
* `message`—Optional field, only if error, contains description of error
* `vault`—vaultname to use in recovery,
* `type`—backup/restore
* `err`—None if no error, last 5 lines of log if status=Failed
* `task_id`—task_id of the task   

An example is given below:  

```
{"status": "Successful", "vault": "20170927T1122", "type": "backup", "err": "None", "task_id": "a592eeb6-abac-4d98-b638-75a820e333b1"}
```

## List Backups

To list backups, use the following command:

```
curl -XGET localhost:8080/listbackups
```

This command returns a json list of backup names.

## Get Backup Information

To get the backup information, use the following command:

```
curl -XGET localhost:8080/listbackups/<vault_id>
```

This command returns a JSON string with stats about particular backup:

* `ts`—UNIX timestamp of backup
* `spent_time`—time spent on backup (in ms)
* `db_list`—List of backed up databases
* `id`—vault name
* `size`—Size of backup in bytes
* `evictable`—whether backup is evictable
* `locked`—whether backup is locked (either process isn't finished, or it failed somehow)
* `exit_code`—exit code of backup script
* `failed`—whether backup failed or not
* `valid`—is backup valid or not

An example is given below:

```
{"ts": 1514282821000, "spent_time": "5066ms", "db_list": "Sorry, no information on databases available", "id": "20171226T100701", "size": "36647b", "evictable": true, "locked": false, "exit_code": 0, "failed": false, "valid": true}
```

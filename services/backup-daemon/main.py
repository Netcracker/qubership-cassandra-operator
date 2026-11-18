#!/usr/bin/env python3
# Copyright 2024-2025 NetCracker Technology Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import os
import traceback
import argparse
import src.aws_restore
import logging
import src.backup_and_restore
import src.cassandra_client
import src.marker
import src.os_utils


def get_secret(path: str, fallback: str = "") -> str:
    try:
        with open(path, "r", encoding="utf-8") as file:
            value = file.read().strip()

            if value:
                return value

    except Exception:
        pass

    return fallback


TLS_ENABLED = src.os_utils.str_to_bool(os.getenv("TLS_ENABLED", False))

CASSANDRA_USERNAME = get_secret("/var/run/secrets/cassandra/username")
CASSANDRA_PASSWORD = get_secret("/var/run/secrets/cassandra/password")

aws_access_key = get_secret("/var/run/secrets/aws/accessKey")
aws_secret_key = get_secret("/var/run/secrets/aws/secretKey")
aws_region = get_secret("/var/run/secrets/aws/region")



def parse_args():
    parser = argparse.ArgumentParser(description="Backup and Restore")

    parser.add_argument('action', choices=['backup', 'restore', 'list-dbs', 'set-marker', 'get-marker'],
                        help='Action to perform')
    parser.add_argument('-f', dest='vault', help='Vault option')
    parser.add_argument('-d','--dbs', dest='databases', help='Databases option')
    parser.add_argument('-m','--dbmap', dest='dbmap', help='Dbmap option')
    parser.add_argument('-restore_roles', dest='restore_roles', help='Do we need to replace roles from backup', default=True)
    parser.add_argument('-restore_timestamp', dest='restore_timestamp',
                        help='Restore timestamp option')
    parser.add_argument('-ks_name', dest='ks_name',
                        help='KeySpace name option')
    parser.add_argument('-table', dest='table', help='Table option')
    parser.add_argument('-restored_table_name', dest='restored_table_name',
                        help='Restored table name option')
    parser.add_argument('-value', '--marker-value', dest='marker_value', help='Marker value')

    return parser.parse_args()


def main():
    logging.basicConfig(level=logging.INFO,
                        format="[%(asctime)s][%(levelname)s][class=%(name)s][thread=%(thread)d] %(message)s",
                        datefmt="%Y-%m-%dT%H:%M:%S%z")

    hosts_file_path = '/opt/backup/cassandra_hosts/hosts'
    hosts_template_path = '/opt/backup/hosts_template'
    src.os_utils.create_hosts_inventory(hosts_file_path, hosts_template_path)

    args = parse_args()

    if args.action == 'backup':
        try:
            src.backup_and_restore.cluster_backup(args.databases, args.vault,
                                                  TLS_ENABLED, CASSANDRA_USERNAME, CASSANDRA_PASSWORD)
        except Exception as e:
            logging.error(f"Backup has failed: {e}")
            exit(1)
    elif args.action == 'restore':
        restore = src.backup_and_restore.Restore(
            args.vault, args.dbmap, args.databases, args.restore_roles)
        try:
            restore.restore()
        except Exception as e:
            logging.error(f"Restore has failed: {e}")
            logging.error(traceback.format_exc())
            exit(1)
    elif args.action == 'aws-restore':
        src.aws_restore.run_restore_aws(
            args.ks_name, args.table, args.restored_table_name, aws_access_key, aws_secret_key, aws_region)
    elif args.action == 'list-dbs':
        try:
            print("\n".join(src.backup_and_restore.list_databases(args.vault)))
        except Exception as e:
            logging.error(f"ListDB has failed: {e}")
            exit(1)
    elif args.action == 'set-marker':
        try:
            hosts = src.os_utils.reformat_hostnames(os.getenv('CASSANDRA_HOSTS'))
            client = src.cassandra_client.CassandraClient(
                hosts,
                username=CASSANDRA_USERNAME,
                password=CASSANDRA_PASSWORD,
                tls_enabled=TLS_ENABLED,
                connect_timeout=int(os.getenv('CONNECT_TIMEOUT', 20)),
                request_timeout=int(os.getenv('REQUEST_TIMEOUT', 20))
            )
            src.marker.set_marker(client, args.marker_value)
            client.close()
        except Exception as e:
            logging.error(f"Set marker has failed: {e}")
            exit(1)
    elif args.action == 'get-marker':
        try:
            hosts = src.os_utils.reformat_hostnames(os.getenv('CASSANDRA_HOSTS'))
            client = src.cassandra_client.CassandraClient(
                hosts,
                username=CASSANDRA_USERNAME,
                password=CASSANDRA_PASSWORD,
                tls_enabled=TLS_ENABLED,
                connect_timeout=int(os.getenv('CONNECT_TIMEOUT', 20)),
                request_timeout=int(os.getenv('REQUEST_TIMEOUT', 20))
            )
            value = src.marker.get_marker(client)
            print(value)
            client.close()
        except Exception as e:
            logging.error(f"Get marker has failed: {e}")
            exit(1)
    else:
        logging.error("Invalid action:", args.action)
        exit(1)


if __name__ == "__main__":
    main()

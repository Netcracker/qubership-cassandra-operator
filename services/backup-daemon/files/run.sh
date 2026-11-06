#!/bin/sh
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


if [ "$TLS_ENABLED" = "true" ]; then
    keytool -import -file $TLS_ROOTCERT -noprompt -alias cassandra -storepass "cassandra" -keystore /opt/cassandra/truststore.jks
fi

if [ "$CASSANDRA_MAJOR_VERSION" = "4" ]; then
    cp -R /opt/downloads/apache-cassandra-"$CASSANDRA4_DIR"/* "$CASSANDRA_HOME"/
else
    cp -R /opt/downloads/apache-cassandra-"$CASSANDRA3_DIR"/* "$CASSANDRA_HOME"/
fi

debug_params=""
if [ "$REMOTE_DEBUG" = "true" ]; then
    debug_params="-m debugpy --listen localhost:5678"
fi

exec /opt/backup/backup-daemon
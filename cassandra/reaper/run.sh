#!/bin/bash
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

set -x
set -e


if [[ $REAPER_CASS_NATIVE_PROTOCOL_SSL_ENCRYPTION_ENABLED == "true" ]]; then
  keytool -import -trustcacerts -noprompt -alias ca_alias -file /usr/ssl/ca.crt -keystore /tmp/truststore.jks -storepass reaper_password
  openssl pkcs12 -export -in /usr/ssl/tls.crt -inkey /usr/ssl/tls.key -out /tmp/client.p12 -CAfile /usr/ssl/ca.crt -name client-cert -passout pass:reaper_password
fi

cp --remove-destination /etc/cassandra-reaper-temp/cassandra-reaper.yml /etc/cassandra-reaper/config/
#call original entrypoint
exec /usr/local/bin/entrypoint.sh cassandra-reaper
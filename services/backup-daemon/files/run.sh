#!/bin/sh

cp -pr /app/backup/. /opt/backup/

mkdir /opt/backup/.ssh && chmod 700 /opt/backup/.ssh
echo "$SSH_PRIVATE_KEY" >/opt/backup/.ssh/id_rsa
chmod 600 /opt/backup/.ssh/id_rsa

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
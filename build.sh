#!/usr/bin/env bash
set -e

TARGET=target
DISTR_NAME=deploy

export CGO_ENABLED=0
export GOPROXY="https://artifactorycn.qubership.org/pd.sandbox-staging.go.group"
export GOSUMDB=off


SCRIPTS=scripts
DIST_FILE="${SCRIPTS}/migration-artifacts.zip"
DIST_CONTENT="migration-artifacts"

rm -rf ./${SCRIPTS}
mkdir ${SCRIPTS}
zip -qr "$DIST_FILE" "$DIST_CONTENT"


go build -o ./bin/cassandra-operator -gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go

docker build -t cassandra-operator .
for id in $DOCKER_NAMES
do
    docker tag cassandra-operator $id
done

mkdir -p deployments/charts/cassandra-operator

cp -R ./charts/helm/cassandra-operator/* deployments/charts/cassandra-operator/
cp ./charts/helm/cassandra-operator/deployment-configuration.json deployments/deployment-configuration.json

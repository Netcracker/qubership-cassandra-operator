FROM --platform=$BUILDPLATFORM golang:1.23.0-alpine3.20 AS builder

ENV GOSUMDB=off GOPRIVATE=github.com/Netcracker

RUN apk add --no-cache git
RUN --mount=type=secret,id=GH_ACCESS_TOKEN git config --global url."https://$(cat /run/secrets/GH_ACCESS_TOKEN)@github.com/".insteadOf "https://github.com/"

COPY . /workspace

WORKDIR /workspace

RUN go mod tidy

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o ./bin/cassandra-operator \
-gcflags all=-trimpath=${GOPATH} -asmflags all=-trimpath=${GOPATH} ./main.go

#FROM alpine:3.20.3

ENV WORKDIR=/opt/operator/

ENV OPERATOR=/usr/local/bin/cassandra-operator \
    USER_UID=1001 \
    USER_NAME=cassandra-operator

RUN echo 'https://dl-cdn.alpinelinux.org/alpine/v3.20/main/' > /etc/apk/repositories \
    && apk add --no-cache openssl \
    && apk update \
    && apk upgrade

COPY --from=builder /workspace/bin/cassandra-operator ${OPERATOR}
COPY build/bin /usr/local/bin
#COPY bin/gocql /usr/local/bin/

RUN chmod +x /usr/local/bin/entrypoint
RUN  chmod +x /usr/local/bin/user_setup && /usr/local/bin/user_setup

RUN mkdir -p deployments/charts/cassandra-operator

COPY /charts/helm/cassandra-operator/ deployments/charts/cassandra-operator/
COPY /charts/helm/cassandra-operator/deployment-configuration.json deployments/deployment-configuration.json

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

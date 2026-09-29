The following topics are described in this chapter:

- [Overview](#overview)
  - [Netcracker Cassandra Delivery and Features](#netcracker-cassandra-delivery-and-features)
- [Cassandra Components](#cassandra-components)
  - [Cassandra Operator](#cassandra-operator)
  - [Cassandra](#cassandra)
  - [Cassandra Reaper](#cassandra-reaper)
  - [Cassandra Backup Daemon](#cassandra-backup-daemon)
  - [Cassandra DBaaS Adapter](#cassandra-dbaas-adapter)
  - [Robot Tests](#robot-tests)
- [Supported Deployment Schemes](#supported-deployment-schemes)
  - [On-Prem](#on-prem)
    - [Non-HA Deployment Scheme](#non-ha-deployment-scheme)
    - [HA Deployment Scheme](#ha-deployment-scheme)
    - [DR Deployment Scheme](#dr-deployment-scheme)
    - [Google Cloud](#google-cloud)
    - [AWS](#aws)
    - [Azure](#azure)

# Overview

Apache Cassandra is an open source NoSQL distributed database trusted by thousands of companies for scalability and high availability without compromising performance. Linear scalability and proven fault-tolerance on commodity hardware or cloud infrastructure makes it the perfect platform for mission-critical data.

Cassandra was designed to meet emerging largescale, both in data footprint and query volume, storage requirements. As applications began to require full global replication and always available low-latency reads and writes, it became imperative to design a new kind of database model as the relational database systems of the time struggled to meet the new requirements of global scale applications.

Systems like Cassandra are designed for these challenges and seek the following design objectives:

* Full multi-master database replication
* Global availability at low latency
* Scaling out on commodity hardware
* Linear throughput increase with each additional processor
* Online load balancing and cluster growth
* Partitioned key-oriented queries
* Flexible schema

## Netcracker Cassandra Delivery and Features

The Netcracker platform provides Cassandra deployment to Kubernetes/OpenShift using their own helm chart with an operator and additional features.

The deployment procedure and additional features include the following:

* Support of Netcracker deployment jobs for HA scheme and different configurations. For more information, refer to the [Cassandra Operator Deployment](installation_guide.md) chapter.
* Backup and restore of keyspaces. <!-- #GFCFilterMarkerStart# -->For more information, refer to [Cassandra Backup Daemon Guide](https://git.netcracker.com/PROD.Platform.Databases/cassandra-backup-daemon/-/blob/master/readme.md).<!-- #GFCFilterMarkerEnd# -->
* Monitoring integration with Grafana Dashboard and Prometheus Alerts. For more information, refer to the [Cassandra Operator Monitoring](grafana_dashboard.md) chapter in _Cloud Platform Monitoring Guide_.
* Integration with DBaaS.
* Cassandra Reaper as a sidecar container. For more information, refere to the _Cassandra Reaper_ documentation at [http://cassandra-reaper.io/docs/](http://cassandra-reaper.io/docs/).
* Disaster Recovery scheme with multi-DC configuration. 

# Cassandra Components

The Cassandra components are shown in the following image:

![Cassandra Components](/docs/public/images/cassandra_components.png)

## Cassandra Operator

The Cassandra Operator is a mandatory microservice written with Operator-SDK and designed specifically for Kubernetes environments.
It simplifies the deployment and management of Cassandra clusters, which are critical for distributed coordination.
In addition to deploying the Cassandra cluster, the operator also takes care of managing supplementary services, ensuring seamless integration, and efficient resource utilization.
Cassandra Operator also performs an upgrade scenario without Cassandra downtime and allows to scale the Cassandra cluster.

## Cassandra

Cassandra is a custom docker image distribution with additional tools to enhance troubleshooting capabilities and provides configuration flexibility.
The solution includes a Prometheus exporter that monitors and collects vital metrics, optimizing performance, and ensuring a robust Cassandra cluster.

## Cassandra Reaper

Cassandra Reaper is an open source tool that aims to schedule and orchestrate repairs of Cassandra clusters. 

## Cassandra Backup Daemon

Cassandra Backup Daemon is a microservice that offers a convenient REST API for performing backups and restores of Cassandra keyspaces.
It enables users to initiate full or granular backups and restores programmatically, making it easier to automate these processes.
Additionally, the daemon allows users to schedule regular backups, ensuring data protection and disaster recovery.
Furthermore, it offers the capability to store backups on remote S3 storage, providing a secure and scalable solution for long-term data retention.

## Cassandra DBaaS Adapter

The Cassandra DBaaS adapter is a microservice for integration with DBaaS that allows to manage logical databases through API.

## Robot Tests

Robot Tests is a microservice that performs integration testing after all components of Cassandra deployment are installed.

# Supported Deployment Schemes

The supported deployment schemes are described in the below sections.

## On-Prem

The deployment schemes for On-Prem are specified below.

### Non-HA Deployment Scheme

It is the same as the HA Deployment Scheme but with a single Cassandra replica.

### HA Deployment Scheme

![Cassandra HA](/docs/public/images/cassandra_ha.png)

### DR Deployment Scheme

The Disaster Recovery scheme of Cassandra deployment assumes that two Cassandra Datacenters are deployed for both sides on separate Kubernetes environments with pod-to-pod connectivity between them.

Cassandra works in the active/active scheme and does not change its state during DR scenarios.
By default, Cassandra supports asynchronous replication between data centers. This means that when data is written to the local data center, it is first stored locally and then asynchronously replicated to other data centers in the cluster. 
However, it is important to note that Cassandra also provides options for synchronous replication between data centers if strong consistency is required. This is achieved using features like "QUORUM" consistency level, where a write operation is acknowledged only after it has been replicated to a specified number of data centers.
Synchronous replication ensures that data is replicated and acknowledged across multiple data centers before considering the write operation as successful. However, this can introduce additional latency and potential performance implications, as writes must wait for confirmation from remote data centers.

![Cassandra DR](/docs/public/images/cassandra_DR.png)

### Google Cloud

Not Applicable; the default HA scheme is used for the deployment to Google Cloud.

### AWS

Not Applicable; the default HA scheme is used for the deployment to AWS.

### Azure

Not Applicable; the default HA scheme is used for the deployment to Azure.

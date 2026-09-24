The following topics are described in this guide:

[[_TOC_]]

# Prerequisites

The prerequisites for the installation process are specified in this section.

## Common 

It is highly recommended to deploy Cassandra along with platform logging and monitoring infrastructure components, as they are necessary to troubleshoot in case of any performance or technical issues.

For information about the hardware prerequisites, refer to [Hardware Prerequisites](docs/installation_guide.md#hardware-prerequisites).

The prerequisites to deploy the Cassandra-Operator are as follows:

* Deployer user (SA) must have the following Role bound:
  
  ```
  apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: nc-role
  rules:
    - apiGroups:
        - netcracker.com
      resources:
        - '*'
      verbs:
        - create
        - get
        - list
        - patch
        - update
        - watch
        - delete
  ```

* The project or namespace should be created.
* The cloud administrator should create the Custom Resource Definition (CRD).
* If Dynamic Volume Provisioning is not available, all the persistent volumes should be created manually.
* If pre-created PVs are used in OpenShift, the project must be annotated with the same UID that is used for the PV.
* If deployed to OpenShift with restricted SCC, the project supplemental group annotation must have same UID as in the parameters **podSecurityContext.runAsUser** and **podSecurityContext.fsGroup**.
  ```
  For the default values of podSecurityContext.runAsUser and podSecurityContext.fsGroup, the supplemental group annotation should be set as shown below:

    oc annotate --overwrite ns cassandra openshift.io/sa.scc.supplemental-groups=999/999

  ```
* If the Pod Security Policy is enabled on the Kubernetes (K8s) cluster, it is mandatory to set the **podSecurityContext.fsGroup** and **podSecurityContext.runAsUser** parameters. For more information, refer to [https://kubernetes.io/docs/concepts/policy/pod-security-policy/](https://kubernetes.io/docs/concepts/policy/pod-security-policy/).
* In case of Prometheus Monitoring stack deployment, you need have the rights to create the **integreatly.org/v1alpha1** and **monitoring.coreos.com/v1** objects.

The following is an example of such role:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: generic-monitoring-role
rules:
  - apiGroups:
      - "monitoring.coreos.com"
    resources:
      - servicemonitors
      - prometheusrules
    verbs:
      - get
      - list
      - create
      - update
      - delete
      - watch
      - patch
  - apiGroups:
      - "integreatly.org"
    resources:
      - grafanadashboards
    verbs:
      - get
      - list
      - create
      - update
      - delete
      - watch
      - patch

```

###  Apply New Custom Resource Definition Version

The process of applying new custom resource definition version is described in the below sections.

#### Automated CRD Upgrade

The upgrade of CRD happens automatically through the pre-deploy scripts.

**Note**: Automation CRD upgrade requires the following cluster rights for the deploy user:

```yaml
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "create", "patch"]
```

To disable this feature, add the `DISABLE_CRD: true` parameter(The DISABLE_CRD option is supported from release 2.16.0 onward).

**Note**: `DISABLE_CRD=true` is intended only for restricted environments where cluster-wide permissions (`apiextensions.k8s.io`) are unavailable. **This flag is deprecated for all other deployment scenarios.** When CRDs are delivered as a dedicated CRD application, use that application instead of disabling CRD installation. See [Installation](#installation) for the recommended CRD application deployment order.

#### Manual CRD Upgrade

You can find multiple CRDs in the `charts/helm/cassandra-operator/` directory:

* `crds/crd.yaml` - CRD for Kubernetes version 1.22+ and OpenShift 4.9+.

Apply the new version of CRD using the following command:

`kubectl replace -f charts/helm/cassandra-operator/crds/<crd_name>.yaml`

Specify `DISABLE_CRD=true;` in the `CUSTOM_PARAMS` parameter of the App Deployer Job.

### Host

Cassandra provides recommended host configuration settings for the production. For more details, refer to the official documentation links mentioned at the end of this section.

The following host prerequisites have to be implemented before the deployment: 

1. Synchronize clocks - Synchronize the clocks on all nodes, using NTP (Network Time Protocol) or other methods. This is required because Cassandra only overwrites a column if there is another version whose timestamp is more recent.

2. Increase user resource limits - All containers, by default, inherit user limits from the Docker daemon. In the production environments, DSE expects the following changes to ulimit:

```
ulimit -n 100000 # nofile: max number of open files
ulimit -l unlimited # memlock: maximum locked-in-memory address space
```

Add the following line to /etc/sysctl.conf:

```
vm.max_map_count = 1048575
```

For installations on Debian and Ubuntu operating systems, the `pam_limits.so` module is not enabled by default. Edit the **/etc/pam.d/su** file and uncomment this line:

```
session    required   pam_limits.so
```

This change to the PAM configuration file ensures that the system reads the files in the /etc/security/limits.d directory.
To make the changes to take effect, reboot the server or run the following command:

```
sudo sysctl -p
```

To confirm that the limits are applied to the Cassandra process, run the following command, where pid is the process ID of the currently running Cassandra process:

```
cat /proc/pid/limits
```

3. TCP settings - To handle thousands of concurrent connections used by Cassandra, DataStax recommends these settings to optimize the Linux network stack. Add these settings to /etc/sysctl.conf.

```
net.core.rmem_max = 16777216
        net.core.wmem_max = 16777216
        net.core.rmem_default = 16777216
        net.core.wmem_default = 16777216
        net.core.optmem_max = 40960
        net.ipv4.tcp_rmem = 4096 87380 16777216
        net.ipv4.tcp_wmem = 4096 65536 16777216
```

To set immediately (depending on your distribution):

```
sudo sysctl -p /etc/sysctl.conf
sudo sysctl -p /etc/sysctl.d/filename.conf
```

Make sure that new settings persist after the reboot.

**Caution**: Depending on your environment, some of the following settings may not be persisted after the reboot. Check with your system administrator to ensure that they are viable for your environment.

4. Disable zone_reclaim_mode on NUMA systems - The Linux kernel can be inconsistent in enabling/disabling zone_reclaim_mode. This can result into odd performance problems.
Random huge CPU spikes resulting in large increases in latency and throughput.
Programs hanging indefinitely apparently doing nothing.
Symptoms appearing and disappearing suddenly.
After a reboot, the symptoms generally do not show again for some time.
To ensure that zone_reclaim_mode is disabled, run the following command:

```
echo 0 > /proc/sys/vm/zone_reclaim_mode
```

5. Check the Java Hugepages setting - Many modern Linux distributions ship with Transparent Hugepages enabled by default. When Linux uses Transparent Hugepages, the kernel tries to allocate memory in large chunks (usually 2MB), rather than 4K. This can improve performance by reducing the number of pages the CPU must track. However, some applications still allocate the memory based on 4K pages. This can cause noticeable performance problems when Linux tries to defrag 2MB pages. To disable defrag, run the following command on the Docker host:

```
echo never | sudo tee /sys/kernel/mm/transparent_hugepage/defrag
```

Links:

https://docs.datastax.com/en/cassandra-oss/3.x/cassandra/install/installRecommendSettings.html

https://docs.datastax.com/en/dse/5.1/dse-dev/datastax_enterprise/config/configRecommendedSettings.html

https://docs.datastax.com/en/docker/doc/docker/dockerRecommendedSettings.html


## HWE

Cassandra services resources can be selected during deployment using parameter `global.profile` that takes values: `small`, `medium`, `large`.

### Small

Recommended for development purposes, PoC, and demos.


| Module                   | CPU Requests | RAM Requests | CPU Limits | RAM Limits | Storage, Gb |
| ------------------------ | ------------ | ------------ | ---------- | ---------- | ----------- |
| cassandra                | 250m         | 1Gi          | 500m       | 2Gi        | 5Gi         |
| operator                 | 50m          | 64Mi         | 100m       | 128Mi      | -           |
| backup                   | 150m         | 256Mi        | 250m       | 512Mi      | 10Gi        |
| dbaas                    | 20m          | 32Mi         | 100m       | 128Mi      | -           |
| prometheusExporter       | 200m         | 128Mi        | 300m       | 256Mi      | -           |
| robotTests               | 200m         | 128Mi        | 200m       | 256Mi      | -           |
| Total                    | 2            | 4Gi          | 2620m      | 7Gi        | 25Gi        |

### Medium

Recommended for deployments with average load.

| Module                   | CPU Requests | RAM Requests | CPU Limits | RAM Limits | Storage, Gb |
| ------------------------ | ------------ | ------------ | ---------- | ---------- | ----------- |
| cassandra                | 1            | 2Gi          | 2          | 4Gi        | 50Gi        |
| operator                 | 50m          | 64Mi         | 100m       | 128Mi      | -           |
| backup                   | 150m         | 256Mi        | 1          | 1Gi        | 100Gi       |
| dbaas                    | 20m          | 32Mi         | 100m       | 128Mi      | -           |
| prometheusExporter       | 200m         | 128Mi        | 300m       | 256Mi      | -           |
| robotTests               | 200m         | 128Mi        | 200m       | 256Mi      | -           |
| Total                    | 4            | 7Gi          | 7          | 14Gi       | 250Gi       |

### Large

Recommended for deployments with high workload and large amount of data.

| Module                   | CPU Requests | RAM Requests | CPU Limits | RAM Limits | Storage, Gb |
| ------------------------ | ------------ | ------------ | ---------- | ---------- | ----------- |
| cassandra                | 2            | 4Gi          | 8          | 16Gi       | 500Gi       |
| operator                 | 50m          | 64Mi         | 100m       | 128Mi      | -           |
| backup                   | 150m         | 256Mi        | 2          | 2Gi        | 1000Gi      |
| dbaas                    | 20m          | 32Mi         | 100m       | 128Mi      | -           |
| prometheusExporter       | 200m         | 128Mi        | 300m       | 256Mi      | -           |
| robotTests               | 200m         | 128Mi        | 200m       | 256Mi      | -           |
| Total                    | 7            | 13Gi         | 27         | 51Gi       | 2TB         |


# Parameters

The [values.yaml](charts/helm/cassandra-operator/values.yaml) file contains the description and example of each parameter and its default value. Some parameters are self-explanatory.

The following sections provide the list of parameters.

**Note**: The parameters in the bold are mandatory.

## Operator

The list of Operator parameters is as follows:

| Parameter                                                        |  Mandatory | Type     | Default 			| Description 																															   |
|------------------------------------------------------------------|------------|----------|--------------------|------------------------------------------------------------------------------------------------------------------------------------------|
| `tls.enabled`                                                    | false      | bool     | false        	    | If TLS encryption should be enabled.                                                                                 |
| `tls.optional`                                                   | false      | bool     | false        	    | If TLS encryption is optional. If `true`, the Cassandra cluster accepts TLS and non-TLS connections.             |
| `tls.rootCASecretName`                                           | false      | string   | root-ca            | The name of the Kubernetes secret that holds a CA certificate, a Signed Cassandra certificate, and a private key.           |
| `tls.rootCAFileName`                                             | false      | string   | ca.crt        	    | The key in the Kubernetes secret `tls.rootCASecretName` that holds the CA certificate.                                 |
| `tls.privateKeyFileName`                                         | false      | string   | tls.key            | The key in the Kubernetes secret `tls.rootCASecretName` that holds the private key.                                    |
| `tls.signedCRTFileName`                                          | false      | string   | tls.crt            | The key in the Kubernetes secret `tls.rootCASecretName` that holds the Signed Cassandra certificate.                   |
| `tls.keystorePass`                                               | false      | string   | cassandra          | The password to Cassandra keystore.                                                                                    |
| `tls.generateCerts.enabled`                                      | false      | bool     | false              | If certificate needs to be generated by Cert Manager.|
| `tls.generateCerts.clusterIssuerName`                            | false      | string   | ""                 | The cluster issuer name to generate certificate by Cert Manager.|
| `tls.generateCerts.duration`                                     | false      | string   | 365                | The certificate validity period. |
| `tls.generateCerts.subjectAlternativeName.additionalDnsNames`    | false      | string[] | [ ]                | The additional DNS names to be set in certificate. |
| `tls.generateCerts.subjectAlternativeName.additionalIpAddresses` | false      | string[] | [ ]                | The additional IP addresses to be set in certificate.|
| `tls.generateSelfSignedCRTSecret`                                | false      | bool     | false        	    | If a Kubernetes secret with TLS assets should be created for testing.                                                |
| `waitTimeout`                                                    | false      | int      | 700        	      | The timeout of main installation steps.                                                                              |
| `gocqlConnectTimeout`                                            | false      | int      | 20           	    | The connect timeout in seconds for gocql driver (Dbaas adapter and Backup daemon)                                                                   |
| `gocqlTimeout`                                                   | false      | int      | 20        	        | The (operations) timeout in seconds for gocql driver (Dbaas adapter and Backup daemon)                                                    |
| `deletePVConUninstall`                                           | false      | bool     | false       	    | If PVCs needs to be deleted when `helm uninstall` executed.                                                                              |
| `pvc.metadata.annotations`                                       | false      | map[string]string |        | The annotations to add to all PVC metadata.                                                                                              |
| `imagePullPolicy`                                                | false      | string   | IfNotPresent       | If the image should be pulled prior to starting the container. Values: Always - always pull the image; IfNotPre  |
| `operator.name`                                                  | false      | string   | operator-service   | The name of the operator.                                                                                            |
| `operator.resources.requests.cpu`                                | false      | Quantity | 50m        		| The minimum number of CPUs the operator should use.                                                                  |
| `operator.resources.requests.memory`                             | false      | Quantity | 64Mi        		| The minimum amount of memory the operator should use.                                                                |
| `operator.resources.limits.cpu`                                  | false      | Quantity | 100m        		| The maximum number of CPUs the operator can use.                                                                               |
| `operator.resources.limits.memory`                               | false      | Quantity | 128Mi       		| The maximum amount of memory the operator can use.                                                                   |
| `operator.priorityClassName`                                     | false      | string   |         			| The priority class for an operator pod.                                                                              |
| `operator.nodeLabels`                                            | false      | string   |                    | The node selectors for an operator pod. |
| `securityContext.fsGroup`                                        | false      | int      | 999        		| The fsGroup of containers. It should be used in case of Kubernetes installation and the value must be set to `999`.      |
| `securityContext.runAsGroup`                                     | false      | int      | 999               | The group to run container.                                                                                          |
| `securityContext.runAsUser`                                      | false      | int      | 999        		| The user to run the container under.                                                                                     |
| `securityContext.supplementalGroups`                             | false      | int      |         			| The supplementalGroups of containers.                                                                                |
| `policies.tolerations[$idx].key`                                 | false      | string   | key        		| The taint key the toleration applies to.                                                                             |
| `policies.tolerations[$idx].operator`                            | false      | string   | operator        	| The key relationship to the value.                                                                                   |
| `policies.tolerations[$idx].value`                               | false      | string   | value        	    | The taint value the toleration matches to.                                                                           |
| `policies.tolerations[$idx].effect`                              | false      | string   | NoSchedule         | The taint effect to match.                                                                                           |
| `policies.tolerations[$idx].tolerationSeconds`                   | false      | int      |         			| The period the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint.  |

## Cassandra

The list of Cassandra parameters is as follows:

| Parameter                                                                     | Mandatory   | Type                | Default    | Description 																																															      |
|-------------------------------------------------------------------------------|-------------|---------------------|---------   |------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `cassandra.install`                                                           | false       | bool    			| true       |	If Cassandra needs to be installed.                                                                                                                                                   |
| `cassandra.ipV6`                                                              | false       | bool    			| false      | If ipV6 needs to be enabled.                                                                                                                                                           |
| `cassandra.hostNework`                                                        | false       | bool    			| false      | If the hostNetwork needs to be used for Cassandra pods.                                                                                                                                |
| `cassandra.auditLogEnabled`                                                   | false       | bool    			| false      | If audit logging is enabled. The default value is "false".                                                                                                                               |
| `cassandra.dataCenters[$idx].name`                                            | false       | []string  			|            | The name of the datacenter.                                                                                                                                                            |
| `cassandra.deploymentSchema.dataCenters[$idx].deploy`                         | false       | bool     			|            | If the datacenter needs to be deployed. The default is set to "True".                                                                                                                    |
| `cassandra.deploymentSchema.dataCenters[$idx].replicas`                       | false       | int     			| 3          | The number of Cassandra replicas in the datacenter.                                                                                                                                    |
| `cassandra.deploymentSchema.dataCenters[$idx].seeds`                          | false       | int     			| 1          | The number of Cassandra seeds in the datacenter.                                                                                                                                       |
| `cassandra.deploymentSchema.dataCenters[$idx].seedList`                       | false       | []string			|            | The exact list of seeds. If this parameter is set, the cassandra.deploymentSchema.dataCenters[$idx].seeds parameter is ignored.                                                        |
| `cassandra.deploymentSchema.dataCenters[$idx].racks`                          | false       | []string			|            | The list of racks in the datacenter the Cassandra replicas should be deployed to. All replicas are deployed to rack1 by default.                                                       |
| `cassandra.deploymentSchema.dataCenters['dc1'].clusterDomain`                 | false       | string  			|            | The DNS name of the Kubernetes cluster. For example, in the address cassandra0-0.cassandra-service.cassandra-namespace.svc.cluster.local, the cluster.local is cluster domain. It is only applicable for deployment to multiple Kubernetes. |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.waitPvcBound`           | false       | bool    			|            | If the operator needs to wait for PVC binding. The default is set to "false".                                                                                                                |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.size`                   | false       | []string			| 5Gi        | The size of PVCs in the datacenter.                                                                                                                                                    |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.volumes`                | false       | []string			|            | The list of persistence volumes in the datacenter. For auto-provision, leave the value blank for this parameter.                                                                       |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.nodeLabels`             | false       | []map[string]string |            | The labels to map the replicas to the nodes. To set the node name, use the kubernetes.io/hostname label.                                                                                       |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.storageClasses`         | false       | []string  			|            | The name of the storage class that needs to be used for PVC creation. For hostpath PVs, leave the value blank for this parameter.                                                      |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.matchLabelSelectors`    | false       | []map[string]string	|            | The key:value pair of PVs to bind to Cassandra PVCs.                                                                                                                                   |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.mountSettings.name`     | false       | sting      			|            | The name of the volume that should be mounted with the Cassandra replica.                                                                                                               |
| `cassandra.deploymentSchema.dataCenters[$idx].storage.mountSettings.mountPath`| false       | string     			|            | The path of the volume that should be mounted to the Cassandra replica.                                                                                                                 |
| `cassandra.flavor`                              | false     | string            | small   | The flavor of Cassandra deployment resources. Possible values are  `small`, `medium`, `large`.           |
| `cassandra.resources.requests.cpu`                                            | false       | Quantity    		| 250m       | The minimum number of CPUs Cassandra should use.                                                                                                                                   |
| `cassandra.resources.requests.memory`                                         | false       | Quantity    		| 1Gi        | The minimum amount of memory Cassandra should use.                                                                                                                                 |
| `cassandra.resources.limits.cpu`                                              | false       | Quantity 			| 500m       | The maximum number of CPUs Cassandra can use.                                                                                                                                          |
| `cassandra.resources.limits.memory`                                           | false       | Quantity 			| 2Gi        | The maximum amount of memory Cassandra can use.                                                                                                                                        |
| `cassandra.priorityClassName`                                                 | false       | string   			|            | The priority class for Cassandra replicas.                                                                                                                                             |
| `cassandra.pdb`                                                               | false       | bool       			| false      | If Cassandra Pod Disruption Budgets need to be created.                                                                                                                                |
| `cassandra.username`                                                          | false       | string     			| admin      | The Cassandra database username in secret. Do not use the symbol "-" or capital letters in this parameter value.                                                                           |
| `cassandra.password`                                                          | false       | string     			| admin      | The Cassandra database password in the secret. Do not use the symbol "-" in this parameter value.                                                                                              |
| `cassandra.commitlogArchiving.enabled`                                        | false       | boolean             | false                                                           | If archiving of commit logs need to be enabled.                                                                                                           |
| `cassandra.commitlogArchiving.archive_command`                                | false       | string              | /bin/cp -f %path /var/lib/cassandra/commitlog_archives/archives/%name | The command to archive a commit log segment.                                                                                                         |
| `cassandra.commitlogArchiving.restore_command`                                | false       | string              | /bin/cp -f %from %to                                                  | The command to restore an archived commit log.                                                                                                       |
| `cassandra.commitlogArchiving.restore_directories`                            | false       | string              | /var/lib/cassandra/commitlog_archives/archives/                       | The restore directory location.                                                                                                                      |
| `cassandra.commitlogArchiving.storage.size`                                   | false       | string              |                                                                 | The size of the PVC for the commitlog volume.                                                                                                             |
| `cassandra.commitlogArchiving.storage.storageClasses`                         | false       | []string            |                                                                 | The storage class for PVC creation for the commitlog volume. For hostpath PVs, leave the value blank.                                                    |
| `cassandra.commitlogArchiving.storage.volumes`                                | false       | []string            |                                                                 | The list of pre-created persistent volumes for the commitlog. For auto-provision, leave the value blank.                                                  |
| `cassandra.commitlogArchiving.storage.nodeLabels`                             | false       | []map[string]string |                                                                 | The labels to map the commitlog PVC to specific nodes. To set the node name, use the `kubernetes.io/hostname` label.                                      |
| `cassandra.commitlogArchiving.storage.matchLabelSelectors`                    | false       | []map[string]string |                                                                 | The key:value pair of PVs to bind to the commitlog PVC.                                                                                                   |
| `cassandra.commitlogArchiving.storage.waitPvcBound`                           | false       | bool                | false                                                           | If the operator needs to wait for PVC binding before proceeding.                                                                                          |

## Cassandra Reaper

The list of Cassandra Reaper parameters is as follows:

| Parameter            				 |	Mandatory| Type              | Default    | Description 	                                                                                                                                                                                                                                                                |
|-----------------------------------|-----------|-------------------|---------   |---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `reaper.install`     				 | false     | bool              | false      |   If the Cassandra Reaper needs to be installed.                                                                                                                                                                                                            |
| `reaper.port`     				     | false     | int               | 8080/8443 | Cassandra Reaper port.                                                                                                                                                                                                                                        |
| `reaper.username`    				 | false     | string            | admin      |   The Cassandra Reaper username to access the Web UI. Do not use the symbol "-" in this parameter value.                                                                                                                                                            |
| `reaper.password`    				 | false     | string            | admin      |   The Cassandra Reaper password to access the Web UI. Do not use the symbol "-" in this parameter value.                                                                                                                                                            |
| `reaper.ingressHost` 				 | false     | string            |            |   The ingress host address to access the Web UI of Cassandra Reaper.                                                                                                                                                                                            |
| `reaper.envs`        				 | false     | map[string]string |            |   The map of environment variables of the Cassandra Reaper container. For more information about additional environment variables, refer to [https://hub.docker.com/r/thelastpickle/cassandra-reaper/](https://hub.docker.com/r/thelastpickle/cassandra-reaper/) |
| `reaper.resources.limits.memory`  | false     | Quantity          | 2Gi        |   The memory limit of the Cassandra Reaper replica.                                                                                                                                                                                                             |
| `reaper.resources.limits.cpu`     | false     | Quantity          | 5m00       |   The CPU limit of the Cassandra Reaper replica.                                                                                                                                                                                                                |
| `reaper.resources.requests.memory` | false     | Quantity          | 256Mi      |   The memory request of the Cassandra Reaper replica.                                                                                                                                                                                                           |
| `reaper.resources.requests.cpu`   | false     | Quantity          | 150m       |   The CPU request of the Cassandra Reaper replica.                                                                                                                                                                                                              |

Example of reaper.envs setting:

```
reaper:
  envs:
    REAPER_REPAIR_RUN_THREADS: 10
    REAPER_REPAIR_PARALELLISM: SEQUENTIAL
```

## DBaaS Cassandra Adapter

The list of DBaaS Cassandra Adapter parameters is as follows:

| Parameter                                               | Mandatory   | Type                | Default                                             | Description 																																															                                                                                       |
|-------------------------------------------------------  |-------------|---------------------|-----------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `dbaas.pdb`                                             | false       | bool                | false                                               | IF PDB needs to be used. |
| `dbaas.install`                                         | false       | bool                | true                                                | If DBaaS adapter needs to be installed.                                                                                                                                                                                                                                  |
| `dbaas.apiVersion`                                      | false       | string              | v1                                                  | The DBaaS adapter REST API version.                                                                                                                                                                                                                                          |
| `dbaas.multiUsers`                                      | false       | bool                | false                                               | If DBaaS adapter needs to create multiple users on a Create DB request.                                                                                                                                                                                                    |
| `dbaas.allDCTopologyStrategy`                           | false       | bool                | false                                               | If keyspaces created through DBaaS should have full NetworkTopologyStrategy. The default values is "false".                                                                                                                                                                  |
| `dbaas.topologyStrategy`                                | false       | string              | {'class':'SimpleStrategy','replication_factor': 1 } | The default topology strategy for keyspaces created using DBaaS. The default values is `"{'class':'SimpleStrategy','replication_factor': 1 }"`. This parameter is ignored if `dbaas.allDCTopologyStrategy` is `true`. The value can be overridden in a request body to DBaaS.|
| `dbaas.dbaasStreamingRoleName`                          | false       | string              | streaming											| The role name for streaming. |
| `dbaas.dbaasStreamingRoles`                             | false       | string[]            | ALL                                                 | The set of permissions for a streaming role. |
| `dbaas.tls.dbaasAdapterCASecretName`                    | false       | string              | dbaas-adapter-certificate                           | The name of the secret to store a certificate. |
| `dbaas.tls.duration`                                    | false       | string[]            | 365                                                 | The certificate validity period. |
| `dbaas.tls.subjectAlternativeName.additionalDnsNames`   | false       | string[]            | [ ]                                                 | The additional DNS names to be set in a certificate. |
| `dbaas.tls.subjectAlternativeName.additionalIpAddresses`| false       | string[]            | [ ]                                                 | The additional IP addresses to be set in a certificate. |
| `dbass.externalCassandra.enabled`                       | false       | bool                | false                                               | The DBaaS adapter needs to connect to an external Cassandra cluster.|
| `dbass.externalCassandra.host`                          | false       | string              |                                                     | The external Cassandra cluster hostname. |
| `dbass.externalCassandra.port`                          | false       | string              |                                                     | The external Cassandra cluster port. |
| `dbass.externalCassandra.username`                      | false       | string              |                                                     | The external Cassandra cluster username. |
| `dbass.externalCassandra.password`                      | false       | string              |                                                     | The external Cassandra cluster password. |
| `dbass.externalCassandra.defaultKeyspace`               | false       | string              | system                                              | The external Cassandra cluster keyspace for DBaaS adapter to connect to. |
| `dbass.externalCassandra.consistency`                   | false       | string              | QUORUM                                              | The consistency for connection to an external Cassandra cluster. |
| `dbass.externalCassandra.useTLS`                        | false       | bool                | true                                                | If an external Cassandra cluster uses TLS. |
| `dbaas.resources.limits.memory`                         | false       | Quantity            | 64Mi                                                | The memory limit of a DBaaS adapter replica.                                                                                                                                                                                                                               |
| `dbaas.resources.limits.cpu`                            | false       | Quantity            | 20m                                                 | The CPU limit of a DBaaS adapter replica.                                                                                                                                                                                                                                  |
| `dbaas.resources.requests.memory`                       | false       | Quantity            | 32Mi                                                | The memory request of a DBaaS adapter replica.                                                                                                                                                                                                                             |
| `dbaas.resources.requests.cpu`                          | false       | Quantity            | 20m                                                 | The CPU request of a DBaaS adapter replica.                                                                                                                                                                                                                                |
| `dbaas.priorityClassName`                               | false       | string              |                                                     | The priority class for a DBaaS adapter replica.                                                                                                                                                                                                                            |
| `dbaas.nodeLabels`                                      | false       | map[string]string   |                                                     | The additional node labels for a DBaaS adapter replica.                                                                                                                                                                                                                    |
| `dbaas.adapter.username`                                | false       | string              | dbaas-aggregator                                    | The username for the database adapter.                                                                                                                                                                                                                                   |
| `dbaas.adapter.password`                                | false       | string              | dbaas-aggregator                                    | The password for the database adapter. Optional parameters for DBaaS installation if multiple physical databases for DBaaS are used.                                                                                                                                    |
| `dbaas.aggregator.physicalDatabaseIdentifier`           | false       | string              |                                                     | The database identifier in the DBaaS aggregator. The default is set to the Kubernetes namespace name.                                                                                                                                                                    |
| `dbaas.aggregator.dbaasAggregatorRegistrationAddress`   | false       | string              | http://dbaas-aggregator.dbaas:8080                  | The address of the aggregator where the adapter registers its physical database cluster. The default is set to ```http://dbaas-aggregator.dbaas:8080/```.                                                                                                                |
| `dbaas.aggregator.username`                             | false       | string              | cluster-dba                                         | The username for database registration.                                                                                                                                                                                                                                  |
| `dbaas.aggregator.password`                             | false       | string              | Bnmq5567_PO                                         | The password for database registration.                                                                                                                                                                                                                                  |

## Cassandra Backup Daemon

The list of Cassandra Backup Daemon parameters is as follows:

| Parameter                                                       |	Mandatory    | Type                | Default                          | Description 				                                                                                                                         |
|-----------------------------------------------------------------|--------------|---------------------|----------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| `backupDaemon.s3.endpointUrl`                                   | false        | string              |                                  | The S3 URL.                                                                                                                       |
| `backupDaemon.s3.accessKeyId`                                   | false        | string              |                                  | The S3 access key ID.                                                                                                             |
| `backupDaemon.s3.accessKeySecret`                               | false        | string              |                                  | The S3 access key secret.                                                                                                         |
| `backupDaemon.s3.enabled`                                       | false        | bool                |  false                           | If TLS encryption should be enabled.                                                                                          |
| `backupDaemon.s3.bucketName`                                    | false        | string              |                                  | The bucket name in an S3 storage to connect to. |
| `backupDaemon.username`                                         | false        | string              | backup                           | The Cassandra backup daemon API username. Do not use the symbol "-" in this parameter value.                                      |
| `backupDaemon.password`                                         | false        | string              | backup                           | The Cassandra backup daemon API password. Do not use the symbol "-" in this parameter value.                                      |
| `backupDaemon.nodeLabels`                                       | false        | []map[string]string |                                  | The labels to map the replicas to the nodes. To set the node name, use the `kubernetes.io/hostname` label.                            |
| `backupDaemon.resources.limits.memory`                          | false        | Quantity            | 384Mi                            | The maximum amount of memory for a backup-daemon replica.                                                                       |
| `backupDaemon.resources.limits.cpu`                             | false        | Quantity            | 250m                             | The maximum number of CPUs for a backup-daemon replica.                                                                         |
| `backupDaemon.resources.requests.memory`                        | false        | Quantity            | 128Mi                            | The minimum amount of memory for a backup-daemon replica.                                                                       |
| `backupDaemon.resources.requests.cpu`                           | false        | Quantity            | 150m                             | The minimum number of CPUs for a backup-daemon replica.                                                                         |
| `backupDaemon.priorityClassName`                                | false        | string              |                                  | The priority class for a backup-daemon replica.                                                                                 |
| `backupDaemon.storage.size`                                     | false        | string              | 5Gi                              | The size of PVCs.                                                                                                             |
| `backupDaemon.storage.waitPvcBound`                             | false        | bool                |                                  | If the operator needs to wait for PVC binding. The default is set to "false".                                                     |
| `backupDaemon.storage.storageClasses`                           | false        | []string            |                                  | The name of the storage class that needs to be used for PVC creation. For hostpath PVs, leave the value blank for this parameter. |
| `backupDaemon.storage.volumes`                                  | false        | []string            |                                  | The list of persistence volumes. For auto-provision, leave the value blank for this parameter.                                |
| `backupDaemon.storage.matchLabelSelectors`                      | false        | string              |                                  | The PV labels to be set as selectors in a backup PVC. 
| `backupDaemon.backupSchedule`                                   | false        | string              | 0 0 * * *                        | The backup schedule in cron pattern.|
| `backupDaemon.evictionPolicy`                                   | false        | string              | 0/1h,3d/7d,1m/1m,1y/delete       | The backup eviction policy. The default value is `"0/1h,3d/7d,1m/1m,1y/delete"`.                               |
| `backupDaemon.granularEvictionPolicy`                           | false        | string              | 7d/delete                        | The granular backup eviction policy. The default value is `"7d/delete"`.                                       |
| `backupDaemon.granularBackupSchedule`                           | false        | string              | 0 3 * * *                        | The granular backup schedule in cron pattern. |
| `backupDaemon.granularBackupScheduledDbs`                       | false        | []string            |                                  | The list of dbs for scheduled granular backup. |
| `backupDaemon.pdb`                                              | false        | bool                | false                            | If PDB is used.
| `backupDaemon.tls.backupDaemonCASecretName`                     | false        | string              | backup-daemon-certificate        | The secret name where a certificate is stored.
| `backupDaemon.tls.duration`                                     | false        | int                 | 365                              | The certificate validity period.
| `backupDaemon.tls.subjectAlternativeName.additionalDnsNames`    | false        | string[]            | [ ]                              | The additional DNS names to be set in a certificate.
| `backupDaemon.tls.subjectAlternativeName.additionalIpAddresses` | false        | string[]            | [ ]                              | The additional IP addresses to be set in a certificate.
| `backupDaemon.install`                                          | false        | bool                | true                             | If Cassandra Backup Daemon needs to be installed.
| `backupDaemon.cqlConnectTimeout`                                | false        | int                 | 10                               | The connect timeout from Cassandra Backup Daemon to the Cassandra cluster.
| `backupDaemon.cqlRequestTimeout`                                | false        | int                 | 20                               | The request timeout from Cassandra Backup Daemon to the Cassandra cluster.

## Monitoring Agent

The list of Monitoring Agent parameters is as follows:

| Parameter                                      					  |	Mandatory | Type   	       | Default                                                    | Description 				                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
|---------------------------------------------------------------------|-----------|----------------|------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `monitoringAgent.install`                                           | false     | bool           | true                                                       | If backup daemon needs to be installed.|
| `monitoringAgent.prometheus.metricRelabelings.cassandra[$idx]`      | false     | string         |                                                            | The `RelabelConfig` array for Prometheus ServiceMonitor to dynamically rewrite the Cassandra metrics. For more information, refer to the [https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#relabelconfig](https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#relabelconfig). For example of parameters, see [Cassandra Prometheus Metrics Relabeling](#cassandra-prometheus-metrics-relabeling). |
| `monitoringAgent.prometheus.alerts.common.cpuThreshold`             | false     | int            | 95                                                         | The Cassandra high CPU usage alert threshold.                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `monitoringAgent.prometheus.alerts.common.memThreshold`             | false     | int            | 95                                                         | The Cassandra high memory usage alert threshold.                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `monitoringAgent.prometheus.alerts.common.usedSpaceThreshold`       | false     | int            | 50                                                         | The Cassandra high disk usage alert threshold.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `monitoringAgent.prometheus.alerts.backup.usedSpaceThreshold`       | false     | int            | 80                                                         | The Cassandra Backup Daemon high disk usage alert threshold.                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `monitoringAgent.prometheus.alerts.backup.usedInodesThreshold`      | false     | int            | 80                                                         | The Cassandra Backup Daemon high Inodes usage alert threshold.                                                                                                                                                                                                                                                                                                                                                                                                                       |
| `monitoringAgent.prometheus.alerts.cassandra.nodeUnavailableWaitFor`| false     | string         | 5m                                                         | The timeout before Cassandra node is considered unavailable.                                                                                                                                                                                                                                                                                                                                                                                                                         |
| `monitoringAgent.nodeLabels`                                        | false     | string          |                                                            | The labels on a node.                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `monitoringAgent.resources.limits.memory`                           | false     | Quantity       | 128Mi                                                      | The maximum amount of memory for a monitoring agent replica.                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `monitoringAgent.resources.limits.cpu`                              | false     | Quantity       | 200m                                                       | The maximum number of CPUs for a monitoring agent replica.                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `monitoringAgent.resources.requests.memory`                         | false     | Quantity       | 256Mi                                                      | The minimum amount of memory for a monitoring agent replica.                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `monitoringAgent.resources.requests.cpu`                            | false     | Quantity       | 200m                                                       | The minimum number of CPUs for a monitoring agent replica.                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `monitoringAgent.priorityClassName`                                 | false     | string         |                                                            | The priority class for a monitoring agent replica.                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `monitoringAgent.monitoringInterval`                                | false     | string         | 20s                                                        | The monitoring interval in seconds.                                                                                                                                                                                                                                                                                                                                                                                   
## Consul Registration

The list of Consul Registration parameters is as follows:


| Parameter                             | Mandatory   | Type   	      | Default                 | Description 																																															   																					  |
|---------------------------------------|-------------|---------------|-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `consulRegistration.Enabled`          | false       | bool          | false                   | If the Consul services from the `consulDiscoverySettings` section registration is enabled or not.                                                                                                                                                                       |
| `consulRegistration.AclEnabled`       | false       | bool          | false                   | If the Consul ACL is enabled. The default value is "false".                                                                                                                                                                                                            |
| `consulRegistration.AuthMethod`       | false       | string        | consul-k8s-auth-method  | The name of the Consul authentication method. If the Consul has hardcoded authentication method name, then that is used as the default value for this parameter (`consul-k8s-auth-method`). Otherwise, the parameter value should be equal to the Consul authentication method name.|
| `consulRegistration.Port`             | false       | string        | 8500                    | The  Consul host. |                     
| `consulRegistration.Host`             | false       | string        |                         | The Consul host. If the value is not specified, the `nodeIp` is used. `consulRegistration.Port` is the Consul port. The default value is `8500`.                                                                                                       |

## Consul Discovery Settings

This section provides information about the Consul Discovery Settings.

### Cassandra Discovery Settings

The list of Cassandra Discovery Settings parameters is as follows:

| Parameter                                      						  | Mandatory 	| Type   	         | Default | Description 																																													|
|-------------------------------------------------------------------------|-------------|--------------------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `consulDiscoverySettings.cassandra.Enabled`                             | false       | bool               | false   | If the operator registers or deregisters Cassandra service in Consul. The value should be equal to "true" to register Cassandra service in Consul or "false" to deregister. |
| `consulDiscoverySettings.cassandra.Name`                                | false       | string             |         | The desired name that is used for Cassandra in Consul. If the value is not specified, the default value `{namespace}-cassandra` is used.                                     |
| `consulDiscoverySettings.cassandra.Tags`                                | false       | []string           |         | The tags for Cassandra in Consul.                                                                                                                                           |
| `consulDiscoverySettings.cassandra.Meta`                                | false       | map[string]string  |         | The meta for the Cassandra in Consul.                                                                                                                                           |
| `consulDiscoverySettings.cassandra.Check.DeregisterCriticalServiceAfter`| false       | string             | 100s    | The time after which the Service is deregistered from Consul.                                                                                                                    |
| `consulDiscoverySettings.cassandra.Check.Interval`                      | false       | string             | 10s     | The check interval for the Service check.                                                                                                                                   |
| `consulDiscoverySettings.cassandra.Check.Timeout`                       | false       | string             | 1s      | The check timeout of the Service check.                                                                                                                                     |

## Robot Tests

The list of Robot Tests parameters is as follows:

| Parameter                                | Mandatory | Type        | Default             | Description 																																									 	 |
|------------------------------------------|--------------|--------- |---------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `robotTests.install`                     | false        | bool     | false               | If Robot tests needs to be installed.                                                                                                                           |
| `robotTests.tags`                        | false        | string   |smokeORdbaasORbackup | The tags of Robot tests. The possible values are smoke, backup, DBaaS.                                                                                               |
| `robotTests.replicationFactor`           | false        | int      | 3                   | The replication factor for test keyspaces. It must be a value between 1 and Cassandra's first data center `cassandra.deploymentSchema.dataCenters[0].replicas` value.  |
| `robotTests.nodeLabels`                  | false        | string   |                     | The additional node labels for Robot tests' replica.                                                                                                             |
| `robotTests.resources.requests.cpu`      | false        | Quantity | 200m                | The CPU request of Robot tests' replica.                                                                                                                         |
| `robotTests.resources.requests.memory`   | false        | Quantity | 128Mi               | The memory request of Robot tests' replica.                                                                                                                      |
| `robotTests.resources.limits.cpu`        | false        | Quantity | 200m                | The CPU limit of Robot tests' replica.                                                                                                                           |
| `robotTests.resources.limits.memory`     | false        | Quantity | 256Mi               | The memory limit of Robot tests' replica.                                                                                                                        |

**Note**: If you run the full scope of the tests, the default timeout for the App Deployer job might not be enough. To override this timeout, use the optional `CUSTOM_TIMEOUT_MIN` parameter.

## Parameters Examples

The parameter examples for the different scenarios of deployment are given below.

### Cassandra Registration in Consul  

```
consulRegistration:
  AclEnabled: true
  Enabled: true
consulDiscoverySettings:
  cassandra:
    Check:
      DeregisterCriticalServiceAfter: 100s
      Interval: 15s
      Timeout: 1s
    Enabled: true
    Name: cassandra-custom
    Tags:
      - additionalTagTest
    Meta:
      additionalMeta1: metaValue1
      additionalMeta2: metaValue2
```

### Cassandra Prometheus Metrics Relabeling

#### Drop Metrics by Regex

```
monitoringAgent:
  install: true
  metricCollector: prometheus
  prometheus:
    metricRelabelings:
      cassandra:
        - action: drop
          regex: 'cassandra_table_disk_space_bytes'
          sourceLabels: [__name__]
        - action: drop
          regex: 'cassandra_another_metrics.*'
          sourceLabels: [__name__]
```

#### Keep Only Certain Tables

```
- action: keep
     regex: cyclist_name|
     sourceLabels:
       - table 
```

# Installation

the installation procedure is described in the below sections.

**Important!**
Starting with release 2024.3-2.0.0 Cassandra is delivered as 2 versions: one containing Cassandra operator and the second containing Cassandra Services Operator and Supplementary services (Cassandra Backup Daemon, Cassabdra Dbaas Adapter, Robot tests, Monitoring configuration).

Both versions are mentioned in the Release notes in App Deployer Versions section. They use the same set of parameters in CMDB as the previous versions.

In case of Clean Install Cassandra operator must be installed first, then Cassandra Services is installed with Rolling Update mode.
In case of Rolling Update from older versions the installation can be performed in any order, but Cassandra first is preferable.

**Note**: For Database services where CRDs are delivered as a dedicated CRD application (separate from the main microservice), the CRD application must be installed and fully synced before the main microservice installation begins. The main microservice deployment relies on the CRDs being pre-installed in the cluster. Do not install the main microservice until the CRD application sync is complete.

In restricted environments where cluster-wide permissions are unavailable, the dedicated CRD application must not be installed. Use `DISABLE_CRD=true` in such environments instead (see [Automated CRD Upgrade](#automated-crd-upgrade)).


### App Deployer

This section describes the deployment of Cassandra on Kubernetes/OpenShift using App Deployer.

You need to have the artifacts ready for App Deployer.

1. Choose the artifact version for the App Deployer (version as a string). <!-- #GFCFilterMarkerStart# --> You can get the version from [https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/releases](https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/releases).<!-- #GFCFilterMarkerEnd# --> Use this link as the value for the `ARTIFACT_DESCRIPTOR_VERSION` parameter.
1. Navigate to [groovy.deploy.v3](https://cloud-deployer.netcracker.com/job/INFRA/job/groovy.deploy.v3/).
1. Specify the values for the following parameters:
   * `PROJECT` - The target namespace for the installation procedure.
   * `OPENSHIFT_CREDENTIALS` - The credentials of the user on behalf of whom the deployment process needs to run.
   * `DEPLOY_MODE` - The mode of deployment, either `Rolling Update` or `Clean Install`. The `Clean Install` mode deletes everything from the project before deployment.
   * `ARTIFACT_DESCRIPTOR_VERSION` - The version as a string. <!-- #GFCFilterMarkerStart# --> You can get the version from [https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/releases](https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/releases).<!-- #GFCFilterMarkerEnd# -->
   * `CUSTOM_PARAMS` - The custom parameters that overwrite values from **values.yaml**. For more information, see [Deployment Parameters](#deployment-parameters) and [Parameters Examples](#parameters-examples).
   * `DEPLOY_W_HELM: true` - It must be set in `CUSTOM_PARAMS` or in CMDB.
1. Click `Build` button to start deployment.

**Note**: It is also possible to set `--skip-crds`.

### Helm

Before you start with the manual deployment of Cassandra service using Helm, ensure you have Helm 3 release.

Alternatively, you can install the operator using the Helm chart from the `charts/helm/cassandra-operator` folder.
Install the Helm CLI on your machine. For more information about Helm v3.0.0, refer to [https://github.com/helm/helm/releases/tag/v3.0.0](https://github.com/helm/helm/releases/tag/v3.0.0).

To use a CRD-based configuration, specify the **values.yaml** file as shown in the following example.

Also, you have to specify proper microservice images in the **values.yaml** file. The list of microservices can be found in `Microservice versions` section of each release. 
<!-- #GFCFilterMarkerStart# -->
The Releases page can be found at [https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/releases](https://git.netcracker.com/PROD.Platform.Databases/cassandra-operator/-/releases).
<!-- #GFCFilterMarkerEnd# --> 
Follow each microservice tag link and find the `Artifacts` section with the **Docker image** name.

Note that only the set of microservices' versions from one preferred release should be used. Otherwise, the stable working of microservices is not guaranteed.

1. Clone the project to a local machine using the following command. 
 
   ```git clone git@git.netcracker.com/PROD.Platform.Databases/cassandra-operator.git```

2. Navigate to the Cassandra operator directory using the following command. 

   ```cd charts/helm/cassandra-operator```

3. Login to OpenShift using the following command.

   ```oc login https://openshift_url:8443```

4. Deploy the operator using `helm install cassandra-operator`.

#####  Helm Job Parameters

The list of Helm job parameters is as follows:

The `CLOUD_URL` is the URL of the OpenShift or Kubernetes server. For example, ```https://search.openshift.sdntest.example.com:6443```. It is a **mandatory** parameter.

The `CLOUD_NAMESPACE` is the OpenShift project name. It is a **mandatory** parameter. 

The `CLOUD_TOKEN` is the cloud token. It is a **mandatory** parameter. 

The `DESCRIPTOR_URL` is the URL to Application Manifest (.mf) or Deployment Descriptor (.json, .zip). It is a **mandatory** parameter.

The `DEPLOYMENT_PARAMETERS` is the **custom_values.yml** content.

The `DEPLOYMENT_MODE` is the main helm command: `install` or `upgrade`. The `auto` mode tries to determine the mode automatically. It is a **mandatory** parameter.

## On-Prem

### HA Scheme

An example of parameters of HA scheme is as follows:

```
cassandra:
  install: true
  deploymentSchema:
    dataCenters:
      - name: dc1
        replicas: 3
        seeds: 1
        storage:
          nodeLabels:
            - 'kubernetes.io/hostname': dr311-arbiter-node-left-1
            - 'kubernetes.io/hostname': dr311-arbiter-node-left-2
            - 'kubernetes.io/hostname': dr311-arbiter-node-arb-left-1
          size: 1Gi
          volumes:
            - cassandra-0
            - cassandra-1
            - cassandra-2
          storageClasses:
            - ""
backupDaemon:
  install: true
  storage:
    nodeLabels:
      - 'kubernetes.io/hostname': dr311-arbiter-node-right-1
    size: 1Gi
    volumes:
      - cassandra-backup
dbaas:
  install: true
monitoringAgent:
  install: true
robotTests:
  install: true
```

### DR

#### Standard DR Scheme

This section describes the deployment of Cassandra on two separate Kubernetes instances with Pod-to-Pod IP connectivity between them.

The following image depicts a high-level architecture of Cassandra deployment.

![Deployment Architecture](/docs/public/images/calico.png)

Key notes:

* The BGP and Calico CNI plugins must be configured to provide Pod-to-Pod connectivity between Kubernetes clusters.
* A separate deployment job must be run consecutively.
* The Backup daemon takes backups from all Cassandra replicas across all Kubernetes clusters.

To deploy Cassandra pods with Calico:

1. Run the deployment job [App Deployer](#app-deployer) on the first Kubernetes instance. For more information, see [calico_deploy_dc1.yaml](examples/calico_deploy_dc1.yaml):

   * The `cassandra.ipV6` parameter must be set to `true` if IPv6 protocol is used.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].clusterDomain` parameter must be set to the DNS name of the first Kubernetes.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `true`.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `false`.

2. Run the deployment job [App Deployer](#app-deployer) on the second Kubernetes instance. For more information, see [calico_deploy_dc2.yaml](examples/calico_deploy_dc2.yaml):

   * The `cassandra.ipV6` parameter must be set to `true` if IPv6 protocol is used.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].clusterDomain` parameter must be set to the DNS name of the second Kubernetes.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `false`.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `true`.

**Note**: The `cassandra.deploymentSchema.dataCenters['<dc_name>'].deploy` parameter changes between installations.

#### Deployment on Two Kubernetes Instances With HostNetwork

This section describes the deployment of Cassandra on two separate Kubernetes instances with hostNetwork.

##### Deploy Cassandra Pods With HostNetwork

To deploy Cassandra pods with hostNetwork:

1. In the PSP used for Cassandra deployment, `oob-host-network-psp`, you need to set `hostNetwork: true` and the `hostPorts` range should include the ports 9042, 7199, and 7000.

2. Create Role and Role binding to bind the role to a service account that is used to deploy cassandra-operator:

   **Role**
   
   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: Role
   metadata:
     name: cassandra-hostnetwork-role
     namespace: cassandra
   rules:
   - apiGroups:
     - policy
     resourceNames:
     - oob-host-network-psp
     resources:
     - podsecuritypolicies
     verbs:
     - use
   ```

   **Role binding**

   ```
   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
     name: cassandra-hostnetwork-role-binding
     namespace: cassandra
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: Role
     name: cassandra-hostnetwork-role
   subjects:
   - kind: ServiceAccount
     name: cassandra-operator
   ```

3. Run the deployment job [App Deployer](#app-deployer) on the first Kubernetes instance. For more information, see [hostnetwork_deploy_dc1.yaml](examples/hostnetwork_deploy_dc1.yaml):

   * The `cassandra.hostNetwork` parameter must be set to `true`.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `true`.
   * The `cassandra.deploymentSchema.dataCenters[$idx].seedList` parameter should include at least one IP address from each Kubernetes cluster node to which Cassandra is deployed.
    * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `false`.

4. Run the deployment job [App Deployer](#app-deployer) on the second Kubernetes instance. For more information, see [hostnetwork_deploy_dc2.yaml](examples/hostnetwork_deploy_dc2.yaml):

   * The `cassandra.hostNetwork` parameter must be set to `true`.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `false`.
   * The `cassandra.deploymentSchema.dataCenters[$idx].seedList` parameter should include at least one IP address from each Kubernetes cluster node to which Cassandra is deployed.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `true`.

**Note**: The `cassandra.deploymentSchema.dataCenters['<dc_name>'].deploy` parameter changes between installations.

##### Specify Listen and Broadcast Addresses Multi-Site

The Cassandra parameters are as follows:

The `listen_address` parameter specifies the IP address that Cassandra binds to for connecting to other Cassandra nodes. For a single node cluster, you can use the default setting (localhost). Never specify 0.0.0.0; it is always incorrect. For more information, refer to the official documentation at [https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__listen_address](https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__listen_address).

The `broadcast_address` parameter specifies the IP address using which a node should contact other nodes in the cluster. It allows the public and private address to be different. For example, use the `broadcast_address` parameter in topologies where not all nodes have access to other nodes by their private IP addresses. If this section is absent, it is set to the value of the `listenAddress`. For more information, refer to the official documentation at [https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__broadcast_address](https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__broadcast_address).

The `rpc_address` parameter specifies the listen address for client connections. It should be IP address, hostname, or 0.0.0.0 to listen on all configured interfaces, but you must set the `broadcast_rpc_address` to a value other than 0.0.0.0. For more information, refer to the official documentation at [https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__rpc_address](https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__rpc_address).

The `broadcast_rpc_address` parameter specifies the RPC address to broadcast to drivers and other Cassandra nodes. You cannot set it to 0.0.0.0. If this section is absent, it is set to the value of the `rpc_address`. If `rpc_address` is set to 0.0.0.0, this property must be set. For more information, refer to the official documentation at [https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__broadcast_rpc_address](https://docs.datastax.com/en/cassandra-oss/3.0/cassandra/configuration/configCassandra_yaml.html#configCassandra_yaml__broadcast_rpc_address).

The `externalAddress` parameter combines the `broadcast_address` and `broadcast_rpc_address` parameters so they must not be set if `externalAddress` is used mutually exclusively. After the deployment, you can see the `broadcast_address` and `broadcast_rpc_address` parameters instead of `externalAddress` in the custom resources file.

The `internalAddress` parameter combines the `listen_address` and `rpc_address` parameters so they must not be set if `internalAddress` is used mutually exclusively. After the deployment, you can see the `listen_address` and `rpc_address` parameters instead of `internalAddress` in the custom resources file.

The values' vars description is as follows:

The `cassandra.hostNetwork` if set to true allows to use host IP on the pod interface.

The `cassandra.deploymentSchema.dataCenters[$idx].seedList` specifies at least one 'public' IP (floating IP) of seed node from each site.

The `cassandra.deploymentSchema.dataCenters[$idx].broadcastAddress` specifies a list of 'public' IP (floating IP) of each node from the datacenter. These addresses are announced to other Cassandra nodes.

The `cassandra.deploymentSchema.dataCenters[$idx].listenAddress` specifies a list of 'private' IP (host IP) of each node from the datacenter. This network is used for communication between nodes from the current datacenter.

The `cassandra.deploymentSchema.dataCenters[$idx].rpcBroadcastAddress` specifies a list of 'public' IP (floating IP) of each node from the datacenter. These addresses are announced to clients (cqlsh).

The `cassandra.deploymentSchema.dataCenters[$idx].rpcListenAddress` specifies a list of 'private' IP (host IP) of each node from the datacenter. These addresses may be used for local connection between a client (cqlsh) and Cassandra inside the current datacenter.

The `cassandra.deploymentSchema.dataCenters[$idx].externalAddress` specifies a list of 'public' IP (floating IP) of each node from the datacenter. It replaces `cassandra.deploymentSchema.dataCenters[$idx].broadcastAddress` and `cassandra.deploymentSchema.dataCenters[$idx].rpcBroadcastAddress` in **values.yaml**.

The `cassandra.deploymentSchema.dataCenters[$idx].internalAddress` specifies a list of 'public' IP (floating IP) of each node from the datacenter. It replaces `cassandra.deploymentSchema.dataCenters[$idx].listenAddress` and `cassandra.deploymentSchema.dataCenters[$idx].rpcListenAddress` in **values.yaml**.

The deployment architecture is shown in the following image.

![Deployment Architecture](/docs/public/images/cassandra_2_tenants_nat.png)

Cluster 1

|Node|IP Address|Floating IP|
|---|---|---|
|Cas_Serv1| 192.168.0.1|10.10.0.1|
|Cas_Serv2| 192.168.0.2|10.10.0.2|
|Cas_Serv3| 192.168.0.3|10.10.0.3|
|Client|192.168.0.4|None|

Cluster 2

|Node|IP Address|Floating IP|
|---|---|---|
|Cas_Serv1|192.168.10.1|10.10.10.1|
|Cas_Serv2|192.168.10.2|10.10.10.2|  
|Cas_Serv3|192.168.10.3|10.10.10.3|
|Client|192.168.10.4|None|

When a client connects to a Cassandra node within the same cluster to internal (to its host IP) or external (to its floating IP - NAT) network, it receives the rpc addresses from the external network because they are defined in the `rpcBroadcastAddress` parameter. Still the client can connect to the Cassandra node by the internal address but only inside one cluster. To connect to the Cassandra node from another cluster, only an external IP should be used.

In a server-server connection, Cassandra selects the preferred_ip field from the system.peers table. If it is not null, then Cassandra uses this address to connect. If null, then Cassandra uses the address from the peer column, which contains the external IP. Therefore, inside a cluster, Cassandra communicates with other nodes by an internal network.

An example of the system.peers table is shown in the following image.

![system.peers Table](/docs/public/images/cassandra_system.peers.png)

**Note**: If you use the default credentials like cassandra:cassandra, then it is better to add the `listen_on_broadcast_address: true` parameter to be able to connect to Cassandra, otherwise it returns an authentication error. If you use different credentials, do not add the `listen_on_broadcast_address: true` parameter. The reason for this behavior is currently unclear. It is better to use non-default credentials.

Config example to deploy DC1: [different_tenants_dc1.yaml](examples/different_tenants_dc1.yaml)

Config example to deploy DC2: [different_tenants_dc2.yaml](examples/different_tenants_dc2.yaml)

The following is a short list of an address example:

```
      externalAddress:
        - 10.101.203.166
        - 10.101.203.167
        - 10.101.203.168
      internalAddress:
        - 192.168.11.10
        - 192.168.11.11
        - 192.168.11.12
```

For more information, refer to the official documentation at [https://docs.datastax.com/en/archived/cassandra/2.0/cassandra/configuration/configMultiNetworks.html](https://docs.datastax.com/en/archived/cassandra/2.0/cassandra/configuration/configMultiNetworks.html).


#### Deployment on Two Kubernetes Instances in GKE

This section describes the deployment of Cassandra on two separate Kubernetes instances in GKE.

##### Prerequisites

 * MSC is configured to provide Pod-to-Pod IP connectivity. For more information, refer to [https://cloud.google.com/kubernetes-engine/docs/how-to/multi-cluster-services](https://cloud.google.com/kubernetes-engine/docs/how-to/multi-cluster-services).
 * Namespaces for Cassandra are created with the same name in both Kubernetes instances.

##### Cassandra Deployment

Currently, there is an issue in GKE that prevents Headless Services from being exported through ServiceExport. The following method is an alternative to deploy Cassandra.

1. Run the [App Deployer](#app-deployer) deployment job on the first Kubernetes instance. For more information, see [gke_dr_deploy_dc1.yaml](examples/gke_dr_deploy_dc1.yaml):

   * The `publicCloud` parameter must be set to `GKE`.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].seedList` parameter must be set to 
     ```
     - cassandra0-0.cassandra.<cassandra-namespace>.svc.cluster.local
     ```
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `true`.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `false`.

1. After the first job is completed, run the [App Deployer](#app-deployer) deployment job on the second Kubernetes instance. For more information, see [gke_dr_deploy_dc2.yaml](examples/gke_dr_deploy_dc2.yaml):

   * The `publicCloud` parameter must be set to `GKE`.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].seedList` parameter must be set to 
     ```
     - cassandra0-0.cassandra.<cassandra-namespace>.svc.cluster.local
     - <IP of cassandra0-0 from DC1>
     ```
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `false`.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `true`.

**Note**: The `cassandra.deploymentSchema.dataCenters['<dc_name>'].deploy` parameter changes between installations.

### Deployment With Separate Volumes for Commitlog and Data Files

Multiple volumes can be attached to Cassandra replicas by setting multiple `storage` elements.

It is required to set the `mountSettings.name` and `mountSettings.mountPath` parameters of each `storage` element except the first one that is used for data files and has predefined name `data` and path `/var/lib/cassandra/data`. Mount names and paths should not overlap. The `nodeLabels` parameter is applicable only for the first `storage` element, the others are not counted.

Alternatively, a dedicated commitlog volume can be provisioned automatically by enabling `cassandra.commitlogArchiving.enabled` and configuring the `cassandra.commitlogArchiving.storage` sub-block. This approach provisions a separate PVC for the commitlog without requiring manual `storage` list entries.

The list of parameters for the commitlog storage block is as follows:

| Parameter                                              | Mandatory | Type                | Default | Description                                                                                                    |
|--------------------------------------------------------|-----------|---------------------|---------|----------------------------------------------------------------------------------------------------------------|
| `cassandra.commitlogArchiving.enabled`                 | false     | boolean             | false   | If a separate volume for the commitlog should be enabled.                                                      |
| `cassandra.commitlogArchiving.storage.size`            | false     | string              |         | The size of the PVC for the commitlog volume.                                                                  |
| `cassandra.commitlogArchiving.storage.storageClasses`  | false     | []string            |         | The storage class for PVC creation. For hostpath PVs, leave blank.                                            |
| `cassandra.commitlogArchiving.storage.volumes`         | false     | []string            |         | The list of pre-created persistent volumes. For auto-provision, leave blank.                                   |
| `cassandra.commitlogArchiving.storage.nodeLabels`      | false     | []map[string]string |         | The labels to map the commitlog PVC to specific nodes. Use the `kubernetes.io/hostname` label to set the node. |
| `cassandra.commitlogArchiving.storage.matchLabelSelectors` | false | []map[string]string |         | The key:value pair of PVs to bind to the commitlog PVC.                                                        |
| `cassandra.commitlogArchiving.storage.waitPvcBound`    | false     | bool                | false   | If the operator needs to wait for PVC binding before proceeding.                                               |

An example using the `commitlogArchiving.storage` block is as follows:

```
cassandra:
  commitlogArchiving:
    enabled: true
    storage:
      size: 1Gi
      storageClasses:
        - standard
      volumes: []
      nodeLabels: []
      matchLabelSelectors: []
      waitPvcBound: false
```

An example using multiple `storage` elements is as follows:

```
cassandra:
  configuration:  |-
    commitlog_directory: /var/lib/cassandra/customcommitlog
  deploymentSchema:
    dataCenters:
    - replicas: 2
      storage:
      - nodeLabels:
        - kubernetes.io/hostname: dr311dev-node-left-1
        - kubernetes.io/hostname: dr311dev-node-left-2
        size:
        - 1Gi
        volumes:
        - cassandra-0
        - cassandra-1
      - mountSettings:
          mountPath: /var/lib/cassandra/customcommitlog
          name: commitlog
        size:
        - 2Gi
        volumes:
        - cassandra-0-log
        - cassandra-1-log
    - replicas: 1
      storage:
      - nodeLabels:
        - kubernetes.io/hostname: dr311dev-node-left-3
        size:
        - 1Gi
        volumes:
        - cassandra-2
      - mountSettings:
          mountPath: /var/lib/cassandra/customcommitlog
          name: commitlog
        size:
        - 2Gi
        storageClasses:
        - csi-sc-cinderplugin
```

### Deployment of Backup Daemon With S3 Storage

Cassandra Backup Daemon can be configured to save backups to S3 storage.

It is required to set the `backupDaemon.s3.enabled` parameter to `true` and set the `backupDaemon.s3.bucketName`, `backupDaemon.s3.endpointUrl`, `backupDaemon.s3.endpointUrl`, `backupDaemon.s3.accessKeyId`, and `backupDaemon.s3.accessKeySecret` parameters.

Cassandra Backup Daemon with S3 still requires a local storage as a clipboard. It can be hostPath PV or a dynamic storage. The storage is configured in the `backupDaemon.storage` parameter.

An example of parameters for Cassandra backup daemon with S3 is as follows:

```
backupDaemon:
  install: true
  storage:
    size: 1Gi
    storageClasses:
      - local-path
  storageDirectory: /cassandra/backup-storage
  s3:
    enabled: true
    secretName: cassandra-backup-s3-credentials
    bucketName: backup
    accessKeyId: minio
    accessKeySecret: *****
    endpointUrl: http://minio-tenant-4.paas-miniha-kubernetes.openshift.sdntest.netcracker.com
```

### Cassandra Reaper Installation 

This section describes how to install the Cassandra Reaper repair tool in a Cassandra cluster.

For more information about Cassandra Reaper, refer to the official documentation at [http://cassandra-reaper.io/docs/](http://cassandra-reaper.io/docs/).

Cassandra Reaper is installed in a Sidecar mode. Therefore, every Cassandra pod has two containers, one with Cassandra process and one with Cassandra Reaper process.
 
To enable Cassandra Reaper, you need to specify the parameters from the [Cassandra Reaper Parameters](#cassandra-reaper-parameters) section in the microservice_deployer-helm or the App Deployer job:

After all the services are deployed, you can access the Cassandra Reaper Web UI using the address specified in the `reaper.ingressHost` parameter plus /webui/index.html. The credentials for the Web UI are specified in the `reaper.username` and `reaper.password` parameters.

#### Re-encrypt Route In Openshift Without NGINX Ingress Controller

Automatic re-encrypt Route creation is not supported out of box, need to perform the following steps:

1. Disable Ingress in deployment parameters: `reaper.ingressHost: ""`.

   Deploy with enabled reaper Ingress leads to incorrect Ingress and Route configuration.

2. Create Route manually. You can use the following template as an example:

   ```yaml
   kind: Route
   apiVersion: route.openshift.io/v1
   metadata:
     annotations:
       route.openshift.io/termination: reencrypt
     name: <specify-uniq-route-name>
     namespace: <specify-namespace-where-reaper-is-installed>
   spec:
     host: <specify-your-target-host-here>
     to:
       kind: Service
       name: reaper 
       weight: 100
     port:
       targetPort: http
     tls:
       termination: reencrypt
       destinationCACertificate: <place-CA-certificate-here-from-reaper-TLS-secret>
       insecureEdgeTerminationPolicy: Redirect
   ```

**NOTE**: If you can't access the reaper host after Route creation because of "too many redirects" error, then one of the possible root
causes is there is HTTP traffic between balancers and the cluster. To resolve that issue it's necessary to add the Route name to
the exception list at the balancers,
[see documentation](https://git.netcracker.com/PROD.Platform.HA/ocp-4-support/-/blob/master/documentation/Maintenance.md#configure-tls-offload-at-the-load-balancer-nodes)

### Diagnostic Tool SJK

Starting from version 1.8.0, Cassandra pods are deployed with diagnostic tool SJK. For more information, refer to the official documentation at [https://github.com/aragozin/jvm-tools](https://github.com/aragozin/jvm-tools).

SJK jar file is located at `/usr/share/java/sjk-plus-0.17.jar`.

## Azure

This section describes the deployment of Cassandra DBaaS Adapter to Microsoft Azure Kubernetes to work with Cosmos DB.

The following table describes the limitations of Cassandra DBaaS Adapter with Cosmos DB.

| Operation           | Support | Comment                                             |
|---------------------|---------|-----------------------------------------------------|
| Create Database     | No      | CREATE ROLE statement is not supported in Cosmos DB.|
| Create User         | No      | CREATE ROLE statement is not supported in Cosmos DB.|
| List Databases      | Yes     |                                                     |
| Describe   Database | No      | LIST statement is not supported in Cosmos DB.       |
| Update   metadata   | Yes     |                                                     |
| Drop   database     | Yes     |                                                     |
| Drop   user         | No      |                                                     |

#### Deployment of Cassandra DBaaS Adapter to Work With External Cassandra Instance Like Azure Cosmos DB or Amazon Keyspaces

To deploy Cassandra DBaaS Adapter to AKS or Amazon EKS, run the [App Deployer](#app-deployer) deployment job with the following parameters:

* `cassandra.install` parameter must be set to `false`.
* `backupDaemon.install` parameter must be set to `false`.
* `dbaas.install` parameter must be set to `true`.
* `dbaas.externalCassandra.install` parameter must be set to `true`.
* `dbaas.externalCassandra.host` parameter must be set to Cosmos DB/Amazon Keyspaces hostname.
* `dbaas.externalCassandra.port` parameter must be set to Cosmos DB/Amazon Keyspaces port.
* `dbaas.externalCassandra.username` parameter must be set to Cosmos DB/Amazon Keyspaces username.
* `dbaas.externalCassandra.password` parameter must be set to Cosmos DB/Amazon Keyspaces password.

# Upgrade 

This section provides the information about the upgrade procedure from one operator version to another version.
The Upgrade procedure is identical to a clean installation. The only difference is that DEPLOY_MODE needs to be set to Rolling Update. If needed, change the deployment parameters.

### Prerequisites

Ensure you use the same type of deployer for previous and current installation.
For example, if App Deployer was used for previous installation, it should be used for the current installation as well.

#### CRD Upgrade

The upgrade of CRD happens automatically through pre-deploy scripts.

**Note**: Automation CRD upgrade requires the corresponding permissions for a deploy user:

```yaml
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "create", "patch"]
```

To disable this feature, add the `DISABLE_CRD: true` parameter(The DISABLE_CRD option is supported from release 2.16.0 onward).

**Note**: `DISABLE_CRD=true` is intended only for restricted environments where cluster-wide permissions (`apiextensions.k8s.io`) are unavailable. **This flag is deprecated for all other deployment scenarios.** When CRDs are delivered as a dedicated CRD application, use that application instead of disabling CRD installation. In restricted environments without cluster-wide permissions, the dedicated CRD application must not be installed — use `DISABLE_CRD=true` for the main microservice application instead.

### Upgrade Using App Deployer

Set **DEPLOY_MODE** to **Rolling Update** and repeat the steps from [Deployment Using App Deployer](#app-deployer).

## Upgrade Procedure From Cassandra 3.11.X to Cassandra 4.0.X

Cassandra can be upgraded from binary version 3.11.X to 4.0.X using Cassandra Operator of the version starting with 1.28.0.

To perform the upgrade, run the [App Deployer Job](#app-deployer) with Rolling Upgrade DEPLOY_MODE.


### Limitations

* The backups taken before the upgrade are not restored in the new Cassandra 4 cluster.

## Upgrade Procedure to Scheme With Separate Commit Log PV

To run the update:

1. Execute ```nodetool flush``` on each replica.
2. Ensure that the .db files are created on each replica:

   ```
   sh-4.2$ ls -la ./var/lib/cassandra/data/keyspace1/standard1-45a9bc70ad7611eba4d231804e83a3a5/
   total 8004
   drwxr-xr-x. 3 cassandra cassandra     229 May  5 07:59 .
   drwxr-xr-x. 4 cassandra cassandra     105 May  5 07:48 ..
   drwxr-xr-x. 2 cassandra cassandra       6 May  5 07:48 backups
   -rw-r--r--. 1 cassandra cassandra     468 May  5 07:59 mc-1-big-CRC.db
   -rw-r--r--. 1 cassandra cassandra 7564700 May  5 07:59 mc-1-big-Data.db
   -rw-r--r--. 1 cassandra cassandra       9 May  5 07:59 mc-1-big-Digest.crc32
   -rw-r--r--. 1 cassandra cassandra   41128 May  5 07:59 mc-1-big-Filter.db
   -rw-r--r--. 1 cassandra cassandra  549938 May  5 07:59 mc-1-big-Index.db
   -rw-r--r--. 1 cassandra cassandra   10287 May  5 07:59 mc-1-big-Statistics.db
   -rw-r--r--. 1 cassandra cassandra    5706 May  5 07:59 mc-1-big-Summary.db
   -rw-r--r--. 1 cassandra cassandra      80 May  5 07:59 mc-1-big-TOC.txt
   ```

   In ```/var/lib/cassandra/data```, you can find all the keyspaces and table names as subdirectories. If you are not sure about the data directory, execute ```find -name backups``` and in the output, you can find paths with keyspaces names:

   ```
   sh-4.2$ find -name backups
   ./var/lib/cassandra/data/keyspace1/standard1-45a9bc70ad7611eba4d231804e83a3a5/backups
   ./var/lib/cassandra/data/keyspace1/counter1-47bee0d0ad7611ebbc7eeb61d405b136/backups
   ```

3. Run the Helm Deployer CI with the `upgrade` mode and parameters as described in the [Multi-Storage Schema for separate Cassandra Commitlog Volume](#multi-storage-schema-for-separate-cassandra-commitlog-volume) section.
4. Wait until all nodes are up and then check the data.

For example:

```
admin@cqlsh> select *from keyspace1.standard1 limit 2;

 key                    | C0                                                                     | C1                                                                     | C2               | C3                                                                     | C4
------------------------+------------------------------------------------------------------------+------------------------------------------------------------------------+------------------------------------------------------------------------+------------------------------------------------------------------------+------------------------------------------------------------------------
 0x4c4e5032304e34334d30 | 0xfb073d21b291e3030d2be5d3217fe78e75f6dcd1fb8d690ed26677bc13b52cf5e0de | 0xfe069ee8cb3ed5589e7b15891b6fada8d719cb68257fd31b118db7a157bdd185e615 | 0xfc9b745e4beb31506821babde05d31e13e0c0f15e73807f028b9406960058c1450a9 | 0x1288207dc2f2527b2e49e7794a0b67b23a9dbe489913aa12b96519dd95abd0ec5d43 | 0xa09c829ae45d230939d8b003a86ed37e2624581181070f529891e2751dee8c5c4386
 0x343131504b3850323631 | 0xe47ca02cc31a43e7b77d06979fb0788b6f900766abad58ff4d76330f2f006a64f052 | 0x112b68a452dcbfd42d8680b11cf598b6fdd57850daaa1523b06cf33f94e8a3505da0 | 0xa12a44e2aa7e86138c50a49a3aa21058877bb610164321bef504629a41f14dd04ed4 | 0xd9d0ae5483a47575d1232cf1aa3cd868bb82e41a436b2c67a33cf88c0a9b0eabb79a | 0x57a9aa5deac4c8595f58f3082d796a8056a4d34d2233c529455341b4295deafa7edb

(2 rows)
```

## How to deploy with DEPLOY_W_HELM true over false?

App Deployer does not support migration from `DEPLOY_W_HELM: false` to `DEPLOY_W_HELM: true`.

If you need it, you have to delete all resources that belong to the current installation.

For example:

```bash
kubectl delete all,secrets,configmaps,ingresses,serviceaccounts,roles,rolebindings,grafanadashboard,prometheusrule,servicemonitor --all --namespace=<namespace_name>
```

Then install using App Deployer and `DEPLOY_W_HELM: true`.

## Migration from Joint to Separate Deployment

Starting from `2.0.0` version the installation of Cassandra and its supplementary services consists of 2 separate steps:

1. Installation of `Cassandra` application with Apache Cassandra and its operator.
2. Installation of `Cassandra Services` application with supplementary services (backup daemon, dbaas adapter, robot tests).

There are two ways for updating Cassandra from joint to separate scheme - `automatic` and `manual`. Both are described below.

### Automatic Way

The required steps can be performed automatically using the parameter `ENABLE_MIGRATION: true`. It is enabled by default.
It requires full access to namespace and previous Cassandra should have been installed with `DEPLOY_W_HELM: true` mode.

### Manual Way

The following steps can be performed manually:

1. Verify that previous `cassandra` is installed with Helm mode.

    ```yaml
    helm list
    ```

    If you do not see any cassandra related Helm releases like `cassandra-operator` or `cassandra-operator-{NAMESPACE}`
    it means `cassandra` was installed without Helm.
    In that case only one option of migrations is possible - remove all resources following the guide
    [How to deploy with DEPLOY_W_HELM true over false](#how-to-deploy-with-deploy_w_helm-true-over-false).
2. If Helm release exists, you need to remove the linkage between deployments and services to keep them running during migration:

    ```bash
    kubectl get deployments,services,statefulsets -l ${SELECTOR} -o name | xargs -I {} $kubectl patch {} -p '{"metadata":{"ownerReferences":null}}'
    ```

After that you need to install `Cassandra` and then `Cassandra Services` services by following the [App Deployer](#app-deployer) using `Rolling Update`.

# TLS Encryption

## Platform Cassandra Distributive

_New in Cassandra Operator 1.29.0_

Secure communication between a client machine and a database cluster can be provided using TLS encryption (mTLS encryption is not supported). By default, TLS encryption is disabled. To enable it, set the `tls.enabled` parameter to `true`.    
Use the `tls.optional=true` parameter to allow both TLS and non-TLS encrypted connections to Cassandra.

The following image shows a TLS encrypted communication in the Cassandra namespace:

![Deployment Architecture](/docs/public/images/cassandra_tls.png)

This configuration provides a TLS encrypted connection between a client machine and a database. The DBaaS adapter and Backup daemon also use a secure connection to Cassandra and accept REST requests on the TLS port 8443. 

To connect to Cassandra using cqlsh over TLS, execute the following commands:

```
export SSL_CERTFILE=/var/lib/cassandra/configuration/ca.crt
cqlsh -u <cassandra_username> -p <cassandra_password> --ssl
```

#### Set Certificates Manually

To pass pre-generated certificates as deploy parameters use the following parameters:

`tls.certificates.ca_crt` - a base 64 encoded CA certificate

`tls.certificates.tls_crt` - a base 64 encoded certificate for Cassandra, DNS name is `cassandra.<namespace>.svc`

`tls.certificates.tls_key` - a base 64 encoded key for Cassandra

`tls.backup.certificates.tls_crt` - a base 64 encoded certificate for Backup Daemon, DNS name is `cassandra-backup-daemon.<namespace>.svc`

`tls.backup.certificates.tls_key` - a base 64 encoded key for Backup Daemon

`tls.dbaas.certificates.tls_crt` - a base 64 encoded certificate for Dbaas Adapter, DNS name is `dbaas-cassandra-adapter.<namespace>.svc`

`tls.dbaas.certificates.tls_key` - a base 64 encoded key for Dbaas Adapter


Example: 

```
backupDaemon:
  tls:
    certificates:
      tls_key: "LS0tLS1CRUdJTiB...tLS0tCg=="
      tls_crt: "LS0tLS1CRUd...LS0tLS0K"
dbaas:
  tls:
    certificates:
      tls_key: "LS0tLS1CRU...S0tLQo="
      tls_crt: "LS0tLS1C...UtLS0tLQo="
tls:
  enabled: true
  generateCerts:
    enabled: false
  certificates:
    tls_key: "LS0tLS1CRUdJ...0tLS0tCg=="
    tls_crt: "LS0tLS1CRU...tLS0tCg=="
    ca_crt: "LS0tLS1CRU...S0tLS0K" 
```

## Amazon Keyspaces

Amazon Keyspaces only accept secure connections using Transport Layer Security (TLS).

Refer to the official documentation at [https://docs.aws.amazon.com/keyspaces/latest/devguide/encryption-in-transit.html](https://docs.aws.amazon.com/keyspaces/latest/devguide/encryption-in-transit.html) to get the Starfield digital certificate to configure the Cassandra driver or cqlsh. 

# Scaling

This section describes how scale up or scale down Cassandra.

## Add Cassandra Datacenter From a Separate Kubernetes Cluster - Upgrade Procedure From non-DR to DR

The procedure is similar to [Deployment on Two Kubernetes Instances With Calico](#deployment-on-two-kubernetes-instances-with-calico). The only difference is that the first Cassandra Datacenter is already deployed, and only the second one needs to be deployed.

### Prerequisites

Cassandra is not deployed to a new Kubernetes cluster.

### Procedure

1. Run the [App Deployer](#app-deployer) deployment job on the new Kubernetes instance.

   * The `dataCenters` section must contain all datacenters, including the already deployed ones.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].clusterDomain` parameter must be set to the DNS name of the new Kubernetes.
   * The `cassandra.deploymentSchema.dataCenters['dc1'].deploy` parameter must be set to `false`.
   * The `cassandra.deploymentSchema.dataCenters['dc2'].deploy` parameter must be set to `true`.

2. After the job is successfully completed, validate that Cassandra Datacenters are visible to each other by executing the `nodetool status` command in any `cassandraX-0` pod.

Following is an example of the `cassandra` section of deployment values:

```
cassandra:
  deploymentSchema:
    dataCenters:
      - name: dc1
        replicas: 3
        deploy: false
        seeds: 2
        clusterDomain: cluster-1.local
        storage:
          size: 1Gi
          mountSettings:
            name: data
            mountPath: /var/lib/cassandra/data
          storageClasses:
            - nfs-external
      - name: dc2
        replicas: 3
        seeds: 2
        deploy: true
        clusterDomain: cluster-2.local
        storage:
          size: 1Gi
          mountSettings:
            name: data
            mountPath: /var/lib/cassandra/data
          storageClasses:
            - nfs-external
```

## Scaling Up

Cassandra deployment can be scaled up within a data center. To scale up Cassandra, run an update of Cassandra using microservice_deployer-helm or App Deployer:

1. In the `cassandra.deploymentSchema.dataCenters[$idx].replicas` deployment parameter, specify the desired number of replicas.
1. In case of a pre-created (hostPath) PV, in the `cassandra.deploymentSchema.dataCenters[$idx].storage.volumes` deployment parameter, add a new volume name at the end of list.
1. In case of a pre-created (hostPath) PV, in the `cassandra.deploymentSchema.dataCenters[$idx].storage.nodeLabels` deployment parameter, add a new node label for the new replica.
1. In case of pre-defined racks, in the `cassandra.deploymentSchema.dataCenters[$idx].racks` deployment parameter, add a new or existing rack name for the new replica.
1. Do not change the remaining parameters.
1. Run the job.
    
## Scaling Down

Cassandra deployment can be scaled down within a data center. To scale down Cassandra, run an update of Cassandra using microservice_deployer-helm or App Deployer.

1. Login to the Kubernetes/OpenShift Cassandra project and on any running Cassandra pod, execute the `nodetool status` command. 
   
   The output of the command contains mapping of replicas' IP address and their Host ID as shown in the example below.

   ```
    Datacenter: dc1
    ===============
    Status=Up/Down
    |/ State=Normal/Leaving/Joining/Moving
    --  Address        Load       Tokens       Owns (effective)  Host ID                               Rack
    UN  10.131.60.149  304.71 KiB  256          100.0%            62e36be8-5891-49d4-a20a-9303b6eba7e8  rack1
    UN  10.131.6.74    512.75 KiB  256          100.0%            504ef706-00f7-4e1c-a83f-dc4494784bfe  rack1
    UN  10.129.187.14  257.24 KiB  256          100.0%            acde4590-a49a-4221-be5f-5f6d1333fafd  rack1
   ``` 
   
1. Using the IP address, determine the host ID of the replica to be removed.
1. In the `cassandra.deploymentSchema.dataCenters[$idx].removeNodes` deployment parameter, set the key:value pair, where key is the index of replica to remove and the value is its host ID derived from step 2. There might be multiple pairs.
1. In case of pre-selected seeds, in the `cassandra.deploymentSchema.dataCenters[$idx].seedList` deployment parameter, replace the replicas' hosts to be removed with the existing ones.
1. Do not change the remaining parameters.
1. Run the job.
   
   For example, to remove cassandra1-0 replica with IP address 10.131.6.74, the following parameters must be specified:
   
   ```
   cassandra:
     install: true
     deploymentSchema:
       dataCenters:
         - name: dc1
           replicas: 3
           seeds: 2
           removeNodes:
             - 1: 504ef706-00f7-4e1c-a83f-dc4494784bfe
           storage:
             nodeLabels:
               - 'kubernetes.io/hostname': k8s-miniha-1-master-1
               - 'kubernetes.io/hostname': k8s-miniha-1-master-2
               - 'kubernetes.io/hostname': k8s-miniha-1-master-3
             size: 1Gi
             volumes:
               - cassandra-0
               - cassandra-1
               - cassandra-2
             storageClasses:
               - ""
   ```

## AWS Keyspaces

AWS Keyspaces allow you to configure automatic scaling. For more information, refer to [https://docs.aws.amazon.com/keyspaces/latest/devguide/autoscaling.html](https://docs.aws.amazon.com/keyspaces/latest/devguide/autoscaling.html).

For information about the cost calculation model, refer to [https://aws.amazon.com/keyspaces/pricing/](https://aws.amazon.com/keyspaces/pricing/).

For information about limits, refer to [https://docs.aws.amazon.com/keyspaces/latest/devguide/quotas.html](https://docs.aws.amazon.com/keyspaces/latest/devguide/quotas.html).

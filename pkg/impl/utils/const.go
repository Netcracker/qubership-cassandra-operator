// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils //todo package name confuses

const Reaper = "reaper"

const CassandraLb = "cassandra-lb"
const Cassandra = "cassandra"
const CassandraMetrics = "cassandra-metrics"
const DCFormat = "dc-%s"
const CassandraDCFormat = Cassandra + "-" + DCFormat
const CassandraReplicaNameFormat = Cassandra + "%v"
const CassandraCluster = "cassandra-cluster"

var SystemKeyspaces = []string{"system_traces", "system_distributed", "system_auth"}

const PVNodesFormat = "pvNodeNames%v"

const KubernetesHelperImpl = "kubernetesHelperImpl"
const CassandraHelperImpl = "cassandraHelperImpl"

const KubeHostName = "kubernetes.io/hostname"
const Name = "name"
const Service = "service"
const App = "app"
const Data = "data"
const Username = "username"
const Password = "password"

const ContextPasswordKey = "ctxPasswordKey"

const CassandraDCSeedsFormat = "cassandraSeeds%v"
const CassandraSeeds = "cassandraSeeds"

const ConfigurationPath = "/var/lib/cassandra/configuration/"
const Configuration = "configuration"
const CassandraEnv = "cassandra-env"
const CassandraJVM = "cassandra-jvm"
const CassandraConfiguration = "cassandra-configuration"
const CassandraCommitlogArchiving = "cassandra-commiglog-archiving"
const CassandraLogback = "cassandra-logback"
const CassandraReaper = "cassandra-reaper"
const Config = "config"

const BashCommand = "bash"

const DefaultClusterDomain = "cluster.local"
const StatefulSetPodNameTemplate = "%s-0"
const ClusterDomainTemplate = "%s.svc.%s"

const ReplicaNumber = "replica_number"

const TriesCount = "triesCount"
const RetryTimeoutSec = "retryTimeout"

const CassandraDCPvcNameFormat = "cassandra-data-dc%v"

const RootCertPath = "/usr/ssl/"

const AccessKey = "accessKey"
const SecretKey = "secretKey"
const Region = "region"

// http
const FlushURI = "flush"

const Charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// labels
const (
	AppName              = "app.kubernetes.io/name"
	AppInstance          = "app.kubernetes.io/instance"
	AppVersion           = "app.kubernetes.io/version"
	AppComponent         = "app.kubernetes.io/component"
	AppManagedBy         = "app.kubernetes.io/managed-by"
	AppManagedByOperator = "app.kubernetes.io/managed-by-operator"
	AppProcByOperator    = "app.kubernetes.io/processed-by-operator"
	AppTechnology        = "app.kubernetes.io/technology"
	AppPartOf            = "app.kubernetes.io/part-of"
	CloneModeType        = "clone-mode-type"
)

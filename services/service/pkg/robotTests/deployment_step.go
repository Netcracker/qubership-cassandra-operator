package robotTests

import (
	"fmt"
	"strconv"

	v1 "github.com/Netcracker/qubership-cassandra-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-cassandra-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	coreUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	robotTestsAtpStorageSecretName = "cassandra-robot-tests-atp-storage-secret"
	robotTestsPodSecretsMountPath  = "/etc/secrets/robot-tests-pod-secrets"
)

type RobotDeployment struct {
	core.DefaultExecutable
}

func (r *RobotDeployment) Execute(ctx core.ExecutionContext) error {

	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1.CassandraSupplService)
	robot := spec.Spec.RobotTests
	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	credsManager := ctx.Get(utils.ContextCredsManager).(utils.CredsManagerI)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	// Secret volumes and mounts
	secretVolumes := map[string]string{
		spec.Spec.Cassandra.SecretName: "/var/run/secrets/cassandra",
	}

	if spec.Spec.Backup.Install {
		secretVolumes[spec.Spec.Backup.SecretName] =
			"/var/run/secrets/backup"
	}

	if spec.Spec.Dbaas.Install {
		secretVolumes[spec.Spec.Dbaas.Adapter.SecretName] =
			"/var/run/secrets/dbaas-adapter"
	}

	secretVolumeMode := int32(256)
	volumes := []v12.Volume{}
	volumeMounts := []v12.VolumeMount{}

	for secretName, mountPath := range secretVolumes {

		volumeName := utils.SanitizeName(secretName)

		volumes = append(volumes, v12.Volume{
			Name: volumeName,
			VolumeSource: v12.VolumeSource{
				Secret: &v12.SecretVolumeSource{
					SecretName:  secretName,
					DefaultMode: &secretVolumeMode,
				},
			},
		})

		volumeMounts = append(volumeMounts, v12.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  true,
		})
	}

	// Environment variable Start
	envs := []v12.EnvVar{
		{
			Name: "NAMESPACE",
			ValueFrom: &v12.EnvVarSource{
				FieldRef: &v12.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	currentDc := utils.NewStream(spec.Spec.Cassandra.DeploymentSchema.DataCenters).FindFirst(func(dc interface{}) bool {
		return dc.(*v1.DataCenter).Deploy
	}).(*v1.DataCenter)

	envs = append(envs,
		coreUtils.GetPlainTextEnvVar("CASSANDRA_HOST", core.OptionalString(spec.Spec.Cassandra.Host, fmt.Sprintf("%s.%s", utils.Cassandra, request.Namespace))),
		coreUtils.GetPlainTextEnvVar("CASSANDRA_PORT", core.OptionalString(strconv.Itoa(spec.Spec.Cassandra.Port), "9042")),
		coreUtils.GetPlainTextEnvVar("TEST_KEYSPACES_REPLICATION_FACTOR", strconv.Itoa(robot.ReplicationFactor)),
		coreUtils.GetPlainTextEnvVar("ATTEMPTS_NUMBER", strconv.Itoa(robot.AttemptsNumber)),
		coreUtils.GetPlainTextEnvVar("PROMETHEUS_URL", robot.PrometheusUrl),
		coreUtils.GetPlainTextEnvVar("TAGS", robot.Tags),
		coreUtils.GetPlainTextEnvVar("WAIT_TIMEOUT", strconv.Itoa(spec.Spec.WaitTimeout)),
		coreUtils.GetPlainTextEnvVar("DC_NAME", currentDc.Name),
		coreUtils.GetPlainTextEnvVar("DBAAS_ADAPTER_API_VERSION", spec.Spec.Dbaas.ApiVersion),
		coreUtils.GetPlainTextEnvVar("PORT", fmt.Sprint(utils.GetHTTPPort(spec.Spec.TLS.Enabled))),
		coreUtils.GetPlainTextEnvVar("CONFIG_NAME", "cassandra-tests-config"),
		coreUtils.GetPlainTextEnvVar("SUPPLEMENTARY_CONFIG_NAME", "supplementary-tests-config"),

		// todo better place or variables
		coreUtils.GetPlainTextEnvVar("STATUS_CUSTOM_RESOURCE_PATH", fmt.Sprintf("apps/v1/%s/deployments/robot-tests", request.Namespace)),
		coreUtils.GetPlainTextEnvVar("STATUS_WRITING_ENABLED", "true"),
	)

	if spec.Spec.Backup.Install {
		envs = append(envs,
			coreUtils.GetPlainTextEnvVar("BACKUP_HOST", fmt.Sprintf("%s.%s.svc", utils.BackupDaemon, request.Namespace)),
		)
	}

	if spec.Spec.Dbaas.Install {
		envs = append(envs,
			coreUtils.GetPlainTextEnvVar("DBAAS_HOST", fmt.Sprintf("%s.%s.svc", utils.DbaasName, request.Namespace)),
		)
	}

	if robot.AtpReport.Enabled {
		volumes, volumeMounts, envs = appendAtpReportPodConfig(robot, volumes, volumeMounts, envs)
	}

	// Environment variable End

	dc := RobotTemplate(
		request.Namespace,
		robot.DockerImage,
		*robot.Resources,
		robot.NodeLabels,
		envs,
		spec.Spec.RobotTests.Args,
		volumeMounts,
		volumes,
	)

	secretNames := []string{spec.Spec.Cassandra.SecretName}
	if robot.AtpReport.Enabled {
		secretNames = append(secretNames, robotTestsAtpStorageSecretName)
	}
	err := credsManager.AddCredHashToPodTemplate(
		secretNames,
		&dc.Spec.Template,
	)
	if err != nil {
		log.Error(fmt.Sprintf("can't add secret HASH to annotations for %s", dc.Name), zap.Error(err))
		return err
	}

	utils.TLSClientSpecUpdate(
		&dc.Spec.Template.Spec,
		utils.RootCertPath,
		spec.Spec.TLS,
	)

	err = helperImpl.DeleteDeploymentAndPods(
		dc.Name,
		request.Namespace,
		spec.Spec.WaitTimeout,
	)

	core.PanicError(
		err,
		log.Error,
		"RobotTests deployment config processing failed",
	)

	labels := utils.BasicLabels{
		AppName:       utils.Robot,
		AppComponent:  "operator",
		AppTechnology: "python",
	}

	err = utils.CreateRuntimeObjectContextWrapper(
		ctx,
		dc,
		dc.ObjectMeta,
		labels,
	)

	core.PanicError(
		err,
		log.Error,
		"RobotTests deployment config processing failed",
	)

	log.Debug("Waiting for robot tests ready")

	err = helperImpl.WaitForTestsReady(
		dc.Name,
		dc.Namespace,
		spec.Spec.WaitTimeout,
	)

	core.PanicError(err, log.Error, "RobotTests failed")

	return nil
}

func appendAtpReportPodConfig(
	robot v1.RobotTests,
	volumes []v12.Volume,
	volumeMounts []v12.VolumeMount,
	envs []v12.EnvVar,
) ([]v12.Volume, []v12.VolumeMount, []v12.EnvVar) {
	atpSecretMode := int32(420)
	volumes = append(volumes, v12.Volume{
		Name: "robot-tests-pod-secrets",
		VolumeSource: v12.VolumeSource{
			Projected: &v12.ProjectedVolumeSource{
				DefaultMode: &atpSecretMode,
				Sources: []v12.VolumeProjection{
					{
						Secret: &v12.SecretProjection{
							LocalObjectReference: v12.LocalObjectReference{
								Name: robotTestsAtpStorageSecretName,
							},
							Items: []v12.KeyToPath{
								{Key: "atp-storage-username", Path: "ATP_STORAGE_USERNAME"},
								{Key: "atp-storage-password", Path: "ATP_STORAGE_PASSWORD"},
							},
						},
					},
				},
			},
		},
	})
	volumeMounts = append(volumeMounts, v12.VolumeMount{
		Name:      "robot-tests-pod-secrets",
		MountPath: robotTestsPodSecretsMountPath,
		ReadOnly:  true,
	})

	envs = append(envs,
		coreUtils.GetPlainTextEnvVar("INTEGRATION_TESTS_SECRETS_DIR", robotTestsPodSecretsMountPath),
		coreUtils.GetPlainTextEnvVar("ATP_REPORT_ENABLED", "true"),
		coreUtils.GetPlainTextEnvVar("ATP_STORAGE_PROVIDER", robot.AtpReport.AtpStorage.Provider),
		coreUtils.GetPlainTextEnvVar("ATP_STORAGE_SERVER_URL", robot.AtpReport.AtpStorage.ServerUrl),
		coreUtils.GetPlainTextEnvVar("ATP_STORAGE_SERVER_UI_URL", robot.AtpReport.AtpStorage.ServerUiUrl),
		coreUtils.GetPlainTextEnvVar("ATP_STORAGE_BUCKET", robot.AtpReport.AtpStorage.Bucket),
		coreUtils.GetPlainTextEnvVar("ATP_STORAGE_REGION", robot.AtpReport.AtpStorage.Region),
		coreUtils.GetPlainTextEnvVar("ATP_REPORT_VIEW_UI_URL", robot.AtpReportViewUiUrl),
		coreUtils.GetPlainTextEnvVar("ENVIRONMENT_NAME", robot.EnvironmentName),
		coreUtils.GetPlainTextEnvVar("ENABLE_JIRA_INTEGRATION", strconv.FormatBool(robot.EnableJiraIntegration)),
	)

	return volumes, volumeMounts, envs
}

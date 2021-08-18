package converter

const (
	TagMessage                = "message"
	TagSpanKind               = "span.kind"
	TagStatusCode             = "status.code"
	TagStatusMsg              = "status.message"
	TagError                  = "error"
	TagHTTPStatusCode         = "http.status_code"
	TagHTTPStatusMsg          = "http.status_message"
	TagZipkinCensusCode       = "census.status_code"
	TagZipkinCensusMsg        = "census.status_description"
	TagZipkinOpenCensusMsg    = "opencensus.status_description"
	TagW3CTraceState          = "w3c.tracestate"
	TagServiceNameSource      = "otlp.service.name.source"
	TagInstrumentationName    = "otlp.instrumentation.library.name"
	TagInstrumentationVersion = "otlp.instrumentation.library.version"

	ResourceNoServiceName = "OTLPResourceNoServiceName"

	OCAttributeProcessStartTime        = "opencensus.starttime"
	OCAttributeProcessID               = "opencensus.pid"
	OCAttributeExporterVersion         = "opencensus.exporterversion"
	OCAttributeResourceType            = "opencensus.resourcetype"
	OCAttributeSameProcessAsParentSpan = "opencensus.same_process_as_parent_span"
	OCTimeEventMessageEventType        = "opencensus.timeevent.messageevent.type"
	OCTimeEventMessageEventID          = "opencensus.timeevent.messageevent.id"
	OCTimeEventMessageEventUSize       = "opencensus.timeevent.messageevent.usize"
	OCTimeEventMessageEventCSize       = "opencensus.timeevent.messageevent.csize"
)

const (
	AttributeComponent    = "component"
	AttributeEnduserID    = "enduser.id"
	AttributeEnduserRole  = "enduser.role"
	AttributeEnduserScope = "enduser.scope"
	AttributeNetHostIP    = "net.host.ip"
	AttributeNetHostName  = "net.host.name"
	AttributeNetHostPort  = "net.host.port"
	AttributeNetPeerIP    = "net.peer.ip"
	AttributeNetPeerName  = "net.peer.name"
	AttributeNetPeerPort  = "net.peer.port"
	AttributeNetTransport = "net.transport"
	AttributePeerService  = "peer.service"
)

const (
	AttributeCloudAccount               = "cloud.account.id"
	AttributeCloudProvider              = "cloud.provider"
	AttributeCloudRegion                = "cloud.region"
	AttributeCloudZone                  = "cloud.zone"
	AttributeCloudInfrastructureService = "cloud.infrastructure_service"
	AttributeContainerID                = "container.id"
	AttributeContainerImage             = "container.image.name"
	AttributeContainerName              = "container.name"
	AttributeContainerTag               = "container.image.tag"
	AttributeDeploymentEnvironment      = "deployment.environment"
	AttributeFaasID                     = "faas.id"
	AttributeFaasInstance               = "faas.instance"
	AttributeFaasName                   = "faas.name"
	AttributeFaasVersion                = "faas.version"
	AttributeHostID                     = "host.id"
	AttributeHostImageID                = "host.image.id"
	AttributeHostImageName              = "host.image.name"
	AttributeHostImageVersion           = "host.image.version"
	AttributeHostName                   = "host.name"
	AttributeHostType                   = "host.type"
	AttributeK8sCluster                 = "k8s.cluster.name"
	AttributeK8sContainer               = "k8s.container.name"
	AttributeK8sCronJob                 = "k8s.cronjob.name"
	AttributeK8sCronJobUID              = "k8s.cronjob.uid"
	AttributeK8sDaemonSet               = "k8s.daemonset.name"
	AttributeK8sDaemonSetUID            = "k8s.daemonset.uid"
	AttributeK8sDeployment              = "k8s.deployment.name"
	AttributeK8sDeploymentUID           = "k8s.deployment.uid"
	AttributeK8sJob                     = "k8s.job.name"
	AttributeK8sJobUID                  = "k8s.job.uid"
	AttributeK8sNamespace               = "k8s.namespace.name"
	AttributeK8sNodeName                = "k8s.node.name"
	AttributeK8sNodeUID                 = "k8s.node.uid"
	AttributeK8sPod                     = "k8s.pod.name"
	AttributeK8sPodUID                  = "k8s.pod.uid"
	AttributeK8sReplicaSet              = "k8s.replicaset.name"
	AttributeK8sReplicaSetUID           = "k8s.replicaset.uid"
	AttributeK8sStatefulSet             = "k8s.statefulset.name"
	AttributeK8sStatefulSetUID          = "k8s.statefulset.uid"
	AttributeOSType                     = "os.type"
	AttributeOSDescription              = "os.description"
	AttributeProcessCommand             = "process.command"
	AttributeProcessCommandLine         = "process.command_line"
	AttributeProcessExecutableName      = "process.executable.name"
	AttributeProcessExecutablePath      = "process.executable.path"
	AttributeProcessID                  = "process.pid"
	AttributeProcessOwner               = "process.owner"
	AttributeServiceInstance            = "service.instance.id"
	AttributeServiceName                = "service.name"
	AttributeServiceNamespace           = "service.namespace"
	AttributeServiceVersion             = "service.version"
	AttributeTelemetryAutoVersion       = "telemetry.auto.version"
	AttributeTelemetrySDKLanguage       = "telemetry.sdk.language"
	AttributeTelemetrySDKName           = "telemetry.sdk.name"
	AttributeTelemetrySDKVersion        = "telemetry.sdk.version"
)

func GetResourceSemanticConventionAttributeNames() []string {
	return []string{
		AttributeCloudAccount,
		AttributeCloudProvider,
		AttributeCloudRegion,
		AttributeCloudZone,
		AttributeCloudInfrastructureService,
		AttributeContainerID,
		AttributeContainerImage,
		AttributeContainerName,
		AttributeContainerTag,
		AttributeDeploymentEnvironment,
		AttributeFaasID,
		AttributeFaasInstance,
		AttributeFaasName,
		AttributeFaasVersion,
		AttributeHostID,
		AttributeHostImageID,
		AttributeHostImageName,
		AttributeHostImageVersion,
		AttributeHostName,
		AttributeHostType,
		AttributeK8sCluster,
		AttributeK8sContainer,
		AttributeK8sCronJob,
		AttributeK8sCronJobUID,
		AttributeK8sDaemonSet,
		AttributeK8sDaemonSetUID,
		AttributeK8sDeployment,
		AttributeK8sDeploymentUID,
		AttributeK8sJob,
		AttributeK8sJobUID,
		AttributeK8sNamespace,
		AttributeK8sNodeName,
		AttributeK8sNodeUID,
		AttributeK8sPod,
		AttributeK8sPodUID,
		AttributeK8sReplicaSet,
		AttributeK8sReplicaSetUID,
		AttributeK8sStatefulSet,
		AttributeK8sStatefulSetUID,
		AttributeOSType,
		AttributeOSDescription,
		AttributeProcessCommand,
		AttributeProcessCommandLine,
		AttributeProcessExecutableName,
		AttributeProcessExecutablePath,
		AttributeProcessID,
		AttributeProcessOwner,
		AttributeServiceInstance,
		AttributeServiceName,
		AttributeServiceNamespace,
		AttributeServiceVersion,
		AttributeTelemetryAutoVersion,
		AttributeTelemetrySDKLanguage,
		AttributeTelemetrySDKName,
		AttributeTelemetrySDKVersion,
	}
}
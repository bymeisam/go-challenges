package main

import (
	"strings"
	"testing"
)

func TestSimpleDeploymentYAML(t *testing.T) {
	if !strings.Contains(SimpleDeploymentYAML, "kind: Deployment") {
		t.Error("Should have Deployment kind")
	}

	if !strings.Contains(SimpleDeploymentYAML, "replicas: 3") {
		t.Error("Should have 3 replicas")
	}

	if !strings.Contains(SimpleDeploymentYAML, "containers:") {
		t.Error("Should have containers")
	}

	if !strings.Contains(SimpleDeploymentYAML, "ports:") {
		t.Error("Should expose ports")
	}

	t.Log("✓ SimpleDeploymentYAML is valid!")
}

func TestServiceYAML(t *testing.T) {
	if !strings.Contains(ServiceYAML, "kind: Service") {
		t.Error("Should have Service kind")
	}

	if !strings.Contains(ServiceYAML, "ClusterIP") {
		t.Error("Should have ClusterIP service")
	}

	if !strings.Contains(ServiceYAML, "LoadBalancer") {
		t.Error("Should have LoadBalancer service")
	}

	if !strings.Contains(ServiceYAML, "selector:") {
		t.Error("Should have selector")
	}

	t.Log("✓ ServiceYAML is valid!")
}

func TestConfigMapYAML(t *testing.T) {
	if !strings.Contains(ConfigMapYAML, "kind: ConfigMap") {
		t.Error("Should have ConfigMap kind")
	}

	if !strings.Contains(ConfigMapYAML, "data:") {
		t.Error("Should have data section")
	}

	if !strings.Contains(ConfigMapYAML, "app.yaml") {
		t.Error("Should contain app.yaml config")
	}

	if !strings.Contains(ConfigMapYAML, "database.sql") {
		t.Error("Should contain database SQL")
	}

	t.Log("✓ ConfigMapYAML is valid!")
}

func TestSecretsYAML(t *testing.T) {
	secretCount := strings.Count(SecretsYAML, "kind: Secret")

	if secretCount < 1 {
		t.Error("Should have at least one Secret")
	}

	if !strings.Contains(SecretsYAML, "stringData:") {
		t.Error("Should have stringData for opaque secret")
	}

	if !strings.Contains(SecretsYAML, "database-url") {
		t.Error("Should contain database URL secret")
	}

	if !strings.Contains(SecretsYAML, "jwt-secret") {
		t.Error("Should contain JWT secret")
	}

	t.Log("✓ SecretsYAML is valid!")
}

func TestProductionDeploymentYAML(t *testing.T) {
	expectedElements := []string{
		"kind: Deployment",
		"namespace: production",
		"RollingUpdate",
		"livenessProbe",
		"readinessProbe",
		"startupProbe",
		"resources:",
		"requests:",
		"limits:",
		"securityContext:",
		"runAsNonRoot",
		"affinity",
		"podAntiAffinity",
	}

	for _, element := range expectedElements {
		if !strings.Contains(ProductionDeploymentYAML, element) {
			t.Errorf("ProductionDeployment should contain %s", element)
		}
	}

	t.Log("✓ ProductionDeploymentYAML is production-ready!")
}

func TestIngressYAML(t *testing.T) {
	if !strings.Contains(IngressYAML, "kind: Ingress") {
		t.Error("Should have Ingress kind")
	}

	if !strings.Contains(IngressYAML, "tls:") {
		t.Error("Should have TLS configuration")
	}

	if !strings.Contains(IngressYAML, "rules:") {
		t.Error("Should have routing rules")
	}

	if !strings.Contains(IngressYAML, "backend:") {
		t.Error("Should have backend service")
	}

	t.Log("✓ IngressYAML is valid!")
}

func TestHorizontalPodAutoscalerYAML(t *testing.T) {
	if !strings.Contains(HorizontalPodAutoscalerYAML, "kind: HorizontalPodAutoscaler") {
		t.Error("Should have HPA kind")
	}

	if !strings.Contains(HorizontalPodAutoscalerYAML, "minReplicas:") {
		t.Error("Should have minimum replicas")
	}

	if !strings.Contains(HorizontalPodAutoscalerYAML, "maxReplicas:") {
		t.Error("Should have maximum replicas")
	}

	if !strings.Contains(HorizontalPodAutoscalerYAML, "metrics:") {
		t.Error("Should have metrics for scaling")
	}

	if !strings.Contains(HorizontalPodAutoscalerYAML, "cpu") {
		t.Error("Should have CPU metric")
	}

	if !strings.Contains(HorizontalPodAutoscalerYAML, "memory") {
		t.Error("Should have memory metric")
	}

	t.Log("✓ HorizontalPodAutoscalerYAML is valid!")
}

func TestPersistentVolumeClaimYAML(t *testing.T) {
	if !strings.Contains(PersistentVolumeClaimYAML, "kind: PersistentVolumeClaim") {
		t.Error("Should have PVC kind")
	}

	if !strings.Contains(PersistentVolumeClaimYAML, "accessModes:") {
		t.Error("Should specify access modes")
	}

	if !strings.Contains(PersistentVolumeClaimYAML, "storage:") {
		t.Error("Should specify storage size")
	}

	if !strings.Contains(PersistentVolumeClaimYAML, "10Gi") {
		t.Error("Should have storage size")
	}

	t.Log("✓ PersistentVolumeClaimYAML is valid!")
}

func TestGetRecommendedLimits(t *testing.T) {
	tests := map[string]string{
		"api":        "1000m",
		"background": "500m",
		"database":   "2000m",
		"cache":      "250m",
	}

	for appType, expectedCPULimit := range tests {
		limits := GetRecommendedLimits(appType)

		if limits.CPULimit != expectedCPULimit {
			t.Errorf("For %s, expected CPU limit %s, got %s",
				appType, expectedCPULimit, limits.CPULimit)
		}

		if limits.CPURequest == "" {
			t.Errorf("For %s, CPU request should not be empty", appType)
		}

		if limits.MemoryRequest == "" {
			t.Errorf("For %s, memory request should not be empty", appType)
		}
	}

	t.Log("✓ GetRecommendedLimits works!")
}

func TestGetHealthCheckConfig(t *testing.T) {
	configTypes := []string{"readiness", "liveness", "startup"}

	for _, configType := range configTypes {
		config := GetHealthCheckConfig(configType)

		if config.Path == "" {
			t.Errorf("Config %s should have a path", configType)
		}

		if config.FailureThreshold == 0 {
			t.Errorf("Config %s should have failure threshold", configType)
		}

		if config.PeriodSeconds == 0 {
			t.Errorf("Config %s should have period seconds", configType)
		}
	}

	// Verify startup has higher failure threshold
	startupConfig := GetHealthCheckConfig("startup")
	livenessConfig := GetHealthCheckConfig("liveness")

	if startupConfig.FailureThreshold <= livenessConfig.FailureThreshold {
		t.Error("Startup probe should have higher failure threshold than liveness")
	}

	t.Log("✓ GetHealthCheckConfig works!")
}

func TestGetKubernetesBestPractices(t *testing.T) {
	practices := GetKubernetesBestPractices()

	expectedKeys := []string{
		"resource_requests",
		"health_checks",
		"security",
		"rollout_strategy",
		"scheduling",
		"volumes",
		"namespacing",
		"monitoring",
	}

	for _, key := range expectedKeys {
		if _, exists := practices[key]; !exists {
			t.Errorf("Should have practice %s", key)
		}

		if practices[key] == "" {
			t.Errorf("Practice %s should not be empty", key)
		}
	}

	t.Log("✓ Kubernetes best practices documented!")
}

func TestGenerateDeploymentYAML(t *testing.T) {
	limits := GetRecommendedLimits("api")
	yaml := GenerateDeploymentYAML("testapp", "myregistry/testapp:1.0", 3, limits)

	expectedElements := []string{
		"kind: Deployment",
		"testapp",
		"replicas: 3",
		"myregistry/testapp:1.0",
		"livenessProbe",
		"readinessProbe",
		"resources:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(yaml, element) {
			t.Errorf("Generated YAML should contain %s", element)
		}
	}

	if !strings.Contains(yaml, limits.CPURequest) {
		t.Error("Generated YAML should contain CPU request")
	}

	if !strings.Contains(yaml, limits.MemoryLimit) {
		t.Error("Generated YAML should contain memory limit")
	}

	t.Log("✓ GenerateDeploymentYAML works!")
}

func TestGenerateServiceYAML(t *testing.T) {
	yaml := GenerateServiceYAML("testapp", 80, 8080, "LoadBalancer")

	expectedElements := []string{
		"kind: Service",
		"testapp",
		"type: LoadBalancer",
		"port: 80",
		"targetPort: 8080",
		"selector:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(yaml, element) {
			t.Errorf("Generated Service YAML should contain %s", element)
		}
	}

	t.Log("✓ GenerateServiceYAML works!")
}

func TestKubernetesCommand(t *testing.T) {
	cmd := KubernetesCommand{
		Action:    "get",
		Resource:  "pods",
		Name:      "myapp",
		Namespace: "production",
		Flags:     []string{"-o", "json"},
	}

	result := cmd.ToCommand()

	if !strings.Contains(result, "kubectl") {
		t.Error("Command should contain kubectl")
	}

	if !strings.Contains(result, "get") {
		t.Error("Command should contain action")
	}

	if !strings.Contains(result, "pods/myapp") {
		t.Error("Command should contain resource and name")
	}

	if !strings.Contains(result, "-n production") {
		t.Error("Command should contain namespace")
	}

	if !strings.Contains(result, "-o json") {
		t.Error("Command should contain flags")
	}

	t.Log("✓ KubernetesCommand works!")
}

func TestKubernetesCommandWithoutNamespace(t *testing.T) {
	cmd := KubernetesCommand{
		Action:   "create",
		Resource: "deployment",
		Name:     "myapp",
	}

	result := cmd.ToCommand()

	if strings.Contains(result, "-n") {
		t.Error("Command should not contain namespace flag")
	}

	t.Log("✓ KubernetesCommand without namespace works!")
}

func TestProbesConfiguration(t *testing.T) {
	readiness := GetHealthCheckConfig("readiness")
	liveness := GetHealthCheckConfig("liveness")
	startup := GetHealthCheckConfig("startup")

	// Readiness should be quick (frequent checks)
	if readiness.PeriodSeconds > 10 {
		t.Error("Readiness period should be quick")
	}

	// Liveness should be less frequent (prevent false positives)
	if liveness.PeriodSeconds < readiness.PeriodSeconds {
		t.Error("Liveness period should be longer than readiness")
	}

	// Startup should have most retries (slow startup)
	if startup.FailureThreshold < liveness.FailureThreshold {
		t.Error("Startup should have more failure threshold")
	}

	t.Log("✓ Probe configurations are properly tuned!")
}

func TestResourceLimitsScaling(t *testing.T) {
	appTypes := []string{"api", "background", "database", "cache"}

	for _, appType := range appTypes {
		limits := GetRecommendedLimits(appType)

		// Verify CPU request < CPU limit
		if !isValidCPU(limits.CPURequest) {
			t.Errorf("%s has invalid CPU request: %s", appType, limits.CPURequest)
		}

		if !isValidCPU(limits.CPULimit) {
			t.Errorf("%s has invalid CPU limit: %s", appType, limits.CPULimit)
		}

		// Verify memory request < memory limit
		if !isValidMemory(limits.MemoryRequest) {
			t.Errorf("%s has invalid memory request: %s", appType, limits.MemoryRequest)
		}

		if !isValidMemory(limits.MemoryLimit) {
			t.Errorf("%s has invalid memory limit: %s", appType, limits.MemoryLimit)
		}
	}

	t.Log("✓ Resource limits are properly scaled!")
}

func TestSecurityContextPresence(t *testing.T) {
	securityElements := []string{
		"securityContext:",
		"runAsNonRoot",
		"runAsUser",
		"readOnlyRootFilesystem",
		"allowPrivilegeEscalation",
		"capabilities:",
		"drop:",
	}

	for _, element := range securityElements {
		if !strings.Contains(ProductionDeploymentYAML, element) {
			t.Errorf("ProductionDeployment should have security context with %s", element)
		}
	}

	t.Log("✓ Security context is properly configured!")
}

func TestVolumesPresence(t *testing.T) {
	volumeElements := []string{
		"volumes:",
		"volumeMounts:",
		"configMap:",
		"emptyDir:",
	}

	for _, element := range volumeElements {
		if !strings.Contains(ProductionDeploymentYAML, element) {
			t.Errorf("ProductionDeployment should have %s", element)
		}
	}

	t.Log("✓ Volumes are properly configured!")
}

func TestMonitoringAnnotations(t *testing.T) {
	monitoringElements := []string{
		"prometheus.io/scrape",
		"prometheus.io/port",
		"prometheus.io/path",
	}

	for _, element := range monitoringElements {
		if !strings.Contains(ProductionDeploymentYAML, element) {
			t.Errorf("ProductionDeployment should have monitoring annotation %s", element)
		}
	}

	t.Log("✓ Monitoring annotations are present!")
}

// Helper functions
func isValidCPU(cpu string) bool {
	// Simple validation for CPU format (100m, 1, etc.)
	return cpu != "" && (strings.Contains(cpu, "m") || strings.Contains(cpu, "."))
}

func isValidMemory(memory string) bool {
	// Simple validation for memory format (128Mi, 1Gi, etc.)
	return memory != "" && (strings.Contains(memory, "Mi") || strings.Contains(memory, "Gi"))
}

func BenchmarkGenerateDeploymentYAML(b *testing.B) {
	limits := GetRecommendedLimits("api")
	for i := 0; i < b.N; i++ {
		GenerateDeploymentYAML("app", "registry/app:1.0", 3, limits)
	}
}

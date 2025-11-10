package main

import (
	"fmt"
)

// ========== Kubernetes YAML Templates ==========

// SimpleDeploymentYAML is a basic Kubernetes deployment
const SimpleDeploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: default
  labels:
    app: myapp
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1
    spec:
      containers:
      - name: myapp
        image: myregistry.azurecr.io/myapp:1.0.0
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: PORT
          value: "8080"
        - name: LOG_LEVEL
          value: "info"
`

// ServiceYAML exposes the deployment
const ServiceYAML = `
apiVersion: v1
kind: Service
metadata:
  name: myapp
  namespace: default
  labels:
    app: myapp
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: myapp
---
apiVersion: v1
kind: Service
metadata:
  name: myapp-lb
  namespace: default
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: myapp
`

// ConfigMapYAML for application configuration
const ConfigMapYAML = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
  namespace: default
  labels:
    app: myapp
data:
  app.yaml: |
    server:
      port: 8080
      readTimeout: 30s
      writeTimeout: 30s
    database:
      host: postgres.default.svc.cluster.local
      port: 5432
      maxConnections: 20
    logging:
      level: info
      format: json
  database.sql: |
    CREATE TABLE IF NOT EXISTS users (
      id SERIAL PRIMARY KEY,
      email VARCHAR(255) UNIQUE NOT NULL,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
`

// SecretsYAML for sensitive data
const SecretsYAML = `
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secrets
  namespace: default
type: Opaque
stringData:
  database-url: "postgres://user:password@postgres:5432/myapp"
  jwt-secret: "your-super-secret-jwt-key-change-in-production"
  api-key: "api-key-for-external-services"
---
apiVersion: v1
kind: Secret
metadata:
  name: docker-registry
  namespace: default
type: kubernetes.io/dockercfg
data:
  .dockercfg: eyJteXJlZ2lzdHJ5LmF6dXJlY3IuaW8iOnsidXNlcm5hbWUiOiJ1c2VybmFtZSIsInBhc3N3b3JkIjoicGFzc3dvcmQiLCJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20ifX0=
`

// ProductionDeploymentYAML with health checks and resource limits
const ProductionDeploymentYAML = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
  labels:
    app: myapp
    version: v1
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      # Pod scheduling
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - myapp
              topologyKey: kubernetes.io/hostname

      # Security context for pod
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault

      # Image pull secrets
      imagePullSecrets:
      - name: docker-registry

      containers:
      - name: myapp
        image: myregistry.azurecr.io/myapp:1.0.0
        imagePullPolicy: IfNotPresent

        # Ports
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        - containerPort: 9090
          name: metrics
          protocol: TCP

        # Environment variables
        env:
        - name: PORT
          value: "8080"
        - name: LOG_LEVEL
          value: "info"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: jwt-secret
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace

        # Volume mounts
        volumeMounts:
        - name: config
          mountPath: /app/config
          readOnly: true
        - name: cache
          mountPath: /app/cache

        # Resource limits and requests
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi

        # Liveness probe - is app alive?
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3

        # Readiness probe - is app ready for traffic?
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          successThreshold: 1
          failureThreshold: 2

        # Startup probe - has app started?
        startupProbe:
          httpGet:
            path: /startup
            port: 8080
          failureThreshold: 30
          periodSeconds: 1

        # Security context for container
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL

      # Volumes
      volumes:
      - name: config
        configMap:
          name: myapp-config
      - name: cache
        emptyDir:
          sizeLimit: 1Gi

      # Pod lifecycle
      terminationGracePeriodSeconds: 30
      restartPolicy: Always
      dnsPolicy: ClusterFirst
`

// IngressYAML for HTTP routing
const IngressYAML = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  namespace: production
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - myapp.example.com
    secretName: myapp-tls
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: myapp
            port:
              number: 80
`

// HorizontalPodAutoscalerYAML for auto-scaling
const HorizontalPodAutoscalerYAML = `
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 15
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 2
        periodSeconds: 15
      selectPolicy: Max
`

// PersistentVolumeClaimYAML for data persistence
const PersistentVolumeClaimYAML = `
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: myapp-data
  namespace: production
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 10Gi
`

// ========== Kubernetes Patterns and Helpers ==========

// ResourceLimits defines CPU and memory limits
type ResourceLimits struct {
	CPURequest    string // e.g., "100m"
	CPULimit      string // e.g., "500m"
	MemoryRequest string // e.g., "128Mi"
	MemoryLimit   string // e.g., "512Mi"
}

// GetRecommendedLimits returns production-recommended limits
func GetRecommendedLimits(appType string) ResourceLimits {
	limits := map[string]ResourceLimits{
		"api": {
			CPURequest:    "200m",
			CPULimit:      "1000m",
			MemoryRequest: "256Mi",
			MemoryLimit:   "1Gi",
		},
		"background": {
			CPURequest:    "100m",
			CPULimit:      "500m",
			MemoryRequest: "128Mi",
			MemoryLimit:   "512Mi",
		},
		"database": {
			CPURequest:    "500m",
			CPULimit:      "2000m",
			MemoryRequest: "512Mi",
			MemoryLimit:   "2Gi",
		},
		"cache": {
			CPURequest:    "100m",
			CPULimit:      "250m",
			MemoryRequest: "64Mi",
			MemoryLimit:   "256Mi",
		},
	}

	if limits, ok := limits[appType]; ok {
		return limits
	}

	// Default
	return ResourceLimits{
		CPURequest:    "100m",
		CPULimit:      "500m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "512Mi",
	}
}

// HealthCheckConfig defines probe settings
type HealthCheckConfig struct {
	Path                 string
	InitialDelaySeconds  int
	TimeoutSeconds       int
	PeriodSeconds        int
	SuccessThreshold     int
	FailureThreshold     int
}

// GetHealthCheckConfig returns probe configuration
func GetHealthCheckConfig(probeType string) HealthCheckConfig {
	configs := map[string]HealthCheckConfig{
		"readiness": {
			Path:                "/ready",
			InitialDelaySeconds: 10,
			TimeoutSeconds:      3,
			PeriodSeconds:       5,
			SuccessThreshold:    1,
			FailureThreshold:    2,
		},
		"liveness": {
			Path:                "/health",
			InitialDelaySeconds: 30,
			TimeoutSeconds:      5,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    3,
		},
		"startup": {
			Path:                "/startup",
			InitialDelaySeconds: 0,
			TimeoutSeconds:      3,
			PeriodSeconds:       1,
			SuccessThreshold:    1,
			FailureThreshold:    30,
		},
	}

	if config, ok := configs[probeType]; ok {
		return config
	}

	return configs["readiness"]
}

// ========== Kubernetes Best Practices Documentation ==========

// GetKubernetesBestPractices returns documentation
func GetKubernetesBestPractices() map[string]string {
	return map[string]string{
		"resource_requests": `
Always set resource requests and limits:
- Requests: Kubernetes uses for scheduling
- Limits: Maximum resources the pod can use
- CPU: 100m = 0.1 CPU core
- Memory: Mi (mebibyte), Gi (gibibyte)
Example: 100m CPU request, 500m CPU limit`,

		"health_checks": `
Implement three probe types:
- Readiness: Is app ready to handle traffic? (quick check)
- Liveness: Is app healthy? (slower check)
- Startup: Has app started? (slow startup apps)
Each probe has different timing requirements`,

		"security": `
Kubernetes security context:
- runAsNonRoot: true (no root user)
- readOnlyRootFilesystem: true (when possible)
- allowPrivilegeEscalation: false
- drop: ALL (drop all capabilities)
Use network policies to restrict traffic`,

		"rollout_strategy": `
RollingUpdate strategy for zero-downtime:
- maxSurge: Pods above desired count during update
- maxUnavailable: Pods that can be unavailable
Example: maxSurge: 1 (add 1 pod, remove 1 when ready)`,

		"scheduling": `
Pod affinity for optimal placement:
- Pod Anti-affinity: Spread pods across nodes
- Node affinity: Schedule on specific nodes
- Taints and tolerations: Node preferences
Example: Spread replicas across different nodes`,

		"volumes": `
Kubernetes volume types:
- emptyDir: Temporary storage, pod-scoped
- configMap: Configuration data
- secret: Sensitive data
- persistentVolumeClaim: Persistent storage
- downwardAPI: Pod metadata as files`,

		"namespacing": `
Use namespaces for isolation:
- Separate environments (dev, staging, prod)
- Isolate teams or services
- Resource quotas per namespace
- Network policies per namespace`,

		"monitoring": `
Prometheus monitoring integration:
- Add annotations for scraping:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
- Export metrics in Prometheus format`,
	}
}

// GenerateDeploymentYAML creates custom deployment
func GenerateDeploymentYAML(name, image string, replicas int, limits ResourceLimits) string {
	template := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: %s
            memory: %s
          limits:
            cpu: %s
            memory: %s
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
`
	return fmt.Sprintf(template,
		name, name, replicas, name, name, name, image,
		limits.CPURequest, limits.MemoryRequest, limits.CPULimit, limits.MemoryLimit)
}

// GenerateServiceYAML creates custom service
func GenerateServiceYAML(name string, port, targetPort int, serviceType string) string {
	template := `
apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  type: %s
  ports:
  - port: %d
    targetPort: %d
  selector:
    app: %s
`
	return fmt.Sprintf(template, name, serviceType, port, targetPort, name)
}

// KubernetesCommand represents kubectl commands
type KubernetesCommand struct {
	Action    string // create, apply, delete, get, describe
	Resource  string // deployment, service, pod
	Name      string
	Namespace string
	Flags     []string
}

// ToCommand converts to kubectl command string
func (k *KubernetesCommand) ToCommand() string {
	cmd := fmt.Sprintf("kubectl %s %s/%s", k.Action, k.Resource, k.Name)

	if k.Namespace != "" {
		cmd += fmt.Sprintf(" -n %s", k.Namespace)
	}

	for _, flag := range k.Flags {
		cmd += fmt.Sprintf(" %s", flag)
	}

	return cmd
}

func main() {}

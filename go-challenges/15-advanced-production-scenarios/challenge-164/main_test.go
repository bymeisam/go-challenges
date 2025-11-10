package main

import (
	"testing"
	"time"
)

func TestCreateVersion(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances := []*Instance{
		{ID: "i1", Status: "starting"},
		{ID: "i2", Status: "starting"},
	}

	version, err := dc.CreateVersion("v1.0.0", instances)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if version.ID == "" {
		t.Fatal("Expected non-empty version ID")
	}

	if version.Status != "pending" {
		t.Fatalf("Expected pending status, got %s", version.Status)
	}
}

func TestBlueGreenDeployment(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances := []*Instance{
		{ID: "i1", Status: "starting"},
		{ID: "i2", Status: "starting"},
	}

	version, _ := dc.CreateVersion("v1.0.0", instances)

	// Make instances healthy
	for _, instance := range version.Instances {
		instance.Status = "healthy"
	}

	// Simulate health checks
	dc.healthChecker.CheckVersion(version)

	time.Sleep(100 * time.Millisecond)

	err := dc.DeployBlueGreen(version.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if dc.activeVersion != version.ID {
		t.Fatal("Expected version to be active")
	}
}

func TestCanaryDeployment(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances := []*Instance{
		{ID: "i1", Status: "healthy"},
		{ID: "i2", Status: "healthy"},
	}

	version, _ := dc.CreateVersion("v1.0.1", instances)

	err := dc.DeployCanary(version.ID, 10, 1*time.Second)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	traffic := dc.trafficRouter.GetTrafficDistribution()
	if len(traffic) == 0 {
		t.Fatal("Expected traffic to be routed")
	}
}

func TestInvalidCanaryPercent(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances := []*Instance{{ID: "i1", Status: "healthy"}}
	version, _ := dc.CreateVersion("v1.0.1", instances)

	err := dc.DeployCanary(version.ID, 0, 1*time.Second)
	if err == nil {
		t.Fatal("Expected error for invalid canary percent")
	}

	err = dc.DeployCanary(version.ID, 100, 1*time.Second)
	if err == nil {
		t.Fatal("Expected error for canary percent >= 100")
	}
}

func TestRollback(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances1 := []*Instance{{ID: "i1", Status: "healthy"}}
	version1, _ := dc.CreateVersion("v1.0.0", instances1)

	instances2 := []*Instance{{ID: "i2", Status: "healthy"}}
	version2, _ := dc.CreateVersion("v1.0.1", instances2)

	dc.DeployBlueGreen(version1.ID)
	dc.DeployBlueGreen(version2.ID)

	err := dc.RollbackToVersion(0)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if dc.activeVersion != version1.ID {
		t.Fatal("Expected rollback to version1")
	}
}

func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker()

	instance := &Instance{ID: "i1", Status: "starting"}
	version := &DeploymentVersion{
		ID:        "v1",
		Instances: []*Instance{instance},
	}

	hc.CheckVersion(version)

	time.Sleep(100 * time.Millisecond)

	healthy := hc.GetHealthyCount("v1")
	if healthy == 0 {
		t.Fatal("Expected healthy count")
	}
}

func TestTrafficRouter(t *testing.T) {
	tr := NewTrafficRouter()

	tr.SwitchTraffic("v1", 1.0)
	tr.SwitchTraffic("v2", 0.0)

	distribution := tr.GetTrafficDistribution()

	if distribution["v1"] != 100.0 {
		t.Fatalf("Expected 100%% traffic to v1, got %f", distribution["v1"])
	}
}

func TestGracefulShutdown(t *testing.T) {
	instance := &Instance{ID: "i1", Status: "healthy", Connections: 0}
	gs := NewGracefulShutdown(instance, 1*time.Second)

	err := gs.Shutdown()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if instance.Status != "terminated" {
		t.Fatalf("Expected terminated status, got %s", instance.Status)
	}
}

func TestMetricsMonitor(t *testing.T) {
	monitor := NewMetricsMonitor()

	monitor.RecordError("v1")
	monitor.RecordError("v1")
	monitor.RecordLatency("v1", 10*time.Millisecond)
	monitor.RecordLatency("v1", 20*time.Millisecond)

	errorRate := monitor.GetErrorRate("v1")
	if errorRate != 2 {
		t.Fatalf("Expected 2 errors, got %f", errorRate)
	}
}

func TestGetStatus(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances := []*Instance{{ID: "i1", Status: "healthy"}}
	version, _ := dc.CreateVersion("v1.0.0", instances)

	dc.versionsMu.Lock()
	dc.activeVersion = version.ID
	dc.versionsMu.Unlock()

	status := dc.GetStatus()

	if status["active_version"] != "v1.0.0" {
		t.Fatal("Expected version in status")
	}

	if status["status"] != "pending" {
		t.Fatal("Expected pending status")
	}
}

func TestDeploymentEvents(t *testing.T) {
	dc := NewDeploymentCoordinator()

	instances := []*Instance{{ID: "i1", Status: "healthy"}}
	version, _ := dc.CreateVersion("v1.0.0", instances)

	dc.DeployBlueGreen(version.ID)

	dc.logMu.RLock()
	logCount := len(dc.deploymentLog)
	dc.logMu.RUnlock()

	if logCount == 0 {
		t.Fatal("Expected deployment events logged")
	}
}

func TestMultipleVersions(t *testing.T) {
	dc := NewDeploymentCoordinator()

	for i := 0; i < 5; i++ {
		instances := []*Instance{{ID: "i" + string(rune(i))}}
		dc.CreateVersion("v1.0."+string(rune(i)), instances)
	}

	dc.versionsMu.RLock()
	versionCount := len(dc.versions)
	dc.versionsMu.RUnlock()

	if versionCount != 5 {
		t.Fatalf("Expected 5 versions, got %d", versionCount)
	}
}

func TestTrafficDistribution(t *testing.T) {
	tr := NewTrafficRouter()

	tr.SwitchTraffic("v1", 0.7)
	tr.SwitchTraffic("v2", 0.3)

	for i := 0; i < 100; i++ {
		tr.RouteRequest("v1")
	}

	for i := 0; i < 30; i++ {
		tr.RouteRequest("v2")
	}

	total := tr.totalRequests
	if total != 130 {
		t.Fatalf("Expected 130 total requests, got %d", total)
	}
}

func TestInstanceStatusTransition(t *testing.T) {
	instance := &Instance{
		ID:     "i1",
		Status: "starting",
	}

	instance.Status = "healthy"
	if instance.Status != "healthy" {
		t.Fatal("Expected healthy status")
	}

	instance.Status = "draining"
	if instance.Status != "draining" {
		t.Fatal("Expected draining status")
	}
}

func TestConcurrentDeployment(t *testing.T) {
	dc := NewDeploymentCoordinator()

	done := make(chan bool, 3)

	// Simulated concurrent deployments
	for i := 0; i < 3; i++ {
		go func(index int) {
			instances := []*Instance{{ID: "i" + string(rune(index))}}
			version, _ := dc.CreateVersion("v1.0."+string(rune(index)), instances)
			_ = dc.DeployBlueGreen(version.ID)
			done <- true
		}(i)
	}

	for i := 0; i < 3; i++ {
		<-done
	}

	dc.versionsMu.RLock()
	versionCount := len(dc.versions)
	dc.versionsMu.RUnlock()

	if versionCount != 3 {
		t.Fatalf("Expected 3 versions after concurrent deploys, got %d", versionCount)
	}
}

// ========== Benchmarks ==========

func BenchmarkBlueGreenDeployment(b *testing.B) {
	dc := NewDeploymentCoordinator()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		instances := []*Instance{{ID: "i1", Status: "healthy"}}
		version, _ := dc.CreateVersion("v"+string(rune(i)), instances)
		dc.DeployBlueGreen(version.ID)
	}
}

func BenchmarkTrafficRouting(b *testing.B) {
	tr := NewTrafficRouter()
	tr.SwitchTraffic("v1", 1.0)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tr.RouteRequest("v1")
	}
}

func BenchmarkHealthChecking(b *testing.B) {
	hc := NewHealthChecker()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		instance := &Instance{ID: "i" + string(rune(i))}
		version := &DeploymentVersion{
			ID:        "v" + string(rune(i)),
			Instances: []*Instance{instance},
		}
		hc.CheckVersion(version)
	}
}

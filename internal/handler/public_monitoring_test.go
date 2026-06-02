package handler

import (
	"net/http"
	"testing"

	"raevtar/internal/model"
)

func TestPublicDashboardShowsSafeExtendedNodeHealth(t *testing.T) {
	app := newPublicTestApp(t)
	recordExtendedPublicMetric(t, app)

	status, body := getBody(t, app, "/dashboard", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	for _, want := range []string{"System Health", "CPU", "12.5%", "Load", "0.0 / 0.3 / 0.2", "4c", "RAM", "256.0 / 1024.0 MB · 25.0%", "Disk", "8.0 / 32.0 GB · 25.0%", "Temp", "48.5°C", "Uptime", "1h 0m", "1 sample", "100% online in recent samples"} {
		assertContains(t, body, want)
	}
	assertPublicMonitoringRedaction(t, body)
}

func TestPublicServerDetailShowsSafeExtendedNodeHealth(t *testing.T) {
	app := newPublicTestApp(t)
	recordExtendedPublicMetric(t, app)

	status, body := getBody(t, app, "/dashboard/1/live", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	for _, want := range []string{"System Health", "CPU percent", "12.5%", "Load 0.0 / 0.3 / 0.2 · 4 cores", "RAM", "256.0 / 1024.0 MB · 25.0%", "Disk", "8.0 / 32.0 GB · 25.0%", "Temperature", "48.5°C", "Uptime", "1h 0m", "Latest sample", "1 sample", "100% online in recent samples"} {
		assertContains(t, body, want)
	}
	assertPublicMonitoringRedaction(t, body)
}

func TestPublicServerDetailDoesNotRenderHistoricalSampleRows(t *testing.T) {
	app := newPublicTestApp(t)
	recordExtendedPublicMetric(t, app)
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:           77.7,
		CPULoad1:             1.1,
		CPULoad5:             1.2,
		CPULoad15:            1.3,
		CPUCores:             8,
		RAMUsedMB:            512,
		RAMTotalMB:           1024,
		DiskUsedGB:           16,
		DiskTotalGB:          32,
		TemperatureC:         55.5,
		TemperatureAvailable: true,
		UptimeSeconds:        7200,
		Online:               false,
	}); err != nil {
		t.Fatalf("record second metrics: %v", err)
	}

	status, body := getBody(t, app, "/dashboard/1/live", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	for _, want := range []string{"System Health", "77.7%", "1.1 / 1.2 / 1.3", "8 cores", "2 samples", "captured", "50% online in recent samples", `hx-get="/dashboard/1/live"`, `hx-trigger="every 15s"`, `hx-swap="outerHTML"`} {
		assertContains(t, body, want)
	}
	for _, hidden := range []string{"Metrics Terakhir", "12.5%", "0.0 / 0.3 / 0.2", "48.5°C", "1h 0m", "Jan 2 15:04"} {
		assertNotContains(t, body, hidden)
	}
	assertPublicMonitoringRedaction(t, body)
}

func recordExtendedPublicMetric(t *testing.T, app *publicTestApp) {
	t.Helper()
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:           12.5,
		CPULoad1:             0,
		CPULoad5:             0.3,
		CPULoad15:            0.2,
		CPUCores:             4,
		RAMUsedMB:            256,
		RAMTotalMB:           1024,
		DiskUsedGB:           8,
		DiskTotalGB:          32,
		TemperatureC:         48.5,
		TemperatureAvailable: true,
		UptimeSeconds:        3600,
		Online:               true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}
}

func assertPublicMonitoringRedaction(t *testing.T, body string) {
	t.Helper()
	for _, leak := range []string{"127.0.0.1", "127.0.0.1:9100", "port 9100", ">local<", "agent token", "raevtar-agent.sh", "POST /api/v1/servers", "install", "cron", "audit", "token"} {
		assertNotContains(t, body, leak)
	}
}

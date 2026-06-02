package pages

import (
	"strconv"
	"strings"
	"time"

	"raevtar/internal/model"
)

const (
	freshSignalWindow = 3 * time.Minute
	staleSignalWindow = 15 * time.Minute
)

type IndexData struct {
	CurrentPath string
	Posts       []model.Post
	PostCount   int
	Servers     []model.Server
	Categories  []model.Category
	Domain      string
}

type BlogListData struct {
	CurrentPath string
	Posts       []model.Post
	Categories  []model.Category
	CurrentCat  string
	Page        int
	TotalPages  int
}

type BlogPostData struct {
	CurrentPath string
	Post        *model.Post
	Categories  []model.Category
}

type LabData struct {
	CurrentPath   string
	Categories    []model.Category
	PostCount     int
	CategoryCount int
	ServerCount   int
}

type DocsData struct {
	CurrentPath string
	Categories  []model.Category
	PostCount   int
}

type DashboardData struct {
	CurrentPath     string
	Servers         []model.Server
	ServerSummaries []PublicServerSummary
	Categories      []model.Category
	PlatformHealth  PublicHostHealthData
	RefreshedAt     time.Time
}

type PublicHostHealthData struct {
	CPULoad     string
	CPUCores    string
	RAMUsage    string
	RAMPercent  string
	DiskUsage   string
	DiskPercent string
	Temperature string
}

type PublicServerSummary struct {
	Server  model.Server
	Metrics []model.ServerMetric
}

type ServerDetailData struct {
	CurrentPath string
	Server      *model.Server
	Metrics     []model.ServerMetric
	Categories  []model.Category
	RefreshedAt time.Time
}

type NotFoundData struct {
	CurrentPath string
	Categories  []model.Category
}

func PortText(port int) string {
	return strconv.Itoa(port)
}

func IDText(id int64) string {
	return strconv.FormatInt(id, 10)
}

func MetricText(value float64) string {
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func MetricStatusText(online bool) string {
	if online {
		return "Online"
	}
	return "Offline"
}

func LastSeenText(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "No signal yet"
	}
	return lastSeen.Format("Jan 2 15:04")
}

func LastSignalAgeText(lastSeen *time.Time, now time.Time) string {
	if lastSeen == nil {
		return "No signal yet"
	}
	return AgeText(now.Sub(*lastSeen)) + " ago"
}

func LatestMetricTimestampText(metrics []model.ServerMetric) string {
	if len(metrics) == 0 || metrics[0].RecordedAt.IsZero() {
		return "No metrics yet"
	}
	return metrics[0].RecordedAt.Format("Jan 2 15:04:05 UTC")
}

func RefreshTimeText(refreshedAt time.Time) string {
	if refreshedAt.IsZero() {
		return "Unknown"
	}
	return refreshedAt.Format("Jan 2 15:04:05 UTC")
}

func FreshnessCauseHint(lastSeen *time.Time, metrics []model.ServerMetric, now time.Time) string {
	if lastSeen == nil {
		return "No telemetry received yet. Waiting for first agent signal."
	}
	if len(metrics) == 0 {
		return "Signal timestamp exists, but no metrics sample is stored yet."
	}

	age := now.Sub(*lastSeen)
	if age < 0 {
		age = 0
	}
	switch {
	case age < freshSignalWindow:
		return "Telemetry is fresh. Latest agent signal arrived recently."
	case age < staleSignalWindow:
		return "Telemetry is delayed. Agent may be between scheduled reports."
	default:
		return "Telemetry is offline. No recent agent signal has reached Raevtar."
	}
}

func DashboardStatusHint(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "No agent signal has arrived yet."
	}
	age := time.Since(*lastSeen)
	if age < 0 {
		age = 0
	}
	switch {
	case age < 3*time.Minute:
		return "Agent checked in recently."
	case age < 15*time.Minute:
		return "Agent signal is delayed; one cron run may have been missed."
	default:
		return "No recent agent signal reached Raevtar."
	}
}

func OnlineServerCount(servers []model.Server) int {
	count := 0
	for _, server := range servers {
		if server.LastSeen != nil && time.Since(*server.LastSeen) < 3*time.Minute {
			count++
		}
	}
	return count
}

func StaleServerCount(servers []model.Server) int {
	count := 0
	for _, server := range servers {
		if server.LastSeen != nil && time.Since(*server.LastSeen) >= 3*time.Minute && time.Since(*server.LastSeen) < 15*time.Minute {
			count++
		}
	}
	return count
}

func OfflineServerCount(servers []model.Server) int {
	count := 0
	for _, server := range servers {
		if server.LastSeen == nil || time.Since(*server.LastSeen) >= 15*time.Minute {
			count++
		}
	}
	return count
}

func MetricAvailabilityText(metrics []model.ServerMetric) string {
	if len(metrics) == 0 {
		return "No availability window yet"
	}
	online := 0
	for _, metric := range metrics {
		if metric.Online {
			online++
		}
	}
	percent := (online * 100) / len(metrics)
	return strconv.Itoa(percent) + "% online in recent samples"
}

func DashboardFreshnessReason(lastSeen *time.Time) string {
	return DashboardFreshnessReasonAt(lastSeen, time.Now())
}

func DashboardFreshnessReasonAt(lastSeen *time.Time, now time.Time) string {
	if lastSeen == nil {
		return "Why: no last agent signal has reached Raevtar yet."
	}
	age := now.Sub(*lastSeen)
	if age < 0 {
		age = 0
	}
	switch {
	case age < freshSignalWindow:
		return "Why: last agent signal is inside the <3m online window."
	case age < staleSignalWindow:
		return "Why: last agent signal is older than 3m but still inside the 15m stale window."
	default:
		return "Why: last agent signal is older than 15m, so public status is offline."
	}
}

func FreshnessWindowText(lastSeen *time.Time, now time.Time) string {
	if lastSeen == nil {
		return "No signal recorded"
	}
	age := now.Sub(*lastSeen)
	if age < 0 {
		age = 0
	}
	switch {
	case age < freshSignalWindow:
		return "Fresh window: <3m"
	case age < staleSignalWindow:
		return "Stale window: 3-15m"
	default:
		return "Offline window: 15m+"
	}
}

func MetricSampleCountText(metrics []model.ServerMetric) string {
	if len(metrics) == 1 {
		return "1 sample"
	}
	return strconv.Itoa(len(metrics)) + " samples"
}

func MetricWindowText(metrics []model.ServerMetric) string {
	if len(metrics) == 0 {
		return "No history yet"
	}
	newest := metrics[0].RecordedAt
	oldest := metrics[len(metrics)-1].RecordedAt
	if newest.IsZero() || oldest.IsZero() {
		return "History window unknown"
	}
	return AgeText(newest.Sub(oldest)) + " captured"
}

func MetricRecordedAgeText(recordedAt time.Time, now time.Time) string {
	if recordedAt.IsZero() {
		return "unknown age"
	}
	return AgeText(now.Sub(recordedAt)) + " ago"
}

func AgeText(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	if duration < time.Minute {
		return "<1m"
	}
	minutes := int(duration.Minutes())
	if minutes < 60 {
		return strconv.Itoa(minutes) + "m"
	}
	hours := minutes / 60
	if hours < 24 {
		return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes%60) + "m"
	}
	days := hours / 24
	return strconv.Itoa(days) + "d " + strconv.Itoa(hours%24) + "h"
}

func CreatedText(createdAt time.Time) string {
	if createdAt.IsZero() {
		return "Unknown"
	}
	return createdAt.Format("Jan 2 2006")
}

func RAMUsageText(used, total float64) string {
	if total <= 0 {
		return MetricText(used) + " MB"
	}
	return MetricText(used) + " / " + MetricText(total) + " MB"
}

func DashboardServerSummaries(data DashboardData) []PublicServerSummary {
	if len(data.ServerSummaries) > 0 || len(data.Servers) == 0 {
		return data.ServerSummaries
	}
	summaries := make([]PublicServerSummary, 0, len(data.Servers))
	for _, server := range data.Servers {
		summaries = append(summaries, PublicServerSummary{Server: server})
	}
	return summaries
}

func CPULoadText(metric model.ServerMetric) string {
	if metric.CPUCores <= 0 {
		return "N/A"
	}
	return MetricText(metric.CPULoad1) + " / " + MetricText(metric.CPULoad5) + " / " + MetricText(metric.CPULoad15)
}

func CPUCoresText(metric model.ServerMetric) string {
	if metric.CPUCores <= 0 {
		return "N/A"
	}
	return strconv.FormatInt(metric.CPUCores, 10)
}

func ResourcePercentText(used, total float64) string {
	if total <= 0 {
		return "N/A"
	}
	return MetricText(used/total*100) + "%"
}

func RAMHealthText(metric model.ServerMetric) string {
	if metric.RAMTotalMB <= 0 {
		return MetricText(metric.RAMUsedMB) + " MB / N/A"
	}
	return RAMUsageText(metric.RAMUsedMB, metric.RAMTotalMB) + " · " + ResourcePercentText(metric.RAMUsedMB, metric.RAMTotalMB)
}

func DiskHealthText(metric model.ServerMetric) string {
	if metric.DiskTotalGB <= 0 {
		return MetricText(metric.DiskUsedGB) + " GB / N/A"
	}
	return MetricText(metric.DiskUsedGB) + " / " + MetricText(metric.DiskTotalGB) + " GB · " + ResourcePercentText(metric.DiskUsedGB, metric.DiskTotalGB)
}

func TemperatureText(metric model.ServerMetric) string {
	if !metric.TemperatureAvailable {
		return "N/A"
	}
	return MetricText(metric.TemperatureC) + "°C"
}

func PublicServerLabel(server model.Server) string {
	if server.Tags != "" {
		return "tagged node"
	}
	return "registered node"
}

func UptimeText(seconds int64) string {
	if seconds <= 0 {
		return "Unknown"
	}
	d := time.Duration(seconds) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return strconv.Itoa(days) + "d " + strconv.Itoa(hours) + "h"
	}
	if hours > 0 {
		return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"
	}
	return strconv.Itoa(minutes) + "m"
}

func ReadMinutes(markdown string) string {
	words := len(strings.Fields(markdown))
	if words == 0 {
		return "<1"
	}
	minutes := words / 200
	if minutes < 1 {
		return "<1"
	}
	return strconv.Itoa(minutes)
}

func TagClass(name string) string {
	switch name {
	case "auto":
		return "text-xs px-2 py-0.5 font-bold border-2 border-retro-ink bg-retro-sageLight text-retro-ink"
	case "commissioned":
		return "text-xs px-2 py-0.5 font-bold border-2 border-retro-ink bg-retro-wheat text-retro-ink"
	default:
		return "text-xs px-2 py-0.5 font-bold border-2 border-retro-ink bg-retro-paper text-retro-ink"
	}
}

func ServerStatusText(lastSeen *time.Time) string {
	return ServerStatusTextAt(lastSeen, time.Now())
}

func ServerStatusTextAt(lastSeen *time.Time, now time.Time) string {
	switch {
	case lastSeen != nil && now.Sub(*lastSeen) < freshSignalWindow:
		return "Online"
	case lastSeen != nil && now.Sub(*lastSeen) < staleSignalWindow:
		return "Stale"
	default:
		return "Offline"
	}
}

func DashboardStatusText(lastSeen *time.Time) string {
	return DashboardStatusTextAt(lastSeen, time.Now())
}

func DashboardStatusTextAt(lastSeen *time.Time, now time.Time) string {
	if lastSeen == nil {
		return "Offline — no signal"
	}
	return ServerStatusTextAt(lastSeen, now) + " — " + LastSignalAgeText(lastSeen, now)
}

func StatusDotClass(lastSeen *time.Time) string {
	return StatusDotClassAt(lastSeen, time.Now())
}

func StatusDotClassAt(lastSeen *time.Time, now time.Time) string {
	switch {
	case lastSeen != nil && now.Sub(*lastSeen) < freshSignalWindow:
		return "bg-retro-sage"
	case lastSeen != nil && now.Sub(*lastSeen) < staleSignalWindow:
		return "bg-retro-wheat"
	default:
		return "bg-retro-blush"
	}
}

func MetricDotClass(online bool) string {
	if online {
		return "bg-retro-sage"
	}
	return "bg-retro-blush"
}

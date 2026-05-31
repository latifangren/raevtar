package pages

import (
	"strconv"
	"strings"
	"time"

	"raevtar/internal/model"
)

type IndexData struct {
	CurrentPath string
	Posts       []model.Post
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

type DashboardData struct {
	CurrentPath       string
	Servers           []model.Server
	Categories        []model.Category
	CanRegisterServer bool
	CanViewServerInfo bool
	CSRFToken         string
}

type ServerDetailData struct {
	CurrentPath       string
	Server            *model.Server
	Metrics           []model.ServerMetric
	Categories        []model.Category
	CanManageServer   bool
	CanViewServerInfo bool
	RefreshedAt       time.Time
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
	case age < 3*time.Minute:
		return "Telemetry is fresh. Latest agent signal arrived recently."
	case age < 15*time.Minute:
		return "Telemetry is delayed. Agent may be between scheduled reports."
	default:
		return "Telemetry is offline. No recent agent signal has reached Raevtar."
	}
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
	switch {
	case lastSeen != nil && time.Since(*lastSeen) < 3*time.Minute:
		return "Online"
	case lastSeen != nil && time.Since(*lastSeen) < 15*time.Minute:
		return "Stale"
	default:
		return "Offline"
	}
}

func DashboardStatusText(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "Offline"
	}
	return ServerStatusText(lastSeen) + " — " + lastSeen.Format("Jan 2 15:04")
}

func StatusDotClass(lastSeen *time.Time) string {
	switch {
	case lastSeen != nil && time.Since(*lastSeen) < 3*time.Minute:
		return "bg-retro-sage"
	case lastSeen != nil && time.Since(*lastSeen) < 15*time.Minute:
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

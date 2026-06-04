package admin

import (
	"strconv"
	"strings"
	"time"

	"raevtar/internal/model"
)

type LoginData struct{}

type DashboardData struct {
	CurrentPath string
	CSRFToken   string
	PostCount   int
	ServerCount int
	UserCount   int
	OnlineCount int
	Servers     []model.Server
	Stats       HostStatsData
}

type HostStatsData struct {
	CPULoad1       string
	CPULoad5       string
	CPUCores       int
	RAMUsed        string
	RAMTotal       string
	RAMPercent     float64
	DiskUsed       string
	DiskTotal      string
	DiskPercent    float64
	Temperature    string
	TempValue      float64
	CPULoadPercent float64
}

type UsersData struct {
	CurrentPath string
	CSRFToken   string
	Users       []UserRow
	RoleOptions []RoleOption
}

type UserRow struct {
	User      model.User
	CanDelete bool
}

type RoleOption struct {
	Value    string
	Selected bool
}

type AuditData struct {
	CurrentPath string
	CSRFToken   string
	Logs        []model.AuditLog
}

type PostsData struct {
	CurrentPath string
	CSRFToken   string
	Posts       []model.Post
	Categories  []model.Category
	MediaAssets []model.MediaAsset
}

type PostEditData struct {
	CurrentPath string
	CSRFToken   string
	Post        *model.Post
	Categories  []model.Category
	MediaAssets []model.MediaAsset
}

type MediaData struct {
	CurrentPath string
	CSRFToken   string
	Assets      []model.MediaAsset
}

type ServersData struct {
	CurrentPath            string
	CSRFToken              string
	Servers                []model.Server
	AgentURLExample        string
	GeneratedTokenServerID int64
	GeneratedAgentToken    string
}

type ServerDetailData struct {
	CurrentPath     string
	CSRFToken       string
	Server          *model.Server
	Metrics         []model.ServerMetric
	Logs            []model.AuditLog
	AgentURLExample string
}

type EditorialInboxData struct {
	CurrentPath string
	CSRFToken   string
	Now         time.Time
	Items       []model.EditorialInboxItem
	Counts      map[string]int
	Summary     *model.EditorialInboxSummary
	Categories  []model.Category
	Modes       []string
	Statuses    []string
}

func AgentTokenStatus(server model.Server) string {
	if server.AgentTokenHash == "" {
		return "token missing"
	}
	return "token ready"
}

func AgentURL(serverID int64) string {
	return "$RAEVTAR_URL/api/v1/servers/" + IDText(serverID) + "/ping"
}

func AgentRunCommand(serverID int64, url, token string) string {
	if token == "" {
		token = "paste-token-here"
	}
	return "RAEVTAR_URL=" + url + " RAEVTAR_SERVER_ID=" + IDText(serverID) + " RAEVTAR_AGENT_TOKEN=" + token + " /usr/local/bin/raevtar-agent.sh"
}

func AgentInstallCommand(url string) string {
	url = strings.TrimRight(url, "/")
	return "sudo install -d /usr/local/bin && curl -fsSL " + url + "/static/agent/raevtar-agent.sh | sudo tee /usr/local/bin/raevtar-agent.sh >/dev/null && sudo chmod +x /usr/local/bin/raevtar-agent.sh"
}

func AgentCronLine(serverID int64, url, token string) string {
	if token == "" {
		token = "paste-token-here"
	}
	return "*/1 * * * * RAEVTAR_URL=" + url + " RAEVTAR_SERVER_ID=" + IDText(serverID) + " RAEVTAR_AGENT_TOKEN=" + token + " /usr/local/bin/raevtar-agent.sh >/dev/null 2>&1"
}

func ServerMetricCountText(metrics []model.ServerMetric) string {
	if len(metrics) == 1 {
		return "1 sample"
	}
	return strconv.Itoa(len(metrics)) + " samples"
}

func ServerMetricLatestText(metrics []model.ServerMetric) string {
	if len(metrics) == 0 || metrics[0].RecordedAt.IsZero() {
		return "No metrics yet"
	}
	return metrics[0].RecordedAt.Local().Format("02 Jan 2006 15:04:05")
}

func ServerMetricAvailabilityText(metrics []model.ServerMetric) string {
	if len(metrics) == 0 {
		return "No availability data"
	}
	online := 0
	for _, metric := range metrics {
		if metric.Online {
			online++
		}
	}
	return strconv.Itoa((online*100)/len(metrics)) + "% online"
}

func LastPayloadSummaryText(metrics []model.ServerMetric, now time.Time) string {
	if len(metrics) == 0 {
		return "No payload received yet"
	}
	metric := metrics[0]
	return "CPU " + MetricText(metric.CPUPercent) + "% · RAM " + RAMUsageText(metric.RAMUsedMB, metric.RAMTotalMB) + " · Disk " + MetricText(metric.DiskUsedGB) + " GB · Uptime " + UptimeText(metric.UptimeSeconds) + " · " + MetricRecordedAgeText(metric.RecordedAt, now)
}

func LastPayloadStateText(metrics []model.ServerMetric) string {
	if len(metrics) == 0 {
		return "No payload"
	}
	if metrics[0].Online {
		return "Online payload"
	}
	return "Offline payload"
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

func RAMUsageText(used, total float64) string {
	if total <= 0 {
		return MetricText(used) + " MB"
	}
	return MetricText(used) + " / " + MetricText(total) + " MB"
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

func MetricDotClass(online bool) string {
	if online {
		return "bg-retro-sage"
	}
	return "bg-retro-blush"
}

func MetricStateText(online bool) string {
	if online {
		return "Online"
	}
	return "Offline"
}

func MetricTimelineText(metrics []model.ServerMetric, index int) string {
	if index < 0 || index >= len(metrics) {
		return "Sample recorded"
	}
	current := metrics[index]
	if index == len(metrics)-1 {
		return MetricStateText(current.Online) + " sample recorded"
	}
	previous := metrics[index+1]
	if current.Online != previous.Online {
		return MetricStateText(previous.Online) + " to " + MetricStateText(current.Online)
	}
	return MetricStateText(current.Online) + " sample continued"
}

func CountText(value int) string {
	return strconv.Itoa(value)
}

func IDText(id int64) string {
	return strconv.FormatInt(id, 10)
}

func BytesText(value int64) string {
	if value < 1024 {
		return strconv.FormatInt(value, 10) + " B"
	}
	if value < 1024*1024 {
		return strconv.FormatFloat(float64(value)/1024, 'f', 1, 64) + " KB"
	}
	return strconv.FormatFloat(float64(value)/(1024*1024), 'f', 1, 64) + " MB"
}

func PortText(port int) string {
	return strconv.Itoa(port)
}

func PercentText(percent float64) string {
	return strconv.FormatFloat(percent, 'f', 0, 64) + "%"
}

func MetricText(value float64) string {
	return strconv.FormatFloat(value, 'f', 1, 64)
}

func IsOnline(lastSeen *time.Time) bool {
	return lastSeen != nil && time.Since(*lastSeen) < 3*time.Minute
}

func IsStale(lastSeen *time.Time) bool {
	return lastSeen != nil && time.Since(*lastSeen) >= 3*time.Minute && time.Since(*lastSeen) < 15*time.Minute
}

func StatusText(lastSeen *time.Time) string {
	switch {
	case IsOnline(lastSeen):
		return "Online"
	case IsStale(lastSeen):
		return "Stale"
	default:
		return "Offline"
	}
}

func LastSeenShort(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "-"
	}
	return lastSeen.Local().Format("15:04")
}

func LastSeenLong(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "Never"
	}
	return lastSeen.Local().Format("02 Jan 2006 15:04")
}

func StatusDotClass(lastSeen *time.Time) string {
	switch {
	case IsOnline(lastSeen):
		return "bg-retro-sage"
	case IsStale(lastSeen):
		return "bg-retro-wheat"
	default:
		return "bg-retro-blush"
	}
}

func StatusBadgeClass(lastSeen *time.Time) string {
	switch {
	case IsOnline(lastSeen):
		return "bg-retro-sage text-retro-cream"
	case IsStale(lastSeen):
		return "bg-retro-wheat text-retro-ink"
	default:
		return "bg-retro-blush text-retro-ink"
	}
}

func HeroStatusText(data DashboardData) string {
	switch {
	case data.ServerCount == 0:
		return "No servers registered"
	case data.OnlineCount == 0:
		return "No servers online"
	default:
		return "All systems operational"
	}
}

func HeroStatusClass(data DashboardData) string {
	switch {
	case data.ServerCount == 0:
		return "text-retro-muted"
	case data.OnlineCount == 0:
		return "text-retro-ink"
	default:
		return "text-retro-sage"
	}
}

func HeroDotClass(data DashboardData) string {
	switch {
	case data.ServerCount == 0:
		return "bg-retro-wheat"
	case data.OnlineCount == 0:
		return "bg-retro-blush"
	default:
		return "bg-retro-sage"
	}
}

func MetricToneClass(percent float64, high, mid float64) string {
	switch {
	case percent > high:
		return "bg-retro-blush"
	case percent > mid:
		return "bg-retro-wheat"
	default:
		return "bg-retro-sage"
	}
}

func TextToneClass(percent float64, high, mid float64) string {
	switch {
	case percent > high:
		return "text-retro-ink"
	case percent > mid:
		return "text-retro-muted"
	default:
		return "text-retro-sage"
	}
}

func MeterWidthStyle(percent float64) map[string]string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	return map[string]string{"width": strconv.FormatFloat(percent, 'f', 0, 64) + "%"}
}

func RoleBadgeClass(role string) string {
	switch role {
	case model.RoleOwner:
		return "bg-retro-blush text-retro-ink"
	case model.RoleAdmin:
		return "bg-retro-wheat text-retro-ink"
	case model.RoleOperator:
		return "bg-retro-sage text-retro-cream"
	case model.RoleReadonly:
		return "bg-retro-paper text-retro-muted"
	default:
		return "bg-retro-paper text-retro-ink"
	}
}

func ActionBadgeClass(action string) string {
	switch {
	case strings.Contains(action, "LOGIN"):
		return "bg-retro-sage text-retro-cream"
	case strings.Contains(action, "CREATE"):
		return "bg-retro-wheat text-retro-ink"
	case strings.Contains(action, "DELETE"):
		return "bg-retro-blush text-retro-ink"
	default:
		return "bg-retro-paper text-retro-ink"
	}
}

func CategoryBadgeClass(slug string) string {
	switch slug {
	case "ai-agent":
		return "bg-retro-blush text-retro-ink"
	case "security":
		return "bg-retro-ink text-retro-cream"
	case "kernel-embedded":
		return "bg-retro-wheat text-retro-ink"
	case "devops":
		return "bg-retro-sage text-retro-cream"
	case "tools":
		return "bg-retro-paper text-retro-ink"
	default:
		return "bg-retro-paper text-retro-ink"
	}
}

func TagsInput(tags []model.Tag) string {
	parts := make([]string, 0, len(tags))
	for _, tag := range tags {
		parts = append(parts, tag.Name)
	}
	return strings.Join(parts, ", ")
}

func TagList(tags string) []string {
	parts := strings.Split(tags, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func Initials(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "?"
	}
	if len(runes) > 2 {
		runes = runes[:2]
	}
	return strings.ToUpper(string(runes))
}

func EditorialStatusBadgeClass(status string) string {
	switch status {
	case model.EditorialStatusApproved:
		return "bg-retro-sage text-retro-cream"
	case model.EditorialStatusQueued:
		return "bg-retro-wheat text-retro-ink"
	case model.EditorialStatusPaused:
		return "bg-retro-paper text-retro-ink"
	case model.EditorialStatusRunning:
		return "bg-retro-ink text-retro-cream"
	case model.EditorialStatusFailed:
		return "bg-retro-blush text-retro-ink"
	case model.EditorialStatusDone:
		return "bg-retro-sage text-retro-cream"
	case model.EditorialStatusCancelled:
		return "bg-retro-paper text-retro-muted"
	default:
		return "bg-retro-paper text-retro-ink"
	}
}

func EditorialModeBadgeClass(mode string) string {
	switch mode {
	case model.EditorialModeScheduled:
		return "bg-retro-ink text-retro-cream"
	case model.EditorialModeOpportunistic:
		return "bg-retro-wheat text-retro-ink"
	case model.EditorialModeCampaign:
		return "bg-retro-sage text-retro-cream"
	case model.EditorialModeSeed:
		return "bg-retro-blush text-retro-ink"
	default:
		return "bg-retro-paper text-retro-ink"
	}
}

func EditorialTimeText(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Local().Format("02 Jan 2006 15:04")
}

func EditorialInputValue(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format("2006-01-02T15:04")
}

func EditorialInputValuePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return EditorialInputValue(*value)
}

func EditorialIDValuePtr(value *int64) string {
	if value == nil {
		return ""
	}
	return IDText(*value)
}

func EditorialEscalationText(deadline *time.Time, status string, now time.Time) string {
	if deadline == nil {
		return "On track"
	}
	if deadline.Before(now) && (status == model.EditorialStatusApproved || status == model.EditorialStatusRunning || status == model.EditorialStatusDone) {
		return "Overdue"
	}
	return "On track"
}

func EditorialEscalationBadgeClass(deadline *time.Time, status string, now time.Time) string {
	if EditorialEscalationText(deadline, status, now) == "Overdue" {
		return "bg-retro-blush text-retro-ink"
	}
	return "bg-retro-paper text-retro-muted"
}

func EditorialFairnessStateText(streak int, opened bool) string {
	if opened {
		return "Autonomous slot open next claim"
	}
	return "Non-urgent streak " + CountText(streak)
}

func EditorialDurationText(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	return AgeText(duration)
}

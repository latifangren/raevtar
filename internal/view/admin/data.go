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
}

type ServersData struct {
	CurrentPath            string
	CSRFToken              string
	Servers                []model.Server
	AgentURLExample        string
	GeneratedTokenServerID int64
	GeneratedAgentToken    string
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

func CountText(value int) string {
	return strconv.Itoa(value)
}

func IDText(id int64) string {
	return strconv.FormatInt(id, 10)
}

func PortText(port int) string {
	return strconv.Itoa(port)
}

func PercentText(percent float64) string {
	return strconv.FormatFloat(percent, 'f', 0, 64) + "%"
}

func IsOnline(lastSeen *time.Time) bool {
	return lastSeen != nil && time.Since(*lastSeen) < 2*time.Minute
}

func IsStale(lastSeen *time.Time) bool {
	return lastSeen != nil && time.Since(*lastSeen) >= 2*time.Minute && time.Since(*lastSeen) < 10*time.Minute
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

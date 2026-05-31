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
	if lastSeen == nil {
		return "Unknown"
	}
	return "Online"
}

func DashboardStatusText(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "Unknown"
	}
	return "Online — " + lastSeen.Format("Jan 2 15:04")
}

func StatusDotClass(lastSeen *time.Time) string {
	if lastSeen == nil {
		return "bg-retro-wheat"
	}
	return "bg-retro-sage"
}

func MetricDotClass(online bool) string {
	if online {
		return "bg-retro-sage"
	}
	return "bg-retro-blush"
}

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
	CurrentPath string
	Servers     []model.Server
	Categories  []model.Category
}

type ServerDetailData struct {
	CurrentPath string
	Server      *model.Server
	Metrics     []model.ServerMetric
	Categories  []model.Category
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
		return "text-xs px-2 py-0.5 font-bold border-2 border-black bg-violet-300 text-black"
	case "commissioned":
		return "text-xs px-2 py-0.5 font-bold border-2 border-black bg-amber-300 text-black"
	default:
		return "text-xs px-2 py-0.5 font-bold border-2 border-black bg-neutral-200 text-black"
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
		return "bg-neutral-300"
	}
	return "bg-emerald-400"
}

func MetricDotClass(online bool) string {
	if online {
		return "bg-emerald-400"
	}
	return "bg-rose-400"
}

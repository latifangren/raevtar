package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"raevtar/internal/model"
	"raevtar/internal/service"
)

func (h *Handler) apiRecordPostReadTime(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid post id"})
		return
	}

	var req struct {
		Seconds int `json:"seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	// security/correctness limits: pings must report standard heartbeats
	if req.Seconds <= 0 || req.Seconds > 60 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid seconds value"})
		return
	}

	hash := sha256.Sum256([]byte(clientIP(r)))
	ipHash := hex.EncodeToString(hash[:8])

	if err := h.svc.Blog.RecordPostReadTime(postID, ipHash, req.Seconds); err != nil {
		internalServerJSON(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) apiListPosts(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	posts, _, err := h.svc.Blog.ListPosts(cat, 1, 100)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, posts)
}

func (h *Handler) apiListProjects(w http.ResponseWriter, r *http.Request) {
	featuredOnly := strings.EqualFold(r.URL.Query().Get("featured"), "true")
	state := r.URL.Query().Get("state")
	sort := r.URL.Query().Get("sort")
	projects, _, err := h.svc.Projects.ListProjects(1, 100, service.ProjectListOptions{
		FeaturedOnly: featuredOnly,
		State:        state,
		Sort:         sort,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidProjectSort) || errors.Is(err, service.ErrInvalidProjectState) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) apiSearch(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil && strings.TrimSpace(r.URL.Query().Get("page")) != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid page"})
		return
	}
	if page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil && strings.TrimSpace(r.URL.Query().Get("page_size")) != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid page_size"})
		return
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	results, err := h.svc.Search.SearchPublic(service.SearchOptions{
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Scope:    strings.TrimSpace(r.URL.Query().Get("scope")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid search scope") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"query":       results.Query,
		"scope":       results.Scope,
		"page":        results.Page,
		"page_size":   results.PageSize,
		"total":       results.Total,
		"total_pages": results.TotalPages,
		"paginated":   results.Paginated,
		"counts": map[string]int{
			"posts":    results.PostCount,
			"projects": results.ProjectCount,
			"pages":    results.PageCount,
		},
		"posts":    results.Posts,
		"projects": results.Projects,
		"pages":    results.Pages,
	})
}

func (h *Handler) apiListProjectUpdates(w http.ResponseWriter, r *http.Request) {
	project, err := h.svc.Projects.GetPublishedProject(r.PathValue("slug"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	items, err := h.svc.Projects.ListProjectTimeline(project.ID, true, 100)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) apiListProjectChangelog(w http.ResponseWriter, r *http.Request) {
	project, err := h.svc.Projects.GetPublishedProject(r.PathValue("slug"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	items, err := h.svc.Projects.ListProjectChangelog(project.ID, true, 100)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) apiListProjectRelations(w http.ResponseWriter, r *http.Request) {
	project, err := h.svc.Projects.GetPublishedProject(r.PathValue("slug"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	items, err := h.svc.Projects.GetResolvedProjectRelations(project.ID, true)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) apiListProjectShowcase(w http.ResponseWriter, r *http.Request) {
	project, err := h.svc.Projects.GetPublishedProject(r.PathValue("slug"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	items, err := h.svc.Projects.ListProjectShowcase(project.ID, true)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) apiCreateProjectUpdate(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var input model.ProjectUpdateEntryCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Projects.CreateProjectUpdate(projectID, input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) apiUpdateProjectUpdate(w http.ResponseWriter, r *http.Request) {
	updateID, err := strconv.ParseInt(r.PathValue("updateID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var input model.ProjectUpdateEntryUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Projects.UpdateProjectUpdate(updateID, input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrProjectUpdateNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) apiDeleteProjectUpdate(w http.ResponseWriter, r *http.Request) {
	updateID, err := strconv.ParseInt(r.PathValue("updateID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Projects.DeleteProjectUpdate(updateID); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrProjectUpdateNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) apiCreateProjectRelation(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var input model.ContentRelationCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Projects.CreateProjectRelation(projectID, input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) apiDeleteProjectRelation(w http.ResponseWriter, r *http.Request) {
	relationID, err := strconv.ParseInt(r.PathValue("relationID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Projects.DeleteProjectRelation(relationID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) apiCreateProjectShowcase(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var input model.ProjectShowcaseItemCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Projects.CreateProjectShowcase(projectID, input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) apiUpdateProjectShowcase(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var input model.ProjectShowcaseItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Projects.UpdateProjectShowcase(itemID, input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrProjectShowcaseNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) apiDeleteProjectShowcase(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.svc.Projects.DeleteProjectShowcase(itemID); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrProjectShowcaseNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) apiCreatePost(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	var input model.PostCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	post, err := h.svc.Blog.CreatePost(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPostInput) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

func (h *Handler) apiCreateProject(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	var input model.ProjectCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	project, err := h.svc.Projects.CreateProject(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProjectInput) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *Handler) apiUpdateProject(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	id, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var input model.ProjectUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	project, err := h.svc.Projects.UpdateProject(id, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProjectInput) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrProjectNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *Handler) apiDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := h.svc.Projects.DeleteProject(id); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) apiListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *Handler) apiCreateServer(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	var input struct {
		Name string `json:"name"`
		Host string `json:"host"`
		Port int    `json:"port"`
		Tags string `json:"tags"`
	}
	// Accept both JSON and form-encoded (HTMX)
	ct := r.Header.Get("Content-Type")
	if len(ct) >= 16 && ct[:16] == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			if isBodyTooLarge(err) {
				writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			if isBodyTooLarge(err) {
				writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid form"})
			return
		}
		input.Name = r.FormValue("name")
		input.Host = r.FormValue("host")
		if p, err := strconv.Atoi(r.FormValue("port")); err == nil {
			input.Port = p
		}
		input.Tags = r.FormValue("tags")
	}
	if input.Name == "" || input.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and host required"})
		return
	}
	if input.Port == 0 {
		input.Port = 22
	}

	s, token, err := h.svc.Monitor.CreateServerWithAgentToken(input.Name, input.Host, input.Port, input.Tags)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"server":      s,
		"agent_token": token,
	})
}

func (h *Handler) apiListServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func (h *Handler) apiGetServer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	server, err := h.svc.Monitor.GetServer(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerJSON(w, r, err)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, server)
}

func (h *Handler) apiRecordMetrics(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if !h.canRecordMetrics(id, r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid agent token"})
		return
	}

	var m model.ServerMetric
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if err := h.svc.Monitor.RecordMetrics(id, m); err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "server not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidMetricPayload) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) apiGetPendingCommands(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if !h.canRecordMetrics(id, r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid agent token"})
		return
	}

	cmds, err := h.svc.CommandQ.PendingCommands(id)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	if cmds == nil {
		cmds = []model.ServerCommand{}
	}
	writeJSON(w, http.StatusOK, cmds)
}

type commandResultRequest struct {
	CommandID int64  `json:"command_id"`
	Result    string `json:"result"`
	Failed    bool   `json:"failed"`
}

func (h *Handler) apiReportCommandResult(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if !h.canRecordMetrics(id, r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid agent token"})
		return
	}

	var req commandResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.CommandID == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "command_id required"})
		return
	}

	if req.Failed {
		err = h.svc.CommandQ.FailCommand(req.CommandID, req.Result)
	} else {
		err = h.svc.CommandQ.CompleteCommand(req.CommandID, req.Result)
	}
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) apiBootstrap(w http.ResponseWriter, r *http.Request) {
	serverID, err := strconv.ParseInt(r.PathValue("serverID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid server id", http.StatusBadRequest)
		return
	}
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	if !h.svc.Monitor.VerifyAgentToken(serverID, token) {
		http.Error(w, "invalid server or token", http.StatusForbidden)
		return
	}

	url := agentURLExample(h.cfg.Domain)

	script := `#!/bin/sh
set -eu

RAEVTAR_URL='` + url + `'
RAEVTAR_SERVER_ID='` + strconv.FormatInt(serverID, 10) + `'
RAEVTAR_AGENT_TOKEN='` + token + `'

echo ""
echo "=== Raevtar Agent Setup ==="
echo "Server ID : ${RAEVTAR_SERVER_ID}"
echo "Server URL: ${RAEVTAR_URL}"
echo ""

# Step 1: Detect OS and package manager
echo "[1/5] Detecting operating system..."
OS=""
PKG_INSTALL=""
if command -v apk >/dev/null 2>&1; then
	OS="alpine (apk)"; PKG_INSTALL="apk add --no-cache"
elif command -v apt-get >/dev/null 2>&1; then
	OS="debian/ubuntu (apt)"; PKG_INSTALL="apt-get install -y"
elif command -v opkg >/dev/null 2>&1; then
	OS="openwrt (opkg)"; PKG_INSTALL="opkg install"
else
	echo "[!] Unknown package manager. Dependencies may be missing."
	echo "    Install curl manually and re-run this script."
fi
echo "  => Detected: ${OS:-unknown}"
echo ""

# Step 2: Install dependencies
echo "[2/5] Checking dependencies..."
if command -v curl >/dev/null 2>&1; then
	echo "  => curl: OK"
else
	echo "  => curl: installing..."
	if [ -n "${PKG_INSTALL}" ]; then
		${PKG_INSTALL} curl
		echo "  => curl: installed"
	else
		echo "  => curl: FAILED - install curl manually"
	fi
fi

if command -v jq >/dev/null 2>&1; then
	echo "  => jq: OK"
else
	echo "  => jq: installing..."
	if [ -n "${PKG_INSTALL}" ]; then
		${PKG_INSTALL} jq 2>/dev/null && echo "  => jq: installed" || echo "  => jq: skipped (not critical)"
	else
		echo "  => jq: skipped (install manually if needed)"
	fi
fi
echo ""

# Step 3: Install agent script
echo "[3/5] Installing agent script..."
AGENT_DIR="/usr/local/bin"
AGENT_PATH="${AGENT_DIR}/raevtar-agent.sh"
mkdir -p "${AGENT_DIR}"
if curl -fsSL -o "${AGENT_PATH}" "${RAEVTAR_URL}/static/agent/raevtar-agent.sh"; then
	chmod +x "${AGENT_PATH}"
	echo "  => Agent installed: ${AGENT_PATH}"
else
	echo "  => FAILED: could not download agent script"
	exit 1
fi
echo ""

# Step 4: Save config
echo "[4/5] Saving agent configuration..."
CONF_DIR="${HOME:-/tmp}/.raevtar"
mkdir -p "${CONF_DIR}"
cat > "${CONF_DIR}/agent.env" << EOF
RAEVTAR_URL=${RAEVTAR_URL}
RAEVTAR_SERVER_ID=${RAEVTAR_SERVER_ID}
RAEVTAR_AGENT_TOKEN=${RAEVTAR_AGENT_TOKEN}
EOF
echo "  => Config saved: ${CONF_DIR}/agent.env"
echo ""

# Step 5: Setup cron
echo "[5/5] Setting up periodic ping via cron..."
CRON_LINE="*/5 * * * * RAEVTAR_URL=${RAEVTAR_URL} RAEVTAR_SERVER_ID=${RAEVTAR_SERVER_ID} RAEVTAR_AGENT_TOKEN=${RAEVTAR_AGENT_TOKEN} ${AGENT_PATH} >/dev/null 2>&1"
EXISTING_CRON=$(crontab -l 2>/dev/null || true)
if echo "${EXISTING_CRON}" | grep -qF "${AGENT_PATH}"; then
	echo "  => Cron entry already exists, skipping"
else
	(echo "${EXISTING_CRON}"; echo "${CRON_LINE}") | crontab - 2>/dev/null && echo "  => Cron: added (every 5 minutes)" || echo "  => Cron: skipped (no crontab access)"
fi
echo ""

# Test ping
echo "=== Testing connection ==="
if ${AGENT_PATH}; then
	echo "  => [OK] Ping successful. Server is reporting."
else
	echo "  => [FAIL] Ping failed. Check RAEVTAR_URL and network."
fi

echo ""
echo "=== Setup Complete ==="
echo "Next ping in 5 minutes (via cron)."
echo "Manual run: ${AGENT_PATH}"
echo "Config:     ${CONF_DIR}/agent.env"
`

	w.Header().Set("Content-Type", "text/x-shellscript")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(script))
}

func (h *Handler) canRecordMetrics(serverID int64, r *http.Request) bool {
	token, ok := bearerToken(r)
	if !ok {
		return false
	}
	if h.cfg.AdminKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(h.cfg.AdminKey)) == 1 {
		return true
	}
	return h.svc.Monitor.VerifyAgentToken(serverID, token)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// adminAuth is middleware that validates Authorization: Bearer <key>
// using constant-time comparison to prevent timing attacks.
func (h *Handler) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.cfg.AdminKey == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "admin key not configured"})
			return
		}

		provided, ok := bearerToken(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid Authorization header"})
			return
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(h.cfg.AdminKey)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid admin key"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	return token, token != ""
}

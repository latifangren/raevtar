# DEPLOYMENT — Raevtar

Cara deploy Raevtar di Linux/macOS: systemd (Linux), SQLite lokal, dan Cloudflare Tunnel opsional. Ini runbook operator; agent/assistant tidak boleh menjalankan restart/deploy kecuali user minta eksplisit.

## Prasyarat

- **Go** (1.26+) — untuk build
- **cloudflared** — hanya jika pakai Cloudflare Tunnel
- **Node/npm** — hanya untuk regenerasi Tailwind CSS (HTMX sudah vendored)

Install Go lewat package manager pilihan Anda:

```bash
# Alpine
apk add go

# Debian/Ubuntu
apt-get install golang

# macOS (Homebrew)
brew install go
```

## Build

```bash
cd /home/latif/raevtar
make build
```

`make build` menjalankan `go run github.com/a-h/templ/cmd/templ@v0.3.906 generate`, Tailwind CLI, lalu `go build`.

## System Service (systemd)

Biar jalan otomatis pas boot + restart kalo crash.
Set `RAEVTAR_ENV=production`; kalau `RAEVTAR_ADMIN_KEY` atau `RAEVTAR_ADMIN_PASS` kosong, Raevtar akan gagal start supaya endpoint admin/API tidak hidup tanpa secret.

### 1. Generate service file (opsional)

```bash
make generate-service
# This creates raevtar.service with your user/home paths
```

Atau buat manual:

```bash
sudo tee /etc/systemd/system/raevtar.service << 'EOF'
[Unit]
Description=Raevtar personal platform
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
User=$USER
WorkingDirectory=$HOME/raevtar
ExecStart=$HOME/raevtar/raevtar
Environment=RAEVTAR_ENV=production
Environment=RAEVTAR_ADMIN_KEY=<isi-admin-key>
Environment=RAEVTAR_ADMIN_USER=admin
Environment=RAEVTAR_ADMIN_PASS=<isi-password-admin>
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
```

### 2. Aktifin & start

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now raevtar
```

### 3. Cek status

```bash
sudo systemctl status raevtar
journalctl -u raevtar -f        # live log
journalctl -u raevtar --no-pager -n 50   # 50 line terakhir
```

## SQLite Backup (systemd timer)

Backup otomatis tiap hari jam 03:00.

### 1. Service backup

```bash
sudo tee /etc/systemd/system/raevtar-backup.service << 'EOF'
[Unit]
Description=Raevtar SQLite backup

[Service]
Type=exec
User=latif
ExecStart=/home/latif/raevtar/cron/backup.sh
EOF
```

### 2. Timer (gantinya cron)

```bash
sudo tee /etc/systemd/system/raevtar-backup.timer << 'EOF'
[Unit]
Description=Daily Raevtar backup at 03:00

[Timer]
OnCalendar=*-*-* 03:00:00
Persistent=true

[Install]
WantedBy=timers.target
EOF
```

### 3. Aktifin

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now raevtar-backup.timer
systemctl list-timers raevtar-backup.timer
```

## Cloudflare Tunnel

```bash
# Login ke Cloudflare
cloudflared tunnel login

# Bikin tunnel
cloudflared tunnel create raevtar

# Route DNS
cloudflared tunnel route dns raevtar raevtar.tech

# Config
mkdir -p ~/.cloudflared
cat > ~/.cloudflared/config.yml << EOF
tunnel: <tunnel-id>
credentials-file: /home/latif/.cloudflared/<tunnel-id>.json

ingress:
  - hostname: raevtar.tech
    service: http://localhost:8080
  - service: http_status:404
EOF

# Tunnel service (systemd atau langsung)
cloudflared tunnel run raevtar  # jalan di terminal dulu buat test
```

Opsional trusted proxy untuk audit IP/rate limit kalau request dari tunnel masuk dari localhost:

```bash
RAEVTAR_TRUSTED_PROXY_CIDRS=127.0.0.1/32,::1/128
```

Atau pake `cloudflared service install` buat systemd:

```bash
sudo cloudflared service install
# Config wajib di /etc/cloudflared/config.yml
```

Jangan aktifkan trusted proxy untuk IP publik sembarang. Default Raevtar sengaja mengabaikan `X-Forwarded-For`; hanya `CF-Connecting-IP`/forwarded header dari CIDR tepercaya yang dipakai untuk audit log dan rate limit.

### Hardening env vars (opsional)

| Variable | Default | Keterangan |
|----------|---------|------------|
| `RAEVTAR_RATE_LIMIT_REQUESTS` | `60` | Max requests per window per IP |
| `RAEVTAR_RATE_LIMIT_WINDOW` | `60s` | Rate limit window |
| `RAEVTAR_READ_TIMEOUT` | `10s` | HTTP read timeout |
| `RAEVTAR_WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `RAEVTAR_IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `RAEVTAR_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown timeout |
| `RAEVTAR_MAX_UPLOAD_MB` | `6` | Max upload size in MB |
| `RAEVTAR_LOGIN_FAILURE_LIMIT` | `5` | Max login failures per user/IP |
| `RAEVTAR_LOGIN_IP_FAILURE_LIMIT` | `20` | Max login failures per IP |
| `RAEVTAR_DISK_ROOT` | `/` | Filesystem root for disk stats |

## Public Docs

`/docs` dan `/lab/docs` adalah halaman Templ public-safe. File Swagger shell lama tidak dikirim lagi; `static/openapi.json` tetap read-only dan hanya mendokumentasikan endpoint publik.

## Service Status

```bash
sudo systemctl status raevtar
sudo systemctl status raevtar-backup.timer
journalctl -u raevtar -f
```

## Update

Jalankan hanya saat memang mau deploy versi baru.

```bash
cd /home/latif/raevtar
git pull
make build
sudo systemctl restart raevtar
```

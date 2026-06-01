# DEPLOYMENT — Raevtar

Cara deploy Raevtar di setup postmarketOS/Linux yang dipakai di mesin ini: systemd, SQLite lokal, dan Cloudflare Tunnel.

## Prasyarat

```bash
apk add go cloudflared
```

Build juga butuh Node/npm untuk Tailwind CLI kalau `static/css/style.css` mau diregenerate.

## Build

```bash
cd /home/latif/raevtar
make build
```

`make build` menjalankan `go run github.com/a-h/templ/cmd/templ@v0.3.906 generate`, Tailwind CLI, lalu `go build`.

## System Service (systemd)

Biar jalan otomatis pas hp boot + restart kalo crash.
Set `RAEVTAR_ENV=production`; kalau `RAEVTAR_ADMIN_KEY` atau `RAEVTAR_ADMIN_PASS` kosong, Raevtar akan gagal start supaya endpoint admin/API tidak hidup tanpa secret.

### 1. Buat service file

```bash
sudo tee /etc/systemd/system/raevtar.service << 'EOF'
[Unit]
Description=Raevtar personal platform
After=network-online.target
Wants=network-online.target

[Service]
Type=exec
User=latif
WorkingDirectory=/home/latif/raevtar
ExecStart=/home/latif/raevtar/raevtar
Environment=RAEVTAR_ENV=production
Environment=RAEVTAR_ADMIN_KEY=<isi-admin-key>
Environment=RAEVTAR_ADMIN_USER=admin
Environment=RAEVTAR_ADMIN_PASS=<isi-password-admin>
Restart=on-failure
RestartSec=5
# Journald capture stdout/stderr otomatis — gak perlu redirect manual

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

## Public Docs

`/docs` dan `/lab/docs` adalah halaman Templ public-safe. File Swagger shell lama tidak dikirim lagi; `static/openapi.json` tetap read-only dan hanya mendokumentasikan endpoint publik.

## Service Status

```bash
sudo systemctl status raevtar
sudo systemctl status raevtar-backup.timer
journalctl -u raevtar -f
```

## Update

```bash
cd /home/latif/raevtar
git pull
make build
sudo systemctl restart raevtar
```

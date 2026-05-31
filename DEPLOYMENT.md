# DEPLOYMENT — Raevtar

Cara deploy Raevtar di postmarketOS (Alpine dengan systemd) atau Linux lain.

## Prasyarat

```bash
apk add go cloudflared
```

## Build

```bash
cd /home/latif/raevtar
go build -o raevtar ./cmd/server/
```

## System Service (systemd)

Biar jalan otomatis pas hp boot + restart kalo crash.

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
Environment=RAEVTAR_ADMIN_KEY=<isi-admin-key>
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

Atau pake `cloudflared service install` buat systemd:

```bash
sudo cloudflared service install
# Config wajib di /etc/cloudflared/config.yml
```

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
go build -o raevtar ./cmd/server/
sudo systemctl restart raevtar
```

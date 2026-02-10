# FleetLens Agent

A lightweight, cross-platform agent that collects system information from virtual machines and reports it to a FleetLens server.

## Collected Data

| Category | Details |
|----------|---------|
| **System** | Hostname, operating system, architecture, serial number |
| **Network** | IP addresses, network interfaces (name, IPs, MAC address, type) |
| **Software** | Installed applications (name, version, publisher, install date) |
| **Users** | Local user accounts (username, display name, admin status, last login) |

### JSON Payload Example

```json
{
  "hostname": "vm-web-01",
  "ip_addresses": ["192.168.1.10"],
  "serial_number": "VMware-42 1a 7b ...",
  "operating_system": "linux",
  "architecture": "amd64",
  "network_interfaces": [
    {
      "name": "eth0",
      "ip_addresses": ["192.168.1.10", "fe80::1"],
      "mac_address": "00:0c:29:1a:7b:3e",
      "type": "ethernet"
    }
  ],
  "software": [
    {
      "name": "nginx",
      "version": "1.24.0-2",
      "publisher": "Debian Nginx Maintainers",
      "installed_at": ""
    }
  ],
  "users": [
    {
      "username": "admin",
      "display_name": "Admin User",
      "is_admin": true,
      "last_login": "2026-02-10 09:15:30"
    }
  ]
}
```

## Installation

### From GitHub Releases

Download the latest binary for your platform from the [Releases](https://github.com/Pitchouneee/fleetlens-agent/releases) page.

Available builds: Linux (amd64/arm64), Windows (amd64/arm64), macOS (amd64/arm64).

### From Source

```bash
git clone https://github.com/Pitchouneee/fleetlens-agent.git
cd fleetlens-agent
go build -o fleetlens-agent .
```

## Usage

```bash
fleetlens-agent -api <FLEETLENS_SERVER_URL> [-interval <DURATION>]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-api` | *(required)* | FleetLens API base URL (e.g. `http://fleetlens.example.com:8080`) |
| `-interval` | `24h` | Interval between data collection cycles (e.g. `30m`, `1h`, `24h`) |

The API URL can also be set via the `FLEETLENS_API_URL` environment variable.

### Examples

```bash
# Default: collect and send every 24 hours
fleetlens-agent -api http://fleetlens.example.com:8080

# Collect every hour
fleetlens-agent -api http://fleetlens.example.com:8080 -interval 1h

# Using environment variable
export FLEETLENS_API_URL=http://fleetlens.example.com:8080
fleetlens-agent
```

## Setup as a System Service

### Linux (systemd)

Create `/etc/systemd/system/fleetlens-agent.service`:

```ini
[Unit]
Description=FleetLens Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/fleetlens-agent -api http://fleetlens.example.com:8080
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Then enable and start:

```bash
sudo cp fleetlens-agent /usr/local/bin/
sudo systemctl daemon-reload
sudo systemctl enable --now fleetlens-agent
```

### macOS (launchd)

Create `~/Library/LaunchAgents/com.fleetlens.agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.fleetlens.agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/fleetlens-agent</string>
        <string>-api</string>
        <string>http://fleetlens.example.com:8080</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Then load:

```bash
cp fleetlens-agent /usr/local/bin/
launchctl load ~/Library/LaunchAgents/com.fleetlens.agent.plist
```

### Windows (NSSM)

Using [NSSM](https://nssm.cc/) (Non-Sucking Service Manager):

```powershell
nssm install FleetLensAgent C:\Program Files\FleetLens\fleetlens-agent.exe -api http://fleetlens.example.com:8080
nssm start FleetLensAgent
```

Or with PowerShell natively:

```powershell
New-Service -Name "FleetLensAgent" `
  -BinaryPathName '"C:\Program Files\FleetLens\fleetlens-agent.exe" -api http://fleetlens.example.com:8080' `
  -DisplayName "FleetLens Agent" `
  -StartupType Automatic `
  -Description "FleetLens system information collection agent"

Start-Service FleetLensAgent
```

## Notes

- The agent sends data to `POST {api}/api/agents`
- Serial number collection may require elevated privileges (run as root/admin)
- The agent runs the first collection immediately on startup, then repeats at the configured interval
- On API failure, the agent logs an error and retries on the next cycle
- The agent shuts down gracefully on SIGINT/SIGTERM
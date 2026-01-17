# Virus Scanning Configuration Guide

## ClamAV Installation

### Install ClamAV Daemon
```bash
sudo apt-get update
sudo apt-get install -y clamav clamav-daemon
```

### Start ClamAV Service
```bash
sudo systemctl start clamav-daemon
sudo systemctl enable clamav-daemon
```

### Update Virus Definitions
```bash
sudo freshclam
```

### Verify ClamAV is Running
```bash
sudo systemctl status clamav-daemon

# Should show: Active: active (running)
```

### Check ClamAV Socket
```bash
# ClamAV listens on Unix socket or TCP port
ls /var/run/clamav/clamd.ctl  # Unix socket
# OR
netstat -tuln | grep 3310  # TCP port

## Environment Configuration

Add to `.env`:
```
CLAMAV_ADDR=tcp://localhost:3310
# OR for Unix socket:
# CLAMAV_ADDR=unix:///var/run/clamav/clamd.ctl
```

## Testing

### Test with EICAR Virus Test File
```bash
# Download EICAR test file (harmless virus signature)
curl -o eicar.txt https://secure.eicar.org/eicar.com.txt

# Try uploading - should be rejected
curl -X POST http://localhost:8081/api/v1/documents \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@eicar.txt" \
  -F "title=Virus Test" \
  -F "collection_id=YOUR_COLLECTION_ID"

# Expected response:
# {"error": "Virus scan failed: file eicar.txt is infected with: Eicar-Test-Signature"}
```

### Test with Clean File
```bash
# Should succeed
curl -X POST http://localhost:8081/api/v1/documents \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@clean-document.pdf" \
  -F "title=Clean Document" \
  -F "collection_id=YOUR_COLLECTION_ID"

# Expected: 200 OK with document response
```

## Troubleshooting

### ClamAV Not Running
```bash
# Check logs
sudo journalctl -u clamav-daemon -n 50

# Restart service
sudo systemctl restart clamav-daemon
```

### Virus Database Outdated
```bash
# Update manually
sudo freshclam

# Enable automatic updates
sudo systemctl enable clamav-freshclam
sudo systemctl start clamav-freshclam
```

### Connection Issues
```bash
# Check ClamAV is listening
sudo netstat -tlnp | grep clam

# Test connection
clamdscan /bin/ls
# Should return: /bin/ls: OK
```

## Production Notes

- ClamAV virus definitions should be updated daily
- Scan timeout: Default 30 seconds (configurable)
- Max file size: Default 25MB (configurable in clamd.conf)
- For large files, consider async scanning with job queue

#!/bin/bash
# PLC/OES Monitoring Platform - Maintenance Script
# Handles automated backups and log pruning

BACKUP_DIR="/data/backups"
LOG_DIR="/var/log/plc-monitor"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

echo "--- Starting Maintenance ($DATE) ---"

# 1. Backup PostgreSQL
echo "Backing up PostgreSQL..."
docker exec postgres pg_dump -U plc_user plc_monitor > $BACKUP_DIR/postgres_$DATE.sql
gzip $BACKUP_DIR/postgres_$DATE.sql

# 2. Backup InfluxDB Metadata
echo "Backing up InfluxDB metadata..."
docker exec influxdb influx backup $BACKUP_DIR/influx_$DATE

# 3. Prune Old Backups (keep 30 days)
echo "Pruning backups older than 30 days..."
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
find $BACKUP_DIR -name "influx_*" -type d -mtime +30 -exec rm -rf {} +

# 4. Rotate Logs
echo "Rotating application logs..."
# Assuming logrotate is configured for the log directory
logrotate -f /etc/logrotate.d/plc-monitor

echo "--- Maintenance Complete ---"

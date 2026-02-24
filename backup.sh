#!/bin/bash
# backup.sh - Backup system data

BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

echo "💾 Starting backup to $BACKUP_DIR..."

# PostgreSQL Backup
if [ "$(docker ps -q -f name=plc-postgres)" ]; then
    echo "📦 Backing up PostgreSQL..."
    docker-compose exec -T postgres pg_dump -U plc_user plc_monitor | gzip > "$BACKUP_DIR/postgres.sql.gz"
fi

# InfluxDB Backup
if [ "$(docker ps -q -f name=plc-influxdb)" ]; then
    echo "📦 Backing up InfluxDB..."
    docker-compose exec -T influxdb influx backup /tmp/backup
    docker cp plc-influxdb:/tmp/backup "$BACKUP_DIR/influxdb"
fi

# Copy exports and configs
echo "📦 Backing up exports and configs..."
cp -r exports "$BACKUP_DIR/exports" 2>/dev/null || true
cp -r config "$BACKUP_DIR/config" 2>/dev/null || true
cp .env "$BACKUP_DIR/.env" 2>/dev/null || true

echo "✅ Backup completed at $BACKUP_DIR"

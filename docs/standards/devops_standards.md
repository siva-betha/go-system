# DevOps & Deployment Standards

Standards for maintaining high availability and observability.

## 1. Containerization
- **Base Images**: Use Alpine or Distroless for security and size.
- **Layers**: Multi-stage builds enforced for Go and Next.js.
- **Naming**: `monitoring-<service>:<version>`

## 2. Observability (SLA)
- **Metrics**: Prometheus `/metrics` endpoint on all services.
- **Uptime**: 99.9% target for Edge Collector; 99% for Analytics.
- **Latency**: <100ms for P99 UI updates.

## 3. Monitoring Dashboards
| Dashboard | Primary KPIs |
|---|---|
| **System Health** | CPU, RAM, Disk, Error Rates. |
| **PLC Performance** | ADS Latency, Packet Loss, Polling Status. |
| **Data Engine** | Kafka Lag, InfluxDB Insert Rate, Zstd Ratio. |

## 4. Disaster Recovery
- **Backups**: Daily snapshots of PostgreSQL and InfluxDB metadata.
- **Retention**: 30 days of daily backups; 12 months of monthly summaries.
- **RTO/RPO**: 4-hour Recovery Time Objective (RTO).

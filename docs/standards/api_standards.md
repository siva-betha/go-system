# API Standards

This document defines the interface standards for the PLC/OES Monitoring Platform.

## 1. REST API Standards
- **Version**: Always use `/api/v1/` prefix.
- **Format**: All responses MUST be JSON.
- **Errors**: Use standard HTTP status codes (401, 403, 404, 500) with a JSON body:
  ```json
  { "error": "Reason for failure", "code": "ERR_CODE" }
  ```

### Key Endpoints
- `GET /api/v1/data/scalars`: Retrieve time-series data.
- `GET /api/v1/data/spectra`: Retrieve spectral OES data.
- `POST /api/v1/analytics/endpoint/detect`: Trigger endpoint analysis.

## 2. WebSocket Standards
- **Connection**: `ws://<host>/api/v1/stream`
- **Protocol**: JSON payloads for updates.
- **Message Types**:
  - `scalar_update`: Real-time sensor data.
  - `spectrum_update`: Compressed spectral preview.
  - `alert`: System or process anomalies.

## 3. Data Models (Simplified)
### PLC Scalar
```json
{
  "timestamp": "ISO8601",
  "values": { "pressure": 45.2, "power": 1250 }
}
```
### OES Spectrum
```json
{
  "timestamp": "ISO8601",
  "intensities": [123, 456, ...],
  "sequence": 1001
}
```

# Technology Stack & Libraries

## Backend (Go & Fiber)
- **Runtime**: Go 1.25+
- **Major Frameworks**:
  - Web Server: [Fiber v3](github.com/gofiber/fiber/v3)
  - Industrial Connectivity: [Gouads](github.com/robinson/gouads) (TwinCAT ADS implementation)
  - Messaging: [Confluent Kafka Go v2](github.com/confluentinc/confluent-kafka-go)
- **Database Drivers**:
  - PostgreSQL: [pgx/v5](github.com/jackc/pgx/v5)
  - Time-Series: [InfluxDB Client Go v2](github.com/influxdata/influxdb-client-go/v2)
- **Security & Identity**:
  - JWT: [golang-jwt/jwt/v5](github.com/golang-jwt/jwt/v5)
  - ID Generation: [google/uuid](github.com/google/uuid)
  - Password Hashing: [golang.org/x/crypto](golang.org/x/crypto) (Argon2id)
- **Utilities**:
  - Validation: [go-playground/validator/v10](github.com/go-playground/validator/v10)
  - Environment: [godotenv](github.com/joho/godotenv)
  - High-Efficiency Compression: [klauspost/compress](github.com/klauspost/compress) (Zstd, Gzip)
  - Migrations: [golang-migrate/migrate/v4](github.com/golang-migrate/migrate/v4)

## Analytics & Signal Processing (Python)
- **Runtime**: Python 3.10+
- **Scientific Stack**:
  - [NumPy](https://numpy.org/): High-performance numerical arrays.
  - [Pandas](https://pandas.pydata.org/): Industrial time-series analysis.
  - [SciPy](https://scipy.org/): Signal processing & spectral analysis.
  - [Statsmodels](https://www.statsmodels.org/): Exploratory data analysis.
- **Advanced Modeling**:
  - [Scikit-learn](https://scikit-learn.org/): Anomaly detection & classification.
  - [Ruptures](https://github.com/deepcharles/ruptures): Statistical changepoint detection.
  - [PyMannKendall](https://github.com/yumitdmr/pyMannKendall): Non-parametric trend detection.
- **Integration**:
  - Messaging: [Kafka-python](github.com/dpkp/kafka-python)
  - Database: [SQLAlchemy](www.sqlalchemy.org) & [Psycopg2](psycopg.org)
  - Metrics: [Prometheus Client](github.com/prometheus/client_python)

## Frontend (Next.js)
- **Framework**: [Next.js 16 (App Router)](https://nextjs.org/)
- **Core**: React 19, TypeScript 5
- **UI & Visualization**:
  - [Tailwind CSS 4](https://tailwindcss.com/): Utility-first styling.
  - [Lucide React](https://lucide.dev/): Engineering iconography.
  - [Recharts](https://recharts.org/): Industrial data plotting.
  - Styling Utilities: [clsx](github.com/lukeed/clsx) & [tailwind-merge](github.com/dcastil/tailwind-merge)

## Infrastructure & Industrial Layer
- **Relational DB**: PostgreSQL 15 (Metadata & RBAC)
- **Time-Series DB**: InfluxDB 2.7 (High-frequency OES data)
- **Caching**: Redis 7 (Session and configuration cache)
- **Messaging Bus**: Apache Kafka & Zookeeper (Real-time data streaming)
- **Observability**: Prometheus (Metrics) & Grafana (Visualization & Alerts)
- **Orchestration**: Docker, Docker Compose, Kubernetes
- **Industrial Protocol**: Beckhoff ADS (TwinCAT 3 Integration)

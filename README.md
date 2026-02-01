# Proxmox PVE Summary Exporter

This project is a Proxmox VE (PVE) Summary Exporter that collects status metrics from your Proxmox nodes and exports them for monitoring. It includes a setup with Prometheus and Grafana for visualization.

## Configuration

 The exporter is configured using a YAML file located at `config/proxmox-node.yml`.

### Example Configuration (`config/proxmox-node.yml`)

Create a `config` directory and add a `proxmox-node.yml` file with the following structure:

```yaml
monitoring:
  # Node 1
  - identify: "Homelabs"
    # Host PVE URL
    host: "https://192.168.1.2:8006"
    # Credentials (Ticket Authentication)
    ticket:
      username: "root"
      password: "your_password"
      realm: "pam"
      # new-format: "1" is often required for newer Proxmox versions
      new-format: "1"
```

| Field | Description |
| :--- | :--- |
| `identify` | A unique label for the Proxmox node/cluster. |
| `host` | The URL of the Proxmox VE API (usually port 8006). |
| `ticket.username` | The username for authentication (e.g., `root`). |
| `ticket.password` | The password for the user. |
| `ticket.realm` | The authentication realm (e.g., `pam`, `pve`). |
| `ticket.new-format` | Set to `"1"` to use the new ticket format. |

## Deployment with Docker Compose

You can deploy the exporter along with Prometheus and Grafana using Docker Compose.

### `compose.yml`

```yaml
services:
  # Services PVE API Status Exporter
  pve-summary-exporter:
    image: ghcr.io/ernestoyoofi/pve-summary-exporter:latest
    container_name: pve-summary-exporter
    restart: unless-stopped
    environment:
      - PORT=8007
      - HOST=0.0.0.0
    ports:
      - 8007:8007
    user: root
    volumes:
      - ./config/proxmox-node.yml:/app/config/proxmox-node.yml
    networks:
      - pve-summary-exporter

  # Services Prometheus
  prometheus:
    image: prom/prometheus:latest
    container_name: pve-summary-exporter-prometheus
    depends_on:
      - pve-summary-exporter
    restart: unless-stopped
    ports:
      - 9090:9090
    command:
      - --config.file=/etc/prometheus/prometheus.yml
      - --storage.tsdb.path=/prometheus
      - --storage.tsdb.retention.time=1y
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    networks:
      - pve-summary-exporter

  # Services Grafana  
  grafana:
    image: grafana/grafana:latest
    container_name: pve-summary-exporter-grafana
    depends_on:
      - pve-summary-exporter
      - prometheus
    restart: unless-stopped
    environment:
      GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH: /etc/grafana/provisioning/dashboards/pve-summary-exporter/pve-summary-exporter.json
    ports:
      - 3010:3000
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources:ro
      - ./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards:ro
    networks:
      - pve-summary-exporter

networks:
  pve-summary-exporter:

volumes:
  grafana-data:
  prometheus-data:
```

### Running the Project

1. Ensure your `config/proxmox-node.yml`, `prometheus/prometheus.yml` and Grafana provisioning files are in place.
2. Start the services:

```bash
docker compose up -d
```

3. Access the services:
    - **Exporter Metrics**: `http://localhost:8007` (or configured host/port)
    - **Prometheus**: `http://localhost:9090`
    - **Grafana**: `http://localhost:3010` (Default login: `admin` / `admin`)

## Running Locally (Development)

To run the exporter directly on your machine without Docker:

1. **Prerequisites**: Ensure you have [Node.js](https://nodejs.org/) (v18+) or [Bun](https://bun.sh/) installed.

2. **Install Dependencies**:

   Using npm:
   ```bash
   npm install
   ```
   Or using Bun:
   ```bash
   bun install
   ```

3. **Start the Application**:

   ```bash
   # Using Node
   node server.js

   # Using Bun
   bun server.js
   ```

   The server will start on port `8007` by default.

## Environment Variables

You can configure the exporter using the following environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | The port the server listens on. | `8007` |
| `HOST` | The host address to bind to. | `0.0.0.0` |

# Clash Tracing Dashboard

An example of a clash tracing exporter API.

### Screenshot

![screenshot](./screenshot/screenshot.jpg)

### How to use

1. modify `docker-compose.yaml` and start (`docker-compose up -d`)
2. setup Grafana (add datasource)
3. import `panels/dashboard.json` and `panels/logs.json` to Grafana

> If you are using arm64 machines (like Mac M1), use `influxdb:1.8` image in `docker-compose.yml`.
> The alpine version does not yet have an image built for arm64.

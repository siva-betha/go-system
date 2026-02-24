from influxdb_client import InfluxDBClient
import os

def get_influx_client(config):
    return InfluxDBClient(
        url=config.get('url', os.Getenv('INFLUX_URL')),
        token=config.get('token', os.Getenv('INFLUX_TOKEN')),
        org=config.get('org', 'plc-org')
    )

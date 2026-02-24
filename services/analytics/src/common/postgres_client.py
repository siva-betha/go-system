import psycopg2
from psycopg2.extras import RealDictCursor
import os

def get_postgres_client(config):
    conn_str = config.get('conn_str', os.Getenv('DB_URL'))
    return psycopg2.connect(conn_str, cursor_factory=RealDictCursor)

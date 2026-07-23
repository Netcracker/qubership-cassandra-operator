import logging
import re
import json
from src.cassandra_client import CassandraClient

log = logging.getLogger(__name__)

MARKER_KEYSPACE = "backup_markers"
MARKER_TABLE = "markers"
MARKER_KEY = "cloud-backuper-marker"

def normalize_marker(value: str) -> str:
    value = value.strip()

    if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
        value = value[1:-1].strip()

    try:
        obj = json.loads(value)
        if isinstance(obj, dict) and "marker" in obj:
            return obj["marker"]
    except Exception:
        pass
    match = re.fullmatch(r"\{marker:\s*(.*)\}", value)
    if match:
        return match.group(1).strip()

    return value

def ensure_schema(client: CassandraClient):
    client.execute_query(
        f"CREATE KEYSPACE IF NOT EXISTS {MARKER_KEYSPACE} "
        f"WITH replication = {{'class': 'SimpleStrategy', 'replication_factor': 1}}"
    )
    client.execute_query(
        f"CREATE TABLE IF NOT EXISTS {MARKER_KEYSPACE}.{MARKER_TABLE} "
        f"(marker_key text PRIMARY KEY, marker_value text, created_at timestamp)"
    )


def set_marker(client: CassandraClient, value: str):
    ensure_schema(client)
    value = normalize_marker(value)
    log.info("Setting marker")
    log.info(f"Normalized marker: {repr(value)}")
    client.execute_query(
        f"INSERT INTO {MARKER_KEYSPACE}.{MARKER_TABLE} "
        f"(marker_key, marker_value, created_at) "
        f"VALUES ('{MARKER_KEY}', '{value}', toTimestamp(now()))"
    )
    log.info("Marker set successfully")

def get_marker(client: CassandraClient) -> str:
    ensure_schema(client)
    log.info("Getting marker")
    rows = client.execute_query(
        f"SELECT marker_value FROM {MARKER_KEYSPACE}.{MARKER_TABLE} WHERE marker_key = '{MARKER_KEY}'"
    )
    result = list(rows)
    if not result:
        raise ValueError("No marker found")
    value = result[0].marker_value
    log.info("Marker retrieved")
    return value

import pytest
from src import marker
from . import mocks


class Row:
    def __init__(self, marker_value):
        self.marker_value = marker_value


class MarkerMockClient(mocks.MockCassandraClient):
    def __init__(self, stored_value=None):
        super().__init__()
        self._stored_value = stored_value
        self.executed_queries = []

    def execute_query(self, query):
        self.executed_queries.append(query)
        if "SELECT" in query:
            if self._stored_value is not None:
                return [Row(self._stored_value)]
            return []
        return []


@pytest.fixture(autouse=True)
def mock_cassandra(mocker):
    mocker.patch('src.os_utils.reformat_hostnames', return_value=None)


def test_set_marker():
    client = MarkerMockClient()
    marker.set_marker(client, "my-value")
    insert_queries = [q for q in client.executed_queries if "INSERT INTO" in q]
    assert len(insert_queries) == 1
    assert marker.MARKER_KEY in insert_queries[0]
    assert "my-value" in insert_queries[0]


def test_set_marker_replaces_existing():
    client = MarkerMockClient()
    marker.set_marker(client, "first-value")
    marker.set_marker(client, "second-value")
    insert_queries = [q for q in client.executed_queries if "INSERT INTO" in q]
    assert len(insert_queries) == 2
    assert "second-value" in insert_queries[1]


def test_get_marker_success():
    client = MarkerMockClient(stored_value="my-value")
    result = marker.get_marker(client)
    assert result == "my-value"
    select_queries = [q for q in client.executed_queries if "SELECT" in q]
    assert len(select_queries) == 1


def test_get_marker_not_found():
    client = MarkerMockClient(stored_value=None)
    with pytest.raises(ValueError, match="No marker found"):
        marker.get_marker(client)


def test_ensure_schema_creates_keyspace_and_table():
    client = MarkerMockClient()
    marker.ensure_schema(client)
    assert any("CREATE KEYSPACE" in q for q in client.executed_queries)
    assert any("CREATE TABLE" in q for q in client.executed_queries)

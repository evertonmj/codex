import os
import pathlib
from codex_sdk import CodexClient


import pytest

@pytest.mark.smoke
def test_codex_sdk_smoke():
    db_path = pathlib.Path(__file__).parent / 'test-db.db'
    if db_path.exists():
        db_path.unlink()

    client = CodexClient(str(db_path))
    client.set('test-key', {'hello': 'world'})
    value = client.get('test-key')
    assert value['hello'] == 'world'

    assert client.has('test-key') is True
    assert 'test-key' in client.keys()

    client.delete('test-key')
    assert client.has('test-key') is False

    client.clear()

    if db_path.exists():
        db_path.unlink()

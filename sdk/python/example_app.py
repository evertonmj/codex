from codex_sdk import CodexClient
import pathlib

db_path = pathlib.Path(__file__).parent / 'example-db.db'
client = CodexClient(str(db_path))

# Set a value
client.set('greeting', {'msg': 'Hello from Python!'})

# Get the value
value = client.get('greeting')
print('Greeting:', value['msg'])

# List all keys
print('All keys:', client.keys())

# Delete the key
deleted = client.delete('greeting')
print('Deleted:', deleted)

# Cleanup
db_path.unlink(missing_ok=True)

# CodexDB Python SDK

Pequeno wrapper para `codex-cli` como processo externo.

## Instalação

Use `pip` ou apenas copie `codex_sdk.py` para seu projeto.

## Exemplo

```python
from codex_sdk import CodexClient

client = CodexClient(file='test.db')
client.set('user:1', {'name': 'Alice', 'age': 30})
print(client.get('user:1'))
print(client.keys())
client.delete('user:1')
```

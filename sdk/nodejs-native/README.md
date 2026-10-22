# CodexDB embarcado para Node.js

Biblioteca compilada carregada no processo Node.js. Não inicia servidor nem executa a CLI. API síncrona; operações de disco bloqueiam o event loop. Para trabalho intenso, use um Worker e mantenha a instância naquele Worker.

## Compilar e importar

Na raiz do repositório, com Go, compilador C e headers do Node instalados:

```sh
make build-node
# Caso os headers estejam em outra pasta:
make build-node NODE_INCLUDE_DIR=/caminho/para/include/node
npm install ./sdk/nodejs-native --ignore-scripts
```

```js
const { CodexDatabase } = require('codexdb-native');
const db = new CodexDatabase({ file: './data.json' });
try {
  db.set('user', { name: 'João', age: 30 });
  console.log(db.get('user'));
  console.log(db.keys());
} finally {
  db.close();
}
```

O pacote também pode ser importado por caminho: `require('./sdk/nodejs-native')`. A compilação gera `build/codex.node`, `build/libcodex.so` (Linux) ou `build/libcodex.dylib` (macOS) e `build/libcodex.h`. Distribua o addon e a biblioteca juntos. O pacote é local; esta alteração não publica no npm nem baixa binários automaticamente. Quem consome o pacote já compilado não precisa de Go ou compilador.

Cada combinação de sistema operacional, arquitetura e libc exige seu próprio build. O script suporta Linux e macOS; Linux amd64 foi validado. Em macOS, forneça o diretório dos headers do Node. O addon Windows ainda não está implementado; a API C pode ser compilada como DLL com a toolchain adequada.

## API

- `new CodexDatabase({ file?, encryptionKey?, ledger? })`: abre/cria o banco. Caminho padrão: `.codex-data/default.db`.
- `set(key, value)`, `get(key)`, `delete(key)`, `has(key)`, `keys()`, `clear()`.
- `close()`: libera o bloqueio; pode ser chamado novamente.

Valores devem ser serializáveis em JSON. Ausência de chave em `get()` lança erro com `code = 'NOT_FOUND'`; JSON `null` continua sendo um valor válido. Outros códigos incluem `LOCKED`, `INVALID_KEY` e `CODEX_ERROR`. Sempre feche a instância. Apenas uma instância/processo pode abrir o mesmo arquivo por vez.

## Criptografia opcional

Sem `encryptionKey`, o snapshot contém JSON legível com checksum de integridade. A variável `CODEX_KEY` não afeta esta API embarcada. Para ativar AES-GCM:

```js
const db = new CodexDatabase({
  file: './encrypted.db',
  encryptionKey: process.env.MY_DATABASE_KEY
});
```

A chave deve ter 16, 24 ou 32 bytes UTF-8. A aplicação deve validar que a variável está preenchida quando exigir criptografia. Use a mesma chave e o mesmo modo em todas as reaberturas. Arquivos existentes não são descriptografados automaticamente; snapshots antigos em base64 são lidos e passam ao formato JSON legível na próxima escrita, mantendo a configuração de criptografia. Versões antigas do CodexDB não leem o novo formato de snapshot.

`ledger: true` mantém o formato append-only binário existente, mesmo sem criptografia; para um arquivo JSON legível, use o snapshot padrão.

## Outras linguagens

A biblioteca exporta uma API C estável com duas funções (`CodexCall`, `CodexFree`). Veja [o protocolo](../../docs/NATIVE_LIBRARY.md) e [o exemplo Python com ctypes](../python-native/test_native.py).

```sh
make test-native
```

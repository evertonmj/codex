# Biblioteca compilada embarcada

Uma biblioteca compartilhada é o artefato apropriado para importar o CodexDB dentro de outro processo. Um executável separado continua disponível pelas CLIs existentes. A biblioteca usa `go build -buildmode=c-shared` e exporta uma API C; pode ser chamada por linguagens com suporte a FFI (C/C++, Python ctypes, Rust extern, Java JNI/JNA, C# P/Invoke e outras). Node.js tem um addon Node-API, sem dependência de FFI de terceiros.

Referências oficiais: [Go build modes](https://pkg.go.dev/cmd/go), [cgo exports](https://pkg.go.dev/cmd/cgo) e [Node-API](https://nodejs.org/api/n-api.html).

## Build

```sh
make build-shared                           # bin/libcodex.so e bin/libcodex.h
make build-shared SHARED_EXT=dylib           # macOS, com toolchain local
make build-shared SHARED_EXT=dll             # Windows, com toolchain local
make build-node                             # addon + biblioteca para Node.js
make test-native                            # testes reais Node.js e Python
```

Exige CGO, compilador C e Go na compilação. Compile em cada plataforma/arquitetura desejada. Não há um único binário universal. O consumo do artefato compilado dispensa Go e servidor. Linux amd64 foi verificado; os demais builds exigem validação na plataforma de destino.

## ABI versão 1

Inclua o header `libcodex.h` gerado pelo build. As funções públicas são:

```c
char *CodexCall(char *request);
void CodexFree(char *response);
```

`request` é uma string JSON UTF-8 terminada em NUL, apenas lida durante a chamada. A resposta é uma string JSON UTF-8 terminada em NUL alocada pela biblioteca. Copie/leia a resposta e chame `CodexFree` exatamente uma vez, inclusive quando `ok` for falso. Não use `free` de outra biblioteca. Nenhum ponteiro Go atravessa a ABI. `CodexCall(NULL)` retorna erro de protocolo; `CodexFree(NULL)` é permitido.

Todas as chamadas são síncronas e serializadas, incluindo fechamento. O caller mantém os handles até `close`, para liberar os bloqueios. Os handles são strings decimais opacas, locais à biblioteca/processo e não persistem entre reinicializações.

Requisições:

```json
{"op":"open","file":"./data.json"}
{"op":"open","file":"./encrypted.db","encryptionKey":"01234567890123456789012345678901","ledger":false}
{"op":"set","handle":"1","key":"user","value":{"name":"João"}}
{"op":"get","handle":"1","key":"user"}
{"op":"has","handle":"1","key":"user"}
{"op":"keys","handle":"1"}
{"op":"delete","handle":"1","key":"user"}
{"op":"clear","handle":"1"}
{"op":"close","handle":"1"}
```

Cada linha representa uma chamada independente. `open` retorna um handle; `keys` retorna array ordenado; `get` retorna o valor JSON exato; `has` retorna booleano; as mutações e `close` retornam `true` em caso de sucesso. Respostas:

```json
{"ok":true,"result":"1"}
{"ok":true,"result":{"name":"João"}}
{"ok":false,"result":null,"error":"key not found: missing","code":"NOT_FOUND"}
```

Códigos: `NOT_FOUND`, `LOCKED`, `INVALID_KEY`, `CODEX_ERROR`. Em erros, use `ok`/`code`; `result` não representa sucesso. Snapshot é o padrão; criptografia só é ativada por chave explícita. As chaves têm 16, 24 ou 32 bytes UTF-8 e não são senhas derivadas automaticamente. O modo ledger mantém a estrutura binária de frames. Uma primeira entrada ilegível retorna erro, para evitar abrir como vazio um banco com chave/configuração incorreta. A recuperação de corrupção em entradas posteriores mantém o prefixo válido.

## Snapshot JSON versão 2

```json
{
  "checksum": "sha256-do-data-compactado",
  "data": {
    "format": 2,
    "values": {
      "user": {"name":"João"},
      "enabled": true
    }
  }
}
```

O checksum continua sendo verificado antes da leitura. A biblioteca lê os snapshots legados que armazenavam bytes como base64, inclusive com criptografia/compressão configuradas. Na próxima escrita, usa versão 2. Arquivos criptografados exigem sua chave original e continuam criptografados; não há migração implícita para plaintext. A mudança não é legível por versões antigas da biblioteca: guarde backup antes de atualizar se precisar fazer downgrade.

Para remover criptografia de um arquivo existente, abra com a chave original, copie seus valores para um banco novo sem `EncryptionKey` e feche ambos. Para uso Go, `codex.New(path)` grava plaintext e `codex.NewWithOptions(path, codex.Options{EncryptionKey: key})` ativa AES-GCM.

## Verificação

`make test-coverage` mede cobertura agregada dos pacotes `app/...`, `internal/native` e `cmd/codex-shared`, executando também os testes de integração, e falha abaixo de 95%. Os comandos CLI, benchmarks e exemplos existentes não entram nesse denominador; são verificados por `go test -race ./...`. `make test-native` verifica o addon compilado em Node.js e a ABI em Python.

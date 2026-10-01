# API local com Caddy

Para rodar continuamente em um VPS Ubuntu 24.04 com HTTPS em `prsnlspc.xyz`,
veja o [guia de instalação via git clone](../deploy/README.md).

A API usa o pacote `env/filesystem` e os mesmos tipos JSON do cliente existente.
Não há dependências Go novas. Por padrão, usa `~/prsnlspc`, o mesmo diretório do
cliente local. O executável da interface continua funcionando como antes.

Em um terminal, na raiz do projeto:

```sh
go run ./cmd/api
```

Em outro:

```sh
caddy run --config Caddyfile --adapter caddyfile
```

O Caddy recebe em `http://127.0.0.1:8080` e encaminha para a API em
`127.0.0.1:8081`. Configure no arquivo `~/prsnlspc/.options` do cliente:

```text
url: http://127.0.0.1:8080
username:
password:
```

Credenciais vazias são o padrão para uso local. Para exigir credenciais, inicie
com `API_USERNAME` e `API_PASSWORD` e coloque os mesmos valores em `.options`.
O cliente chama `/auth` automaticamente. Cada login bem-sucedido gera um token
de sessão independente. A API mantém no máximo 12 tokens válidos; o 13º login
invalida o token criado há mais tempo, mesmo que ele tenha sido usado recentemente.
Logins inválidos não alteram as sessões. Os tokens ficam em memória, sem expiração
por tempo; reiniciar a API invalida todos. Para obter outro token, autentique
novamente (no cliente atual, reinicie o cliente). A configuração padrão escuta apenas em
loopback; ela não publica seus arquivos na rede.

Para usar outra pasta, passe `-root /caminho/da/pasta`. Isso também funciona com
`go run`, sem depender da localização do binário temporário.

## Contrato

- `POST /auth`: recebe `{"username":"","password":""}` e retorna `{"token":"..."}`.
- `POST /load`: recebe `types.Loading` e retorna `types.Page` ou `null` se os filtros
  não selecionarem páginas. Campos omitidos usam `NewLoading()`; `Api` não causa
  chamadas remotas. Falha ao carregar retorna 404.
- `POST /sync`: recebe **uma `types.Page`**, como `SyncApi` já envia, e retorna 204
  depois de persistir. `draft`, `edit`, `local`, `*api` e `api` gravam;
  `move` move usando `Og.Path`; `doom` exclui; `ghost` não altera arquivos.
  Filhos são enviados separadamente pelo cliente. Diretórios (`Type: "deep"`)
  guardam conteúdo no arquivo `index`, seguindo o filesystem local.

`/load` e `/sync` exigem `Authorization: Bearer TOKEN`. O limite de JSON é 8 MiB.
JSON inválido e erros de sync retornam 400; token inválido retorna 401.
O conteúdo é a fonte persistida; os metadados são derivados pelo filesystem na
leitura. As operações de sync são imediatas, sem o temporizador de patches da UI.
Movimento seguido de gravação e sincronização de múltiplas páginas não são
transacionais. O carregamento mantém a semântica de leitura do filesystem local;
use uma pasta local confiável.

Exemplo após obter o token com `/auth`:

```sh
curl http://127.0.0.1:8080/sync \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"Path":"exemplo","Type":"shallow","Stage":"draft","Content":["Olá"]}'

curl http://127.0.0.1:8080/load \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"Path":"exemplo"}'
```

Validação:

```sh
go test ./...
caddy validate --config Caddyfile --adapter caddyfile
```

Referência: [reverse_proxy do Caddy](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy).

## Exportar conteúdo no navegador

`GET /shhh` retorna `{"files":{"pasta/nota":"conteúdo\n..."}}`, com todos
os arquivos de texto do armazenamento, recursivamente, sem filtros de páginas.
Inclui arquivos `index` e ocultos, exceto `.options` (configuração do cliente).
Links simbólicos e arquivos especiais não são exportados. Pastas vazias não
aparecem. Uma falha de leitura retorna 500, sem exportação parcial.

Abra `https://prsnlspc.xyz/shhh`: o navegador solicita HTTP Basic com
`API_USERNAME` e `API_PASSWORD`, as mesmas credenciais usadas em `/auth`.
Não use a senha do usuário Linux `nausea`. A rota exige ambas as credenciais
configuradas, inclusive em uso local, e envia `Cache-Control: no-store`.
Também é possível usar (o curl solicita a senha):

```sh
curl --user prsnlspc https://prsnlspc.xyz/shhh
```

Após publicar o código no repositório usado pela VPS, execute nela
`sudo systemctl restart prsnlspc-api` para atualizar e compilar a API.

# VPS Ubuntu 24.04

Cliente → `https://prsnlspc.xyz` → Caddy → API em `127.0.0.1:8081`.
O systemd inicia a API no boot e reinicia o processo se ele cair. Antes de cada
início (inclusive reinícios automáticos), atualiza a branch `main` de
`https://github.com/raxcha/env.git` e compila `./cmd/api`.
O instalador é destinado a um VPS dedicado: substitui `/etc/caddy/Caddyfile`,
guardando a configuração anterior em `/etc/caddy/Caddyfile.before-prsnlspc`.

## 1. DNS e rede

No provedor DNS do domínio, configure um registro **A** para `@` apontando para
o IPv4 público do VPS. Se houver **AAAA**, ele deve apontar para o IPv6 funcional
desse mesmo VPS; remova um AAAA antigo ou incorreto.

Libere TCP **80 e 443** no firewall do provedor e no Ubuntu, caso estejam ativos.
Mantenha a porta SSH liberada. A porta 8081 fica apenas em loopback.
Para este VPS, que usa SSH na porta 22022:

```sh
sudo ufw allow 22022/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

O [Caddy emite e renova o certificado automaticamente](https://caddyserver.com/docs/automatic-https)
quando o DNS e o acesso às portas estiverem corretos.

## 2. Instalar via git clone

Entre no VPS por SSH (substitua `IP_DO_VPS`):

```sh
ssh -p 22022 root@IP_DO_VPS
```

Já no VPS:

```sh
sudo apt update
sudo apt install -y git golang-go caddy openssl curl ca-certificates
git clone https://github.com/raxcha/env.git
cd env
sudo bash deploy/install.sh
sudo cat /etc/prsnlspc/api.env
```

Os arquivos de deploy precisam estar publicados no repositório antes do clone.
A atualização automática usa HTTPS sem interação. Se o repositório for privado,
será necessário configurar acesso de leitura ao GitHub também para o usuário
Linux `prsnlspc`; o login Git de root não é compartilhado com esse usuário.
Sem esse acesso, o serviço mantém o binário instalado e registra a falha nos logs.
O projeto exige Go 1.26.0; o instalador usa `GOTOOLCHAIN=auto`, que
[baixa a versão exigida pelo módulo](https://go.dev/doc/toolchain).
A primeira compilação exige acesso à internet e pode demorar.

O instalador gera usuário `prsnlspc` e senha aleatória para a API. As credenciais
ficam em `/etc/prsnlspc/api.env`, legível apenas por root, e são preservadas ao
repetir a instalação. O processo roda como usuário Linux `prsnlspc`, sem
privilégios administrativos. Os dados ficam em `/var/lib/prsnlspc`; faça backup
dessa pasta e do arquivo de credenciais. O VPS começa com uma pasta de dados nova;
o clone não importa seus dados locais.

## 3. Conectar o cliente

Na sua máquina local, edite estes campos em `~/prsnlspc/.options`, mantendo as
outras opções. Use a senha mostrada no VPS:

```text
url: https://prsnlspc.xyz
username: prsnlspc
password: COLE_A_SENHA_GERADA
```

Reinicie o cliente. Ele autentica automaticamente em `/auth`.

## 4. Verificar e manter

```sh
sudo systemctl status prsnlspc-api caddy --no-pager
curl -i -X POST https://prsnlspc.xyz/load
```

O curl deve retornar **401 Unauthorized**: confirma que HTTPS e proxy chegam à
rota protegida sem token. Para validar a autenticação, conecte o cliente com as
credenciais geradas. Abrir `/` no navegador retorna 404: não há página web.

Logs:

```sh
sudo journalctl -u prsnlspc-api -u caddy -n 100 --no-pager
```

Para ativar esta configuração em um VPS já instalado, atualize o clone e execute
o instalador uma vez (depois de publicar as alterações no GitHub):

```sh
git pull --ff-only
sudo bash deploy/install.sh
```

Depois disso, para buscar a versão mais recente e iniciar a API:

```sh
sudo systemctl restart prsnlspc-api
sudo journalctl -u prsnlspc-api -n 50 --no-pager
```

Os logs mostram o commit compilado ou a falha de atualização. Não há polling:
um push no GitHub só entra em execução no próximo início do serviço. O clone
usado pelo serviço fica em `/var/cache/prsnlspc/source`; os caches Go e o binário
ficam em `/var/cache/prsnlspc`. Não edite esse clone; publique alterações pela
sua cópia de desenvolvimento. O serviço faz `git pull --ff-only`, sem resolver
divergências automaticamente. Alterações no instalador, no serviço systemd ou no
Caddy exigem executar o instalador novamente; apenas o código Go é atualizado
nos reinícios.

A atualização tem limite de cinco minutos. Se o download ou a compilação falhar
ou exceder esse limite, o systemd inicia o binário anterior e registra a falha.
O binário é substituído apenas após compilação bem-sucedida. A API fica
indisponível durante a atualização; o Caddy pode retornar 502 nesse intervalo.
Uma compilação bem-sucedida não garante que a versão nova esteja livre de bugs;
não há rollback automático de falhas em execução.

O atualizador roda sem root, com o mesmo usuário da API. A pasta de dados
`/var/lib/prsnlspc` e as credenciais em `/etc/prsnlspc/api.env` são preservadas.
A atualização não migra nem apaga dados; mudanças futuras no comportamento do
código Go continuam exigindo os cuidados normais com backup.
Reiniciar a API invalida os tokens em memória;
reinicie o cliente para autenticar novamente. Para alterar a senha, edite
`/etc/prsnlspc/api.env` com `sudoedit`, reinicie com
`sudo systemctl restart prsnlspc-api` e atualize a senha no cliente.

Referência: [ExecStartPre e reinícios no systemd](https://www.freedesktop.org/software/systemd/man/latest/systemd.service.html).

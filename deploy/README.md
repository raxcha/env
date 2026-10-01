# VPS Ubuntu 24.04

Cliente → `https://prsnlspc.xyz` → Caddy → API em `127.0.0.1:8081`.
O systemd inicia a API no boot e reinicia o processo se ele cair.
O instalador é destinado a um VPS dedicado: substitui `/etc/caddy/Caddyfile`,
guardando a configuração anterior em `/etc/caddy/Caddyfile.before-prsnlspc`.

## 1. DNS e rede

No provedor DNS do domínio, configure um registro **A** para `@` apontando para
o IPv4 público do VPS. Se houver **AAAA**, ele deve apontar para o IPv6 funcional
desse mesmo VPS; remova um AAAA antigo ou incorreto.

Libere TCP **80 e 443** no firewall do provedor e no Ubuntu, caso estejam ativos.
Mantenha a porta SSH liberada. A porta 8081 fica apenas em loopback.
Se usa UFW com SSH na porta padrão:

```sh
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
```

O [Caddy emite e renova o certificado automaticamente](https://caddyserver.com/docs/automatic-https)
quando o DNS e o acesso às portas estiverem corretos.

## 2. Instalar via git clone

Entre no VPS por SSH (substitua `IP_DO_VPS`):

```sh
ssh root@IP_DO_VPS
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
Se o repositório for privado, use sua autenticação GitHub para cloná-lo.
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

Para atualizar, dentro do clone no VPS:

```sh
git pull --ff-only
sudo bash deploy/install.sh
```

Os dados e a senha são preservados. Reiniciar a API invalida os tokens em memória;
reinicie o cliente para autenticar novamente. Para alterar a senha, edite
`/etc/prsnlspc/api.env` com `sudoedit`, reinicie com
`sudo systemctl restart prsnlspc-api` e atualize a senha no cliente.

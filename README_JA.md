# Sub2API デプロイ

<p align="center">
  <a href="README.md"><img src="https://img.shields.io/badge/English-README-0969da" alt="English"></a>
  <a href="README_CN.md"><img src="https://img.shields.io/badge/%E7%AE%80%E4%BD%93%E4%B8%AD%E6%96%87-README-0969da" alt="简体中文"></a>
  <a href="README_JA.md"><img src="https://img.shields.io/badge/%E6%97%A5%E6%9C%AC%E8%AA%9E-README-0969da" alt="日本語"></a>
</p>

このリポジトリは [mizaawa/sub2api](https://github.com/mizaawa/sub2api) で管理されています。デプロイ用イメージ、インストールスクリプト、管理画面のワンクリック更新とロールバックは、すべてこのリポジトリの Release を使用します。

## Docker Compose デプロイ

必要環境：Docker Engine 20.10 以降、Docker Compose v2 以降。

### ワンクリック準備

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/mizaawa/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

スクリプトは `.env` を作成し、`POSTGRES_PASSWORD`、`JWT_SECRET`、`TOTP_ENCRYPTION_KEY` を自動生成します。データは `data`、`postgres_data`、`redis_data` に保存されるため、ディレクトリ単位でバックアップや移行ができます。

コンテナが正常になったら `http://SERVER_IP:8080` を開きます。`.env` に `ADMIN_PASSWORD` を設定していない場合は、アプリケーションログで自動生成されたパスワードを確認してください。

### 手動デプロイ

```bash
git clone https://github.com/mizaawa/sub2api.git
cd sub2api/deploy
cp .env.example .env
chmod 600 .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
```

本番環境では、起動前に `.env` の `POSTGRES_PASSWORD`、`JWT_SECRET`、`TOTP_ENCRYPTION_KEY` に強力なランダム値を設定してください。必要に応じて `ADMIN_EMAIL`、`ADMIN_PASSWORD`、`SERVER_PORT` も設定できます。

### 既存の上流版デプロイからデータを保持して移行

ローカルディレクトリ版 Compose は、アプリケーションデータを `./data`、PostgreSQL を `./postgres_data`、Redis を `./redis_data` に保存します。以下の操作は必ず**既存のデプロイディレクトリ**で実行してください。別のディレクトリを作成したり、`docker-deploy.sh` を再実行したりしないでください。`.env` が置き換えられ、新しい秘密鍵が生成される可能性があります。

```bash
cd /path/to/sub2api-deploy
umask 077
STAMP=$(date +%Y%m%d-%H%M%S)
mkdir -p "backups/$STAMP"
cp .env docker-compose.yml "backups/$STAMP/"
docker compose exec -T postgres sh -ec 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "backups/$STAMP/sub2api.dump"

# sub2api のイメージだけを変更し、postgres と redis は変更しない
sed -i 's#image: weishaw/sub2api:latest#image: ${SUB2API_IMAGE:-ghcr.io/mizaawa/sub2api:latest}#' docker-compose.yml
printf '\nSUB2API_IMAGE=ghcr.io/mizaawa/sub2api:latest\n' >> .env

docker compose config -q
docker compose pull sub2api
docker compose up -d --no-deps sub2api
docker compose ps
```

この手順では `sub2api` コンテナだけを再作成し、PostgreSQL と Redis のコンテナやデータディレクトリは削除しません。既存の `.env` にある `POSTGRES_PASSWORD`、`JWT_SECRET`、`TOTP_ENCRYPTION_KEY`、プロキシ、プロバイダー設定はそのまま保持してください。

### 更新とロールバック

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

特定のイメージバージョンに固定またはロールバックする場合は、`.env` に `SUB2API_IMAGE=ghcr.io/mizaawa/sub2api:<version>` を設定して同じコマンドを実行します。管理画面のワンクリック更新とロールバックは `mizaawa/sub2api` の GitHub Releases からバージョンを取得します。

### よく使うコマンド

```bash
docker compose -f docker-compose.local.yml ps
docker compose -f docker-compose.local.yml logs -f sub2api
docker compose -f docker-compose.local.yml restart
docker compose -f docker-compose.local.yml down
```

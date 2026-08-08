# Sub2API Deployment

This repository is maintained at [mizaawa/sub2api](https://github.com/mizaawa/sub2api). Deployment images, install scripts, and the in-app version update and rollback feature all use this repository's releases.

## Docker Compose

```bash
mkdir -p sub2api-deploy && cd sub2api-deploy
curl -sSL https://raw.githubusercontent.com/mizaawa/sub2api/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f sub2api
```

For a manual deployment:

```bash
git clone https://github.com/mizaawa/sub2api.git
cd sub2api/deploy
cp .env.example .env
chmod 600 .env
mkdir -p data postgres_data redis_data
docker compose -f docker-compose.local.yml up -d
```

Set strong `POSTGRES_PASSWORD`, `JWT_SECRET`, and `TOTP_ENCRYPTION_KEY` values in `.env` before production use. Open `http://SERVER_IP:8080` when containers are healthy.

## Update and rollback

```bash
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

To pin or roll back the Docker image, set `SUB2API_IMAGE=ghcr.io/mizaawa/sub2api:<version>` in `.env` and rerun the commands. The administrator console retrieves releases for one-click updates and rollbacks from `mizaawa/sub2api`.

# Garage S3 - Docker Compose Setup

## Quick Start

```bash
# 1. Avvia il container
docker compose up -d

# 2. Ottieni il Node ID
docker exec garage garage node id

# 3. Crea un layout (zona e capacità in GB)
docker exec garage garage layout assign -z dc1 -c 10G <NODE_ID>

# 4. Applica il layout
docker exec garage garage layout apply --version 1

# 5. Crea una chiave di accesso
docker exec garage garage key create my-key

# 6. Crea un bucket
docker exec garage garage bucket create my-bucket

# 7. Collega bucket alla chiave con permessi full
docker exec garage garage bucket allow my-bucket --read --write --owner --key my-key
```

## Porte Esposte

| Porta | Servizio            | Descrizione                            |
|-------|---------------------|----------------------------------------|
| 3900  | S3 API              | Endpoint compatibile S3 (principale)   |
| 3901  | Admin API           | API di amministrazione REST            |
| 3902  | Web / Static Sites  | Hosting siti statici da bucket         |
| 3903  | K2V API             | Key-Value store API (opzionale)        |

## Accesso S3

Configura il client S3 (es. `aws` CLI, `rclone`, `s3cmd`) con:

- **Endpoint:** `http://localhost:3900`
- **Region:** `garage`
- **Access Key ID / Secret Access Key:** ottenuti dal comando `garage key create`

### Esempio con AWS CLI

```bash
aws --endpoint-url http://localhost:3900 \
    --region garage \
    s3 ls
```

### Esempio con rclone

```ini
[garage]
type = s3
provider = Other
access_key_id = <ACCESS_KEY>
secret_access_key = <SECRET_KEY>
endpoint = http://localhost:3900
region = garage
```

## Admin API

```bash
# Status del cluster
curl http://localhost:3901/v1/status

# Lista bucket
curl http://localhost:3901/v1/bucket?list
```

## Note

- `replication_factor = 1` è adatto per sviluppo/singolo nodo; usa 2 o 3 per produzione.
- Monta `garage.toml` come volume read-only nel container.
- I dati sono persistiti nei volumi Docker `garage_meta` e `garage_data`.

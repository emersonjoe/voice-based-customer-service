#!/bin/sh
# Sobe o servidor carregando o .env local (o trilha não lê .env sozinho).
set -a
. ./.env
set +a
exec trilha dev --addr "${TRILHA_ADDR:-:3210}"

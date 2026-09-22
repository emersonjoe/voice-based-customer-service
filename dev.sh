#!/bin/sh
# Sobe o servidor carregando o .env local (o trilha não lê .env sozinho).
cd "$(dirname "$0")" || exit 1

PORTA="${TRILHA_ADDR:-:3210}"
PORTA_NUM="${PORTA##*:}"
if lsof -ti ":$PORTA_NUM" >/dev/null 2>&1; then
  echo "A porta $PORTA_NUM já está em uso — provavelmente por um servidor antigo"
  echo "iniciado sem o .env. Encerre-o e rode de novo:"
  echo ""
  echo "  lsof -ti :$PORTA_NUM | xargs kill && ./dev.sh"
  exit 1
fi

set -a
. ./.env
set +a
exec trilha dev --addr "$PORTA"

#!/bin/sh
# Despliega Tu Local Web en el VPS (todo corre ahí; Cloudflare es sólo DNS).
#   scripts/desplegar-vps.sh          la web y la API
#   scripts/desplegar-vps.sh web      sólo la web
#   scripts/desplegar-vps.sh api      sólo la API (con sus migraciones)
set -eu
VPS=root@187.41.102.190
RAIZ=$(cd "$(dirname "$0")/.." && pwd)
QUE=${1:-todo}

# Nada se compila en la Mac: lo compila el workflow «Desplegar» de GitHub Actions sobre el commit
# que está en origin/main, y acá sólo se baja y se sube al VPS.
SHA=$(git -C "$RAIZ" rev-parse HEAD)
git -C "$RAIZ" fetch -q origin main
[ "$SHA" = "$(git -C "$RAIZ" rev-parse origin/main)" ] || { echo "HEAD no es origin/main: desplegá desde una copia al día" >&2; exit 1; }
gh workflow run desplegar.yml --ref main
RUN=""
for _ in 1 2 3 4 5 6 7 8 9 10 11 12; do
	sleep 5
	RUN=$(gh run list --workflow=desplegar.yml --branch main --limit 5 --json databaseId,headSha,createdAt \
		-q "[.[] | select(.headSha == \"$SHA\")][0].databaseId")
	[ -n "$RUN" ] && break
done
[ -n "$RUN" ] || { echo "no encontré la corrida de Desplegar para $SHA" >&2; exit 1; }
gh run watch "$RUN" --exit-status >/dev/null
SAL=$(mktemp -d)
gh run download "$RUN" -n despliegue -D "$SAL"
[ "$(cat "$SAL/commit.txt")" = "$SHA" ] || { echo "el artefacto no es de $SHA" >&2; exit 1; }
chmod +x "$SAL"/bin/*

if [ "$QUE" = todo ] || [ "$QUE" = api ]; then
	for c in api worker mostradorctl; do
		scp -q "$SAL/bin/$c" "$VPS:/opt/tulocalweb/bin/$c.nuevo"
	done
	ssh "$VPS" 'cd /opt/tulocalweb/bin && for c in api worker mostradorctl; do mv $c.nuevo $c; done &&
		cd .. && docker compose run --rm --no-deps -T api /app/mostradorctl migrar && docker compose up -d --force-recreate --no-deps api worker'
fi

if [ "$QUE" = todo ] || [ "$QUE" = web ]; then
	rsync -a --delete "$SAL/web/build/" "$VPS:/opt/tulocalweb/web/build/"
	rsync -a --delete "$SAL/web/migraciones-sandbox/" "$VPS:/opt/tulocalweb/web/migraciones/sandbox/"
	rsync -a --delete "$SAL/web/migraciones-estado/" "$VPS:/opt/tulocalweb/web/migraciones/estado/"
	ssh "$VPS" 'cd /opt/tulocalweb && docker compose up -d --force-recreate --no-deps web'
fi
rm -rf "$SAL"

sleep 4
curl -s -o /dev/null -w "tulocalweb.com.ar %{http_code}\n" https://tulocalweb.com.ar/

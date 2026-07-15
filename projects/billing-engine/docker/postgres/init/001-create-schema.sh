#!/bin/sh
set -eu

psql \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --set=admin_user="$POSTGRES_USER" \
  --set=db_name="$POSTGRES_DB" <<'SQL'
CREATE SCHEMA IF NOT EXISTS "billing-engine" AUTHORIZATION :"admin_user";
ALTER DATABASE :"db_name" SET search_path TO "billing-engine", public;
SQL


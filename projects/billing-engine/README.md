# Billing Engine

Local infrastructure is managed with Docker Compose. PostgreSQL stores the data,
and PgBouncer is the application-facing connection pool.

## Start

1. Copy `.env.example` to `.env` and set a non-default administrator password.
2. Start the services:

   ```sh
   make up
   ```

3. Check their health:

   ```sh
   make ps
   ```

Run `make help` to list the available Compose commands. Use `make logs` to
follow service logs and `make down` to stop the stack.

Applications running on the host connect through PgBouncer:

```text
postgresql://billing_admin:<password>@localhost:6432/amartha
```

Other Compose services should use `pgbouncer:5432` instead of `localhost`.
Direct PostgreSQL access remains available at `localhost:5432` for migrations
and administration. Both published ports are bound to loopback only.

The initial database is `amartha`. The initialization script creates the quoted
schema `"billing-engine"` and makes it the database's default schema, ahead of
`public` in `search_path`.

## Pool limits

| Setting | Value | Layer |
| --- | ---: | --- |
| Maximum client connections | 1,000 | PgBouncer |
| Maximum upstream pool size | 100 | PgBouncer |
| Upstream connection lifetime | 15 minutes | PgBouncer |
| Upstream idle timeout | 5 minutes | PgBouncer |
| PostgreSQL connection capacity | 110 | PostgreSQL |

The PostgreSQL limit is deliberately lower than 1,000: PgBouncer accepts up to
1,000 clients and multiplexes them over at most 100 server connections for the
configured user/database pair. Ten extra PostgreSQL slots remain available for
maintenance. These values are configurable in `.env`.

Initialization scripts run only when the data volume is first created. To apply
different initial credentials, database name, or schema initialization to an
existing local instance, remove the volume with `docker compose down -v` first;
that command permanently deletes the local database data.

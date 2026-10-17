# Taaza development foundation

This checkout is under implementation. Existing UI features are not all operational;
the previous payment screen still contains a simulation and must not be used to take
real payments. This milestone establishes dummy configuration and security fixes.

## Local development (Docker Engine + Compose 2.24.4 or later)

```sh
cp .env.example .env
docker compose up --build -d
docker compose run --rm bootstrap-admin
```

Customer: http://localhost:3000, admin: http://localhost:3001.
Test inbox: http://localhost:8025. Use an address such as customer@example.test;
email OTP messages are captured in Mailpit, never delivered externally.
Dummy admin credentials are in .env.example. The bootstrap command creates an admin
only if it does not already exist, and refuses to run in production. Demo values
are intentionally public; use this mode only for local development.

PostgreSQL, Valkey (using the internal service name `redis` for compatibility), and
backend services do not expose host ports. Uploads and database data use volumes.
Development frontend/inbox ports bind to 127.0.0.1 only.

## Production preparation (not a completed release)

```sh
cp .env.production.example .env.production
# Fill real values before running this command:
docker compose --env-file .env.production -f docker-compose.yml -f compose.prod.yml config --quiet
```

Production interpolation rejects missing required values. Customer/admin signing
keys must be distinct strong random values of at least 32 characters. Dummy values
are rejected by JWT validation in production. Mailpit is omitted unless its explicit
development profile is enabled. The app has not completed the production-readiness
plan; do not expose it publicly until the remaining tests and migrations are ready.
Never use `down -v` on production or reuse test data volumes for real data.
Existing credentials committed upstream must be revoked separately: removing files
in this checkout cannot revoke an SMTP credential or erase repository history.

The existing initialization scripts apply only to an empty database. The migration
workflow for existing installations, admin session revocation, production bootstrap,
real payments, asset compilation and full journey testing remain outstanding.

## Tests

Run `go test ./...` inside each Go service. Security regression tests cover JWT
validation, authenticated profile completion, inventory authorization and local email
capture. Database/Compose integration and real-device checks need a usable Docker
host and must not be represented as passed by unit tests.

The standalone test project also runs the Go race detector on a Docker host:

```sh
docker compose -p taaza-tests -f compose.test.yml run --rm go-tests
python -m unittest discover -s tests -v
```

These are code/configuration checks. Full application tests against actual database
and email containers will be added as the remaining workflows are implemented.

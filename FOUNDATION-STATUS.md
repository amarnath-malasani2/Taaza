# Dummy configuration and initial security milestone

Base commit: `7bf28a224b2be25827a7c47380fb0419a4597b45`.
Prepared: 17 October 2026.

## Implemented

- Development-only dummy database/customer/admin settings and example.test identities.
- Mailpit test inbox and SMTP capture support; production rejects capture mode.
- Removed tracked service .env files and the embedded Compose SMTP credential;
  ignored real environment files and prevented service .env files entering images.
- Valkey service with persistent storage, private backend ports, loopback-only local
  frontend/inbox ports and durable upload storage.
- Production override requires separate real settings and HTTPS domains. Its admin
  bootstrap is explicitly disabled at runtime outside development/test.
- Non-root Go runtime images, bounded proxy HTTP calls, same-origin routing and real
  static 404 responses.
- Profile completion requires a verified session and updates only its user ID.
  Signup OTP verification now establishes the session needed for that operation.
- Inventory price/stock/booking/management actions validate administrator identity
  through the admin service; customer tokens, viewer accounts and inactive admins
  cannot use those operations. Stock/product browsing remains public.
- Configured separate admin/customer signing secrets with strict JWT algorithm,
  expiry and identity checks; random session token IDs prevent same-second collisions.
- Cryptographic OTP generation, HMAC storage, atomic single-use verification and a
  five-attempt limit; concurrent-safe bounded local API rate limiting.
- Fresh-install admin schema without baked-in admin credentials and an explicit
  development admin bootstrap command.
- Standalone Compose Go test runner with race detection for an available Docker host.

## Checks actually executed

- `go test ./...`: passed in user, admin, gateway, inventory and email services.
- `go vet ./...`: passed in all five services.
- Linux amd64 builds with CGO disabled: passed for all five service binaries.
- Four Python/Compose configuration tests: passed. They cover private service ports,
  loopback development entry points, missing production settings, production port
  overrides, mail mode, disabled production demo bootstrap and removed .env files.
- Standalone test Compose configuration: validated.
- Git whitespace/error check: passed.

Regressions include account-ID ownership (mocked SQL), anonymous/admin inventory
permissions (HTTP identity test server), JWT expiry/algorithm/dummy-key rejection,
OTP expiration/attempt exhaustion/concurrent consumption (miniredis), concurrent
rate-limit accounting and a local SMTP transaction containing an OTP email.

Toolchain used locally: Go 1.27.2 on Windows; Docker Compose standalone CLI 5.6.0
for configuration inspection only. The local test doubles do not prove real
PostgreSQL/Valkey integration or real email-provider behavior.

## Not yet verified or implemented

This is the first foundation increment, not the completed app or a production release.
There is no usable Docker engine in this environment. Container builds/startup,
PostgreSQL initialization, real Valkey integration, the Linux race-detector runner,
Nginx/Caddy behavior and browser/device journeys have NOT been executed.

Remaining work includes existing-database migrations and constraint reconciliation;
service readiness/recovery and trusted-proxy/distributed rate-limit policies; admin
session revocation and complete role checks throughout the admin service; product/
plan/payment/delivery contracts; atomic dated reservations; page-by-page wiring;
compiled frontend assets; meat/cart/credits workflows; provider adapters and complete
integration/end-to-end/load/recovery checks.

The old payment page still simulates success and other existing screens contain
sample/localStorage behavior. No real payment capability is introduced here. Do not
use this increment to take real payments or publicly launch the app.

Previously exposed upstream credentials require independent revocation. Removing
them from this branch does not invalidate them or remove Git history.

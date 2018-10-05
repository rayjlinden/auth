## v0.3.0 (Released 2018-10-05)

ADDITIONS

- Add `oauth2_client_generations` metric
- Added generic `OPTIONS` handler for CORS pre-flight

IMPROVEMENTS

- Write `X-User-Id` header on `GET /users/login`.
- Validate phone numbers on signup.
- Support `X-Request-Id` for request debugging/tracing.
- Add HTTP method and path to tracing logs
- Return CORS headers if `Origin` is sent.
- Render the `User` object on `GET /users/login`

BUG FIXES

- OAuth2 client credentials response was different than docs

## v0.2.0 (Released 2018-09-26)

IMPROVEMENTS

- Added `/ping` route.

BUG FIXES

- OAuth2 routes should have been prefixed as `/oauth2/`.
- admin: fix pprof profiles (not all work)

## v0.1.0 (Unreleased)

INITIAL RELEASE

- HTTP Server with oauth, user/pass signup (and auth)
- Prometheus metrics and pprof setup

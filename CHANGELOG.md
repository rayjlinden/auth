## v0.5.0 (Released 2019-01-02)

CHANGES

- Removed buntdb in favor of sqlite for OAuth2 client and token storage.

BUG FIXES

- Switch to github.com/moov-io/base's `Time` type.

IMPROVEMENTS

- Better error handling in `main()`.

## v0.4.3 (Released 2018-12-13)

BUG FIXES

- oauth: Send back 403 HTTP status code on invalid tokens. (See: [#61](https://github.com/moov-io/auth/pull/61))

## v0.4.2 (Released 2018-12-11)

BUG FIXES

- oauth: properly associate token with user (See: [#59](https://github.com/moov-io/auth/pull/59))

## v0.4.1 (Released 2018-12-04)

BUG FIXES

- Fixed docker image to boot properly.
- Send down `application/json` content-type on OAuth authorize endpoint
- Fixed `oauth2_token_generations` Prometheus metric

## v0.4.0 (Released 2018-11-29)

ADDITIONS

- Added `GET /auth/check` endpoint that looks at HTTP Cookies and OAuth
- Added `PATCH /users/{userId}` for updating user profile information

BUG FIXES

- Check database `row.Scan` errors

IMPROVEMENTS

- Run as unprivileged user inside Docker container
- Return an empty JSON object (and content-type) for generated clients. (See [go-client #5](https://github.com/moov-io/go-client/issues/5))

## v0.3.1 (Released 2018-10-05)

BUG FIXES

- Don't trample over Content-Type values when writing CORS headers
- Allow all our HTTP Methods with CORS requests

IMPROVEMENTS

- Respond with CORS headers for forward auth calls triggered from preflight requests.

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

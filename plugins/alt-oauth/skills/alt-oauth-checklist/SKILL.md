---
name: alt-oauth-checklist
description: Role-based OAuth 2.1, OpenID Connect Core, RFC 9700 and MCP authorization checklist with a 0-10 scoring rubric and severity levels. Covers clients (web sign-in, CLI and native apps with loopback redirects and refresh, server-to-server), resource servers and MCP servers (RFC 9728 metadata, access-token validation, WWW-Authenticate), and a short authorization-server section. Use when reviewing or building sign-in, token handling, an OAuth client, API or MCP auth, or one app calling another over OAuth.
---

# alt-oauth-checklist

Score each in-scope item 0-10 with `file:line` evidence, or a negative search for something absent. Cite the section anchor from [references/rfc-index.md](references/rfc-index.md), never a bare RFC URL.

## Before reading code

1. **Role.** Client (web, CLI or native, server-to-server), resource server (API or MCP server), or authorization server. A repo can hold several roles; review each separately.
2. **Target.** OAuth 2.1 + OIDC Core 1.0 + RFC 9700 by default. Add the MCP authorization spec when an MCP client or server is in scope.
3. **Profile.** If the repo has its own profile (`docs/oauth-profile.md`, or a file the brief names), it narrows this list: score what it requires, skip what it defers, and flag anything it lists as never or deferred that is live or advertised.

Each item reads: requirement (spec section) then **Verify:** how to check it in code.

## Client: all kinds

- PKCE `S256` on every authorization request; verifier 43-128 chars from a crypto-random source (RFC 7636 Section 4.1, 4.2). **Verify:** grep `code_challenge_method`; no `plain` path; verifier built from `crypto/rand` or the platform CSPRNG.
- `state` crypto-random, single use, bound to the browser session or the CLI's listener, checked before anything else on callback (RFC 6749 Section 10.12). **Verify:** callback handler compares and deletes `state` before reading `code`.
- Issuer check: `iss` in the authorization response equals the recorded issuer by exact string compare, before the code reaches any token endpoint; reject when the AS advertises `authorization_response_iss_parameter_supported` and `iss` is missing; on mismatch ignore `error` too (RFC 9207 Section 2.4, RFC 9700 Section 4.4.2.1). **Verify:** callback reads `iss`; test covers wrong and missing `iss`.
- Mix-up defence with several authorization servers: record which AS each flow started with; keep client credentials and tokens per issuer (OAuth 2.1 Section 7.15, MCP Authorization server binding). **Verify:** flow state stores the issuer; token store is keyed by issuer.
- Discovery: metadata `issuer` identical to the configured issuer, no trailing-slash or case drift (RFC 8414 Section 3.3, OIDC Discovery Section 4.3). **Verify:** discovery loader rejects a mismatch; config is not normalized.
- `resource` sent on both the authorization and the token request when calling an API or MCP server (RFC 8707 Section 2.1, 2.2; MCP Resource parameter implementation). **Verify:** both request builders set `resource` to the canonical server URI.
- ID-token validation: signature with keys from `jwks_uri` and an alg from a pinned list, never `none` and never the header alone; `iss` exact; `aud` contains `client_id` (extra audiences need `azp`); `exp` and `iat` with skew of 5 minutes or less; `nonce` equals the one sent (OIDC Core Section 3.1.3.7). **Verify:** read the verifier; tests for each rejection.
- User identity keyed by `iss` + `sub`, never by email alone (OIDC Core Section 2). **Verify:** account lookup and linking queries.
- Refresh: rotate on every use and store the new token atomically; on `invalid_grant` drop both tokens and sign in again (RFC 6749 Section 6, OAuth 2.1 Section 4.3). **Verify:** concurrent refreshes are serialized or tolerate a single winner.
- Revocation on sign-out or disconnect: revoke the refresh token at the AS, then delete it locally (RFC 7009 Section 2.1). **Verify:** logout path calls the revocation endpoint when advertised.
- JWKS and discovery cached; JWKS re-fetched on an unknown `kid`, rate-limited. **Verify:** cache code and its refetch guard.
- No tokens, codes, verifiers or secrets in logs, errors, URLs or analytics. **Verify:** grep log calls near token handling.

## Client: web (server-side sign-in)

- `redirect_uri` registered exactly, HTTPS outside loopback (RFC 9700 Section 4.1.3). **Verify:** config validation.
- Tokens stay server-side or in `HttpOnly`, `Secure`, `SameSite` cookies; never `localStorage` or URL fragments. **Verify:** cookie flags; no token in a rendered page.
- Post-login `return_to` checked against same origin to avoid open redirects (RFC 9700 Section 4.11.2). **Verify:** the redirect helper and its tests.
- Session fixed after sign-in: a new session id once the user is authenticated. **Verify:** session rotate call on the callback.

## Client: CLI and native

- Loopback redirect on `127.0.0.1` or `[::1]` with an OS-assigned port; not `localhost` (RFC 8252 Section 7.3, 8.3). **Verify:** listener binds the IP literal, port 0.
- Public client: `token_endpoint_auth_method=none`, no shipped secret (RFC 8252 Section 8.5). **Verify:** grep the binary's config for a client secret.
- Listener accepts one callback, then closes; has a timeout. **Verify:** listener lifecycle.
- Tokens stored with user-only permissions (OS keychain, or a file at mode 0600). **Verify:** file mode on write.

## Client: server-to-server

- `client_credentials` only for confidential clients, with `resource` set; no refresh token expected (OAuth 2.1 Section 4.2, RFC 6749 Section 4.4.3). **Verify:** grant builder.
- Client secret from config or a secret store, never committed; compared or sent only over TLS. **Verify:** config tags and grep.
- An app calling another app never forwards the inbound user token; it gets its own token for the target audience (MCP Access token privilege restriction). **Verify:** outbound HTTP clients set their own `Authorization`.

## Resource server and MCP server

- Protected-resource metadata at `/.well-known/oauth-protected-resource<path>`: `resource` equals the URI clients use, `authorization_servers` names the issuer(s) (RFC 9728 Section 2, 3; MCP Authorization server location). **Verify:** handler output against the configured public URL.
- 401 carries `WWW-Authenticate: Bearer resource_metadata="..."` and SHOULD carry `scope` (RFC 9728 Section 5.1); insufficient scope is 403 `error="insufficient_scope"` naming every needed scope (RFC 6750 Section 3.1; MCP Runtime insufficient scope errors). **Verify:** middleware tests assert both headers.
- JWT access-token validation (RFC 9068 Section 4): `typ` is `at+jwt` or `application/at+jwt`, so an ID token is refused; `iss` exact; `aud` contains this resource exactly; signature with AS keys and a pinned alg list, never `none`; `exp` with small skew; any failure is 401 `invalid_token`. **Verify:** one test per rejection.
- Opaque tokens checked by introspection: `active`, `aud` and scope (RFC 7662 Section 2.2). **Verify:** introspection client and cache bound.
- Token read only from the `Authorization` header, never the query string (RFC 6750 Section 2). **Verify:** grep `access_token` in query parsing.
- Scopes enforced per operation or MCP tool, at one seam. **Verify:** map of route or tool to scope; a test for a missing scope.
- No token passthrough to upstream APIs (MCP Token requirements). **Verify:** outbound calls.

## Authorization server (short)

A repo that is the authorization server should keep its own deeper review. Minimum checks here:

- `response_type=code` only; no implicit, hybrid or `password` grant (OAuth 2.1 Section 10, RFC 9700 Section 2.4).
- PKCE `S256` required for every client; constant-time verifier check (RFC 7636 Section 4.6).
- Exact redirect-URI match; loopback matches on any port (RFC 8252 Section 7.3).
- Codes single use, short-lived, bound to client, redirect URI and challenge; replay revokes what the code issued (OAuth 2.1 Section 4.1.3).
- `iss` on every authorization response, success and error (RFC 9207 Section 2).
- `resource` honored; token `aud` equals it exactly (RFC 8707 Section 2).
- Refresh rotation with family revocation on reuse (RFC 9700 Section 4.14).
- Metadata advertises only working features; `code_challenge_methods_supported` in every metadata document (RFC 8414 Section 2).

## Scoring rubric

| Score | Meaning |
|---|---|
| 10 | Spec-clean |
| 8-9 | Right, minor gaps |
| 5-7 | Partial; a real spec deviation |
| 1-4 | Attempted but broken, or a key protection missing |
| 0 | Absent where the spec or profile requires it |

Section score is the mean of its in-scope items; overall is the mean of sections. Report both: one bad row outweighs a good mean.

## Severity

- **Sev 1:** spec violation or security hole. Ship nothing until fixed.
- **Sev 2:** gap against mainstream libraries and providers, such as no rate limit, no revocation on sign-out, no JWKS refetch on `kid` miss.
- **Sev 3:** high-assurance extras: DPoP, PAR, JAR, mTLS, back-channel logout.

Always Sev 1:

- An implicit, hybrid or `password` path that is live.
- `alg=none` accepted, or the alg taken from the token header alone.
- `plain` PKCE accepted, or PKCE optional for any client.
- Wildcard, prefix or normalized redirect-URI matching, or any open redirect.
- `state` or `iss` not checked on the callback.
- An access token accepted without an exact `iss` and `aud` check, or an ID token accepted as an access token.
- A token sent to or accepted from the wrong issuer or audience; a user token passed through to another service.
- A token, code, verifier or secret in a log, an error or a URL.
- Metadata advertising a capability with no working code path.
- An SSRF-able fetch of metadata or a client document.

Unbuilt features are not findings unless something shipped is unsafe without them.

# Modern Authentication & IAM — Combined Feature Checklist

> ✓ A checked box = included in the **citadel** auth MVP. Everything unchecked is deferred post-MVP.

A comprehensive, vendor-agnostic inventory of features found across modern authentication and identity & access management (IAM) systems, aggregated from leading 2026 platforms (Auth0, Okta, Keycloak, Zitadel, Authentik, AWS Cognito, Microsoft Entra ID, Clerk, WorkOS, Descope, Ory, and others). Use it as a reference spec for evaluating or building an authentication system.

---

## 1. Authentication Methods

- [x] Password login — classic email/username plus password sign-in
- [x] User registration — self-service account sign-up
- [ ] Magic link login — passwordless sign-in via a one-time email link
- [ ] Email OTP login — passwordless sign-in with a code sent by email
- [ ] SMS OTP login — passwordless sign-in with a code sent by text
- [ ] Passkeys / WebAuthn / FIDO2 — phishing-resistant public-key login
- [ ] Hardware security keys — roaming FIDO2 authenticators such as YubiKey
- [ ] Platform biometrics — Touch ID / Face ID / Windows Hello via WebAuthn
- [ ] Google login — sign in with a Google account
- [ ] GitHub login — sign in with a GitHub account
- [ ] Apple login — Sign in with Apple
- [ ] Microsoft login — sign in with Microsoft or personal accounts
- [ ] Other social logins — Facebook, X, LinkedIn, and similar providers
- [ ] Generic OIDC provider — configurable upstream OpenID Connect login
- [ ] Generic OAuth 2.0 provider — configurable upstream OAuth login
- [ ] Enterprise SSO via SAML — act as a SAML service provider / relying party
- [ ] Enterprise SSO via OIDC federation — delegate auth to an upstream IdP
- [ ] LDAP / Active Directory login — authenticate against a directory
- [ ] Certificate / mTLS login — client-certificate authentication
- [ ] QR-code login — scan to sign in from another device
- [ ] Anonymous / guest sessions — temporary identities, upgradeable to full accounts

## 2. Multi-Factor Authentication (MFA)

- [ ] TOTP authenticator app — time-based one-time codes (RFC 6238)
- [ ] Recovery / backup codes — single-use fallback codes
- [ ] Email OTP factor — second-factor code delivered by email
- [ ] SMS OTP factor — second-factor code delivered by text
- [ ] Push MFA — approve a login prompt on a trusted device
- [ ] WebAuthn second factor — security key or passkey as a second step
- [ ] Step-up authentication — re-prompt for sensitive actions (ACR/AMR)
- [ ] Adaptive / risk-based MFA — challenge only when risk is elevated
- [ ] MFA enrollment management — users register and manage their factors
- [ ] MFA enforcement policies — require MFA per organization or per policy
- [ ] Trusted-device remembering — skip MFA on recognized devices
- [ ] Factor fallback chains — priority order across multiple factors

## 3. Authorization & Access Control

- [x] Role-based access control (RBAC) — assign roles and permissions
- [ ] Hierarchical roles — role inheritance and composite roles
- [x] Scopes & permissions in tokens — embed granted access in issued tokens
- [ ] Attribute-based access control (ABAC) — decisions from user/resource attributes
- [ ] Relationship-based access control (ReBAC) — Zanzibar-style relationship graphs
- [ ] Policy-based access control (PBAC) — external policy engine such as OPA or Cedar
- [ ] Fine-grained authorization — resource- and row-level access decisions
- [ ] Delegated administration — org admins manage their own users and roles
- [ ] Authorization decision API — a check-access endpoint for applications

## 4. Protocols & Standards

- [ ] OAuth 2.0 authorization server — issue and manage access tokens
- [ ] OAuth 2.1 compliance — mandatory PKCE, no implicit/ROPC, rotating refresh
- [ ] OpenID Connect provider — discovery, JWKS, and userinfo endpoints
- [ ] Authorization Code + PKCE — the recommended browser/app flow (RFC 7636)
- [ ] Client Credentials grant — machine-to-machine authentication
- [ ] Refresh Token grant — obtain new tokens without re-login
- [ ] Device Authorization grant — login for TVs, CLIs, and IoT (RFC 8628)
- [ ] Token Exchange — delegation and impersonation flows (RFC 8693)
- [ ] CIBA — client-initiated backchannel authentication for decoupled devices
- [ ] Pushed Authorization Requests (PAR) — pre-register requests server-side (RFC 9126)
- [ ] DPoP — sender-constrained, proof-of-possession tokens (RFC 9449)
- [ ] JAR / JARM — signed authorization request objects and responses
- [ ] mTLS-bound tokens — certificate-bound access tokens (RFC 8705)
- [ ] FAPI 2.0 profile — high-assurance security for finance and open banking
- [ ] Dynamic Client Registration — programmatic client onboarding (RFC 7591)
- [ ] Client ID Metadata Document (CIMD) — client metadata for MCP / AI agents
- [ ] SAML 2.0 Identity Provider — issue SAML assertions to apps
- [ ] SAML 2.0 Service Provider — consume assertions from upstream IdPs
- [ ] WS-Federation — legacy enterprise federation protocol
- [ ] LDAP / Active Directory endpoint — expose a directory protocol interface
- [ ] Kerberos / SPNEGO — desktop single sign-on in enterprise networks

## 5. Session & Token Management

- [ ] Server-side sessions — secure cookies (HttpOnly, SameSite, Secure)
- [x] JWT access tokens — signed, verifiable tokens (e.g. RS256)
- [x] Refresh tokens — long-lived tokens to renew access
- [x] Refresh-token rotation — rotate and detect reuse of refresh tokens
- [ ] Opaque tokens with introspection — reference tokens validated server-side
- [ ] Token introspection endpoint — check token validity (RFC 7662)
- [ ] Token revocation endpoint — invalidate issued tokens (RFC 7009)
- [x] Configurable token lifetimes — tune expiry and policies per client
- [ ] Single sign-on (SSO) — silent re-auth from an existing session
- [ ] RP-initiated logout — end the session from the relying party
- [ ] Back-channel logout — server-to-server session termination
- [ ] Front-channel logout — browser-driven multi-app logout
- [ ] Active session listing — show a user's signed-in devices
- [ ] Session revocation — sign out a specific session or everywhere
- [ ] Concurrent-session limits — cap simultaneous sessions per user
- [ ] Remember-me sessions — opt-in long-lived sessions
- [ ] Session timeouts — idle and absolute expiration controls

## 6. User Management

- [x] Email verification — confirm ownership of an email address
- [ ] Phone verification — confirm ownership of a phone number
- [x] Password reset — self-service forgot-password flow
- [ ] Profile management — edit name, avatar, and attributes
- [ ] Account recovery — regain access after lost credentials or MFA
- [ ] Identity / account linking — multiple login methods on one account
- [ ] Change email / password — update credentials with re-verification
- [ ] User invitations — admins invite users to join
- [ ] Bulk import / export — onboard or extract users in bulk
- [ ] User migration — lazy password rehash from a legacy store
- [ ] Custom user attributes — public and private metadata
- [ ] Progressive profiling — collect user attributes over time
- [ ] Account deactivation — suspend or disable accounts
- [ ] Account self-deletion — GDPR right-to-erasure flow
- [ ] User search & filtering — admin lookup across users

## 7. Multi-Tenancy & Organizations

- [x] Organizations / tenants — first-class multi-tenant entities
- [x] Tenant data isolation — row-scoping or per-tenant schema/database
- [x] Organization membership — users and roles scoped within an org
- [ ] B2B teams / groups — sub-groups inside an organization
- [ ] Organization invitations — invite members into an org
- [ ] Per-org policies — MFA, password, and session rules per organization
- [ ] Per-org branding — themes and logos per organization
- [ ] Per-tenant SSO — each customer connects their own IdP
- [ ] Domain-based routing — map an email domain to an organization
- [ ] Realm / project separation — full multi-environment isolation

## 8. Security & Threat Protection

- [ ] Brute-force protection — account lockout after repeated failures
- [x] Rate limiting — throttle per IP, account, and endpoint
- [ ] CSRF protection — guard state-changing requests
- [x] Security headers — CSP, X-Frame-Options, HSTS, and nosniff
- [x] Encryption at rest — protect secrets and sensitive data
- [x] Audit logging — record security-relevant auth events
- [x] Enumeration protection — uniform responses and timing
- [ ] Password policies — length, complexity, history, and rotation rules
- [ ] Breached-password detection — check against HaveIBeenPwned (k-anonymity)
- [ ] Bot detection / CAPTCHA — challenge automated abuse
- [x] Signing-key rotation — rotate keys with a multi-key JWKS
- [ ] Anomaly detection — flag suspicious activity
- [ ] Impossible-travel checks — geo-velocity risk signals
- [ ] IP allow / deny lists — restrict access by network
- [ ] Device fingerprinting — recognize and score devices
- [ ] New-device alerts — notify users of new logins or locations
- [ ] Risk scoring engine — quantify login risk for adaptive decisions

## 9. Admin & Management

- [ ] Management API — CRUD for orgs, users, apps, and roles
- [ ] Admin console — dashboard UI for operators
- [ ] Application / client registration — onboard OAuth/OIDC clients
- [ ] Client types — public (PKCE) and confidential clients
- [ ] Client-secret rotation — rotate credentials without downtime
- [ ] Machine-to-machine clients — client-credentials applications
- [ ] API keys — issue, scope, and revoke service keys
- [ ] Service accounts — non-human identities with scoped access
- [ ] Webhooks / events — stream events to external systems
- [ ] Audit-log viewer — browse and export the audit trail
- [ ] Analytics & reporting — sign-ups, logins, and MFA adoption
- [ ] User & role management UI — manage identities and access visually

## 10. Branding & UX

- [ ] Hosted login pages — server-rendered login and registration screens
- [ ] Themeable login UI — customize colors, logo, and CSS
- [ ] Customizable email templates — branded verification and reset emails
- [ ] SMS template customization — branded text-message content
- [ ] Custom domains — vanity issuer and login URLs
- [ ] Localization (i18n) — interfaces and messages in multiple languages
- [x] Embedded / headless mode — API-first with a bring-your-own UI
- [ ] UI components / widgets — drop-in login components and SDK widgets
- [ ] Custom login flows — configurable orchestration of auth steps
- [ ] Consent screen customization — brand and tailor the consent prompt

## 11. Developer Experience

- [x] REST / JSON API — clean programmatic access to all functions
- [ ] Server-side SDKs — official libraries with example apps
- [ ] Frontend SDKs — JS/TS, mobile, and SPA libraries
- [ ] gRPC / GraphQL API — alternative typed or query-based interfaces
- [ ] Admin CLI — command-line administration
- [ ] Terraform / IaC provider — manage configuration as code
- [ ] Documentation & quickstarts — guides and integration walkthroughs
- [ ] Actions / hooks — run custom code at points in the auth flow
- [ ] Plugin system — extend the platform with modules
- [ ] Local dev mode — sandbox with seed data for development

## 12. Identity Federation & Provisioning

- [ ] SCIM inbound provisioning — apps push users into the system
- [ ] SCIM outbound provisioning — push users to downstream apps
- [ ] Just-in-time provisioning — create users on first SSO login
- [ ] Directory sync — sync from LDAP/AD, Google Workspace, or Entra
- [ ] Identity brokering — chain and federate multiple upstream IdPs
- [ ] Claim / attribute mapping — transform claims between systems

## 13. Machine & Workload Identity

- [ ] Client-credentials grant — service-to-service authentication
- [ ] API keys for services — simple programmatic credentials
- [ ] Service accounts — scoped non-human identities
- [ ] mTLS workload auth — certificate-based service authentication
- [ ] Workload identity federation — cloud and Kubernetes workload trust
- [ ] AI-agent identity — MCP authorization server and CIMD support

## 14. Compliance & Privacy

- [ ] Audit trail — immutable record of security-relevant events
- [ ] GDPR consent management — capture and track user consent
- [ ] Data export — user data portability
- [ ] Right to erasure — delete a user's account and data
- [ ] Terms acceptance — versioned ToS and privacy-policy consent
- [ ] Data-retention policies — configurable retention and purging
- [ ] PII encryption — field-level encryption of personal data
- [ ] Data residency — region pinning for stored data
- [ ] SOC 2 / ISO controls — features supporting formal attestations
- [ ] Consent / grant history — per-user record of granted scopes

## 15. Notifications & Communication

- [x] Transactional email — verification, reset, and OTP messages
- [x] Pluggable SMTP — bring your own mail server
- [ ] Email provider integrations — SendGrid, SES, Postmark, and similar
- [ ] SMS delivery — Twilio, Vonage, and other gateways
- [ ] Push notifications — for push MFA and security alerts
- [ ] Per-event notification rules — configure messages per event type

## 16. Deployment & Operations

- [x] Single Docker image — self-contained, plug-and-play container
- [x] docker-compose stack — app and database with one command
- [x] Environment-variable config — twelve-factor configuration
- [x] Database migrations — schema migrations run on startup
- [x] Health endpoints — liveness and readiness probes
- [x] PostgreSQL support — primary relational datastore
- [ ] Additional databases — MySQL, SQLite, or CockroachDB
- [ ] Kubernetes / Helm — manifests and charts for orchestration
- [ ] Stateless scaling — horizontally scalable, share-nothing design
- [ ] High availability — clustering and failover
- [x] Structured logging — machine-parseable log output
- [ ] Metrics — Prometheus-compatible instrumentation
- [ ] Distributed tracing — OpenTelemetry spans across requests
- [ ] Caching layer — Redis or in-memory acceleration
- [ ] Backup & restore — data backup tooling and guidance
- [ ] Zero-downtime upgrades — rolling deployments and migrations

## 17. Advanced & Emerging

- [ ] Continuous authentication — ongoing session risk evaluation
- [ ] Behavioral biometrics — typing and usage-pattern signals
- [ ] Fraud / account-takeover detection — detect and block ATO attempts
- [ ] Liveness / deepfake detection — anti-spoofing for biometric onboarding
- [ ] Decentralized identity — DIDs and verifiable credentials
- [ ] User impersonation — audited "view as user" for support
- [ ] On-behalf-of flows — delegated access between services
- [ ] OAuth 2.1 migration tooling — compatibility mode and upgrade helpers
- [ ] Passkey-first onboarding — passwordless by default with OTP fallback
- [ ] Device trust — managed-device attestation and posture checks

---

## Sources

- [Authentication Trends in 2026: Passkeys, OAuth3, and WebAuthn — C# Corner](https://www.c-sharpcorner.com/article/authentication-trends-in-2026-passkeys-oauth3-and-webauthn/)
- [Passwordless & MFA in 2026: Passkeys, Push MFA, Device Trust — LoginRadius](https://www.loginradius.com/blog/identity/passwordless-and-mfa)
- [Best CIAM Solutions 2026: Passwordless & AI Compared — Corbado](https://www.corbado.com/blog/best-ciam-solutions)
- [Passwordless Authentication Trends: Where We're Headed — Descope](https://www.descope.com/blog/post/passwordless-authentication-trends)
- [What Is CIAM? Complete Guide to Customer Identity 2026 — Gupta Deepak](https://guptadeepak.com/what-is-ciam-a-complete-guide-to-customer-identity-and-access-management-in-2026/)
- [Keycloak vs Okta vs Auth0 vs Authelia vs Cognito vs Authentik — Ritza](https://ritza.co/articles/gen-articles/keycloak-vs-okta-vs-auth0-vs-authelia-vs-cognito-vs-authentik/)
- [Zitadel vs Keycloak — ZITADEL](https://zitadel.com/blog/zitadel-vs-keycloak)
- [Keycloak vs Zitadel: Open-Source IAM Compared — Skycloak](https://skycloak.io/blog/keycloak-vs-zitadel-comparison/)
- [Top Open-Source Auth0 Alternatives in 2026 — Authgear](https://www.authgear.com/post/top-open-source-auth0-alternatives/)
- [OAuth 2.1 Explained: What Changed and Why It Matters — Gupta Deepak](https://guptadeepak.com/ciam-compass/guides/oauth-2-1-explained/)
- [The Many Faces of OAuth 2.0 Token Exchange — Auth0](https://auth0.com/blog/the-many-faces-of-oauth2-token-exchange/)
- [SAML vs OIDC vs OAuth 2.0: 12 Differences — Security Boulevard](https://securityboulevard.com/2026/04/saml-vs-oidc-vs-oauth-2-0-12-differences-every-b2b-engineering-team-should-know/)
- [Keycloak 26.6.0 released — Keycloak](https://www.keycloak.org/2026/04/keycloak-2660-released)

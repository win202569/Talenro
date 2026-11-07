# Windows client — selected direction and four-screen concept draft

Date: 2026-11-07. User selected the first displayed concept (light native,
teal accent, connection-first) and requested expansion to the four core pages.
Only the visual direction is selected; these expanded screens and interaction
proposals await review. No implementation plan or runtime permission follows.

## Screens

- [Connection — selected reference](connection.png)
- [Login — new concept](login.png)
- [Nodes — new concept](nodes.png)
- [Account and plan — new concept](account.png)

All are AI-generated static images, not screenshots of working WinUI code.
Copied originals were SHA256-compared without altering the images. Concept target
was 1440x1024; generated raster size is not a final Windows layout specification.
Logo geometry is provisional generated artwork, not a newly approved brand asset.

## Existing requirements retained

Source: [commercial client specification](../../../superpowers/specs/2026-08-04-vpn-commercial-client-design.md),
[brand](../../../superpowers/specs/2026-08-04-talenro-brand-naming-design.md),
[localization](../../../superpowers/specs/2026-08-04-vpn-localization-design.md).

Windows remains C#/WinUI 3 with a separate service for future tunnel operations;
these images neither choose a web implementation nor create that service.
Desktop navigation retains connection, nodes, traffic, devices, account/plan and
settings. Only four core screens are expanded, not all navigation destinations.
Email/password is primary; no social or SMS login is introduced. Account-level
usage and activated-device counts are distinct from per-device usage and online
counts. Country lock defaults off and must never silently cross countries when on.
Five-language/RTL support remains a requirement; this concept set depicts zh-Hans
only and does not establish localization, accessibility or contrast acceptance.

## Proposed interaction contract — for review

|Page|Primary flow|Required alternate states before a later prototype|
|---|---|---|
|Login|Email/password, show/hide password, sign in; links to reset/create account|Empty/invalid input, submitting, generic rejected credentials, unavailable service; conditional MFA only when required. Demo must not collect or send real credentials|
|Connection|Choose automatic/manual preference then explicitly connect; open nodes from Change|Disconnected, connecting with cancellation, simulated connected visibly labelled, disconnecting, failed/unavailable, quota exhausted hard stop. Never label an unverified real connection as protected|
|Nodes|Search grouped country list, select preference, confirm Use this node, return to connection|Empty search, loading/unavailable list, disabled node, automatic mode. Selection is not an immediate connection command. Manual cross-country selection while lock is on updates the lock country with clear feedback|
|Account/plan|Show account usage and activated devices; navigate to account-security actions|Expired/exhausted entitlement, device limit reached, unavailable/stale data. Account login does not imply VPN credential admission. Delete/revoke actions require explicit confirmation and reauthentication in the future real flow|

If every eligible node in the locked country fails, offer exactly: one-time
cross-country recovery, retry current country, or stay disconnected. One-time
permission does not permanently turn off the country lock. While already connected,
the exact apply-node/change-country confirmation behavior needs definition before
implementation; the depicted node screen is deliberately disconnected.

Password reset, account creation, MFA, device management and deletion are secondary
flows, not silently added pages in this four-screen scope. The future prototype
must disclose unimplemented destinations instead of claiming successful actions.

## Mock-data and visual-review limits

`demo@example.com`, Plus, 12.4/100 GB, 2/5 devices and all locations/quality labels
are illustrative, not pricing, final entitlements, real nodes or live health.
No payments, credentials, network calls, drivers, tunnels or account changes occur.
The connection page's apparent power icon plus button should become one logical
focus/action target or an explicitly decorative icon, not two ambiguous actions.
Node list's unlabeled automatic-protocol column needs an accessible heading in
the next detailed pass. Normalize 账号/账户 terminology and button wording across
screens; do not treat generated glyphs or exact pixel spacing as final copy/tokens.
Country-lock failure dialog, loading/error states, keyboard order, resizing/DPI,
screen-reader names and Persian RTL remain unrendered/unverified.

## Handoff

Next: review these three expanded pages alongside the selected connection page.
After feedback, settle the interaction states and write the bounded design before
any implementation. No code, service, installation, test or build was performed.
Task11/12 and the prior security/runtime blocks remain unchanged;44/34/6/217,
NOT_ACCEPTED. This checkpoint preserves design progress only.

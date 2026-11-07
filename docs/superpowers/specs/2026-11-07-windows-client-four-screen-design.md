# Windows client four-screen interaction design

Date: 2026-11-07. Status: four-screen visuals and written interaction design
approved by user after86dfc037, for implementation planning only. No implementation
or runtime authority. Plan: ../plans/2026-11-07-windows-client-offline-demo.md.

## Intent and scope

Make Talenro's Windows consumer experience understandable before real integration:
sign in, choose a location, explicitly connect, and understand account usage.
The next proposed deliverable is a native C#/WinUI 3 **offline demonstration**
of four screens, not a production VPN client. This is an author proposal, not an
already authorized build. Do not create services, TUN interfaces, routes, IPC,
real authentication, billing, telemetry or network activity for this slice.

Existing authorities: [commercial client](2026-08-04-vpn-commercial-client-design.md),
[brand](2026-08-04-talenro-brand-naming-design.md),
[localization](2026-08-04-vpn-localization-design.md).
Task11/12 and all prior runtime/security gates remain unchanged. Any future build,
launch, test or dependency preparation requires its own bounded approved plan.

## Approaches and selected visual target

Three home concepts were shown: light native/teal with centered connection action;
dark with left-aligned action; light split-view with inline locations. The first
prioritizes one obvious action and calm hierarchy; the second emphasizes contrast;
the third reduces location-navigation steps but adds home-screen density. The user
selected the first and subsequently confirmed its three expanded screens.

The [four saved images](../../design/windows-client/2026-11-07/README.md) are the
visual target, not exact raster-layout requirements. Preserve the light base,
teal action, native title bar, restrained sidebar and generous spacing. Do not
redesign into a web dashboard, add map/marketing panels or real-looking live data.
The generated logo remains a provisional design asset, not final brand approval.

## Navigation and demonstration boundary

Login is outside the authenticated sidebar. The demonstration starts there with
fixed fictional account data and a permanent notice: “离线演示，不会连接 VPN 或提交账号信息”.
Use fixed demo@example.com and a read-only masked password representation; the
demonstration must not request, retain or submit real passwords. The login action
is labelled “进入演示”; production wording “登录” in the visual is for future real
integration. Demo failures are selected scenarios, not genuine account checks.

After demo entry, open Connection. Sidebar keeps Connection, Nodes, Traffic,
Devices, Account and Plan, Settings. Only Connection, Nodes and Account are pages
in this slice; other destinations open a plain “此功能不在本次演示范围内” explanation.
Reset password, registration, MFA, language changes, device management and account
deletion likewise disclose nonimplementation; do not display false success.
No external links open and no actions modify a real account.

All state is in memory, reset when the demo closes. Exiting the current demo
device returns to Login and resets simulated connection state. No credentials,
email values, usage or preferences are logged or persisted. A visibly separate
“演示场景” selector deterministically exposes alternate states; no arbitrary
timeouts, network requests or background workers are needed to simulate failures.

## Four-screen behavior

|Screen|Normal path|Alternate states and safeguards|
|---|---|---|
|Login|Read fixed demonstration fields; enter demo → Connection|Scenario selector shows invalid input, submitting, generic credential rejection, service unavailable and MFA-required explanation. None verifies identity; error copy never reveals whether a real account exists|
|Connection|Disconnected → explicitly Connect → simulated Connecting → simulated Connected → explicitly Disconnect → simulated Disconnecting → Disconnected|Connecting offers Cancel → Disconnected. Completion is an explicit demo scenario event, not evidence of a tunnel. Label every active state “模拟”; no real IP/speed or “已受保护” claim. Failure offers Retry/choose location, never an automatic endless loop|
|Nodes|Search country/node, choose Auto or a preferred node, Apply → return to Connection|Search empty, loading, list unavailable and disabled-node scenarios. Apply updates preference only while disconnected; disabled rows cannot be applied. Back discards uncommitted selection. Selection is separate from connection status|
|Account and Plan|Display demo account usage, example plan and activated-device count; exit demo|Expired plan/quota exhausted block simulated Connect; unavailable/stale entitlement is unknown, not zero or active. Device-limit scenario allows account view but blocks VPN admission. No purchase/upgrade/payment controls|

For a connected demo, applying another node requires a confirmation explaining
the simulated connection will restart. Cancel leaves old state/preferences intact;
confirm atomically commits the preference and enters simulated Connecting. If
country lock is on and the chosen country differs, the same confirmation names
both countries and the new lock country. While connecting/disconnecting, editing
or applying nodes is unavailable with an explanation; cancellation remains usable.

Country lock defaults off. With no known current/selected country, enabling it
requires choosing a country first; never invent a current exit country. In the
demo, lock identity comes from explicitly labelled fixture data. If every eligible
candidate in the locked country fails, offer exactly: once-only cross-country
recovery, retry this country, or stay disconnected. The first permits this one
simulated recovery only and preserves the saved lock preference; retry remains
in-country. No silent cross-country fallback. Normal auto/manual mode does not
override eligibility or quota requirements.

## Data, language and accessibility

Plus, 12.4/100 GB, 2/5 devices and listed locations are mock examples, not final
commercial policy. Usage is account aggregate; device count means activated,
not concurrent. Production values must eventually come from verified services,
but that integration is out of scope. Do not derive a real entitlement from UI state.

Use “账号” for identity and “账户与套餐” for the agreed navigation label. The large
power glyph on Connection is decorative; the labelled button is the sole focusable
connection action. Give node columns visible/accessibility labels, including
“协议” for the automatic-protocol field. Status uses text plus shape/icon, not color
alone. Avoid duplicated “manage devices” links with indistinguishable destinations.

First demo copy is Simplified Chinese, with resource IDs rather than inline
production strings. Preserve the five-language/RTL requirement without claiming
five completed translations. English fallback text and test-only long/RTL strings
belong in the later bounded implementation plan; do not machine-invent approved
Persian translations. Email/IP/technical tokens keep LTR isolation in future RTL.

Author layout proposal: target 1440x1024 logical layout, usable down to 1024x768;
scroll main content at smaller available heights, not the title bar. Keep labels
unclipped under enlarged text/DPI. Support keyboard navigation, visible focus,
named controls and dialog focus return. Validation of DPI, contrast, keyboard,
screen readers and RTL is still future work, not proven by generated images.

## Acceptance boundary and builder discretion

A later plan must cover every normal/alternate state above, Back/Cancel behavior,
connected node-change confirmation, country-lock failures, quota/device refusal,
placeholder destinations, and reset-without-persistence. Test scenarios must be
explicitly synthetic. Visual fidelity and interaction completion are separate
from production security/runtime acceptance. No client implementation starts
merely because these screens are approved.

Builder may choose internal view-model names, spacing/icon details and in-memory
fixture organization within the native design. Do not substitute a browser app,
install SDKs, define production auth tokens or expose tunnel/service interfaces.
Resolve toolchain availability and exact allowed build/test budget in the future
plan rather than assuming the existing machine is ready.

Author calls requiring user review: fixed read-only demo credentials, explicit
scenario progression, in-memory-only state, connected-change confirmation,
1024x768 minimum target and placeholder secondary flows. These make the bounded
demo honest without claiming real capabilities; they are not prior user decisions.

Independent read-only builder check: “nothing important left.” This concerns
design clarity only, not visual QA, security clearance, runtime acceptance or
approval to implement. Link/whitespace checks passed before preservation.

# Releasing GoCode

Publishing a GitHub release with a tag like `v1.0.0` runs
`.github/workflows/release.yml`. The tag sets the version. The workflow tests,
builds every package and attaches them to the release. The workflow is written
by the Giode CLI from `giode.toml`: after changing `[ci]` or a store section,
run `giode ci -f` instead of editing it, then redo the two pins by hand (the
CLI at `giode-cli/cmd/giode@v0.1.35-alpha.1` instead of `@latest`, and
`go-version-file: go.mod` instead of `go-version: stable`).

| Where | What the release gets | Needs |
| --- | --- | --- |
| GitHub, Linux | `.deb` and `.rpm`, x64 and ARM64 | nothing |
| GitHub, Windows | `setup.exe` installers, x64 and ARM64 | nothing; a code signing certificate stops SmartScreen warnings |
| GitHub, macOS | a universal `.dmg` | a Developer ID certificate, or macOS refuses to open it |
| Mac App Store | a `.pkg` to upload with Transporter | an Apple Developer account and the store certificates |
| Microsoft Store | `.msix` packages to upload in Partner Center | a Partner Center account and the reserved name |

There is no Snap Store package: Linux users download from the GitHub release.

## Repository secrets

Set these in GitHub under Settings → Secrets and variables → Actions. Each
`*_BASE64` secret is the file encoded with `base64 -i file | pbcopy` on macOS,
or `base64 -w0 file` on Linux. Without a secret, that step is skipped and the
rest of the release still builds.

| Secret | What it is |
| --- | --- |
| `MACOS_CERT_BASE64`, `MACOS_CERT_PASS` | The "Developer ID Application" certificate as a `.p12`, and its password: signs the `.dmg` |
| `GIODE_NOTARY_ID`, `GIODE_NOTARY_TEAM_ID`, `GIODE_NOTARY_PASS` | Your Apple ID, team ID and an app-specific password: notarizes the `.dmg` |
| `MACOS_STORE_CERT_BASE64`, `MACOS_STORE_CERT_PASS` | The "Apple Distribution" certificate as a `.p12` |
| `MACOS_INSTALLER_CERT_BASE64`, `MACOS_INSTALLER_CERT_PASS` | The "Mac Installer Distribution" certificate as a `.p12` |
| `MACOS_PROFILE_BASE64` | The Mac App Store provisioning profile (`.provisionprofile`) |
| `WINDOWS_CERT_BASE64`, `GIODE_WINDOWS_CERT_PASS` | An Authenticode `.pfx` and its password: signs the Windows installer (optional) |

## Mac App Store

1. Join the Apple Developer Program.
2. In Certificates, Identifiers & Profiles:
   - Register the App ID `dev.gocode.app` (the `app.id` in `giode.toml`).
   - Create an "Apple Distribution" and a "Mac Installer Distribution" certificate.
   - Create a "Mac App Store" provisioning profile for the App ID.
3. In App Store Connect, create the app with the same bundle ID.
4. Add the three Mac App Store secrets above, then publish a release. The
   release gets `GoCode-<version>.pkg`; upload it with the Transporter app.
5. In App Store Connect, fill in the listing:
   - **App Privacy:** "Data Not Collected". GoCode sends chats only to the
     providers the user sets up, never to its developer.
   - **Privacy Policy URL:** https://xmhrnx.github.io/gocode/#privacy
   - **Category:** Developer Tools (already set in the build).
   - **Screenshots and description:** see Store listing below.
   - **Export compliance:** answered in the build
     (`ITSAppUsesNonExemptEncryption = false`).

**The App Sandbox limits GoCode in the store version.** Every Mac App Store app
runs sandboxed, and so do the commands the agent runs:

- **Folders:** the agent can only read and write folders the user opens with the
  folder button, and Downloads. Opened folders are remembered across restarts.
- **Default projects folder:** new chats get their folders inside the app's
  sandbox container (`~/Library/Containers/dev.gocode.app/Data/GoCode`), not in
  your home folder.
- **Other programs:** tools installed outside the system, such as Homebrew's
  `node`, `npx`, `python` or `go`, may fail to run. MCP servers started with
  `npx` are among them.
- **App review:** an app that runs commands the model writes may be questioned
  under guideline 2.5.2.

The notarized `.dmg` on GitHub has none of these limits. Point users who want
the full agent to it.

## Microsoft Store

1. In Partner Center (the same account as EasySend and AnyFile), reserve the
   name "GoCode".
2. `giode.toml` already has the account's values under `[windows.msix]`:
   `publisher = "CN=B40E8F57-A7E0-40BB-8E74-5D051D0B578B"` and
   `publisher_display_name = "Mehran Zahiroddini"`. `identity_name` is set to
   `MehranZahir.GoCode`: once the name is reserved, open the product's Product
   identity page and check that **Package/Identity/Name** matches it exactly;
   if not, change it and commit before the release.
3. Publish a release. It gets `GoCode-<version>-amd64.msix` and
   `GoCode-<version>-arm64.msix`; upload both in a Partner Center submission.
   The Store signs them, so they need no certificate here.
4. GoCode is a full-trust desktop app (it runs commands and reads project
   folders). When the submission asks why it uses the `runFullTrust`
   capability, explain that.
5. **Privacy policy URL:** https://xmhrnx.github.io/gocode/#privacy. Age
   rating: no content of its own; it shows text from the AI models the user
   connects, so answer the questionnaire's user-generated / unrestricted
   content questions accordingly.

## Store listing

**Name:** GoCode

**Subtitle / short description (≤ 80 characters):**
A native coding agent and model chat. No browser engine inside.

**Keywords (Apple, ≤ 100 characters):**
coding agent,ai,llm,code,developer,chat,ollama,openai,claude,mcp,terminal,programming

**Description:**

> GoCode is a coding agent for the desktop, written in Go and drawn straight to
> the GPU. There is no browser engine inside, so it opens instantly and stays
> fast through long chats.
>
> Point it at a project folder and it reads, searches and edits your files and
> runs your builds and tests, streaming as it works. Send messages while it
> works; it picks them up at the next step.
>
> Bring your own model: over 200 providers, local models and custom
> OpenAI-compatible endpoints. Add MCP servers, web search and skills.
>
> Your keys are kept in the system's credential store, and your chats stay on
> your computer. GoCode sends nothing to its developer.

**Category:** Developer Tools (Mac App Store), Developer tools (Microsoft
Store).

**Screenshots:** the website draws its app window in HTML, so there are no
screenshot files yet. Capture the app with a project open and the agent
working: Mac 2880 × 1800 (or 1440 × 900), Microsoft Store 1920 × 1080.

## Linux

The `.deb` and `.rpm` are attached to the release, and the website's download
button offers the right one. They list the system libraries they need, so
`apt` and `dnf` install those too.

## Phones

GoCode on Android and iOS is the same app, with the same ID. On a phone it
doesn't run the agent: it pairs with GoCode on a desktop and follows and
steers the agent there (`mobile_app.go` and `ui/mobile`; desktops build
`desktop_app.go` instead). A release builds both:

| Where | What the release gets | Needs |
| --- | --- | --- |
| Android | an `.apk` to install and an `.aab` for Google Play | a keystore for Play: `ANDROID_KEYSTORE_BASE64` and `GIODE_KEYSTORE_PASS` |
| iOS | an `.ipa` for the App Store | `IOS_CERT_BASE64`, `IOS_CERT_PASS`, `IOS_PROFILE_BASE64` and `IOS_TEAM_ID`; without them CI only checks that it builds |

Make the Android keystore once with
`keytool -genkey -keystore release.keystore -alias gocode -keyalg RSA -validity 10000`
and keep it: every update must be signed with the same key.

- **Pairing** goes through the public Magic Wormhole server. To use your own,
  set `GOCODE_PAIRING_SERVER` to its WebSocket address on both devices.
- **Away from home**, phones reach a desktop through the relay Worker in
  `relay/` while one of its chats is working. Deploy it as `relay/README.md`
  says and put its address in Settings → Phone.
- **iOS** asks for local network access the first time; the reason is set in
  `giode.toml`. Finding a desktop by broadcast on iOS needs Apple's multicast
  networking entitlement. Without it, the phone still reaches the desktop at
  the addresses it had when pairing, and through the relay.

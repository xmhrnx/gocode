# Releasing GoCode

Publishing a GitHub release with a tag like `v1.0.0` runs
`.github/workflows/release.yml`. The tag sets the version. The workflow tests,
builds every package and attaches them to the release. The workflow is written
by the Giode CLI from `giode.toml`: after changing `[ci]` or a store section,
run `giode ci -f` instead of editing it.

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
   - **Category:** Developer Tools (already set in the build).
   - **Screenshots and description:** from the website.
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

1. Open a Partner Center developer account.
2. Reserve the name "GoCode".
3. In the product's Product identity page, copy these into `giode.toml`:
   - **Package/Identity/Name** → `identity_name` under `[windows.msix]`.
   - **Package/Identity/Publisher** (`CN=...`) → `publisher` under `[windows.msix]`.
   - **Package/Properties/PublisherDisplayName** → `publisher` under `[app]`.
4. Commit and publish a release. It gets `GoCode-<version>-amd64.msix` and
   `GoCode-<version>-arm64.msix`; upload both in a Partner Center submission.
   The Store signs them, so they need no certificate here.
5. GoCode is a full-trust desktop app (it runs commands and reads project
   folders). When the submission asks why it uses the `runFullTrust`
   capability, explain that.

## Linux

The `.deb` and `.rpm` are attached to the release, and the website's download
button offers the right one. They list the system libraries they need, so
`apt` and `dnf` install those too.

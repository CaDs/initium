# iOS — SwiftUI + Liquid Glass

This is the native iOS app, replacing the Flutter app dropped on branch
`feat/dropping_flutter`. The Xcode project lives at
`initium/initium.xcodeproj`.

## Toolchain

- **Xcode 26+** (required to build against the iOS 26 SDK — Liquid
  Glass APIs resolve via `#available` at runtime).
- **Deployment target iOS 17.0**.
- **SwiftUI only.** No UIKit controllers for new work.
- **Swift Testing** for unit tests, XCUITest for UI tests.

## Conventions

Load `.claude/skills/initium-mobile/patterns/ios.md` before editing —
that's the authoritative source. Highlights:

- The `initium/` folder is a `PBXFileSystemSynchronizedRootGroup`
  (Xcode 16+). Drop a new `.swift` file in there and it joins the
  target automatically. No pbxproj surgery.
- Liquid Glass is per-surface, always guarded: `#available(iOS 26.0, *)`
  with a `.regularMaterial` fallback. Use the `.liquidGlassCard()`
  modifier in `initium/LiquidGlassCard.swift` as the template.
- State: `@State` for view-local, `@Observable` for shared. No Combine.
- Tabs: `TabView(selection:)` driven by the `AppTab` enum.

## Quick start

```sh
# Build + run on simulator (boots one if needed)
make dev:ios

# Tests
make test:ios

# Override the simulator
make dev:ios IOS_SIM='iPhone 17 Pro'
```

Or open `initium/initium.xcodeproj` in Xcode and press Cmd+R.

## What NOT to do

- Don't reintroduce SwiftData (the template starter used it; we
  removed it — nothing to persist yet).
- Don't add dependencies via SPM without a matching "wire up X" PR.
- Don't hardcode `.glassEffect` without an `#available` guard. Every
  Liquid Glass surface must render acceptably on iOS 17–25.
- Don't add a navigation library. SwiftUI's `NavigationStack` +
  `TabView` cover every routing case we have.

## Files worth knowing

- `initium/initiumApp.swift` — `@main`, hosts `RootView`.
- `initium/RootView.swift` — the `TabView` with three tabs.
- `initium/AppTab.swift` — the enum that drives tab iteration.
- `initium/LiquidGlassCard.swift` — the opt-in Liquid Glass modifier.
- `initium/HomeScreen.swift` — shows the Liquid Glass card in action.
- `initiumTests/initiumTests.swift` — Swift Testing smoke tests.

## Bundle ID

`cads.initium` (Xcode default). Forks should rename in
`initium.xcodeproj` build settings (`PRODUCT_BUNDLE_IDENTIFIER`) for
both the `initium` target and the two test targets.

# iOS pattern (SwiftUI + Liquid Glass)

You are editing the SwiftUI app under `mobile/ios/initium/`. This file
covers the conventions that apply to the iOS side specifically. For the
cross-platform rules see the parent `SKILL.md`.

## Toolchain

- **Xcode 26+** (required to build against the iOS 26 SDK even if the
  deployment target is lower — the `.glassEffect()` API resolves
  conditionally via `#available`).
- **Deployment target iOS 17.0** — gives us `@Observable`, typed
  `NavigationStack`, modern `ScrollView` / `TabView` APIs without
  back-compat complexity. Older iOS versions are intentionally not
  supported.
- **Swift 5.0+** (the pbxproj pins SWIFT_VERSION = 5.0 but Swift 6
  features are available via upcoming-feature flags already set by Xcode).
- **SwiftUI only.** Do not introduce UIKit view controllers unless a
  feature genuinely requires a UIKit-only API; when that happens, wrap
  it in `UIViewControllerRepresentable`.

## Project layout

```
mobile/ios/initium/                           ← Xcode project root
├── initium.xcodeproj/                        ← project file
├── initium/                                  ← app sources (synchronized folder)
│   ├── initiumApp.swift                      ← @main entry
│   ├── RootView.swift                        ← TabView with 3 tabs
│   ├── AppTab.swift                          ← enum Tab definition
│   ├── LiquidGlassCard.swift                 ← `.liquidGlassCard()` modifier
│   ├── HomeScreen.swift
│   ├── MainScreen.swift
│   ├── SettingsScreen.swift
│   └── Assets.xcassets/
├── initiumTests/                             ← Swift Testing unit tests
│   └── initiumTests.swift
└── initiumUITests/                           ← XCUITest UI tests
```

The app target uses `PBXFileSystemSynchronizedRootGroup` (Xcode 16+),
which means **any new `.swift` file dropped into `initium/` is
automatically picked up by the build target** — no manual pbxproj
registration needed. Same for `initiumTests/` and `initiumUITests/`.

## Liquid Glass opt-in (graceful fallback)

Liquid Glass is a **per-surface** treatment, not a global theme. Apply
it where it adds value and let the rest of the app render with standard
SwiftUI materials.

The canonical wrapper is `liquidGlassCard(cornerRadius:)` — see
`mobile/ios/initium/initium/LiquidGlassCard.swift` <!-- expect: liquidGlassCard -->:

```swift
extension View {
    @ViewBuilder
    func liquidGlassCard(cornerRadius: CGFloat = 24) -> some View {
        if #available(iOS 26.0, *) {
            self.glassEffect(in: .rect(cornerRadius: cornerRadius))
        } else {
            self.background(
                .regularMaterial,
                in: RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
            )
        }
    }
}
```

Rules:

- **Always** wrap Liquid Glass APIs in `#available(iOS 26.0, *)` with a
  material fallback. Never ship a surface that only looks acceptable on
  iOS 26.
- `TabView` auto-adopts Liquid Glass on iOS 26+ with no code change —
  the bottom bar we render in `RootView` gets it for free.
- Don't apply `.glassEffect` to text directly; apply it to the
  *container* (a padded VStack, a card), then put text on top.

## Tabs / navigation

`RootView` (see `mobile/ios/initium/initium/RootView.swift`
<!-- expect: struct RootView -->) holds the `TabView` and owns the
`@State private var selection: AppTab` that drives tab switching. Each
tab's content lives in its own screen file and gets a
`NavigationStack` wrapper so it can host detail routes later.

`AppTab` (see `mobile/ios/initium/initium/AppTab.swift`
<!-- expect: enum AppTab -->) is a `Hashable, CaseIterable` enum —
expand it by adding a new case plus `title` and `systemImage` entries.
Never reorder existing cases without reading call sites (cases are used
as `tag` values and hashed into `@SceneStorage` once state persistence
lands).

## State management

- **View-local state**: `@State` for ephemeral UI state (selected tab,
  toggled modal).
- **Shared state**: `@Observable` macro classes, injected via
  `.environment(\.yourService, service)` at the app root and read with
  `@Environment(\.yourService) private var service`. No
  `ObservableObject`, no `@Published`, no Combine.
- **Scene-persisted state**: `@SceneStorage` for things that should
  survive app backgrounding but not reinstall.
- **App-persisted state**: `@AppStorage` for UserDefaults-backed flags.
- **Secure storage**: NOT YET WIRED. When it lands, use Keychain via a
  thin wrapper, not `KeychainAccess` (unnecessary dep).

## Tests

`initiumTests/initiumTests.swift` <!-- expect: @Test --> uses Swift
Testing, not XCTest. Every new test file in `initiumTests/` should
`import Testing` + `@testable import initium` and use `@Test func` +
`#expect(...)`.

Run locally with `make test:ios` (invokes `xcodebuild test` against a
simulator — set `IOS_SIM='iPhone 17'` to override the destination).

UI tests go in `initiumUITests/` and use XCUITest — Swift Testing's UI
story isn't there yet.

## Build

- Debug build for development: `make build:ios` or Cmd+R in Xcode.
- Debug + install + launch on the running simulator: `make dev:ios`.
  Auto-boots a simulator if none is running.
- Archive / distribution: not yet wired (fork concern — add it when
  the product has a real identity).

## Deferred (do NOT scaffold yet)

- SwiftLint / SwiftFormat configuration. Xcode's built-in format is
  fine for an MVP.
- DI container. Swift's `Environment` works for three screens.
- Networking / URLSession wiring. Wait for the first feature.
- Keychain wrapper.
- Google Sign-In SDK integration.
- Localization (Localizable.strings or String Catalog).

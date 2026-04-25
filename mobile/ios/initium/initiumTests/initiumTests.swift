import Testing
@testable import initium

/// Smoke tests for the AppTab enum. AppTab drives the TabView in
/// RootView; bugs here mean a tab doesn't render or has the wrong icon.

struct AppTabTests {

    @Test func allTabsHaveDistinctTitles() {
        let titles = Set(AppTab.allCases.map(\.title))
        #expect(titles.count == AppTab.allCases.count)
    }

    @Test func allTabsHaveSystemImages() {
        for tab in AppTab.allCases {
            #expect(!tab.systemImage.isEmpty)
        }
    }

    @Test func expectedTabsArePresent() {
        #expect(AppTab.allCases.contains(.home))
        #expect(AppTab.allCases.contains(.main))
        #expect(AppTab.allCases.contains(.settings))
    }

    @Test func homeTab_titleAndIcon() {
        #expect(AppTab.home.title == "Home")
        #expect(AppTab.home.systemImage == "house")
    }

    @Test func mainTab_titleAndIcon() {
        #expect(AppTab.main.title == "Main")
        #expect(AppTab.main.systemImage == "square.grid.2x2")
    }

    @Test func settingsTab_titleAndIcon() {
        #expect(AppTab.settings.title == "Settings")
        #expect(AppTab.settings.systemImage == "gearshape")
    }
}

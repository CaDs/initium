import Testing
@testable import initium

struct initiumTests {
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
}

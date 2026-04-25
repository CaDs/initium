import Foundation

enum AppTab: Hashable, CaseIterable {
    case home
    case main
    case settings

    var title: String {
        switch self {
        case .home: return "Home"
        case .main: return "Main"
        case .settings: return "Settings"
        }
    }

    var systemImage: String {
        switch self {
        case .home: return "house"
        case .main: return "square.grid.2x2"
        case .settings: return "gearshape"
        }
    }
}

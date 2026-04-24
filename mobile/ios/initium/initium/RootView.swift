import SwiftUI

struct RootView: View {
    @State private var selection: AppTab = .home

    var body: some View {
        TabView(selection: $selection) {
            HomeScreen()
                .tabItem { Label(AppTab.home.title, systemImage: AppTab.home.systemImage) }
                .tag(AppTab.home)

            MainScreen()
                .tabItem { Label(AppTab.main.title, systemImage: AppTab.main.systemImage) }
                .tag(AppTab.main)

            SettingsScreen()
                .tabItem { Label(AppTab.settings.title, systemImage: AppTab.settings.systemImage) }
                .tag(AppTab.settings)
        }
    }
}

#Preview {
    RootView()
}

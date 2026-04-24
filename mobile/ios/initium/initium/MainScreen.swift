import SwiftUI

struct MainScreen: View {
    var body: some View {
        NavigationStack {
            VStack {
                Text("Main")
                    .font(.largeTitle.weight(.semibold))
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .navigationTitle("Main")
            .navigationBarTitleDisplayMode(.inline)
        }
    }
}

#Preview {
    MainScreen()
}

import SwiftUI

struct HomeScreen: View {
    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 24) {
                    Text("Home")
                        .font(.largeTitle.weight(.semibold))
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal)

                    greetingCard
                        .padding(.horizontal)
                }
                .padding(.vertical, 32)
            }
            .navigationTitle("Home")
            .navigationBarTitleDisplayMode(.inline)
        }
    }

    /// A small greeting surface that demonstrates the Liquid Glass opt-in.
    /// On iOS 26+ it renders with `.glassEffect(...)`; on earlier OSes it
    /// falls back to `.regularMaterial`. Either way the layout is identical,
    /// so forks can freely remove the effect without touching content code.
    private var greetingCard: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Welcome to Initium")
                .font(.title3.weight(.semibold))
            Text("This card uses Liquid Glass on iOS 26+ and a regularMaterial fallback elsewhere.")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .padding(20)
        .frame(maxWidth: .infinity, alignment: .leading)
        .liquidGlassCard()
    }
}

#Preview {
    HomeScreen()
}

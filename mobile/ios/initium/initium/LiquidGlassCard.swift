import SwiftUI

extension View {
    /// Applies a Liquid Glass treatment on iOS 26+, falling back to
    /// a regularMaterial-filled rounded rectangle on older OSes.
    ///
    /// Use this modifier on the *surface* (padding already applied); it
    /// swaps the background treatment without altering layout.
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

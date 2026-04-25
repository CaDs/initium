import Testing
import Foundation
@testable import initium

/// `parseMagicLinkToken(from:)` is the URL parser behind `initium://auth/verify`
/// deep links. It must be strict — a malformed URL must NOT trigger an auth
/// attempt, since that would let any installed app drive our login flow.

struct DeepLinkTests {

    @Test func happyPath_returnsToken() {
        let url = URL(string: "initium://auth/verify?token=abc123")!
        #expect(parseMagicLinkToken(from: url) == "abc123")
    }

    @Test func wrongScheme_returnsNil() {
        let url = URL(string: "https://auth/verify?token=abc")!
        #expect(parseMagicLinkToken(from: url) == nil)
    }

    @Test func wrongHost_returnsNil() {
        let url = URL(string: "initium://other/verify?token=abc")!
        #expect(parseMagicLinkToken(from: url) == nil)
    }

    @Test func wrongPath_returnsNil() {
        let url = URL(string: "initium://auth/login?token=abc")!
        #expect(parseMagicLinkToken(from: url) == nil)
    }

    @Test func missingToken_returnsNil() {
        let url = URL(string: "initium://auth/verify")!
        #expect(parseMagicLinkToken(from: url) == nil)
    }

    @Test func emptyToken_returnsNil() {
        let url = URL(string: "initium://auth/verify?token=")!
        #expect(parseMagicLinkToken(from: url) == nil)
    }

    @Test func extraQueryParams_areIgnored() {
        let url = URL(string: "initium://auth/verify?token=abc&utm_source=email")!
        #expect(parseMagicLinkToken(from: url) == "abc")
    }
}

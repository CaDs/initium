import Foundation

/// Hand-written Codable structs mirroring the schemas in
/// `backend/api/openapi.yaml`. These will be replaced by
/// swift-openapi-generator output once the SPM plugin is wired
/// (follow-up PR). Until then, any spec change must be reflected here.
///
/// Uses a single shared `JSONDecoder` / `JSONEncoder` configured for
/// the backend's conventions (snake_case, ISO-8601 dates).

enum UserRole: String, Codable, Sendable {
    case user
    case admin
}

struct User: Codable, Equatable, Sendable, Identifiable {
    let id: String
    let email: String
    let name: String
    let avatarURL: String
    let role: UserRole
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id, email, name, role
        case avatarURL = "avatar_url"
        case createdAt = "created_at"
    }
}

struct TokenPair: Codable, Equatable, Sendable {
    let accessToken: String
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
    }
}

struct MagicLinkRequest: Codable, Sendable {
    let email: String
}

struct RefreshRequest: Codable, Sendable {
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

struct MobileGoogleRequest: Codable, Sendable {
    let idToken: String

    enum CodingKeys: String, CodingKey {
        case idToken = "id_token"
    }
}

struct MobileVerifyRequest: Codable, Sendable {
    let token: String
}

struct MessageResponse: Codable, Sendable {
    let message: String
}

struct ErrorResponse: Codable, Sendable, Error {
    let code: String
    let message: String
    let requestId: String?

    enum CodingKeys: String, CodingKey {
        case code, message
        case requestId = "request_id"
    }
}

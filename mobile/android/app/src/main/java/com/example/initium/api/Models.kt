package com.example.initium.api

import com.squareup.moshi.Json

/**
 * Hand-written Moshi DTOs mirroring the shapes declared in
 * `backend/api/openapi.yaml`. These will be replaced by
 * openapi-generator output once the Gradle plugin is wired (follow-up
 * PR). Until then, any spec change must be reflected here.
 *
 * Uses `moshi-kotlin` reflection-based adapters (no KSP / codegen).
 * Fine for ~10 small types. If the count crosses ~30, switch to
 * `com.squareup.moshi:moshi-kotlin-codegen` via KSP.
 */

data class User(
    val id: String,
    val email: String,
    val name: String,
    @Json(name = "avatar_url") val avatarUrl: String,
    val role: UserRole,
    @Json(name = "created_at") val createdAt: String,
)

enum class UserRole {
    @Json(name = "user") USER,
    @Json(name = "admin") ADMIN,
}

data class TokenPair(
    @Json(name = "access_token") val accessToken: String,
    @Json(name = "refresh_token") val refreshToken: String,
)

data class MagicLinkRequest(val email: String)

data class RefreshRequest(
    @Json(name = "refresh_token") val refreshToken: String,
)

data class MobileGoogleRequest(
    @Json(name = "id_token") val idToken: String,
)

data class MobileVerifyRequest(val token: String)

data class MessageResponse(val message: String)

data class ErrorResponse(
    val code: String,
    val message: String,
    @Json(name = "request_id") val requestId: String? = null,
)

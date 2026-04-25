package com.example.initium.api

/**
 * In-memory [ApiClient] for unit tests. Each method has a configurable
 * lambda return so tests can simulate success or any subclass of
 * [ApiException]. Tracks call counts + arguments per method so tests
 * can assert on what was invoked.
 */
class FakeApiClient(
    var requestMagicLinkResult: () -> MessageResponse =
        { MessageResponse("magic link sent") },
    var verifyMagicLinkResult: () -> TokenPair =
        { TokenPair(accessToken = "test-access", refreshToken = "test-refresh") },
    var verifyGoogleIdTokenResult: () -> TokenPair =
        { TokenPair(accessToken = "test-access", refreshToken = "test-refresh") },
    var meResult: () -> User = { TestUsers.devUser },
) : ApiClient {

    val requestMagicLinkCalls = mutableListOf<String>()
    val verifyMagicLinkCalls = mutableListOf<String>()
    val verifyGoogleIdTokenCalls = mutableListOf<String>()
    var meCallCount = 0
        private set
    var logoutCallCount = 0
        private set

    override suspend fun requestMagicLink(email: String): MessageResponse {
        requestMagicLinkCalls.add(email)
        return requestMagicLinkResult()
    }

    override suspend fun verifyMagicLink(token: String): TokenPair {
        verifyMagicLinkCalls.add(token)
        return verifyMagicLinkResult()
    }

    override suspend fun verifyGoogleIdToken(idToken: String): TokenPair {
        verifyGoogleIdTokenCalls.add(idToken)
        return verifyGoogleIdTokenResult()
    }

    override suspend fun me(): User {
        meCallCount++
        return meResult()
    }

    override suspend fun logout() {
        logoutCallCount++
    }
}

/** Stable fixture users so assertions don't churn. */
object TestUsers {
    val devUser = User(
        id = "00000000-0000-0000-0000-000000000001",
        email = "dev@initium.local",
        name = "Dev User",
        avatarUrl = "",
        role = UserRole.USER,
        createdAt = "2026-04-25T00:00:00Z",
    )
}

/**
 * In-memory [TokenStorage] for tests. Mirrors [TokenStore]'s behavior
 * without requiring Robolectric or `EncryptedSharedPreferences`.
 */
class FakeTokenStore(
    private var access: String? = null,
    private var refresh: String? = null,
) : TokenStorage {
    override fun save(tokens: TokenPair) {
        access = tokens.accessToken
        refresh = tokens.refreshToken
    }

    override fun accessToken(): String? = access
    override fun refreshToken(): String? = refresh

    override fun clear() {
        access = null
        refresh = null
    }
}

package com.example.initium.api

import kotlinx.coroutines.runBlocking
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

/**
 * Tests `OkHttpApiClient` against a `MockWebServer`. The auth header
 * injection + 401-refresh-retry contract is the most security-relevant
 * part of the client; if it regresses, every authenticated request
 * silently fails or retries with stale tokens.
 */
class OkHttpApiClientTest {

    private lateinit var server: MockWebServer
    private lateinit var storage: FakeTokenStore

    @Before
    fun setUp() {
        server = MockWebServer()
        server.start()
        storage = FakeTokenStore()
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun client(onUnauthorized: () -> Unit = {}): OkHttpApiClient =
        OkHttpApiClient(
            tokenStore = storage,
            baseUrl = server.url("/").toString(),
            onUnauthorized = onUnauthorized,
        )

    @Test
    fun `me attaches Bearer header from token store`() = runBlocking {
        storage.save(TokenPair(accessToken = "ABC", refreshToken = "REF"))
        server.enqueue(MockResponse().setResponseCode(200).setBody(USER_JSON))

        client().me()

        val request: RecordedRequest = server.takeRequest()
        assertEquals("Bearer ABC", request.getHeader("Authorization"))
        assertEquals("/api/me", request.path)
    }

    @Test
    fun `requestMagicLink does NOT attach Authorization (skipAuth)`() = runBlocking {
        storage.save(TokenPair(accessToken = "ABC", refreshToken = "REF"))
        server.enqueue(MockResponse().setResponseCode(200).setBody("""{"message":"sent"}"""))

        client().requestMagicLink("user@example.com")

        val request = server.takeRequest()
        // Magic-link request must be unauthenticated; the user has no
        // session yet by definition.
        assertTrue(
            "magic-link request must NOT have Authorization header",
            request.getHeader("Authorization").isNullOrEmpty(),
        )
    }

    @Test
    fun `401 triggers refresh and retries the original request once`() = runBlocking {
        storage.save(TokenPair(accessToken = "OLD", refreshToken = "REFRESH"))

        // 1) initial me() with OLD → 401
        server.enqueue(MockResponse().setResponseCode(401).setBody(ERROR_JSON))
        // 2) refresh succeeds with NEW
        server.enqueue(MockResponse().setResponseCode(200).setBody(NEW_PAIR_JSON))
        // 3) retried me() with NEW → 200
        server.enqueue(MockResponse().setResponseCode(200).setBody(USER_JSON))

        val user = client().me()

        assertEquals("Dev User", user.name)
        // Token store now holds the new tokens.
        assertEquals("NEW-ACCESS", storage.accessToken())
        assertEquals("NEW-REFRESH", storage.refreshToken())
        assertEquals(3, server.requestCount)
    }

    @Test
    fun `refresh failure invokes onUnauthorized callback`() = runBlocking {
        storage.save(TokenPair(accessToken = "OLD", refreshToken = "REFRESH"))

        server.enqueue(MockResponse().setResponseCode(401).setBody(ERROR_JSON))
        // Refresh itself fails with 401 → callback fires, tokens cleared.
        server.enqueue(MockResponse().setResponseCode(401).setBody(ERROR_JSON))

        var unauthorizedCalled = false
        val client = client(onUnauthorized = { unauthorizedCalled = true })

        var thrown: Throwable? = null
        try {
            client.me()
        } catch (t: Throwable) {
            thrown = t
        }

        assertNotNull("me() must surface an exception when refresh fails", thrown)
        assertTrue(unauthorizedCalled)
        // Tokens cleared so the app drops to login on next observation.
        assertEquals(null, storage.accessToken())
        assertEquals(null, storage.refreshToken())
    }

    @Test
    fun `HTTP error envelope surfaces as ApiException_HttpError`() = runBlocking {
        storage.save(TokenPair(accessToken = "ABC", refreshToken = "REF"))
        server.enqueue(MockResponse().setResponseCode(409).setBody(USED_TOKEN_JSON))

        var thrown: Throwable? = null
        try {
            client().verifyMagicLink("used-token")
        } catch (t: Throwable) {
            thrown = t
        }

        assertTrue(thrown is ApiException.HttpError)
        val http = thrown as ApiException.HttpError
        assertEquals(409, http.status)
        assertEquals("TOKEN_USED", http.envelope?.code)
    }

    private companion object {
        const val USER_JSON = """{
            "id":"00000000-0000-0000-0000-000000000001",
            "email":"dev@initium.local",
            "name":"Dev User",
            "avatar_url":"",
            "role":"user",
            "created_at":"2026-04-25T00:00:00Z"
        }"""

        const val NEW_PAIR_JSON = """{
            "access_token":"NEW-ACCESS",
            "refresh_token":"NEW-REFRESH"
        }"""

        const val ERROR_JSON = """{"code":"INVALID_CREDENTIALS","message":"unauthorized"}"""
        const val USED_TOKEN_JSON = """{"code":"TOKEN_USED","message":"token already used"}"""
    }
}

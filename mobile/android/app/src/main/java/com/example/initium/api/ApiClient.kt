package com.example.initium.api

import com.example.initium.BuildConfig
import com.squareup.moshi.JsonAdapter
import com.squareup.moshi.Moshi
import com.squareup.moshi.kotlin.reflect.KotlinJsonAdapterFactory
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import okhttp3.Authenticator
import okhttp3.Interceptor
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import okhttp3.Route
import okhttp3.logging.HttpLoggingInterceptor
import java.io.IOException
import java.util.concurrent.TimeUnit

private val JSON = "application/json; charset=utf-8".toMediaType()
private const val HEADER_SKIP_AUTH = "X-Initium-Skip-Auth"

/**
 * Thin hand-written API client for the auth + profile endpoints.
 *
 * Concerns:
 *  - Attaches `Authorization: Bearer <access>` to every request except
 *    those marked with the private [HEADER_SKIP_AUTH] (login, magic-link
 *    request, refresh).
 *  - On 401, the [Authenticator] swaps in a fresh token pair from
 *    `/api/auth/refresh` and retries the original request once. OkHttp
 *    serializes authenticator calls per connection; the [refreshMutex]
 *    additionally serializes across connections so a thundering-herd of
 *    401s triggers exactly one refresh.
 *  - All JSON uses Moshi's reflection adapter
 *    ([KotlinJsonAdapterFactory]). Switch to KSP codegen if the model
 *    count grows past ~30.
 *
 * Instances are cheap — create one per-process, hold a reference in
 * your DI graph. Don't construct per-call.
 *
 * @param onUnauthorized Called when a refresh attempt fails. Forks wire
 *   this to their auth state machine so the UI drops to the login
 *   screen.
 */
class ApiClient(
    private val tokenStore: TokenStore,
    baseUrl: String = BuildConfig.API_BASE_URL,
    private val onUnauthorized: () -> Unit = {},
) {

    private val baseUrl = baseUrl.trimEnd('/')
    private val moshi: Moshi = Moshi.Builder().add(KotlinJsonAdapterFactory()).build()
    private val errorAdapter: JsonAdapter<ErrorResponse> = moshi.adapter(ErrorResponse::class.java)
    private val refreshMutex = Mutex()

    private val client: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .addInterceptor(authInterceptor())
        .addInterceptor(
            HttpLoggingInterceptor().apply {
                level = if (BuildConfig.DEBUG) HttpLoggingInterceptor.Level.BASIC else HttpLoggingInterceptor.Level.NONE
            }
        )
        .authenticator(tokenAuthenticator())
        .build()

    // ------------------------------------------------------------------
    // Public API — one suspend fun per endpoint.
    // ------------------------------------------------------------------

    suspend fun requestMagicLink(email: String): MessageResponse =
        post("/api/auth/magic-link", MagicLinkRequest(email), MessageResponse::class.java, skipAuth = true)

    suspend fun verifyMagicLink(token: String): TokenPair {
        val pair = post(
            "/api/auth/mobile/verify",
            MobileVerifyRequest(token),
            TokenPair::class.java,
            skipAuth = true,
        )
        tokenStore.save(pair)
        return pair
    }

    suspend fun verifyGoogleIdToken(idToken: String): TokenPair {
        val pair = post(
            "/api/auth/mobile/google",
            MobileGoogleRequest(idToken),
            TokenPair::class.java,
            skipAuth = true,
        )
        tokenStore.save(pair)
        return pair
    }

    suspend fun me(): User = get("/api/me", User::class.java)

    /**
     * Best-effort logout: POSTs to the backend to revoke the session
     * server-side, then always clears local tokens regardless of the
     * server response.
     */
    suspend fun logout() {
        runCatching {
            post<Any, MessageResponse>("/api/auth/logout", EmptyBody, MessageResponse::class.java)
        }
        tokenStore.clear()
    }

    // ------------------------------------------------------------------
    // Internal plumbing.
    // ------------------------------------------------------------------

    private object EmptyBody

    private suspend fun <T : Any> get(path: String, cls: Class<T>): T =
        withContext(Dispatchers.IO) {
            execute(Request.Builder().url(baseUrl + path).get().build(), cls)
        }

    private suspend fun <B : Any, T : Any> post(
        path: String,
        body: B,
        responseType: Class<T>,
        skipAuth: Boolean = false,
    ): T = withContext(Dispatchers.IO) {
        val bodyJson = if (body === EmptyBody) "{}" else moshi.adapter(body.javaClass).toJson(body)
        val builder = Request.Builder()
            .url(baseUrl + path)
            .post(bodyJson.toRequestBody(JSON))
        if (skipAuth) builder.header(HEADER_SKIP_AUTH, "1")
        execute(builder.build(), responseType)
    }

    private fun <T : Any> execute(request: Request, responseType: Class<T>): T {
        val response = try {
            client.newCall(request).execute()
        } catch (e: IOException) {
            throw ApiException.NetworkError(e.message ?: "network failure", e)
        }
        response.use { resp ->
            val bodyString = resp.body?.string().orEmpty()
            if (!resp.isSuccessful) {
                val envelope = runCatching { errorAdapter.fromJson(bodyString) }.getOrNull()
                throw ApiException.HttpError(resp.code, envelope)
            }
            return moshi.adapter(responseType).fromJson(bodyString)
                ?: throw ApiException.DecodeError("empty body for ${responseType.simpleName}")
        }
    }

    private fun authInterceptor() = Interceptor { chain ->
        val original = chain.request()
        val skip = original.header(HEADER_SKIP_AUTH) != null
        val request = if (skip) {
            original.newBuilder().removeHeader(HEADER_SKIP_AUTH).build()
        } else {
            val token = tokenStore.accessToken()
            if (token.isNullOrEmpty()) {
                original
            } else {
                original.newBuilder()
                    .header("Authorization", "Bearer $token")
                    .build()
            }
        }
        chain.proceed(request)
    }

    private fun tokenAuthenticator() = Authenticator { _: Route?, response: Response ->
        if (response.request.url.encodedPath.endsWith("/api/auth/refresh")) return@Authenticator null
        if (responseChainDepth(response) >= 2) return@Authenticator null

        val refresh = tokenStore.refreshToken() ?: run {
            onUnauthorized()
            return@Authenticator null
        }

        val newPair = runBlocking {
            refreshMutex.withLock {
                // If another thread already refreshed, reuse its result.
                val maybeRefreshed = tokenStore.accessToken()
                val currentlyOnRequest = response.request.header("Authorization")?.removePrefix("Bearer ")
                if (maybeRefreshed != null && maybeRefreshed != currentlyOnRequest) {
                    TokenPair(maybeRefreshed, tokenStore.refreshToken() ?: refresh)
                } else {
                    refreshTokensBlocking(refresh)
                }
            }
        } ?: run {
            onUnauthorized()
            return@Authenticator null
        }

        response.request.newBuilder()
            .header("Authorization", "Bearer ${newPair.accessToken}")
            .build()
    }

    /** Synchronous refresh used from the Authenticator. Null = auth lost. */
    private fun refreshTokensBlocking(refresh: String): TokenPair? {
        val request = Request.Builder()
            .url("$baseUrl/api/auth/refresh")
            .post(
                moshi.adapter(RefreshRequest::class.java)
                    .toJson(RefreshRequest(refresh))
                    .toRequestBody(JSON)
            )
            .header(HEADER_SKIP_AUTH, "1")
            .build()

        // Strip the authenticator to prevent recursion if the refresh itself 401s.
        val bare = client.newBuilder().authenticator(Authenticator.NONE).build()
        val response = runCatching { bare.newCall(request).execute() }.getOrNull() ?: return null
        response.use { resp ->
            if (!resp.isSuccessful) {
                tokenStore.clear()
                return null
            }
            val body = resp.body?.string().orEmpty()
            val pair = runCatching { moshi.adapter(TokenPair::class.java).fromJson(body) }.getOrNull()
            if (pair != null) tokenStore.save(pair)
            return pair
        }
    }

    private fun responseChainDepth(response: Response): Int {
        var count = 1
        var prior = response.priorResponse
        while (prior != null) {
            count++
            prior = prior.priorResponse
        }
        return count
    }
}

/** Polymorphic error wrapper for API failures. Callers pattern-match. */
sealed class ApiException(message: String, cause: Throwable? = null) : Exception(message, cause) {
    class NetworkError(message: String, cause: Throwable) : ApiException(message, cause)
    class HttpError(val status: Int, val envelope: ErrorResponse?) :
        ApiException("HTTP $status${envelope?.let { " (${it.code})" } ?: ""}")

    class DecodeError(message: String) : ApiException(message)
}

package com.example.initium

import android.app.Application
import com.example.initium.api.ApiClient
import com.example.initium.api.OkHttpApiClient
import com.example.initium.api.TokenStore

/**
 * App-level dependency holder. Constructed once by the OS, shared by
 * every Activity/ViewModel via `(context.applicationContext as
 * InitiumApplication)`. No DI container — the graph is tiny.
 *
 * Release builds hard-fail if `DEV_BYPASS_AUTH` is set; matches the
 * iOS / backend guard.
 */
class InitiumApplication : Application() {

    lateinit var tokenStore: TokenStore
        private set

    lateinit var apiClient: ApiClient
        private set

    /** Set by AuthViewModel so ApiClient can broadcast 401 → login. */
    var onUnauthorized: () -> Unit = {}

    override fun onCreate() {
        super.onCreate()

        if (!BuildConfig.DEBUG && BuildConfig.DEV_BYPASS_AUTH) {
            error("DEV_BYPASS_AUTH must not be enabled in release builds.")
        }

        tokenStore = TokenStore(this)
        apiClient = OkHttpApiClient(
            tokenStore = tokenStore,
            onUnauthorized = { onUnauthorized() },
        )
    }
}

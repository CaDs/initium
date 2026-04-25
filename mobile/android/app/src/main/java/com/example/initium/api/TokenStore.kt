package com.example.initium.api

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

/**
 * Persists the access + refresh token pair in EncryptedSharedPreferences.
 *
 * The store survives app restarts but NOT app uninstall. We pair it
 * with a "first launch wipe" flag in regular SharedPreferences so that
 * after a reinstall (where the keystore entries remain but this
 * wrapper's backing prefs are empty) we explicitly clear any residual
 * encrypted entries — prevents stale tokens resurfacing.
 *
 * Thread-safe for single-writer / multi-reader usage. Concurrent
 * writes (e.g. refresh racing logout) are serialized externally by
 * the ApiClient's refresh mutex.
 */
class TokenStore(context: Context) {

    private val prefs: SharedPreferences = run {
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()

        val encrypted = EncryptedSharedPreferences.create(
            context,
            "initium_tokens",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM,
        )

        // First-launch wipe: if the install marker isn't set, clear any
        // residual keystore entries from a prior install.
        val marker = context.getSharedPreferences("initium_install", Context.MODE_PRIVATE)
        if (!marker.getBoolean(KEY_INSTALL_MARKER, false)) {
            encrypted.edit().clear().apply()
            marker.edit().putBoolean(KEY_INSTALL_MARKER, true).apply()
        }

        encrypted
    }

    fun save(tokens: TokenPair) {
        prefs.edit()
            .putString(KEY_ACCESS, tokens.accessToken)
            .putString(KEY_REFRESH, tokens.refreshToken)
            .apply()
    }

    fun accessToken(): String? = prefs.getString(KEY_ACCESS, null)

    fun refreshToken(): String? = prefs.getString(KEY_REFRESH, null)

    fun clear() {
        prefs.edit().remove(KEY_ACCESS).remove(KEY_REFRESH).apply()
    }

    companion object {
        private const val KEY_ACCESS = "access_token"
        private const val KEY_REFRESH = "refresh_token"
        private const val KEY_INSTALL_MARKER = "install_marker"
    }
}

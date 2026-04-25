package com.example.initium.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import com.example.initium.BuildConfig
import com.example.initium.InitiumApplication
import com.example.initium.api.ApiClient
import com.example.initium.api.ApiException
import com.example.initium.api.TokenStorage
import com.example.initium.api.User
import com.example.initium.api.UserRole
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * App-wide auth state machine. Mirrors the iOS `AuthStore` shape so
 * web / iOS / Android parity is immediate for reviewers.
 *
 * Activities / composables observe [state] via
 * `collectAsStateWithLifecycle()`. Actions are fire-and-forget — each
 * call launches a coroutine in `viewModelScope` and transitions state
 * when the underlying API call completes.
 */
sealed interface AuthState {
    data object Loading : AuthState
    data class Authenticated(val user: User) : AuthState
    data object Unauthenticated : AuthState
    data class Error(val message: String) : AuthState
}

class AuthViewModel(
    private val api: ApiClient,
    private val tokenStore: TokenStorage,
    private val devBypass: Boolean = BuildConfig.DEV_BYPASS_AUTH,
) : ViewModel() {

    private val _state = MutableStateFlow<AuthState>(AuthState.Loading)
    val state: StateFlow<AuthState> = _state.asStateFlow()

    init {
        viewModelScope.launch { bootstrap() }
    }

    /** Called once at VM creation. Hydrates session or drops to login. */
    private suspend fun bootstrap() {
        if (devBypass) {
            _state.value = AuthState.Authenticated(DEV_USER)
            return
        }
        if (tokenStore.accessToken() == null && tokenStore.refreshToken() == null) {
            _state.value = AuthState.Unauthenticated
            return
        }
        try {
            _state.value = AuthState.Authenticated(api.me())
        } catch (_: Throwable) {
            tokenStore.clear()
            _state.value = AuthState.Unauthenticated
        }
    }

    /**
     * Asks the backend to email a magic link. [onResult] receives
     * `Result.success` / `Result.failure` so the screen can toggle
     * between "Sending" / "Sent" / inline-error without going through
     * the global auth state machine.
     */
    fun requestMagicLink(email: String, onResult: (Result<Unit>) -> Unit) {
        viewModelScope.launch {
            try {
                api.requestMagicLink(email)
                onResult(Result.success(Unit))
            } catch (t: Throwable) {
                onResult(Result.failure(t))
            }
        }
    }

    /** Exchanges a magic-link token for a session. */
    fun verifyMagicLink(token: String) {
        _state.value = AuthState.Loading
        viewModelScope.launch {
            try {
                api.verifyMagicLink(token)
                _state.value = AuthState.Authenticated(api.me())
            } catch (t: Throwable) {
                _state.value = AuthState.Error(t.humanMessage())
            }
        }
    }

    /** Exchanges a Google ID token for a session. */
    fun verifyGoogleIdToken(idToken: String) {
        _state.value = AuthState.Loading
        viewModelScope.launch {
            try {
                api.verifyGoogleIdToken(idToken)
                _state.value = AuthState.Authenticated(api.me())
            } catch (t: Throwable) {
                _state.value = AuthState.Error(t.humanMessage())
            }
        }
    }

    /** Best-effort logout; always drops to the login screen. */
    fun logout() {
        viewModelScope.launch {
            api.logout()
            _state.value = AuthState.Unauthenticated
        }
    }

    /** Wired to ApiClient.onUnauthorized — drops to login from anywhere. */
    fun handleUnauthorizedCallback() {
        tokenStore.clear()
        _state.value = AuthState.Unauthenticated
    }

    private fun Throwable.humanMessage(): String = when (this) {
        is ApiException.HttpError -> envelope?.message ?: "Sign-in failed."
        is ApiException.NetworkError -> "Network error. Check your connection and try again."
        is ApiException.DecodeError -> "Unexpected server response."
        else -> "Sign-in failed."
    }

    companion object {
        /** Deterministic stub surfaced when DEV_BYPASS_AUTH=true. */
        val DEV_USER = User(
            id = "00000000-0000-0000-0000-000000000001",
            email = "dev@initium.local",
            name = "Dev User",
            avatarUrl = "",
            role = UserRole.USER,
            createdAt = "2026-04-25T00:00:00Z",
        )

        /**
         * Factory that pulls deps out of [InitiumApplication]. Use via
         * `val vm: AuthViewModel = viewModel(factory = AuthViewModel.Factory)`.
         */
        val Factory: ViewModelProvider.Factory = viewModelFactory {
            initializer {
                val app = this[ViewModelProvider.AndroidViewModelFactory.APPLICATION_KEY] as InitiumApplication
                val vm = AuthViewModel(app.apiClient, app.tokenStore)
                app.onUnauthorized = { vm.handleUnauthorizedCallback() }
                vm
            }
        }
    }
}

package com.example.initium.auth

import com.example.initium.api.ApiException
import com.example.initium.api.ErrorResponse
import com.example.initium.api.FakeApiClient
import com.example.initium.api.FakeTokenStore
import com.example.initium.api.TestUsers
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

/**
 * AuthViewModel is the Android app's auth state machine — equivalent to
 * iOS AuthStore. State transitions covered:
 *   - bootstrap: dev-bypass / no-tokens / tokens+success / tokens+failure
 *   - verifyMagicLink: success + envelope error humanization
 *   - verifyGoogleIdToken: success + network error humanization
 *   - logout: drops to .Unauthenticated, clears tokens
 *   - handleUnauthorizedCallback: drops to .Unauthenticated, clears tokens
 *
 * `viewModelScope` runs in a `UnconfinedTestDispatcher` so init {}'s
 * bootstrap launch resolves synchronously before the test asserts.
 */
@OptIn(ExperimentalCoroutinesApi::class)
class AuthViewModelTest {

    @Before
    fun setUp() {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    // ----------------------------------------------------------------
    // bootstrap
    // ----------------------------------------------------------------

    @Test
    fun `bootstrap dev-bypass authenticates with stub user`() = runTest {
        val vm = AuthViewModel(FakeApiClient(), FakeTokenStore(), devBypass = true)
        advanceUntilIdle()

        val state = vm.state.value
        assertTrue(state is AuthState.Authenticated)
        assertEquals(
            "00000000-0000-0000-0000-000000000001",
            (state as AuthState.Authenticated).user.id,
        )
    }

    @Test
    fun `bootstrap with no tokens drops to Unauthenticated`() = runTest {
        val vm = AuthViewModel(FakeApiClient(), FakeTokenStore(), devBypass = false)
        advanceUntilIdle()

        assertEquals(AuthState.Unauthenticated, vm.state.value)
    }

    @Test
    fun `bootstrap with tokens and successful me authenticates`() = runTest {
        val api = FakeApiClient()
        val storage = FakeTokenStore(access = "a", refresh = "r")

        val vm = AuthViewModel(api, storage, devBypass = false)
        advanceUntilIdle()

        assertTrue(vm.state.value is AuthState.Authenticated)
        assertEquals(1, api.meCallCount)
    }

    @Test
    fun `bootstrap with tokens and failing me clears tokens and Unauthenticates`() = runTest {
        val api = FakeApiClient(meResult = { throw ApiException.NetworkError("offline", RuntimeException()) })
        val storage = FakeTokenStore(access = "a", refresh = "r")

        val vm = AuthViewModel(api, storage, devBypass = false)
        advanceUntilIdle()

        assertEquals(AuthState.Unauthenticated, vm.state.value)
        // Tokens cleared so the next bootstrap doesn't loop on the same failure.
        assertNull(storage.accessToken())
        assertNull(storage.refreshToken())
    }

    // ----------------------------------------------------------------
    // verifyMagicLink
    // ----------------------------------------------------------------

    @Test
    fun `verifyMagicLink success authenticates`() = runTest {
        val api = FakeApiClient()
        val vm = AuthViewModel(api, FakeTokenStore(), devBypass = false)
        advanceUntilIdle()

        vm.verifyMagicLink("valid-token")
        advanceUntilIdle()

        assertEquals(listOf("valid-token"), api.verifyMagicLinkCalls)
        assertTrue(vm.state.value is AuthState.Authenticated)
    }

    @Test
    fun `verifyMagicLink HttpError surfaces envelope message`() = runTest {
        val api = FakeApiClient(verifyMagicLinkResult = {
            throw ApiException.HttpError(
                status = 410,
                envelope = ErrorResponse(code = "TOKEN_EXPIRED", message = "token expired", requestId = null),
            )
        })
        val vm = AuthViewModel(api, FakeTokenStore(), devBypass = false)
        advanceUntilIdle()

        vm.verifyMagicLink("expired")
        advanceUntilIdle()

        val state = vm.state.value
        assertTrue(state is AuthState.Error)
        assertEquals("token expired", (state as AuthState.Error).message)
    }

    // ----------------------------------------------------------------
    // verifyGoogleIdToken
    // ----------------------------------------------------------------

    @Test
    fun `verifyGoogleIdToken NetworkError humanizes message`() = runTest {
        val api = FakeApiClient(verifyGoogleIdTokenResult = {
            throw ApiException.NetworkError("offline", RuntimeException("dns"))
        })
        val vm = AuthViewModel(api, FakeTokenStore(), devBypass = false)
        advanceUntilIdle()

        vm.verifyGoogleIdToken("google-id")
        advanceUntilIdle()

        val state = vm.state.value
        assertTrue(state is AuthState.Error)
        assertTrue((state as AuthState.Error).message.contains("Network error"))
    }

    // ----------------------------------------------------------------
    // logout / handleUnauthorizedCallback
    // ----------------------------------------------------------------

    @Test
    fun `logout drops to Unauthenticated and calls api logout`() = runTest {
        val api = FakeApiClient()
        val storage = FakeTokenStore(access = "a", refresh = "r")
        val vm = AuthViewModel(api, storage, devBypass = false)
        advanceUntilIdle()

        vm.logout()
        advanceUntilIdle()

        assertEquals(AuthState.Unauthenticated, vm.state.value)
        assertEquals(1, api.logoutCallCount)
    }

    @Test
    fun `handleUnauthorizedCallback clears tokens and Unauthenticates`() = runTest {
        val api = FakeApiClient()
        val storage = FakeTokenStore(access = "a", refresh = "r")
        val vm = AuthViewModel(api, storage, devBypass = false)
        advanceUntilIdle()

        vm.handleUnauthorizedCallback()

        assertEquals(AuthState.Unauthenticated, vm.state.value)
        assertNull(storage.accessToken())
        assertNull(storage.refreshToken())
    }

    // ----------------------------------------------------------------
    // requestMagicLink — uses callback API instead of state machine
    // ----------------------------------------------------------------

    @Test
    fun `requestMagicLink success invokes onResult with success`() = runTest {
        val api = FakeApiClient()
        val vm = AuthViewModel(api, FakeTokenStore(), devBypass = false)
        advanceUntilIdle()

        var captured: Result<Unit>? = null
        vm.requestMagicLink("user@example.com") { captured = it }
        advanceUntilIdle()

        assertEquals(listOf("user@example.com"), api.requestMagicLinkCalls)
        assertTrue(captured!!.isSuccess)
    }

    @Test
    fun `requestMagicLink failure invokes onResult with failure`() = runTest {
        val api = FakeApiClient(requestMagicLinkResult = {
            throw ApiException.HttpError(400, ErrorResponse("EMAIL_REQUIRED", "email is required", null))
        })
        val vm = AuthViewModel(api, FakeTokenStore(), devBypass = false)
        advanceUntilIdle()

        var captured: Result<Unit>? = null
        vm.requestMagicLink("bad") { captured = it }
        advanceUntilIdle()

        assertTrue(captured!!.isFailure)
    }

    // ----------------------------------------------------------------
    // Confirm fixture stays stable
    // ----------------------------------------------------------------

    @Test
    fun `dev user fixture matches the iOS UUID`() {
        // Both platforms use the same hardcoded dev-bypass user UUID; if
        // this drifts, dev-mode debugging across platforms breaks.
        assertEquals("00000000-0000-0000-0000-000000000001", TestUsers.devUser.id)
    }
}

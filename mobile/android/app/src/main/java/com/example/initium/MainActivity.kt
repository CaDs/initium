package com.example.initium

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.adaptive.navigationsuite.NavigationSuiteScaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.testTag
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.initium.auth.AuthState
import com.example.initium.auth.AuthViewModel
import com.example.initium.ui.screens.LoginScreen
import com.example.initium.ui.theme.InitiumTheme

class MainActivity : ComponentActivity() {

    private lateinit var authViewModel: AuthViewModel

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            InitiumTheme {
                val vm: AuthViewModel = viewModel(factory = AuthViewModel.Factory)
                authViewModel = vm
                RootScreen(vm)

                // If the app was launched directly by a deep link
                // (cold start), consume the intent now.
                val launchIntent = intent
                launchIntent?.let { handleDeepLink(it, vm) }
            }
        }
    }

    /**
     * Warm-start deep link handoff. When the app is already running and
     * the OS routes a new intent to us (e.g. the user taps the magic
     * link in Mailpit), Android calls `onNewIntent` rather than
     * re-creating the activity.
     */
    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        if (this::authViewModel.isInitialized) {
            handleDeepLink(intent, authViewModel)
        }
    }

    private fun handleDeepLink(intent: Intent, vm: AuthViewModel) {
        val data: Uri = intent.data ?: return
        if (data.scheme != "initium" || data.host != "auth" || data.path != "/verify") return
        val token = data.getQueryParameter("token").orEmpty()
        if (token.isEmpty()) return
        vm.verifyMagicLink(token)
    }
}

/**
 * Root composable. Switches between the login screen and the
 * authenticated tab shell based on [AuthViewModel.state].
 */
@Composable
private fun RootScreen(viewModel: AuthViewModel) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    when (val current = state) {
        AuthState.Loading -> LoadingScreen()
        is AuthState.Authenticated -> TabsRoot(user = current.user, onLogout = viewModel::logout)
        AuthState.Unauthenticated -> LoginScreen(viewModel)
        is AuthState.Error -> LoginScreen(viewModel)
    }
}

@Composable
private fun LoadingScreen() {
    Box(
        modifier = Modifier.fillMaxSize().testTag("screen-loading"),
        contentAlignment = Alignment.Center,
    ) {
        CircularProgressIndicator()
    }
}

/**
 * The previously-top-level `NavigationSuiteScaffold`, now nested inside
 * the authenticated branch. Carries the selected tab through state so
 * rotation preserves position.
 */
@Composable
private fun TabsRoot(
    user: com.example.initium.api.User,
    onLogout: () -> Unit,
) {
    var currentTab by rememberSaveable { mutableStateOf(AppTab.HOME) }

    NavigationSuiteScaffold(
        navigationSuiteItems = {
            AppTab.entries.forEach { tab ->
                item(
                    icon = { Icon(tab.icon, contentDescription = tab.label) },
                    label = { Text(tab.label) },
                    selected = tab == currentTab,
                    onClick = { currentTab = tab },
                    modifier = Modifier.testTag("tab-${tab.name}"),
                )
            }
        }
    ) {
        Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
            TabContent(
                tab = currentTab,
                user = user,
                onLogout = onLogout,
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding),
            )
        }
    }
}

@Composable
private fun TabContent(
    tab: AppTab,
    user: com.example.initium.api.User,
    onLogout: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier.testTag("content-${tab.name}"),
        contentAlignment = Alignment.Center,
    ) {
        when (tab) {
            AppTab.HOME -> HomeTabContent(user = user, onLogout = onLogout)
            AppTab.MAIN -> Text(tab.label, style = MaterialTheme.typography.headlineMedium)
            AppTab.SETTINGS -> Text(tab.label, style = MaterialTheme.typography.headlineMedium)
        }
    }
}

@Composable
private fun HomeTabContent(
    user: com.example.initium.api.User,
    onLogout: () -> Unit,
) {
    com.example.initium.ui.screens.HomeProfile(user = user, onLogout = onLogout)
}

enum class AppTab(val label: String, val icon: ImageVector) {
    HOME(label = "Home", icon = Icons.Filled.Home),
    MAIN(label = "Main", icon = Icons.Filled.Star),
    SETTINGS(label = "Settings", icon = Icons.Filled.Settings),
}

package com.example.initium.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Email
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.example.initium.BuildConfig
import com.example.initium.auth.AuthState
import com.example.initium.auth.AuthViewModel

/**
 * Unauthenticated entry point. Offers:
 *  - Magic link by email (POST `/api/auth/magic-link`). Deep link
 *    routed back to `MainActivity.onNewIntent` completes the flow.
 *  - Google Sign-In (disabled in this PR — wired in the follow-up
 *    PR with Credential Manager).
 */
@Composable
fun LoginScreen(
    viewModel: AuthViewModel,
    modifier: Modifier = Modifier,
) {
    val authState by viewModel.state.collectAsState()
    var email by rememberSaveable { mutableStateOf("") }
    var magicLinkStatus by remember { mutableStateOf<MagicLinkStatus>(MagicLinkStatus.Idle) }

    val googleEnabled = BuildConfig.GOOGLE_SERVER_CLIENT_ID.isNotEmpty()

    Column(
        modifier = modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 24.dp, vertical = 40.dp),
        verticalArrangement = Arrangement.spacedBy(24.dp),
    ) {
        Text(
            text = "Welcome to Initium",
            style = MaterialTheme.typography.headlineMedium,
        )
        Text(
            text = "Sign in with Google or request a magic link by email.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        OutlinedButton(
            onClick = { /* wired in the Google Sign-In follow-up */ },
            enabled = googleEnabled,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Sign in with Google")
        }

        DividerWithLabel("or")

        OutlinedTextField(
            value = email,
            onValueChange = { email = it; if (magicLinkStatus is MagicLinkStatus.Failed) magicLinkStatus = MagicLinkStatus.Idle },
            label = { Text("Email") },
            leadingIcon = { Icon(Icons.Filled.Email, contentDescription = null) },
            singleLine = true,
            enabled = magicLinkStatus != MagicLinkStatus.Sending && magicLinkStatus != MagicLinkStatus.Sent,
            keyboardOptions = KeyboardOptions(
                keyboardType = KeyboardType.Email,
                imeAction = ImeAction.Send,
            ),
            modifier = Modifier.fillMaxWidth(),
        )

        Button(
            onClick = {
                magicLinkStatus = MagicLinkStatus.Sending
                viewModel.requestMagicLink(email) { result ->
                    magicLinkStatus = if (result.isSuccess) {
                        MagicLinkStatus.Sent
                    } else {
                        MagicLinkStatus.Failed(
                            result.exceptionOrNull()?.message ?: "Could not send magic link. Try again."
                        )
                    }
                }
            },
            enabled = canSubmit(email, magicLinkStatus),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(buttonLabel(magicLinkStatus))
        }

        when (val status = magicLinkStatus) {
            MagicLinkStatus.Sent -> SentBanner()
            is MagicLinkStatus.Failed -> ErrorBanner(status.message)
            MagicLinkStatus.Idle, MagicLinkStatus.Sending -> {}
        }

        // Surface an error transition from the global auth state (e.g.
        // a failed magic-link verify bubbling up from the deep link).
        (authState as? AuthState.Error)?.let { ErrorBanner(it.message) }
    }
}

@Composable
private fun DividerWithLabel(label: String) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier.fillMaxWidth(),
    ) {
        HorizontalDivider(modifier = Modifier.weight(1f))
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(horizontal = 12.dp),
        )
        HorizontalDivider(modifier = Modifier.weight(1f))
    }
}

@Composable
private fun SentBanner() {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.tertiaryContainer,
        ),
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text("Magic link sent", style = MaterialTheme.typography.titleSmall)
            Spacer(Modifier.height(4.dp))
            Text(
                "Check your inbox — the link expires in 15 minutes.",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onTertiaryContainer,
            )
        }
    }
}

@Composable
private fun ErrorBanner(message: String) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.errorContainer,
        ),
    ) {
        Text(
            text = message,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onErrorContainer,
            modifier = Modifier.padding(12.dp),
        )
    }
}

private sealed interface MagicLinkStatus {
    data object Idle : MagicLinkStatus
    data object Sending : MagicLinkStatus
    data object Sent : MagicLinkStatus
    data class Failed(val message: String) : MagicLinkStatus
}

private fun canSubmit(email: String, status: MagicLinkStatus): Boolean {
    if (status is MagicLinkStatus.Sending || status is MagicLinkStatus.Sent) return false
    return email.contains('@') && email.length > 3
}

private fun buttonLabel(status: MagicLinkStatus): String = when (status) {
    MagicLinkStatus.Idle, is MagicLinkStatus.Failed -> "Send Magic Link"
    MagicLinkStatus.Sending -> "Sending..."
    MagicLinkStatus.Sent -> "Sent"
}

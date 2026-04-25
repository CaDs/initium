package com.example.initium.ui.screens

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.initium.api.User

/**
 * Authenticated Home tab content. Shows the profile card mirroring
 * what `web/src/app/home/page.tsx` renders (email / name / role / id),
 * plus a logout button.
 *
 * Forks that want to keep the original "welcome card" demo screen can
 * add it above the profile card.
 */
@Composable
fun HomeProfile(
    user: User,
    onLogout: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Column(
        modifier = modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(24.dp),
        verticalArrangement = Arrangement.spacedBy(24.dp),
    ) {
        Text(
            text = "Home",
            style = MaterialTheme.typography.headlineMedium,
        )

        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surfaceContainerLow,
            ),
        ) {
            Column(modifier = Modifier.padding(20.dp)) {
                ProfileRow(label = "Email", value = user.email)
                HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
                ProfileRow(label = "Name", value = user.name.ifEmpty { "—" })
                HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
                ProfileRow(label = "Role", value = user.role.name.lowercase())
                HorizontalDivider(modifier = Modifier.padding(vertical = 12.dp))
                ProfileRow(label = "ID", value = user.id, monospace = true)
            }
        }

        OutlinedButton(
            onClick = onLogout,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text("Logout")
        }
    }
}

@Composable
private fun ProfileRow(label: String, value: String, monospace: Boolean = false) {
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Text(
            text = value,
            style = if (monospace) {
                MaterialTheme.typography.bodySmall.copy(fontFamily = androidx.compose.ui.text.font.FontFamily.Monospace)
            } else {
                MaterialTheme.typography.bodyMedium
            },
        )
    }
}

package com.example.initium

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
import androidx.compose.ui.tooling.preview.Preview
import com.example.initium.ui.theme.InitiumTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            InitiumTheme {
                InitiumApp()
            }
        }
    }
}

@Composable
fun InitiumApp() {
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
                modifier = Modifier
                    .fillMaxSize()
                    .padding(innerPadding),
            )
        }
    }
}

@Composable
fun TabContent(tab: AppTab, modifier: Modifier = Modifier) {
    Box(
        modifier = modifier.testTag("content-${tab.name}"),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            text = tab.label,
            style = MaterialTheme.typography.headlineMedium,
        )
    }
}

enum class AppTab(val label: String, val icon: ImageVector) {
    HOME(label = "Home", icon = Icons.Filled.Home),
    MAIN(label = "Main", icon = Icons.Filled.Star),
    SETTINGS(label = "Settings", icon = Icons.Filled.Settings),
}

@Preview(showBackground = true)
@Composable
private fun InitiumAppPreview() {
    InitiumTheme {
        InitiumApp()
    }
}

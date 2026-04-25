package com.example.initium

import androidx.compose.ui.test.assertIsDisplayed
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.performClick
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.example.initium.ui.theme.InitiumTheme
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class InitiumAppTabsTest {

    @get:Rule
    val composeTestRule = createComposeRule()

    @Test
    fun initialTabIsHome() {
        composeTestRule.setContent {
            InitiumTheme { InitiumApp() }
        }

        composeTestRule.onNodeWithTag("content-HOME").assertIsDisplayed()
    }

    @Test
    fun tappingMainTabShowsMainContent() {
        composeTestRule.setContent {
            InitiumTheme { InitiumApp() }
        }

        composeTestRule.onNodeWithTag("tab-MAIN").performClick()
        composeTestRule.onNodeWithTag("content-MAIN").assertIsDisplayed()
        composeTestRule.onNodeWithText("Main").assertIsDisplayed()
    }

    @Test
    fun tappingSettingsTabShowsSettingsContent() {
        composeTestRule.setContent {
            InitiumTheme { InitiumApp() }
        }

        composeTestRule.onNodeWithTag("tab-SETTINGS").performClick()
        composeTestRule.onNodeWithTag("content-SETTINGS").assertIsDisplayed()
        composeTestRule.onNodeWithText("Settings").assertIsDisplayed()
    }
}

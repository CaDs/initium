package com.example.initium

import org.junit.Test
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue

class AppTabUnitTest {

    @Test
    fun allTabsHaveDistinctLabels() {
        val labels = AppTab.entries.map { it.label }.toSet()
        assertEquals(AppTab.entries.size, labels.size)
    }

    @Test
    fun expectedTabsArePresent() {
        val names = AppTab.entries.map { it.name }
        assertTrue(names.contains("HOME"))
        assertTrue(names.contains("MAIN"))
        assertTrue(names.contains("SETTINGS"))
    }
}

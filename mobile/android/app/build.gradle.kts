plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
    jacoco
}

android {
    namespace = "com.example.initium"
    compileSdk {
        version = release(36) {
            minorApiLevel = 1
        }
    }

    defaultConfig {
        applicationId = "com.example.initium"
        minSdk = 24
        targetSdk = 36
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        // Local dev against the backend: Android emulator reaches the host
        // machine via 10.0.2.2 (not localhost). Override with
        //   ./gradlew installDebug -PAPI_BASE_URL=https://staging.example.com
        buildConfigField(
            "String",
            "API_BASE_URL",
            "\"${project.findProperty("API_BASE_URL") ?: "http://10.0.2.2:8000"}\""
        )
        // Mirrors backend DEV_BYPASS_AUTH — when true the app skips real auth
        // and renders a stub user. Hard-fail in release is enforced at runtime.
        buildConfigField(
            "Boolean",
            "DEV_BYPASS_AUTH",
            "${project.findProperty("DEV_BYPASS_AUTH") ?: "false"}"
        )
        // Google Cloud Console OAuth 2.0 Web Client ID — gates the Google
        // Sign-In button on LoginScreen (disabled when empty). The button
        // itself is a stub; the real Credential Manager wiring lands in the
        // follow-up PR that re-introduces androidx.credentials + googleid.
        buildConfigField(
            "String",
            "GOOGLE_SERVER_CLIENT_ID",
            "\"${project.findProperty("GOOGLE_SERVER_CLIENT_ID") ?: ""}\""
        )
    }

    buildTypes {
        debug {
            // Jacoco needs the test-coverage flag to capture .exec data.
            enableUnitTestCoverage = true
        }
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    buildFeatures {
        compose = true
        buildConfig = true
    }
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.graphics)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.material3.adaptive.navigation.suite)
    implementation(libs.androidx.compose.material.icons.core)

    // HTTP + JSON
    implementation(libs.okhttp)
    implementation(libs.okhttp.logging.interceptor)
    implementation(libs.moshi)
    implementation(libs.moshi.kotlin)

    // Secure token storage
    implementation(libs.androidx.security.crypto)

    // Coroutines
    implementation(libs.kotlinx.coroutines.android)

    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.mockwebserver)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.compose.ui.test.junit4)
    debugImplementation(libs.androidx.compose.ui.tooling)
    debugImplementation(libs.androidx.compose.ui.test.manifest)
}

// Jacoco unit-test coverage. Custom task instead of relying on AGP's
// generated `createDebugUnitTestCoverageReport` because that one's API
// is unstable across AGP versions. Reads the .exec file produced by
// `enableUnitTestCoverage = true` and emits HTML + XML.
//
// Excludes generated/Compose noise so the metric reflects the code
// humans wrote: BuildConfig, $$serializer, Compose lambdas, etc.
//
// Floor: 25% lines (phased ramp toward 80%). Below this, the suite
// fails — wired into `make test:android:coverage`.
tasks.register<JacocoReport>("jacocoTestReport") {
    dependsOn("testDebugUnitTest")
    group = "verification"
    description = "Generates Jacoco coverage from JVM unit tests."

    reports {
        html.required.set(true)
        xml.required.set(true)
    }

    val excludes = listOf(
        "**/R.class",
        "**/R$*.class",
        "**/BuildConfig.*",
        "**/Manifest*.*",
        "**/*Test*.*",
        "android/**/*.*",
        "**/*\$Lambda$*.*",
        "**/*\$inlined$*.*",
        "**/databinding/**",
        "**/generated/**",
        "**/ComposableSingletons*.class",
        "**/*Composable*.class",
    )

    classDirectories.setFrom(
        fileTree("${layout.buildDirectory.get()}/intermediates/built_in_kotlinc/debug/compileDebugKotlin/classes") {
            exclude(excludes)
        }
    )
    sourceDirectories.setFrom(files("src/main/java", "src/main/kotlin"))
    executionData.setFrom(
        fileTree(layout.buildDirectory) {
            include("outputs/unit_test_code_coverage/debugUnitTest/testDebugUnitTest.exec")
        }
    )
}

tasks.register<JacocoCoverageVerification>("jacocoCoverageVerification") {
    dependsOn("jacocoTestReport")
    group = "verification"
    description = "Fails the build if line coverage is below the floor."

    violationRules {
        rule {
            limit {
                counter = "LINE"
                minimum = "0.25".toBigDecimal()
            }
        }
    }

    val excludes = listOf(
        "**/R.class",
        "**/R$*.class",
        "**/BuildConfig.*",
        "**/Manifest*.*",
        "**/*Test*.*",
        "android/**/*.*",
        "**/databinding/**",
        "**/generated/**",
        "**/ComposableSingletons*.class",
        "**/*Composable*.class",
    )

    classDirectories.setFrom(
        fileTree("${layout.buildDirectory.get()}/intermediates/built_in_kotlinc/debug/compileDebugKotlin/classes") {
            exclude(excludes)
        }
    )
    sourceDirectories.setFrom(files("src/main/java", "src/main/kotlin"))
    executionData.setFrom(
        fileTree(layout.buildDirectory) {
            include("outputs/unit_test_code_coverage/debugUnitTest/testDebugUnitTest.exec")
        }
    )
}
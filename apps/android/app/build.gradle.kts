import java.util.Properties

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.roborazzi)
}

// Igual que Config/TuLocalWeb.xcconfig en iOS: la app no tiene secretos (la clave publicable de Clerk la da la API
// en /v1/auth/proveedores). Para apuntar a otra API sin tocar esto, poné tlw.api en local.properties (no se versiona).
val locales = Properties().apply {
    val f = rootProject.file("local.properties")
    if (f.exists()) f.inputStream().use { load(it) }
}
fun local(clave: String, porDefecto: String = ""): String = locales.getProperty(clave, porDefecto)
fun cadena(v: String) = "\"" + v.replace("\\", "\\\\").replace("\"", "\\\"") + "\""

android {
    namespace = "ar.com.tulocalweb.app"
    compileSdk = 36

    defaultConfig {
        applicationId = "ar.com.tulocalweb.app"
        minSdk = 26
        targetSdk = 36
        versionCode = (System.getenv("GITHUB_RUN_NUMBER")?.toIntOrNull() ?: 0) + 100
        versionName = "1.0.0"

        buildConfigField("String", "API", cadena(local("tlw.api", "https://api.tulocalweb.com.ar")))
        buildConfigField("String", "WEB", cadena(local("tlw.web", "https://tulocalweb.com.ar")))
        // Firebase sin google-services.json (como Tengo Libre y Caigo): si faltan estos datos, el push queda apagado.
        buildConfigField("String", "FIREBASE_APP_ID", cadena(local("tlw.firebase.appId")))
        buildConfigField("String", "FIREBASE_API_KEY", cadena(local("tlw.firebase.apiKey")))
        buildConfigField("String", "FIREBASE_PROYECTO", cadena(local("tlw.firebase.proyecto")))
        buildConfigField("String", "FIREBASE_REMITENTE", cadena(local("tlw.firebase.remitente")))
    }

    // Firma de release: la llave de subida vive fuera del repo (~/Trabajo/vimo/credenciales-android).
    signingConfigs {
        if (local("tlw.firma.archivo").isNotEmpty()) {
            create("subida") {
                storeFile = file(local("tlw.firma.archivo"))
                storePassword = local("tlw.firma.clave")
                keyAlias = local("tlw.firma.alias", "tulocalweb")
                keyPassword = local("tlw.firma.clave")
            }
        }
    }

    buildTypes {
        release {
            signingConfigs.findByName("subida")?.let { signingConfig = it }
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    buildFeatures {
        compose = true
        buildConfig = true
    }
    packaging {
        resources.excludes += "/META-INF/{AL2.0,LGPL2.1}"
    }
    testOptions {
        unitTests.isReturnDefaultValues = true
        unitTests.isIncludeAndroidResources = true
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

// Las capturas de PantallasTest (claro y oscuro) van a build/capturas: el CI las sube para revisarlas.
roborazzi {
    outputDir.set(layout.buildDirectory.dir("capturas"))
}

dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(libs.androidx.lifecycle.process)

    implementation(platform(libs.compose.bom))
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.graphics)
    implementation(libs.compose.foundation)
    implementation(libs.compose.ui.tooling.preview)
    debugImplementation(libs.compose.ui.tooling)
    debugImplementation(libs.compose.ui.test.manifest)

    implementation(libs.kotlinx.coroutines.android)
    implementation(libs.kotlinx.coroutines.play.services)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.okhttp)
    implementation(libs.coil.compose)
    implementation(libs.coil.network.okhttp)
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.tink.android)
    implementation(platform(libs.firebase.bom))
    implementation(libs.firebase.messaging)
    implementation(libs.androidx.browser)
    implementation(libs.glance.appwidget)
    implementation(libs.zxing.core)
    // Face ID → huella o cara del teléfono; el escáner de Venta con CameraX y el lector de códigos de Google.
    implementation(libs.androidx.biometric)
    implementation(libs.androidx.exifinterface)
    implementation(libs.camerax.camera2)
    implementation(libs.camerax.lifecycle)
    implementation(libs.camerax.view)
    implementation(libs.mlkit.barcode)

    testImplementation(libs.junit)
    testImplementation(libs.kotlinx.coroutines.test)
    testImplementation(libs.okhttp.mockwebserver)
    testImplementation(libs.robolectric)
    testImplementation(platform(libs.compose.bom))
    testImplementation(libs.compose.ui.test.junit4)
    testImplementation(libs.roborazzi)
    testImplementation(libs.roborazzi.compose)
    testImplementation(libs.roborazzi.junit.rule)
}

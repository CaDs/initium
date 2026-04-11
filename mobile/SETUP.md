# Mobile Platform Setup

## Google Sign-In Configuration

### Android

1. Go to [Firebase Console](https://console.firebase.google.com/) or [GCP Console](https://console.cloud.google.com/)
2. Create a project or select your existing one
3. Add an Android app with package name `dev.initium.mobile`
4. Download `google-services.json` and place it in `android/app/`
5. Register your debug SHA-1 fingerprint:
   ```bash
   cd android && ./gradlew signingReport
   ```
6. Copy the SHA-1 from the `debug` variant and add it to your Firebase/GCP app

### iOS

1. In the same Firebase/GCP project, add an iOS app with bundle ID `dev.initium.mobile`
2. Download `GoogleService-Info.plist` and place it in `ios/Runner/`
3. Add the reversed client ID as a URL scheme in `ios/Runner/Info.plist`:
   ```xml
   <key>CFBundleURLTypes</key>
   <array>
     <dict>
       <key>CFBundleURLSchemes</key>
       <array>
         <string>com.googleusercontent.apps.YOUR_CLIENT_ID</string>
       </array>
     </dict>
   </array>
   ```

## Environment Configuration

Copy `.env.example` to `.env` and set `API_BASE_URL` to your backend:

```bash
cp .env.example .env
```

Run with env config:
```bash
flutter run --dart-define-from-file=.env
```

## Dev Bypass Auth

Set `DEV_BYPASS_AUTH=true` in `.env` to skip authentication during development.
The backend must also have `DEV_BYPASS_AUTH=true` in its `.env`.

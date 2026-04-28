const trimSlash = (s: string): string => s.replace(/\/$/, '');

export const config = {
  apiBaseURL: trimSlash(process.env.EXPO_PUBLIC_API_BASE_URL || 'http://localhost:8000'),
  devBypassAuth: process.env.EXPO_PUBLIC_DEV_BYPASS_AUTH === 'true',
  google: {
    iosClientId: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID || '',
    androidClientId: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID || '',
    webClientId: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID || '',
  },
} as const;

export const googleConfigured = (): boolean =>
  config.google.iosClientId !== '' || config.google.androidClientId !== '';

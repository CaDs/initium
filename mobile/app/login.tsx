import * as Google from 'expo-auth-session/providers/google';
import { Redirect } from 'expo-router';
import * as WebBrowser from 'expo-web-browser';
import { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  Text,
  TextInput,
  View,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useAuth, useAuthStatus } from '@/auth/useAuth';
import { config, googleConfigured } from '@/config';

WebBrowser.maybeCompleteAuthSession();

export default function LoginScreen() {
  const status = useAuthStatus();
  const { requestMagicLink, error } = useAuth();
  const [email, setEmail] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [sent, setSent] = useState(false);

  if (status === 'authenticated') return <Redirect href="/(tabs)/home" />;

  const onSubmit = async () => {
    if (!email) return;
    setSubmitting(true);
    setSent(false);
    try {
      await requestMagicLink(email);
      setSent(true);
    } catch {
      // error surfaced via store.error
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-white dark:bg-black">
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        className="flex-1 items-center justify-center px-6"
      >
        <Text className="text-3xl font-semibold text-black dark:text-white">Initium</Text>
        <Text className="mt-2 text-base text-neutral-500">Sign in to continue</Text>

        <View className="mt-10 w-full max-w-sm">
          <TextInput
            autoCapitalize="none"
            autoCorrect={false}
            keyboardType="email-address"
            placeholder="you@example.com"
            placeholderTextColor="#9ca3af"
            value={email}
            onChangeText={setEmail}
            className="rounded-xl border border-neutral-300 bg-white px-4 py-3 text-base text-black dark:border-neutral-700 dark:bg-neutral-900 dark:text-white"
          />

          <Pressable
            onPress={onSubmit}
            disabled={submitting || email.length === 0}
            accessibilityRole="button"
            className="mt-3 items-center rounded-xl bg-black px-4 py-3 disabled:opacity-50 dark:bg-white"
          >
            {submitting ? (
              <ActivityIndicator />
            ) : (
              <Text className="text-base font-medium text-white dark:text-black">
                Email magic link
              </Text>
            )}
          </Pressable>

          {sent && (
            <Text className="mt-3 text-center text-sm text-emerald-600">
              Check your inbox for the sign-in link.
            </Text>
          )}
          {error && (
            <Text className="mt-3 text-center text-sm text-red-600">{error}</Text>
          )}

          <View className="my-6 h-px bg-neutral-200 dark:bg-neutral-800" />

          {googleConfigured() ? <GoogleSignInButton /> : <GoogleDisabledButton />}
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

function GoogleSignInButton() {
  const { verifyGoogle } = useAuth();
  const [, response, promptAsync] = Google.useAuthRequest({
    iosClientId: config.google.iosClientId,
    androidClientId: config.google.androidClientId,
    webClientId: config.google.webClientId,
  });

  useEffect(() => {
    if (response?.type === 'success' && response.params.id_token) {
      void verifyGoogle(response.params.id_token);
    }
  }, [response, verifyGoogle]);

  return (
    <Pressable
      onPress={() => void promptAsync()}
      accessibilityRole="button"
      className="items-center rounded-xl border border-neutral-300 px-4 py-3 dark:border-neutral-700"
    >
      <Text className="text-base font-medium text-black dark:text-white">
        Continue with Google
      </Text>
    </Pressable>
  );
}

function GoogleDisabledButton() {
  return (
    <Pressable
      disabled
      accessibilityRole="button"
      accessibilityState={{ disabled: true }}
      className="items-center rounded-xl border border-neutral-300 px-4 py-3 opacity-50 dark:border-neutral-700"
    >
      <Text className="text-center text-base font-medium text-black dark:text-white">
        Configure GOOGLE_*_CLIENT_ID to enable Google
      </Text>
    </Pressable>
  );
}

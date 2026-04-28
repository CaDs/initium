import { Redirect, useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { ActivityIndicator, Text, View } from 'react-native';

import { useAuth, useAuthStatus } from '@/auth/useAuth';

export default function VerifyScreen() {
  const { token } = useLocalSearchParams<{ token?: string }>();
  const status = useAuthStatus();
  const { verifyMagicLink } = useAuth();
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!token) {
      setFailed(true);
      return;
    }
    void verifyMagicLink(token).catch(() => setFailed(true));
  }, [token, verifyMagicLink]);

  if (status === 'authenticated') return <Redirect href="/(tabs)/home" />;
  if (failed) return <Redirect href="/login" />;

  return (
    <View className="flex-1 items-center justify-center bg-white dark:bg-black">
      <ActivityIndicator />
      <Text className="mt-3 text-sm text-neutral-500">Verifying your sign-in link…</Text>
    </View>
  );
}

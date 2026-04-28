import '../global.css';

import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import { ActivityIndicator, View } from 'react-native';

import { useAuth } from '@/auth/useAuth';

export default function RootLayout() {
  const { status, bootstrap } = useAuth();

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  if (status === 'loading') {
    return (
      <View className="flex-1 items-center justify-center bg-white dark:bg-black">
        <ActivityIndicator />
      </View>
    );
  }

  return (
    <>
      <StatusBar style="auto" />
      <Stack screenOptions={{ headerShown: false }} />
    </>
  );
}

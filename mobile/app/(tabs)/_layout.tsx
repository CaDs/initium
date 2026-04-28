import { Redirect, Tabs } from 'expo-router';

import { useAuthStatus } from '@/auth/useAuth';

export default function TabsLayout() {
  const status = useAuthStatus();
  if (status !== 'authenticated') return <Redirect href="/login" />;

  return (
    <Tabs screenOptions={{ headerShown: false }}>
      <Tabs.Screen name="home" options={{ title: 'Home' }} />
      <Tabs.Screen name="main" options={{ title: 'Main' }} />
      <Tabs.Screen name="settings" options={{ title: 'Settings' }} />
    </Tabs>
  );
}

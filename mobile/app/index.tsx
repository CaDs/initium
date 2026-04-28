import { Redirect } from 'expo-router';

import { useAuthStatus } from '@/auth/useAuth';

export default function Index() {
  const status = useAuthStatus();
  if (status === 'authenticated') return <Redirect href="/(tabs)/home" />;
  return <Redirect href="/login" />;
}

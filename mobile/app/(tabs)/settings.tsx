import { Pressable, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { useAuth, useUser } from '@/auth/useAuth';

export default function SettingsScreen() {
  const user = useUser();
  const { logout } = useAuth();

  return (
    <SafeAreaView className="flex-1 bg-white dark:bg-black">
      <View className="flex-1 px-5 py-6">
        <Text className="text-3xl font-semibold text-black dark:text-white">Settings</Text>
        {user && (
          <Text className="mt-2 text-sm text-neutral-500">Signed in as {user.email}</Text>
        )}

        <Pressable
          onPress={() => void logout()}
          accessibilityRole="button"
          className="mt-8 items-center rounded-xl border border-red-300 bg-red-50 px-4 py-3 dark:border-red-900 dark:bg-red-950"
        >
          <Text className="text-base font-medium text-red-700 dark:text-red-300">
            Log out
          </Text>
        </Pressable>
      </View>
    </SafeAreaView>
  );
}

import { Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { LiquidCard } from '@/ui/LiquidCard';
import { useUser } from '@/auth/useAuth';

export default function HomeScreen() {
  const user = useUser();

  return (
    <SafeAreaView className="flex-1 bg-white dark:bg-black">
      <View className="flex-1 px-5 py-6">
        <Text className="text-3xl font-semibold text-black dark:text-white">Home</Text>
        <Text className="mt-1 text-sm text-neutral-500">Signed-in profile</Text>

        <LiquidCard className="mt-6 border border-neutral-200 dark:border-neutral-800">
          <Row label="Email" value={user?.email ?? '—'} />
          <Row label="Name" value={user?.name ?? '—'} />
          <Row label="Role" value={user?.role ?? '—'} />
          <Row label="ID" value={user?.id ?? '—'} mono />
        </LiquidCard>
      </View>
    </SafeAreaView>
  );
}

function Row({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <View className="flex-row justify-between border-b border-neutral-200/60 py-2 last:border-b-0 dark:border-neutral-800/60">
      <Text className="text-sm text-neutral-500">{label}</Text>
      <Text
        className={`text-sm text-black dark:text-white ${mono ? 'font-mono' : 'font-medium'}`}
      >
        {value}
      </Text>
    </View>
  );
}

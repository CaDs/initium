import { Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

export default function MainScreen() {
  return (
    <SafeAreaView className="flex-1 bg-white dark:bg-black">
      <View className="flex-1 items-center justify-center px-5">
        <Text className="text-2xl font-semibold text-black dark:text-white">Main</Text>
        <Text className="mt-2 text-sm text-neutral-500">
          Replace this tab with the first feature you build.
        </Text>
      </View>
    </SafeAreaView>
  );
}

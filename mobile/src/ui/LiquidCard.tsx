import { BlurView } from 'expo-blur';
import type { ReactNode } from 'react';
import { View } from 'react-native';

type Props = {
  children: ReactNode;
  className?: string;
};

export function LiquidCard({ children, className }: Props) {
  return (
    <View className={`overflow-hidden rounded-2xl ${className ?? ''}`}>
      <BlurView intensity={40} tint="default" className="p-5">
        {children}
      </BlurView>
    </View>
  );
}

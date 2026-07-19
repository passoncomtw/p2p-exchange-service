import React, { useEffect, useRef } from 'react';
import { View, Animated, StyleSheet } from 'react-native';
import { colors } from '@/theme';

type Props = {
  count?: number;
  height?: number;
};

export default function SkeletonList({ count = 3, height = 96 }: Props) {
  const anim = useRef(new Animated.Value(0.3)).current;

  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(anim, { toValue: 0.85, duration: 700, useNativeDriver: true }),
        Animated.timing(anim, { toValue: 0.3, duration: 700, useNativeDriver: true }),
      ])
    ).start();
    return () => anim.stopAnimation();
  }, [anim]);

  return (
    <View style={styles.container}>
      {Array.from({ length: count }).map((_, i) => (
        <Animated.View key={i} style={[styles.skeleton, { height, opacity: anim }]} />
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { gap: 12 },
  skeleton: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
});

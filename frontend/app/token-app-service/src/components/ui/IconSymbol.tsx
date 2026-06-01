/**
 * 統一圖標組件 - 所有平台使用相同的 Ionicons
 * 不需要維護映射表，直接使用 Ionicons 名稱
 * 
 * 完整圖標列表：https://icons.expo.fyi/Index/Ionicons
 */

import Ionicons from '@expo/vector-icons/Ionicons';
import { ComponentProps } from 'react';

export type IconName = ComponentProps<typeof Ionicons>['name'];

export function IconSymbol({
  name,
  size = 24,
  color,
  style,
}: {
  name: IconName;
  size?: number;
  color: string;
  style?: any;
}) {
  return <Ionicons name={name} size={size} color={color} style={style} />;
}

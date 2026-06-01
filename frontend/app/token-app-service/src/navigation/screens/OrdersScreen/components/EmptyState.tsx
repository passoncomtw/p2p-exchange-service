/**
 * EmptyState - 掛單空狀態組件
 */

import React from 'react';
import { View, Text, StyleSheet, Pressable } from 'react-native';

interface EmptyStateProps {
  onCreateOrder: () => void;
}

export default function EmptyState({ onCreateOrder }: EmptyStateProps) {
  return (
    <View style={styles.container}>
      <View style={styles.emptyContainer}>
        <Text style={styles.emptyIcon}>📋</Text>
        <Text style={styles.emptyText}>尚無掛單</Text>
        <Pressable 
          style={({ pressed }) => [
            styles.createButton,
            pressed && styles.createButtonPressed,
          ]}
          onPress={onCreateOrder}
        >
          <Text style={styles.createButtonText}>新增掛單</Text>
        </Pressable>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 40,
  },
  emptyContainer: {
    alignItems: 'center',
    width: '100%',
  },
  emptyIcon: {
    fontSize: 80,
    marginBottom: 24,
    opacity: 0.5,
  },
  emptyText: {
    fontSize: 18,
    color: '#999',
    marginBottom: 32,
    fontWeight: '500',
  },
  createButton: {
    backgroundColor: '#007AFF',
    paddingVertical: 14,
    paddingHorizontal: 40,
    borderRadius: 8,
    width: 200,
    alignItems: 'center',
    shadowColor: '#007AFF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  createButtonPressed: {
    transform: [{ scale: 0.98 }],
    opacity: 0.9,
  },
  createButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});


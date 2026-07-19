import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export type NotificationType = 'success' | 'error' | 'info' | 'warning';

export interface Notification {
  id: string;
  type: NotificationType;
  message: string;
  title?: string;
}

interface NotificationState {
  queue: Notification[];
  nextId: number;
}

const notificationSlice = createSlice({
  name: 'notification',
  initialState: { queue: [], nextId: 0 } as NotificationState,
  reducers: {
    pushNotification(
      state,
      action: PayloadAction<{ type: NotificationType; message: string; title?: string }>,
    ) {
      state.queue.push({ id: String(state.nextId), ...action.payload });
      state.nextId += 1;
    },
    popNotification(state) {
      state.queue.shift();
    },
  },
});

export const { pushNotification, popNotification } = notificationSlice.actions;
export default notificationSlice.reducer;

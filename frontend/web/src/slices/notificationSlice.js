import { createSlice } from '@reduxjs/toolkit'

const notificationSlice = createSlice({
  name: 'notification',
  initialState: { queue: [], nextId: 0 },
  reducers: {
    pushNotification(state, { payload }) {
      state.queue.push({ id: String(state.nextId), ...payload })
      state.nextId += 1
    },
    popNotification(state) {
      state.queue.shift()
    },
  },
})

export const { pushNotification, popNotification } = notificationSlice.actions
export default notificationSlice.reducer

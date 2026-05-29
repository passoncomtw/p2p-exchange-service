import { createSlice } from '@reduxjs/toolkit'

const authSlice = createSlice({
  name: 'auth',
  initialState: {
    isAuth: false,
    token: null,
    username: null,
    error: null,
  },
  reducers: {
    signIn: () => {},         // saga trigger
    signInSuccess: (state, { payload }) => {
      state.isAuth = true
      state.token = payload.token
      state.username = payload.username
      state.error = null
      localStorage.setItem('token', payload.token)
    },
    signInError: (state, { payload }) => {
      state.isAuth = false
      state.token = null
      state.error = payload
    },
    signOut: (state) => {
      state.isAuth = false
      state.token = null
      state.username = null
      state.error = null
      localStorage.removeItem('token')
    },
  },
})

export const { signIn, signInSuccess, signInError, signOut } = authSlice.actions
export default authSlice.reducer

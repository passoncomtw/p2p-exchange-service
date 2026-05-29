import server from 'src/apis/index'

export const loginApi = (payload) =>
  server.post('/backend/auth/login', payload)

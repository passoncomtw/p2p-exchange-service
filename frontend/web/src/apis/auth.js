import server from './';

export const loginResult = (payload) =>
  server.post('/backend/auth/login', payload);

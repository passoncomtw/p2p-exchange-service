import server from './index'

export const membersApi = {
  list: (params) => server.get('/backend/members', { params }).then((r) => r.data.data),
}

import server from './index'

export const fiatWithdrawalsApi = {
  list: (params) =>
    server.get('/backend/fiat-withdrawals', { params }).then((r) => r.data.data),

  review: (id, action, reason) =>
    server.put(`/backend/fiat-withdrawals/${id}/review`, { id, action, reason }).then((r) => r.data),
}

import server from './index'

export const ordersApi = {
  list: (params) => server.get('/backend/orders', { params }).then((r) => r.data.data),
  resolve: (id, action, reason) =>
    server.post(`/backend/orders/${id}/resolve`, { action, reason }).then((r) => r.data),
}

import server from './index'

export const ordersApi = {
  list: (params) => server.get('/backend/orders', { params }).then((r) => r.data.data),
}

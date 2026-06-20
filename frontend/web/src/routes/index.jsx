import { lazy } from 'react'
import { useRoutes, Navigate } from 'react-router-dom'
import Loadable from 'src/components/Loadable'
import MainLayout from 'src/layout/MainLayout'
import UserLayout from 'src/layout/UserLayout'

// 使用者端
const CreateOrderScreen = Loadable(lazy(() => import('src/screens/CreateOrderScreen')))
const MyOrdersScreen = Loadable(lazy(() => import('src/screens/MyOrdersScreen')))

// 後台
const AdminOrdersScreen = Loadable(lazy(() => import('src/screens/AdminOrdersScreen')))
const AdminOrderDetailScreen = Loadable(lazy(() => import('src/screens/AdminOrderDetailScreen')))

// v1 無登入：使用者端與後台皆免登入；目前使用者固定 demo_user。
export default function AppRoutes() {
  return useRoutes([
    // 使用者端
    {
      path: '/',
      element: <UserLayout />,
      children: [
        { index: true, element: <CreateOrderScreen /> },
        { path: 'my-orders', element: <MyOrdersScreen /> },
      ],
    },

    // 後台訂單管理（沿用既有 MainLayout 視覺）
    {
      path: '/admin',
      element: <MainLayout />,
      children: [
        { index: true, element: <AdminOrdersScreen /> },
        { path: 'orders/:id', element: <AdminOrderDetailScreen /> },
      ],
    },

    { path: '*', element: <Navigate to="/" replace /> },
  ])
}

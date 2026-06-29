import { lazy } from 'react'
import { useRoutes, Navigate } from 'react-router-dom'
import Loadable from 'src/components/Loadable'
import ProtectedRoute from 'src/components/ProtectedRoute'
import MainLayout from 'src/layout/MainLayout'

const LoginScreen = Loadable(lazy(() => import('src/screens/LoginScreen')))
const DashboardScreen = Loadable(lazy(() => import('src/screens/DashboardScreen')))
const MemberListScreen = Loadable(lazy(() => import('src/screens/MemberListScreen')))
const OrderListScreen = Loadable(lazy(() => import('src/screens/OrderListScreen')))

export default function AppRoutes() {
  return useRoutes([
    // 公開路由
    {
      path: '/login',
      element: <LoginScreen />,
    },

    // 需要登入的路由，套用 MainLayout
    {
      path: '/',
      element: (
        <ProtectedRoute>
          <MainLayout />
        </ProtectedRoute>
      ),
      children: [
        { index: true, element: <DashboardScreen /> },
        { path: 'members', element: <MemberListScreen /> },
        { path: 'orders', element: <OrderListScreen /> },
      ],
    },

    { path: '*', element: <Navigate to="/" replace /> },
  ])
}

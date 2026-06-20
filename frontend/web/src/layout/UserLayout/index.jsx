import { NavLink, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Box, Typography } from '@mui/material'
import OwlLogo from 'src/components/OwlLogo'
import { tokens } from 'src/theme'

const navLinkStyle = ({ isActive }) => ({
  fontSize: '14px',
  fontWeight: isActive ? 600 : 400,
  color: isActive ? tokens.textPrimary : tokens.textSecondary,
  textDecoration: 'none',
  borderBottom: isActive ? `2px solid ${tokens.primary}` : '2px solid transparent',
  paddingBottom: '4px',
})

// 使用者端版型：頂部品牌列 + 導覽（掛單 / 我的掛單），內容區置中。
const UserLayout = () => {
  const { t } = useTranslation()

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: tokens.bgContent }}>
      <Box
        sx={{
          height: '56px',
          bgcolor: tokens.bgCard,
          borderBottom: `1px solid ${tokens.borderCard}`,
          display: 'flex',
          alignItems: 'center',
          gap: '24px',
          px: '24px',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <OwlLogo size={32} />
          <Typography sx={{ fontSize: '16px', fontWeight: 700, color: tokens.textPrimary }}>
            {t('sidebar.brand')}
          </Typography>
        </Box>

        <Box sx={{ display: 'flex', alignItems: 'center', gap: '20px', flex: 1 }}>
          <NavLink to="/" end style={navLinkStyle}>
            {t('order.nav.create')}
          </NavLink>
          <NavLink to="/my-orders" style={navLinkStyle}>
            {t('order.nav.myOrders')}
          </NavLink>
        </Box>

        <NavLink to="/admin" style={{ fontSize: '13px', color: tokens.textTertiary, textDecoration: 'none' }}>
          {t('order.nav.admin')}
        </NavLink>
      </Box>

      <Box component="main" sx={{ p: '24px', maxWidth: 720, mx: 'auto' }}>
        <Outlet />
      </Box>
    </Box>
  )
}

export default UserLayout

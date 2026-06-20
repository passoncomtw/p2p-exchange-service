import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Box, Collapse, Typography } from '@mui/material'
import {
  ViewList as ViewListIcon,
  ExpandLess,
  ExpandMore,
} from '@mui/icons-material'
import OwlLogo from 'src/components/OwlLogo'
import { tokens } from 'src/theme'

export const SIDEBAR_WIDTH = 160

// v1 後台僅有訂單管理一個項目。
const buildMenuItems = (t) => [
  {
    key: '/admin',
    label: t('order.nav.admin'),
    icon: <ViewListIcon sx={{ fontSize: 18 }} />,
  },
]

const NavItem = ({ item, depth = 0 }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const hasChildren = Boolean(item.children?.length)
  const isActive = !hasChildren && location.pathname === item.key

  const childPaths = hasChildren ? item.children.map((c) => c.key) : []
  const [open, setOpen] = useState(() => childPaths.includes(location.pathname))

  const handleClick = () => {
    if (hasChildren) setOpen((v) => !v)
    else navigate(item.key)
  }

  return (
    <>
      <Box
        onClick={handleClick}
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: '10px',
          px: depth === 0 ? '20px' : '44px',
          height: '44px',
          cursor: 'pointer',
          bgcolor: isActive ? tokens.sidebarActive : 'transparent',
          color: isActive ? tokens.sidebarTextActive : tokens.sidebarText,
          fontWeight: isActive ? 600 : 400,
          fontSize: depth === 0 ? '14px' : '13px',
          '&:hover': { bgcolor: isActive ? tokens.sidebarActive : tokens.sidebarHover },
          userSelect: 'none',
        }}
      >
        {depth === 0 && item.icon}
        <Typography sx={{ fontSize: 'inherit', fontWeight: 'inherit', color: 'inherit', flex: 1, lineHeight: 1 }}>
          {item.label}
        </Typography>
        {hasChildren && (
          open
            ? <ExpandLess sx={{ fontSize: 16, color: tokens.sidebarText }} />
            : <ExpandMore sx={{ fontSize: 16, color: tokens.sidebarText }} />
        )}
      </Box>

      {hasChildren && (
        <Collapse in={open} timeout="auto" unmountOnExit>
          {item.children.map((child) => (
            <NavItem key={child.key} item={child} depth={1} />
          ))}
        </Collapse>
      )}
    </>
  )
}

const Sidebar = () => {
  const { t } = useTranslation()
  const menuItems = buildMenuItems(t)

  return (
    <Box
      sx={{
        width: `${SIDEBAR_WIDTH}px`,
        minWidth: `${SIDEBAR_WIDTH}px`,
        bgcolor: tokens.sidebarBg,
        display: 'flex',
        flexDirection: 'column',
        height: '100vh',
        position: 'fixed',
        top: 0,
        left: 0,
        zIndex: 100,
        overflowY: 'auto',
        overflowX: 'hidden',
        '&::-webkit-scrollbar': { width: '4px' },
        '&::-webkit-scrollbar-thumb': { bgcolor: tokens.scrollbarThumb },
      }}
    >
      {/* Logo */}
      <Box
        sx={{
          height: '56px',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          px: '16px',
          flexShrink: 0,
          borderBottom: `1px solid ${tokens.borderSidebarLogo}`,
        }}
      >
        <OwlLogo size={32} />
        <Box>
          <Typography sx={{ fontSize: '13px', fontWeight: 700, color: tokens.sidebarTextActive, lineHeight: 1.2 }}>
            {t('sidebar.brand')}
          </Typography>
          <Typography sx={{ fontSize: '10px', color: tokens.primary, lineHeight: 1.4 }}>
            {t('sidebar.tagline')}
          </Typography>
        </Box>
      </Box>

      {/* 導航 */}
      <Box sx={{ flex: 1, pt: '8px' }}>
        {menuItems.map((item) => (
          <NavItem key={item.key} item={item} />
        ))}
      </Box>
    </Box>
  )
}

export default Sidebar

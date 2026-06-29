import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Box, Collapse, Typography } from '@mui/material'
import {
  Person as PersonIcon,
  ViewList as ViewListIcon,
  ExpandLess,
  ExpandMore,
} from '@mui/icons-material'

export const SIDEBAR_WIDTH = 160

const SIDEBAR_BG = '#2C2C3C'
const SIDEBAR_HOVER = '#3A3A4E'
const SIDEBAR_ACTIVE = '#1A1A28'
const SIDEBAR_TEXT = '#CCCCCC'
const SIDEBAR_TEXT_ACTIVE = '#FFFFFF'

const OwlLogo = () => (
  <svg width="32" height="32" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect width="48" height="48" rx="8" fill="#FFC107" />
    <circle cx="17" cy="22" r="7" fill="white" />
    <circle cx="31" cy="22" r="7" fill="white" />
    <circle cx="17" cy="22" r="4" fill="#333" />
    <circle cx="31" cy="22" r="4" fill="#333" />
    <circle cx="18" cy="21" r="1.5" fill="white" />
    <circle cx="32" cy="21" r="1.5" fill="white" />
    <path d="M20 30 Q24 34 28 30" stroke="#333" strokeWidth="1.5" strokeLinecap="round" fill="none" />
    <polygon points="24,12 21,17 27,17" fill="#FF8F00" />
  </svg>
)

const buildMenuItems = (t) => [
  {
    key: 'members',
    label: t('sidebar.menu.members'),
    icon: <PersonIcon sx={{ fontSize: 18 }} />,
    children: [
      { key: '/members', label: t('sidebar.menu.memberList') },
    ],
  },
  {
    key: '/orders',
    label: t('sidebar.menu.orders'),
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
          bgcolor: isActive ? SIDEBAR_ACTIVE : 'transparent',
          color: isActive ? SIDEBAR_TEXT_ACTIVE : SIDEBAR_TEXT,
          fontWeight: isActive ? 600 : 400,
          fontSize: depth === 0 ? '14px' : '13px',
          '&:hover': { bgcolor: isActive ? SIDEBAR_ACTIVE : SIDEBAR_HOVER },
          userSelect: 'none',
        }}
      >
        {depth === 0 && item.icon}
        <Typography sx={{ fontSize: 'inherit', fontWeight: 'inherit', color: 'inherit', flex: 1, lineHeight: 1 }}>
          {item.label}
        </Typography>
        {hasChildren && (
          open
            ? <ExpandLess sx={{ fontSize: 16, color: SIDEBAR_TEXT }} />
            : <ExpandMore sx={{ fontSize: 16, color: SIDEBAR_TEXT }} />
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
        bgcolor: SIDEBAR_BG,
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
        '&::-webkit-scrollbar-thumb': { bgcolor: '#444' },
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
          borderBottom: '1px solid rgba(255,255,255,0.06)',
        }}
      >
        <OwlLogo />
        <Box>
          <Typography sx={{ fontSize: '13px', fontWeight: 700, color: '#FFF', lineHeight: 1.2 }}>
            {t('sidebar.brand')}
          </Typography>
          <Typography sx={{ fontSize: '10px', color: '#FFC107', lineHeight: 1.4 }}>
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

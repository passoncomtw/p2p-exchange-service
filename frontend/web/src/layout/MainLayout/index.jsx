import { useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Avatar,
  Box,
  Divider,
  Menu,
  MenuItem,
  Select,
  Typography,
} from '@mui/material'
import { KeyboardArrowDown as ArrowDownIcon, Logout as LogoutIcon } from '@mui/icons-material'
import Sidebar, { SIDEBAR_WIDTH } from 'src/layout/Sidebar'
import { signOut } from 'src/slices/authSlice'
import { changeLanguage, SUPPORTED_LANGS } from 'src/i18n'
import { tokens } from 'src/theme'

const announceLanguageChange = (lang) => {
  const region = document.getElementById('aria-live-region')
  if (!region) return
  const msgs = { 'zh-TW': '語言已切換', 'zh-CN': '语言已切换' }
  region.textContent = msgs[lang] ?? '語言已切換'
}

const Header = () => {
  const dispatch = useDispatch()
  const { t, i18n } = useTranslation()
  const { username } = useSelector((state) => state.auth)
  const [anchorEl, setAnchorEl] = useState(null)

  const initial = username ? username[0].toUpperCase() : 'A'

  return (
    <Box
      sx={{
        height: '56px',
        bgcolor: tokens.bgCard,
        borderBottom: `1px solid ${tokens.borderCard}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'flex-end',
        gap: '16px',
        px: '24px',
        flexShrink: 0,
      }}
    >
      {/* 語系切換 */}
      <Select
        value={i18n.language}
        onChange={(e) => { changeLanguage(e.target.value); announceLanguageChange(e.target.value) }}
        size="small"
        sx={{
          fontSize: '12px',
          '& .MuiOutlinedInput-notchedOutline': { borderColor: tokens.borderSelect },
          '& .MuiSelect-select': { py: '6px', px: '10px' },
        }}
      >
        {SUPPORTED_LANGS.map((lang) => (
          <MenuItem key={lang.code} value={lang.code} sx={{ fontSize: '12px' }}>
            {lang.label}
          </MenuItem>
        ))}
      </Select>

      {/* 帳號下拉 */}
      <Box
        onClick={(e) => setAnchorEl(e.currentTarget)}
        sx={{
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          cursor: 'pointer',
          '&:hover': { opacity: 0.8 },
        }}
      >
        <Avatar sx={{ width: 28, height: 28, fontSize: '13px', fontWeight: 600, bgcolor: tokens.textAvatar }}>
          {initial}
        </Avatar>
        <Typography sx={{ fontSize: '13px', color: tokens.textPrimary }}>{username}</Typography>
        <ArrowDownIcon sx={{ fontSize: 16, color: tokens.textTertiary }} />
      </Box>

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
        PaperProps={{ sx: { minWidth: 140, mt: '4px' } }}
        transformOrigin={{ horizontal: 'right', vertical: 'top' }}
        anchorOrigin={{ horizontal: 'right', vertical: 'bottom' }}
      >
        <MenuItem disabled sx={{ fontSize: '13px', color: tokens.textTertiary, opacity: '1 !important' }}>
          {username}
        </MenuItem>
        <Divider />
        <MenuItem
          onClick={() => { setAnchorEl(null); dispatch(signOut()) }}
          sx={{ fontSize: '13px', color: tokens.danger, gap: '8px' }}
        >
          <LogoutIcon sx={{ fontSize: 16 }} />
          {t('header.logout')}
        </MenuItem>
      </Menu>
    </Box>
  )
}

const MainLayout = () => (
  <Box sx={{ display: 'flex', minHeight: '100vh', bgcolor: tokens.bgContent }}>
    <Sidebar />
    <Box
      sx={{
        flex: 1,
        ml: `${SIDEBAR_WIDTH}px`,
        display: 'flex',
        flexDirection: 'column',
        minHeight: '100vh',
        overflow: 'hidden',
      }}
    >
      <Header />
      <Box component="main" sx={{ flex: 1, p: '24px', overflowY: 'auto' }}>
        <Outlet />
      </Box>
    </Box>
  </Box>
)

export default MainLayout

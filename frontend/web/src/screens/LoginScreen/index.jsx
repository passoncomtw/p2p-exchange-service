import { useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Box,
  Button,
  CircularProgress,
  IconButton,
  InputAdornment,
  InputBase,
  MenuItem,
  Select,
  Typography,
} from '@mui/material'
import { Visibility, VisibilityOff } from '@mui/icons-material'
import { signIn } from 'src/slices/authSlice'
import { changeLanguage, SUPPORTED_LANGS } from 'src/i18n'

const OwlLogo = () => (
  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
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

const FormField = ({ label, value, onChange, type = 'text', placeholder, endAdornment }) => (
  <Box sx={{ mb: '16px' }}>
    <Typography sx={{ fontSize: '12px', color: '#666', mb: '4px', fontWeight: 400 }}>
      {label}
    </Typography>
    <Box
      sx={{
        border: '1px solid #D9D9D9',
        borderRadius: '4px',
        px: '12px',
        py: '8px',
        display: 'flex',
        alignItems: 'center',
        bgcolor: 'white',
        '&:focus-within': { borderColor: '#FFC107' },
      }}
    >
      <InputBase
        fullWidth
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        sx={{
          fontSize: '13px',
          color: '#333',
          '& input::placeholder': { color: '#BFBFBF', fontSize: '13px' },
        }}
      />
      {endAdornment}
    </Box>
  </Box>
)

const LoginScreen = () => {
  const dispatch = useDispatch()
  const { t, i18n } = useTranslation()
  const { isAuth } = useSelector((state) => state.auth)

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  if (isAuth) return <Navigate to="/" replace />

  const handleSubmit = (e) => {
    e.preventDefault()
    if (!username || !password) return
    setSubmitting(true)
    dispatch(signIn({ username, password }))
    setSubmitting(false)
  }

  return (
    <Box
      sx={{
        width: '100vw',
        minHeight: '100vh',
        bgcolor: '#EBEDF2',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {/* 語系切換（右上角） */}
      <Box sx={{ position: 'fixed', top: '16px', right: '20px' }}>
        <Select
          value={i18n.language}
          onChange={(e) => changeLanguage(e.target.value)}
          size="small"
          sx={{
            fontSize: '12px',
            bgcolor: 'white',
            '& .MuiOutlinedInput-notchedOutline': { borderColor: '#D9D9D9' },
            '& .MuiSelect-select': { py: '6px', px: '10px' },
          }}
        >
          {SUPPORTED_LANGS.map((lang) => (
            <MenuItem key={lang.code} value={lang.code} sx={{ fontSize: '12px' }}>
              {lang.label}
            </MenuItem>
          ))}
        </Select>
      </Box>

      {/* 登入卡片 */}
      <Box
        component="form"
        onSubmit={handleSubmit}
        sx={{
          width: '344px',
          bgcolor: 'white',
          borderRadius: '4px',
          p: '32px 28px 20px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
        }}
      >
        {/* Logo */}
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', mb: '24px', gap: '10px' }}>
          <OwlLogo />
          <Box>
            <Typography sx={{ fontSize: '18px', fontWeight: 700, color: '#333', lineHeight: 1.2 }}>
              {t('login.brand')}
            </Typography>
            <Typography sx={{ fontSize: '11px', color: '#999', mt: '2px' }}>
              {t('login.subtitle')}
            </Typography>
          </Box>
        </Box>

        <FormField
          label={t('login.usernameLabel')}
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder={t('login.usernamePlaceholder')}
        />

        <FormField
          label={t('login.passwordLabel')}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          type={showPassword ? 'text' : 'password'}
          placeholder={t('login.passwordPlaceholder')}
          endAdornment={
            <InputAdornment position="end">
              <IconButton
                size="small"
                onClick={() => setShowPassword((v) => !v)}
                sx={{ color: '#BFBFBF', p: '2px' }}
              >
                {showPassword
                  ? <VisibilityOff sx={{ fontSize: 16 }} />
                  : <Visibility sx={{ fontSize: 16 }} />}
              </IconButton>
            </InputAdornment>
          }
        />

        <Button
          type="submit"
          fullWidth
          disabled={submitting}
          sx={{
            mt: '8px',
            height: '36px',
            bgcolor: '#FFC107',
            color: 'white',
            fontSize: '14px',
            fontWeight: 500,
            borderRadius: '4px',
            boxShadow: 'none',
            '&:hover': { bgcolor: '#FFB300', boxShadow: 'none' },
            '&:disabled': { bgcolor: '#FFE082', color: 'white' },
          }}
        >
          {submitting ? <CircularProgress size={18} sx={{ color: 'white' }} /> : t('login.submit')}
        </Button>

        <Typography sx={{ textAlign: 'center', fontSize: '11px', color: '#BFBFBF', mt: '16px' }}>
          {t('login.version')}
        </Typography>
      </Box>
    </Box>
  )
}

export default LoginScreen

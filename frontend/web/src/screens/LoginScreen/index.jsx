import { useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Box,
  Button,
  CircularProgress,
  FormHelperText,
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
import OwlLogo from 'src/components/OwlLogo'
import { tokens } from 'src/theme'

const announceLanguageChange = (lang) => {
  const region = document.getElementById('aria-live-region')
  if (!region) return
  const msgs = { 'zh-TW': '語言已切換', 'zh-CN': '语言已切换' }
  region.textContent = msgs[lang] ?? '語言已切換'
}

const FormField = ({ label, value, onChange, type = 'text', placeholder, endAdornment }) => (
  <Box sx={{ mb: '16px' }}>
    <Typography sx={{ fontSize: '12px', color: tokens.textSecondary, mb: '4px', fontWeight: 400 }}>
      {label}
    </Typography>
    <Box
      sx={{
        border: `1px solid ${tokens.borderInput}`,
        borderRadius: '4px',
        px: '12px',
        py: '8px',
        display: 'flex',
        alignItems: 'center',
        bgcolor: tokens.bgCard,
        '&:focus-within': { borderColor: tokens.primary },
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
          color: tokens.textPrimary,
          '& input::placeholder': { color: tokens.textPlaceholder, fontSize: '13px' },
        }}
      />
      {endAdornment}
    </Box>
  </Box>
)

const LoginScreen = () => {
  const dispatch = useDispatch()
  const { t, i18n } = useTranslation()
  const { isAuth, error } = useSelector((state) => state.auth)

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

  const handleLanguageChange = (lang) => {
    changeLanguage(lang)
    announceLanguageChange(lang)
  }

  return (
    <Box
      sx={{
        width: '100vw',
        minHeight: '100vh',
        bgcolor: tokens.bgLogin,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {/* 語系切換（右上角） */}
      <Box sx={{ position: 'fixed', top: '16px', right: '20px' }}>
        <Select
          value={i18n.language}
          onChange={(e) => handleLanguageChange(e.target.value)}
          size="small"
          sx={{
            fontSize: '12px',
            bgcolor: tokens.bgCard,
            '& .MuiOutlinedInput-notchedOutline': { borderColor: tokens.borderInput },
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
          bgcolor: tokens.bgCard,
          borderRadius: '4px',
          p: '32px 28px 20px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
        }}
      >
        {/* Logo */}
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', mb: '24px', gap: '10px' }}>
          <OwlLogo size={48} />
          <Box>
            <Typography sx={{ fontSize: '18px', fontWeight: 700, color: tokens.textPrimary, lineHeight: 1.2 }}>
              {t('login.brand')}
            </Typography>
            <Typography sx={{ fontSize: '11px', color: tokens.textTertiary, mt: '2px' }}>
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
                aria-label={showPassword ? t('a11y.hidePassword') : t('a11y.showPassword')}
                sx={{ color: tokens.textPlaceholder, p: '2px' }}
              >
                {showPassword
                  ? <VisibilityOff sx={{ fontSize: 16 }} />
                  : <Visibility sx={{ fontSize: 16 }} />}
              </IconButton>
            </InputAdornment>
          }
        />

        {error && (
          <FormHelperText error role="alert" sx={{ mb: '8px', fontSize: '12px' }}>
            {error}
          </FormHelperText>
        )}

        <Button
          type="submit"
          fullWidth
          disabled={submitting}
          sx={{
            mt: '8px',
            height: '36px',
            bgcolor: tokens.primary,
            color: 'white',
            fontSize: '14px',
            fontWeight: 500,
            borderRadius: '4px',
            boxShadow: 'none',
            '&:hover': { bgcolor: tokens.primaryDark, boxShadow: 'none' },
            '&:disabled': { bgcolor: tokens.primaryDisabled, color: 'white' },
          }}
        >
          {submitting ? <CircularProgress size={18} sx={{ color: 'white' }} /> : t('login.submit')}
        </Button>

        <Typography sx={{ textAlign: 'center', fontSize: '11px', color: tokens.textTertiary, mt: '16px' }}>
          {t('login.version')}
        </Typography>
      </Box>
    </Box>
  )
}

export default LoginScreen

import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Box,
  Button,
  Card,
  CardContent,
  MenuItem,
  Snackbar,
  Alert,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from '@mui/material'
import {
  validateCreateOrder,
  calcTotalAmount,
  ASSETS,
  FIATS,
  PAYMENT_METHODS,
} from '@shared'
import { tokens } from 'src/theme'
import { ordersApi } from 'src/apis/v1Orders'

const initialForm = {
  type: 'buy',
  asset: 'USDT',
  fiat: 'TWD',
  price: '',
  quantity: '',
  paymentMethod: 'bank_transfer',
}

const CreateOrderScreen = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [form, setForm] = useState(initialForm)
  const [errors, setErrors] = useState({})
  const [submitting, setSubmitting] = useState(false)
  const [snack, setSnack] = useState({ open: false, severity: 'success', msg: '' })

  const total = useMemo(
    () => calcTotalAmount(Number(form.price), Number(form.quantity)),
    [form.price, form.quantity],
  )

  const update = (key) => (value) => {
    setForm((prev) => ({ ...prev, [key]: value }))
    setErrors((prev) => ({ ...prev, [key]: undefined }))
  }

  const handleSubmit = async () => {
    const input = {
      type: form.type,
      asset: form.asset,
      fiat: form.fiat,
      price: Number(form.price),
      quantity: Number(form.quantity),
      paymentMethod: form.paymentMethod,
    }

    const fieldErrors = validateCreateOrder(input)
    if (fieldErrors.length > 0) {
      const mapped = {}
      fieldErrors.forEach((e) => {
        mapped[e.field] = t(e.messageKey)
      })
      setErrors(mapped)
      return
    }

    setSubmitting(true)
    try {
      await ordersApi.create(input)
      setSnack({ open: true, severity: 'success', msg: t('order.message.createSuccess') })
      setTimeout(() => navigate('/my-orders'), 600)
    } catch {
      setSnack({ open: true, severity: 'error', msg: t('order.message.submitFailed') })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: tokens.textPrimary, mb: '16px' }}>
        {t('order.pageTitle.create')}
      </Typography>

      <Card sx={{ boxShadow: 'none', border: `1px solid ${tokens.borderCard}` }}>
        <CardContent sx={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {/* 買 / 賣 切換 */}
          <ToggleButtonGroup
            exclusive
            value={form.type}
            onChange={(_, v) => v && update('type')(v)}
            fullWidth
          >
            <ToggleButton
              value="buy"
              sx={{
                fontSize: '14px',
                '&.Mui-selected': { bgcolor: tokens.statusActive, color: '#fff', '&:hover': { bgcolor: tokens.statusActive } },
              }}
            >
              {t('order.type.buy')}
            </ToggleButton>
            <ToggleButton
              value="sell"
              sx={{
                fontSize: '14px',
                '&.Mui-selected': { bgcolor: tokens.danger, color: '#fff', '&:hover': { bgcolor: tokens.danger } },
              }}
            >
              {t('order.type.sell')}
            </ToggleButton>
          </ToggleButtonGroup>

          {/* 幣種 / 法幣 */}
          <Box sx={{ display: 'flex', gap: '16px' }}>
            <TextField
              select
              fullWidth
              size="small"
              label={t('order.field.asset')}
              value={form.asset}
              onChange={(e) => update('asset')(e.target.value)}
            >
              {ASSETS.map((a) => (
                <MenuItem key={a} value={a}>{a}</MenuItem>
              ))}
            </TextField>
            <TextField
              select
              fullWidth
              size="small"
              label={t('order.field.fiat')}
              value={form.fiat}
              onChange={(e) => update('fiat')(e.target.value)}
            >
              {FIATS.map((f) => (
                <MenuItem key={f} value={f}>{f}</MenuItem>
              ))}
            </TextField>
          </Box>

          {/* 單價 / 數量 */}
          <Box sx={{ display: 'flex', gap: '16px' }}>
            <TextField
              fullWidth
              size="small"
              type="number"
              label={t('order.field.price')}
              value={form.price}
              onChange={(e) => update('price')(e.target.value)}
              error={Boolean(errors.price)}
              helperText={errors.price}
              inputProps={{ min: 0, step: 'any' }}
            />
            <TextField
              fullWidth
              size="small"
              type="number"
              label={t('order.field.quantity')}
              value={form.quantity}
              onChange={(e) => update('quantity')(e.target.value)}
              error={Boolean(errors.quantity)}
              helperText={errors.quantity}
              inputProps={{ min: 0, step: 'any' }}
            />
          </Box>

          {/* 付款方式 */}
          <TextField
            select
            fullWidth
            size="small"
            label={t('order.field.paymentMethod')}
            value={form.paymentMethod}
            onChange={(e) => update('paymentMethod')(e.target.value)}
            error={Boolean(errors.paymentMethod)}
            helperText={errors.paymentMethod}
          >
            {PAYMENT_METHODS.map((pm) => (
              <MenuItem key={pm} value={pm}>{t(`order.paymentMethod.${pm}`)}</MenuItem>
            ))}
          </TextField>

          {/* 總額（自動計算） */}
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              bgcolor: tokens.bgContent,
              borderRadius: '4px',
              px: '12px',
              py: '10px',
            }}
          >
            <Typography sx={{ fontSize: '13px', color: tokens.textSecondary }}>
              {t('order.field.totalAmount')}
            </Typography>
            <Typography sx={{ fontSize: '16px', fontWeight: 600, color: tokens.textPrimary }}>
              {total.toLocaleString()} {form.fiat}
            </Typography>
          </Box>

          <Button
            variant="contained"
            disableElevation
            disabled={submitting}
            onClick={handleSubmit}
            sx={{ height: '36px', fontSize: '14px' }}
          >
            {t('order.action.submit')}
          </Button>
        </CardContent>
      </Card>

      <Snackbar
        open={snack.open}
        autoHideDuration={2000}
        onClose={() => setSnack((s) => ({ ...s, open: false }))}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert severity={snack.severity} variant="filled" sx={{ fontSize: '13px' }}>
          {snack.msg}
        </Alert>
      </Snackbar>
    </Box>
  )
}

export default CreateOrderScreen

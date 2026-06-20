import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Snackbar,
  Typography,
} from '@mui/material'
import { canComplete } from '@shared'
import { tokens } from 'src/theme'
import { ordersApi } from 'src/apis/v1Orders'
import { TypeTag, StatusTag } from 'src/components/OrderTags'

const formatDateTime = (iso) => new Date(iso).toLocaleString()

const Row = ({ label, children }) => (
  <Box sx={{ display: 'flex', justifyContent: 'space-between', py: '10px', borderBottom: `1px solid ${tokens.borderCard}` }}>
    <Typography sx={{ fontSize: '13px', color: tokens.textSecondary }}>{label}</Typography>
    <Box sx={{ fontSize: '13px', color: tokens.textPrimary }}>{children}</Box>
  </Box>
)

const AdminOrderDetailScreen = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { id } = useParams()
  const [order, setOrder] = useState(null)
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [snack, setSnack] = useState({ open: false, severity: 'success', msg: '' })

  const load = useCallback(async () => {
    try {
      const o = await ordersApi.adminGet(id)
      setOrder(o)
    } catch {
      setSnack({ open: true, severity: 'error', msg: t('order.message.loadFailed') })
    }
  }, [id, t])

  useEffect(() => {
    load()
  }, [load])

  const handleComplete = async () => {
    try {
      await ordersApi.adminComplete(id)
      setSnack({ open: true, severity: 'success', msg: t('order.message.completeSuccess') })
      await load()
    } catch {
      setSnack({ open: true, severity: 'error', msg: t('order.message.submitFailed') })
    } finally {
      setConfirmOpen(false)
    }
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: '16px' }}>
        <Typography sx={{ fontSize: '16px', fontWeight: 600, color: tokens.textPrimary }}>
          {t('order.pageTitle.detail')}
        </Typography>
        <Button onClick={() => navigate('/admin')} sx={{ fontSize: '13px', color: tokens.textSecondary }}>
          {t('order.action.back')}
        </Button>
      </Box>

      {order && (
        <Card sx={{ boxShadow: 'none', border: `1px solid ${tokens.borderCard}` }}>
          <CardContent>
            <Row label={t('order.field.orderId')}>{order.id}</Row>
            <Row label={t('order.field.type')}><TypeTag type={order.type} /></Row>
            <Row label={t('order.field.status')}><StatusTag status={order.status} /></Row>
            <Row label={t('order.field.asset')}>{order.asset}</Row>
            <Row label={t('order.field.fiat')}>{order.fiat}</Row>
            <Row label={t('order.field.price')}>{order.price.toLocaleString()}</Row>
            <Row label={t('order.field.quantity')}>{order.quantity.toLocaleString()}</Row>
            <Row label={t('order.field.totalAmount')}>{order.totalAmount.toLocaleString()} {order.fiat}</Row>
            <Row label={t('order.field.paymentMethod')}>{t(`order.paymentMethod.${order.paymentMethod}`)}</Row>
            <Row label={t('order.field.createdBy')}>{order.createdBy}</Row>
            <Row label={t('order.field.createdAt')}>{formatDateTime(order.createdAt)}</Row>

            {canComplete(order.status) && (
              <Button
                variant="contained"
                disableElevation
                onClick={() => setConfirmOpen(true)}
                sx={{ mt: '16px', height: '36px', fontSize: '14px' }}
              >
                {t('order.action.complete')}
              </Button>
            )}
          </CardContent>
        </Card>
      )}

      <Dialog open={confirmOpen} onClose={() => setConfirmOpen(false)}>
        <DialogTitle sx={{ fontSize: '16px' }}>{t('order.action.complete')}</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ fontSize: '13px' }}>
            {t('order.message.completeConfirm')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmOpen(false)} sx={{ fontSize: '13px', color: tokens.textSecondary }}>
            {t('order.action.dismiss')}
          </Button>
          <Button onClick={handleComplete} sx={{ fontSize: '13px', color: tokens.primaryDeep }}>
            {t('order.action.confirm')}
          </Button>
        </DialogActions>
      </Dialog>

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

export default AdminOrderDetailScreen

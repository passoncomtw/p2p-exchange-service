import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Paper,
  Snackbar,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import { canCancel } from '@shared'
import { tokens } from 'src/theme'
import { ordersApi } from 'src/apis/v1Orders'
import { TypeTag, StatusTag } from 'src/components/OrderTags'

const formatDateTime = (iso) => new Date(iso).toLocaleString()

const MyOrdersScreen = () => {
  const { t } = useTranslation()
  const [orders, setOrders] = useState([])
  const [loading, setLoading] = useState(true)
  const [target, setTarget] = useState(null) // 待取消的訂單
  const [snack, setSnack] = useState({ open: false, severity: 'success', msg: '' })

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const list = await ordersApi.listMine()
      setOrders(list)
    } catch {
      setSnack({ open: true, severity: 'error', msg: t('order.message.loadFailed') })
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    load()
  }, [load])

  const handleCancel = async () => {
    if (!target) return
    try {
      await ordersApi.cancel(target.id)
      setSnack({ open: true, severity: 'success', msg: t('order.message.cancelSuccess') })
      await load()
    } catch {
      setSnack({ open: true, severity: 'error', msg: t('order.message.submitFailed') })
    } finally {
      setTarget(null)
    }
  }

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: tokens.textPrimary, mb: '16px' }}>
        {t('order.pageTitle.myOrders')}
      </Typography>

      <TableContainer component={Paper} sx={{ boxShadow: 'none', border: `1px solid ${tokens.borderCard}` }}>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ '& th': { fontSize: '12px', color: tokens.textSecondary, fontWeight: 600 } }}>
              <TableCell>{t('order.field.type')}</TableCell>
              <TableCell>{t('order.field.asset')}</TableCell>
              <TableCell align="right">{t('order.field.price')}</TableCell>
              <TableCell align="right">{t('order.field.quantity')}</TableCell>
              <TableCell align="right">{t('order.field.totalAmount')}</TableCell>
              <TableCell>{t('order.field.paymentMethod')}</TableCell>
              <TableCell>{t('order.field.status')}</TableCell>
              <TableCell>{t('order.field.createdAt')}</TableCell>
              <TableCell align="right" />
            </TableRow>
          </TableHead>
          <TableBody>
            {orders.map((o) => (
              <TableRow key={o.id} sx={{ '& td': { fontSize: '13px', color: tokens.textPrimary } }}>
                <TableCell><TypeTag type={o.type} /></TableCell>
                <TableCell>{o.asset}</TableCell>
                <TableCell align="right">{o.price.toLocaleString()}</TableCell>
                <TableCell align="right">{o.quantity.toLocaleString()}</TableCell>
                <TableCell align="right">{o.totalAmount.toLocaleString()} {o.fiat}</TableCell>
                <TableCell>{t(`order.paymentMethod.${o.paymentMethod}`)}</TableCell>
                <TableCell><StatusTag status={o.status} /></TableCell>
                <TableCell>{formatDateTime(o.createdAt)}</TableCell>
                <TableCell align="right">
                  {canCancel(o.status) && (
                    <Button
                      size="small"
                      onClick={() => setTarget(o)}
                      sx={{ fontSize: '12px', color: tokens.danger }}
                    >
                      {t('order.action.cancel')}
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {!loading && orders.length === 0 && (
              <TableRow>
                <TableCell colSpan={9} align="center" sx={{ fontSize: '13px', color: tokens.textTertiary, py: '32px' }}>
                  {t('order.message.emptyMine')}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={Boolean(target)} onClose={() => setTarget(null)}>
        <DialogTitle sx={{ fontSize: '16px' }}>{t('order.action.cancel')}</DialogTitle>
        <DialogContent>
          <DialogContentText sx={{ fontSize: '13px' }}>
            {t('order.message.cancelConfirm')}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTarget(null)} sx={{ fontSize: '13px', color: tokens.textSecondary }}>
            {t('order.action.dismiss')}
          </Button>
          <Button onClick={handleCancel} sx={{ fontSize: '13px', color: tokens.danger }}>
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

export default MyOrdersScreen

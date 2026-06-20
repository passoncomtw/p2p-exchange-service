import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Alert,
  Box,
  Paper,
  Snackbar,
  Tab,
  Tabs,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import { ORDER_STATUSES } from '@shared'
import { tokens } from 'src/theme'
import { ordersApi } from 'src/apis/v1Orders'
import { TypeTag, StatusTag } from 'src/components/OrderTags'

const formatDateTime = (iso) => new Date(iso).toLocaleString()
const FILTERS = ['all', ...ORDER_STATUSES] // all | open | completed | cancelled

const AdminOrdersScreen = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [filter, setFilter] = useState('all')
  const [orders, setOrders] = useState([])
  const [loading, setLoading] = useState(true)
  const [snack, setSnack] = useState({ open: false, severity: 'error', msg: '' })

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const list = await ordersApi.adminList(filter === 'all' ? undefined : filter)
      setOrders(list)
    } catch {
      setSnack({ open: true, severity: 'error', msg: t('order.message.loadFailed') })
    } finally {
      setLoading(false)
    }
  }, [filter, t])

  useEffect(() => {
    load()
  }, [load])

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: tokens.textPrimary, mb: '16px' }}>
        {t('order.pageTitle.admin')}
      </Typography>

      <Tabs
        value={filter}
        onChange={(_, v) => setFilter(v)}
        sx={{ mb: '16px', minHeight: 36, '& .MuiTab-root': { fontSize: '13px', minHeight: 36 } }}
      >
        {FILTERS.map((f) => (
          <Tab key={f} value={f} label={t(`order.status.${f}`)} />
        ))}
      </Tabs>

      <TableContainer component={Paper} sx={{ boxShadow: 'none', border: `1px solid ${tokens.borderCard}` }}>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ '& th': { fontSize: '12px', color: tokens.textSecondary, fontWeight: 600 } }}>
              <TableCell>{t('order.field.orderId')}</TableCell>
              <TableCell>{t('order.field.type')}</TableCell>
              <TableCell>{t('order.field.asset')}</TableCell>
              <TableCell align="right">{t('order.field.price')}</TableCell>
              <TableCell align="right">{t('order.field.quantity')}</TableCell>
              <TableCell align="right">{t('order.field.totalAmount')}</TableCell>
              <TableCell>{t('order.field.status')}</TableCell>
              <TableCell>{t('order.field.createdBy')}</TableCell>
              <TableCell>{t('order.field.createdAt')}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {orders.map((o) => (
              <TableRow
                key={o.id}
                hover
                onClick={() => navigate(`/admin/orders/${o.id}`)}
                sx={{ cursor: 'pointer', '& td': { fontSize: '13px', color: tokens.textPrimary } }}
              >
                <TableCell>{o.id}</TableCell>
                <TableCell><TypeTag type={o.type} /></TableCell>
                <TableCell>{o.asset}</TableCell>
                <TableCell align="right">{o.price.toLocaleString()}</TableCell>
                <TableCell align="right">{o.quantity.toLocaleString()}</TableCell>
                <TableCell align="right">{o.totalAmount.toLocaleString()} {o.fiat}</TableCell>
                <TableCell><StatusTag status={o.status} /></TableCell>
                <TableCell>{o.createdBy}</TableCell>
                <TableCell>{formatDateTime(o.createdAt)}</TableCell>
              </TableRow>
            ))}
            {!loading && orders.length === 0 && (
              <TableRow>
                <TableCell colSpan={9} align="center" sx={{ fontSize: '13px', color: tokens.textTertiary, py: '32px' }}>
                  {t('order.message.emptyAdmin')}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>

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

export default AdminOrdersScreen

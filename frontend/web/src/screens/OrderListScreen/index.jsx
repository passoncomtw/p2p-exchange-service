import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Box,
  Typography,
  TextField,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Select,
  MenuItem,
  IconButton,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Divider,
  Alert,
  CircularProgress,
} from '@mui/material'
import {
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
  GavelOutlined as GavelIcon,
} from '@mui/icons-material'
import { ordersApi } from 'src/apis/orders'

const PAGE_SIZE_OPTIONS = [10, 20, 50]

const STATUS_OPTIONS = ['', 'pending', 'paid', 'completed', 'cancelled', 'disputed']

const STATUS_COLOR_MAP = {
  matched: { bg: '#FFF3E0', color: '#E65100' },
  pending: { bg: '#FFF3E0', color: '#E65100' },
  paid: { bg: '#E3F2FD', color: '#1565C0' },
  completed: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FAFAFA', color: '#999' },
  timeout: { bg: '#FAFAFA', color: '#999' },
  disputed: { bg: '#FCE4EC', color: '#C62828' },
}

// ── 詳情 + 申訴仲裁 Dialog ────────────────────────────────────────────────────

const OrderDetailDialog = ({ order, open, onClose, onResolved, t }) => {
  const [action, setAction] = useState('')  // 'complete' | 'refund'
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const resetForm = () => { setAction(''); setReason(''); setError('') }

  const handleClose = () => { resetForm(); onClose() }

  const handleResolve = async () => {
    if (!action) { setError(t('order.dispute.selectAction')); return }
    if (action === 'refund' && !reason.trim()) { setError(t('order.dispute.reasonRequired')); return }
    setSubmitting(true)
    setError('')
    try {
      await ordersApi.resolve(order.id, action, reason.trim())
      resetForm()
      onResolved()
    } catch (e) {
      setError(e?.response?.data?.message || t('order.dispute.resolveFailed'))
    } finally {
      setSubmitting(false)
    }
  }

  if (!order) return null

  const isDisputed = order.status === 'disputed'
  const formatDT = (iso) => iso ? new Date(iso).toLocaleString() : '-'

  const InfoRow = ({ label, value }) => (
    <Box sx={{ display: 'flex', justifyContent: 'space-between', py: '6px', borderBottom: '1px solid #F5F5F5' }}>
      <Typography sx={{ fontSize: '13px', color: '#888', minWidth: 110 }}>{label}</Typography>
      <Typography sx={{ fontSize: '13px', color: '#333', fontWeight: 500, textAlign: 'right', flex: 1 }}>{value}</Typography>
    </Box>
  )

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontSize: '15px', fontWeight: 700, pb: '8px', display: 'flex', alignItems: 'center', gap: '8px' }}>
        {isDisputed && <GavelIcon sx={{ color: '#C62828', fontSize: 20 }} />}
        {t('order.detail.title')}
        <Chip
          label={t(`order.status.${order.status}`, order.status)}
          size="small"
          sx={{
            ml: 'auto',
            bgcolor: STATUS_COLOR_MAP[order.status]?.bg || '#F5F5F5',
            color: STATUS_COLOR_MAP[order.status]?.color || '#666',
            fontWeight: 600,
            fontSize: '12px',
          }}
        />
      </DialogTitle>

      <DialogContent dividers>
        {/* 基本資訊 */}
        <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#999', mb: '8px', textTransform: 'uppercase' }}>
          {t('order.detail.basicInfo')}
        </Typography>
        <InfoRow label={t('order.column.orderNo')} value={order.orderNo} />
        <InfoRow label={t('order.column.type')} value={t(`order.type.${order.listingType}`, order.listingType)} />
        <InfoRow label={t('order.column.status')} value={t(`order.status.${order.status}`, order.status)} />
        <InfoRow label={t('order.column.createdAt')} value={formatDT(order.createdAt)} />
        {order.paymentDeadline && <InfoRow label={t('order.detail.paymentDeadline')} value={formatDT(order.paymentDeadline)} />}
        {order.cancelledAt && <InfoRow label={t('order.detail.cancelledAt')} value={formatDT(order.cancelledAt)} />}
        {order.cancelReason && <InfoRow label={t('order.detail.cancelReason')} value={order.cancelReason} />}

        {/* 交易資訊 */}
        <Divider sx={{ my: '12px' }} />
        <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#999', mb: '8px', textTransform: 'uppercase' }}>
          {t('order.detail.tradeInfo')}
        </Typography>
        <InfoRow label={t('order.column.crypto')} value={`${order.cryptoAmount} ${order.cryptoCurrency}`} />
        <InfoRow label={t('order.column.fiatAmount')} value={`${order.fiatAmount} ${order.fiatCurrency}`} />
        <InfoRow label={t('order.column.price')} value={`${order.price} ${order.fiatCurrency}`} />
        {order.totalFee > 0 && <InfoRow label={t('order.detail.fee')} value={order.totalFee} />}

        {/* 申訴仲裁區 */}
        {isDisputed && (
          <>
            <Divider sx={{ my: '12px' }} />
            <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#C62828', mb: '10px', textTransform: 'uppercase' }}>
              {t('order.dispute.title')}
            </Typography>

            <Typography sx={{ fontSize: '12px', color: '#666', mb: '12px', lineHeight: 1.6 }}>
              {t('order.dispute.hint')}
            </Typography>

            <Box sx={{ display: 'flex', gap: '10px', mb: '12px' }}>
              <Button
                variant={action === 'complete' ? 'contained' : 'outlined'}
                size="small"
                onClick={() => setAction('complete')}
                sx={{
                  flex: 1,
                  fontSize: '13px',
                  fontWeight: 600,
                  ...(action === 'complete' ? { bgcolor: '#2E7D32', '&:hover': { bgcolor: '#1B5E20' }, borderColor: '#2E7D32' } : { borderColor: '#2E7D32', color: '#2E7D32' }),
                }}
              >
                {t('order.dispute.actionComplete')}
              </Button>
              <Button
                variant={action === 'refund' ? 'contained' : 'outlined'}
                size="small"
                onClick={() => setAction('refund')}
                sx={{
                  flex: 1,
                  fontSize: '13px',
                  fontWeight: 600,
                  ...(action === 'refund' ? { bgcolor: '#C62828', '&:hover': { bgcolor: '#B71C1C' }, borderColor: '#C62828' } : { borderColor: '#C62828', color: '#C62828' }),
                }}
              >
                {t('order.dispute.actionRefund')}
              </Button>
            </Box>

            {action === 'complete' && (
              <Alert severity="info" sx={{ fontSize: '12px', mb: '10px' }}>
                {t('order.dispute.completeHint')}
              </Alert>
            )}

            {action === 'refund' && (
              <TextField
                fullWidth
                size="small"
                multiline
                rows={2}
                label={t('order.dispute.reasonLabel')}
                placeholder={t('order.dispute.reasonPlaceholder')}
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                sx={{ mb: '10px', '& .MuiInputBase-input': { fontSize: '13px' }, '& label': { fontSize: '13px' } }}
              />
            )}

            {error && (
              <Alert severity="error" sx={{ fontSize: '12px', mb: '8px' }}>
                {error}
              </Alert>
            )}
          </>
        )}
      </DialogContent>

      <DialogActions sx={{ px: '16px', py: '12px', gap: '8px' }}>
        <Button onClick={handleClose} size="small" sx={{ color: '#666', fontSize: '13px' }}>
          {t('order.detail.close')}
        </Button>
        {isDisputed && (
          <Button
            variant="contained"
            size="small"
            disabled={!action || submitting}
            onClick={handleResolve}
            sx={{
              fontSize: '13px',
              fontWeight: 600,
              bgcolor: '#FFC107',
              color: '#333',
              '&:hover': { bgcolor: '#FFB300' },
              '&:disabled': { bgcolor: '#EEE', color: '#999' },
              boxShadow: 'none',
              minWidth: 90,
            }}
          >
            {submitting ? <CircularProgress size={14} sx={{ color: '#333' }} /> : t('order.dispute.submit')}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  )
}

// ── 主頁面 ────────────────────────────────────────────────────────────────────

const OrderListScreen = () => {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [status, setStatus] = useState('')
  const [searchStatus, setSearchStatus] = useState('')
  const [list, setList] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState(null)

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await ordersApi.list({
        keyword: searchKeyword,
        status: searchStatus,
        limit: pageSize,
        offset: (page - 1) * pageSize,
      })
      setList(res.list || [])
      setTotal(res.total || 0)
    } catch {
      setList([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [searchKeyword, searchStatus, page, pageSize])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleSearch = () => {
    setPage(1)
    setSearchKeyword(keyword)
    setSearchStatus(status)
  }

  const handleReset = () => {
    setKeyword('')
    setStatus('')
    setSearchKeyword('')
    setSearchStatus('')
    setPage(1)
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') handleSearch()
  }

  const formatDateTime = (iso) => {
    if (!iso) return '-'
    return new Date(iso).toLocaleString()
  }

  const renderStatus = (s) => {
    const colors = STATUS_COLOR_MAP[s] || { bg: '#F5F5F5', color: '#666' }
    return (
      <Chip
        label={t(`order.status.${s}`, s)}
        size="small"
        sx={{
          bgcolor: colors.bg,
          color: colors.color,
          fontWeight: 600,
          fontSize: '12px',
          height: '24px',
        }}
      />
    )
  }

  const handleRowClick = (row) => setSelectedOrder(row)

  const handleResolved = () => {
    setSelectedOrder(null)
    fetchData()
  }

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: '#333', mb: '16px' }}>
        {t('page.orders')}
      </Typography>

      <Paper sx={{ p: '16px', mb: '16px', boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '13px', color: '#666', whiteSpace: 'nowrap' }}>
              {t('order.keyword')}
            </Typography>
            <TextField
              size="small"
              placeholder={t('order.keywordPlaceholder')}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              sx={{ width: 200, '& .MuiInputBase-input': { fontSize: '13px', py: '6px' } }}
            />
          </Box>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '13px', color: '#666', whiteSpace: 'nowrap' }}>
              {t('order.statusLabel')}
            </Typography>
            <Select
              size="small"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              displayEmpty
              sx={{ width: 140, fontSize: '13px', '& .MuiSelect-select': { py: '6px' } }}
            >
              <MenuItem value="" sx={{ fontSize: '13px' }}>{t('order.statusAll')}</MenuItem>
              {STATUS_OPTIONS.filter(Boolean).map((s) => (
                <MenuItem key={s} value={s} sx={{ fontSize: '13px' }}>{t(`order.status.${s}`, s)}</MenuItem>
              ))}
            </Select>
          </Box>

          <Button
            variant="text"
            size="small"
            onClick={handleReset}
            sx={{ color: '#666', fontSize: '13px', minWidth: 'auto' }}
          >
            {t('order.reset')}
          </Button>

          <Button
            variant="contained"
            size="small"
            onClick={handleSearch}
            sx={{
              bgcolor: '#FFC107',
              color: '#333',
              fontWeight: 600,
              fontSize: '13px',
              '&:hover': { bgcolor: '#FFB300' },
              boxShadow: 'none',
              px: '20px',
            }}
          >
            {t('order.filter')}
          </Button>
        </Box>
      </Paper>

      <TableContainer component={Paper} sx={{ boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: '#FAFAFA' }}>
              <TableCell sx={thStyle}>{t('order.column.orderNo')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.type')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.crypto')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.fiatAmount')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.price')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.status')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.createdAt')}</TableCell>
              <TableCell sx={thStyle}>{t('order.column.action')}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={8} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('order.loading')}
                </TableCell>
              </TableRow>
            ) : list.length === 0 ? (
              <TableRow>
                <TableCell colSpan={8} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('order.empty')}
                </TableCell>
              </TableRow>
            ) : (
              list.map((row) => (
                <TableRow
                  key={row.id}
                  hover
                  sx={{ cursor: 'pointer' }}
                  onClick={() => handleRowClick(row)}
                >
                  <TableCell sx={tdStyle}>{row.orderNo}</TableCell>
                  <TableCell sx={tdStyle}>{t(`order.type.${row.listingType}`, row.listingType)}</TableCell>
                  <TableCell sx={tdStyle}>{row.cryptoAmount} {row.cryptoCurrency}</TableCell>
                  <TableCell sx={tdStyle}>{row.fiatAmount} {row.fiatCurrency}</TableCell>
                  <TableCell sx={tdStyle}>{row.price}</TableCell>
                  <TableCell sx={tdStyle}>{renderStatus(row.status)}</TableCell>
                  <TableCell sx={tdStyle}>{formatDateTime(row.createdAt)}</TableCell>
                  <TableCell sx={tdStyle}>
                    {row.status === 'disputed' ? (
                      <Chip
                        icon={<GavelIcon sx={{ fontSize: '14px !important' }} />}
                        label={t('order.dispute.arbitrate')}
                        size="small"
                        clickable
                        sx={{ bgcolor: '#FCE4EC', color: '#C62828', fontWeight: 600, fontSize: '11px', height: '22px' }}
                      />
                    ) : (
                      <Typography sx={{ fontSize: '12px', color: '#AAA' }}>
                        {t('order.detail.view')}
                      </Typography>
                    )}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <Box
          sx={{
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            gap: '16px',
            px: '16px',
            py: '10px',
            borderTop: '1px solid #EBEBEB',
          }}
        >
          <Typography sx={{ fontSize: '12px', color: '#999' }}>
            {t('order.pagination.total', { total })} {t('order.pagination.page', { current: page, total: totalPages })}
          </Typography>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '12px', color: '#999' }}>{t('order.pagination.pageSize')}</Typography>
            <Select
              size="small"
              value={pageSize}
              onChange={(e) => { setPageSize(e.target.value); setPage(1) }}
              sx={{ fontSize: '12px', '& .MuiSelect-select': { py: '4px' } }}
            >
              {PAGE_SIZE_OPTIONS.map((n) => (
                <MenuItem key={n} value={n} sx={{ fontSize: '12px' }}>{n}{t('order.pagination.items')}</MenuItem>
              ))}
            </Select>
          </Box>

          <Box sx={{ display: 'flex', gap: '4px' }}>
            <IconButton size="small" disabled={page <= 1} onClick={(e) => { e.stopPropagation(); setPage((p) => p - 1) }}>
              <ChevronLeftIcon fontSize="small" />
            </IconButton>
            <IconButton size="small" disabled={page >= totalPages} onClick={(e) => { e.stopPropagation(); setPage((p) => p + 1) }}>
              <ChevronRightIcon fontSize="small" />
            </IconButton>
          </Box>
        </Box>
      </TableContainer>

      <OrderDetailDialog
        order={selectedOrder}
        open={Boolean(selectedOrder)}
        onClose={() => setSelectedOrder(null)}
        onResolved={handleResolved}
        t={t}
      />
    </Box>
  )
}

const thStyle = { fontSize: '13px', fontWeight: 600, color: '#333', py: '10px' }
const tdStyle = { fontSize: '13px', color: '#555', py: '10px' }

export default OrderListScreen

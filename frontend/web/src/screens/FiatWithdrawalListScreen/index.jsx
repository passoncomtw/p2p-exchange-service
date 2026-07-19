import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Box,
  Typography,
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
  TextField,
  CircularProgress,
  Skeleton,
} from '@mui/material'
import {
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
  CheckCircleOutlined as ApproveIcon,
  CancelOutlined as RejectIcon,
} from '@mui/icons-material'
import { fiatWithdrawalsApi } from 'src/apis/fiatWithdrawals'

const PAGE_SIZE_OPTIONS = [10, 20, 50]

const STATUS_OPTIONS = ['', 'pending', 'approved', 'rejected']

const STATUS_COLOR_MAP = {
  pending:  { bg: '#FFF3E0', color: '#E65100' },
  approved: { bg: '#E8F5E9', color: '#2E7D32' },
  rejected: { bg: '#FAFAFA', color: '#999' },
}

// ── 審核 Dialog ───────────────────────────────────────────────────────────────

const ReviewDialog = ({ item, open, onClose, onDone, t }) => {
  const [action, setAction] = useState('')
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const reset = () => { setAction(''); setReason(''); setError('') }
  const handleClose = () => { reset(); onClose() }

  const handleSubmit = async () => {
    if (!action) { setError(t('fiatWithdrawal.review.selectAction')); return }
    if (action === 'reject' && !reason.trim()) { setError(t('fiatWithdrawal.review.reasonRequired')); return }
    setSubmitting(true)
    setError('')
    try {
      await fiatWithdrawalsApi.review(item.id, action, reason.trim())
      reset()
      onDone()
    } catch (e) {
      setError(e?.response?.data?.message || t('fiatWithdrawal.review.failed'))
    } finally {
      setSubmitting(false)
    }
  }

  if (!item) return null

  const isPending = item.status === 'pending'

  const InfoRow = ({ label, value }) => (
    <Box sx={{ display: 'flex', justifyContent: 'space-between', py: '6px', borderBottom: '1px solid #F5F5F5' }}>
      <Typography sx={{ fontSize: '13px', color: '#888', minWidth: 110 }}>{label}</Typography>
      <Typography sx={{ fontSize: '13px', color: '#333', fontWeight: 500, textAlign: 'right', flex: 1 }}>{value ?? '-'}</Typography>
    </Box>
  )

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle sx={{ fontSize: '15px', fontWeight: 700, pb: '8px', display: 'flex', alignItems: 'center', gap: '8px' }}>
        {t('fiatWithdrawal.review.title')}
        <Chip
          label={t(`fiatWithdrawal.status.${item.status}`, item.status)}
          size="small"
          sx={{
            ml: 'auto',
            bgcolor: STATUS_COLOR_MAP[item.status]?.bg || '#F5F5F5',
            color: STATUS_COLOR_MAP[item.status]?.color || '#666',
            fontWeight: 600,
            fontSize: '12px',
          }}
        />
      </DialogTitle>

      <DialogContent dividers>
        <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#999', mb: '8px', textTransform: 'uppercase' }}>
          {t('fiatWithdrawal.review.applicantInfo')}
        </Typography>
        <InfoRow label={t('fiatWithdrawal.column.id')} value={`#${item.id}`} />
        <InfoRow label={t('fiatWithdrawal.column.userId')} value={item.userId} />
        <InfoRow label={t('fiatWithdrawal.column.createdAt')} value={item.createdAt ? new Date(item.createdAt).toLocaleString() : '-'} />

        <Divider sx={{ my: '12px' }} />
        <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#999', mb: '8px', textTransform: 'uppercase' }}>
          {t('fiatWithdrawal.review.withdrawalInfo')}
        </Typography>
        <InfoRow label={t('fiatWithdrawal.column.amount')} value={`${item.amount} ${item.currency}`} />
        <InfoRow label={t('fiatWithdrawal.column.bankCode')} value={item.bankCode} />
        <InfoRow label={t('fiatWithdrawal.column.bankAccount')} value={item.bankAccount} />
        <InfoRow label={t('fiatWithdrawal.column.accountName')} value={item.accountName} />

        {item.rejectReason && (
          <>
            <Divider sx={{ my: '12px' }} />
            <InfoRow label={t('fiatWithdrawal.column.rejectReason')} value={item.rejectReason} />
          </>
        )}

        {isPending && (
          <>
            <Divider sx={{ my: '12px' }} />
            <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#333', mb: '10px', textTransform: 'uppercase' }}>
              {t('fiatWithdrawal.review.action')}
            </Typography>

            <Box sx={{ display: 'flex', gap: '10px', mb: '12px' }}>
              <Button
                variant={action === 'approve' ? 'contained' : 'outlined'}
                size="small"
                startIcon={<ApproveIcon sx={{ fontSize: '16px !important' }} />}
                onClick={() => setAction('approve')}
                sx={{
                  flex: 1,
                  fontSize: '13px',
                  fontWeight: 600,
                  ...(action === 'approve'
                    ? { bgcolor: '#2E7D32', '&:hover': { bgcolor: '#1B5E20' }, borderColor: '#2E7D32' }
                    : { borderColor: '#2E7D32', color: '#2E7D32' }),
                }}
              >
                {t('fiatWithdrawal.review.approve')}
              </Button>
              <Button
                variant={action === 'reject' ? 'contained' : 'outlined'}
                size="small"
                startIcon={<RejectIcon sx={{ fontSize: '16px !important' }} />}
                onClick={() => setAction('reject')}
                sx={{
                  flex: 1,
                  fontSize: '13px',
                  fontWeight: 600,
                  ...(action === 'reject'
                    ? { bgcolor: '#C62828', '&:hover': { bgcolor: '#B71C1C' }, borderColor: '#C62828' }
                    : { borderColor: '#C62828', color: '#C62828' }),
                }}
              >
                {t('fiatWithdrawal.review.reject')}
              </Button>
            </Box>

            {action === 'approve' && (
              <Alert severity="info" sx={{ fontSize: '12px', mb: '10px' }}>
                {t('fiatWithdrawal.review.approveHint')}
              </Alert>
            )}

            {action === 'reject' && (
              <TextField
                fullWidth
                size="small"
                multiline
                rows={2}
                label={t('fiatWithdrawal.review.reasonLabel')}
                placeholder={t('fiatWithdrawal.review.reasonPlaceholder')}
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                sx={{ mb: '10px', '& .MuiInputBase-input': { fontSize: '13px' }, '& label': { fontSize: '13px' } }}
              />
            )}

            {error && (
              <Alert severity="error" sx={{ fontSize: '12px' }}>
                {error}
              </Alert>
            )}
          </>
        )}
      </DialogContent>

      <DialogActions sx={{ px: '16px', py: '12px', gap: '8px' }}>
        <Button onClick={handleClose} size="small" sx={{ color: '#666', fontSize: '13px' }}>
          {t('fiatWithdrawal.review.close')}
        </Button>
        {isPending && (
          <Button
            variant="contained"
            size="small"
            disabled={!action || submitting}
            onClick={handleSubmit}
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
            {submitting ? <CircularProgress size={14} sx={{ color: '#333' }} /> : t('fiatWithdrawal.review.submit')}
          </Button>
        )}
      </DialogActions>
    </Dialog>
  )
}

// ── 主頁面 ────────────────────────────────────────────────────────────────────

const FiatWithdrawalListScreen = () => {
  const { t } = useTranslation()
  const [status, setStatus] = useState('pending')
  const [searchStatus, setSearchStatus] = useState('pending')
  const [list, setList] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)
  const [selected, setSelected] = useState(null)

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fiatWithdrawalsApi.list({
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
  }, [searchStatus, page, pageSize])

  useEffect(() => { fetchData() }, [fetchData])

  const handleSearch = () => {
    setPage(1)
    setSearchStatus(status)
  }

  const handleReset = () => {
    setStatus('pending')
    setSearchStatus('pending')
    setPage(1)
  }

  const handleDone = () => {
    setSelected(null)
    fetchData()
  }

  const renderStatus = (s) => {
    const colors = STATUS_COLOR_MAP[s] || { bg: '#F5F5F5', color: '#666' }
    return (
      <Chip
        label={t(`fiatWithdrawal.status.${s}`, s)}
        size="small"
        sx={{ bgcolor: colors.bg, color: colors.color, fontWeight: 600, fontSize: '12px', height: '24px' }}
      />
    )
  }

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: '#333', mb: '16px' }}>
        {t('page.fiatWithdrawals')}
      </Typography>

      {/* 篩選列 */}
      <Paper sx={{ p: '16px', mb: '16px', boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '13px', color: '#666', whiteSpace: 'nowrap' }}>
              {t('fiatWithdrawal.statusLabel')}
            </Typography>
            <Select
              size="small"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              displayEmpty
              sx={{ width: 140, fontSize: '13px', '& .MuiSelect-select': { py: '6px' } }}
            >
              <MenuItem value="" sx={{ fontSize: '13px' }}>{t('fiatWithdrawal.statusAll')}</MenuItem>
              {STATUS_OPTIONS.filter(Boolean).map((s) => (
                <MenuItem key={s} value={s} sx={{ fontSize: '13px' }}>{t(`fiatWithdrawal.status.${s}`, s)}</MenuItem>
              ))}
            </Select>
          </Box>

          <Button
            variant="text"
            size="small"
            onClick={handleReset}
            sx={{ color: '#666', fontSize: '13px', minWidth: 'auto' }}
          >
            {t('fiatWithdrawal.reset')}
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
            {t('fiatWithdrawal.filter')}
          </Button>
        </Box>
      </Paper>

      {/* 資料表 */}
      <TableContainer component={Paper} sx={{ boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: '#FAFAFA' }}>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.id')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.userId')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.amount')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.bankCode')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.bankAccount')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.accountName')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.status')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.createdAt')}</TableCell>
              <TableCell sx={thStyle}>{t('fiatWithdrawal.column.action')}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 9 }).map((__, j) => (
                    <TableCell key={j} sx={{ py: '10px' }}>
                      <Skeleton variant="text" sx={{ fontSize: '13px' }} />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : list.length === 0 ? (
              <TableRow>
                <TableCell colSpan={9} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('fiatWithdrawal.empty')}
                </TableCell>
              </TableRow>
            ) : (
              list.map((row) => (
                <TableRow
                  key={row.id}
                  hover
                  sx={{ cursor: 'pointer' }}
                  onClick={() => setSelected(row)}
                >
                  <TableCell sx={tdStyle}>#{row.id}</TableCell>
                  <TableCell sx={tdStyle}>{row.userId}</TableCell>
                  <TableCell sx={tdStyle}>{row.amount} {row.currency}</TableCell>
                  <TableCell sx={tdStyle}>{row.bankCode}</TableCell>
                  <TableCell sx={tdStyle}>{row.bankAccount}</TableCell>
                  <TableCell sx={tdStyle}>{row.accountName}</TableCell>
                  <TableCell sx={tdStyle}>{renderStatus(row.status)}</TableCell>
                  <TableCell sx={tdStyle}>{row.createdAt ? new Date(row.createdAt).toLocaleString() : '-'}</TableCell>
                  <TableCell sx={tdStyle}>
                    {row.status === 'pending' ? (
                      <Chip
                        label={t('fiatWithdrawal.review.actionChip')}
                        size="small"
                        clickable
                        sx={{ bgcolor: '#FFF3E0', color: '#E65100', fontWeight: 600, fontSize: '11px', height: '22px' }}
                      />
                    ) : (
                      <Typography sx={{ fontSize: '12px', color: '#AAA' }}>
                        {t('fiatWithdrawal.review.view')}
                      </Typography>
                    )}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        {/* 分頁 */}
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
            {t('fiatWithdrawal.pagination.total', { total })} {t('fiatWithdrawal.pagination.page', { current: page, total: totalPages })}
          </Typography>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '12px', color: '#999' }}>{t('fiatWithdrawal.pagination.pageSize')}</Typography>
            <Select
              size="small"
              value={pageSize}
              onChange={(e) => { setPageSize(e.target.value); setPage(1) }}
              sx={{ fontSize: '12px', '& .MuiSelect-select': { py: '4px' } }}
            >
              {PAGE_SIZE_OPTIONS.map((n) => (
                <MenuItem key={n} value={n} sx={{ fontSize: '12px' }}>{n}{t('fiatWithdrawal.pagination.items')}</MenuItem>
              ))}
            </Select>
          </Box>

          <Box sx={{ display: 'flex', gap: '4px' }}>
            <IconButton size="small" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              <ChevronLeftIcon fontSize="small" />
            </IconButton>
            <IconButton size="small" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              <ChevronRightIcon fontSize="small" />
            </IconButton>
          </Box>
        </Box>
      </TableContainer>

      <ReviewDialog
        item={selected}
        open={Boolean(selected)}
        onClose={() => setSelected(null)}
        onDone={handleDone}
        t={t}
      />
    </Box>
  )
}

const thStyle = { fontSize: '13px', fontWeight: 600, color: '#333', py: '10px' }
const tdStyle = { fontSize: '13px', color: '#555', py: '10px' }

export default FiatWithdrawalListScreen

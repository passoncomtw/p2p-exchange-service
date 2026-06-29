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
} from '@mui/material'
import {
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
} from '@mui/icons-material'
import { ordersApi } from 'src/apis/orders'

const PAGE_SIZE_OPTIONS = [10, 20, 50]

const STATUS_OPTIONS = ['', 'pending', 'paid', 'completed', 'cancelled', 'disputed']

const STATUS_COLOR_MAP = {
  pending: { bg: '#FFF3E0', color: '#E65100' },
  paid: { bg: '#E3F2FD', color: '#1565C0' },
  completed: { bg: '#E8F5E9', color: '#2E7D32' },
  cancelled: { bg: '#FAFAFA', color: '#999' },
  disputed: { bg: '#FCE4EC', color: '#C62828' },
}

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
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('order.loading')}
                </TableCell>
              </TableRow>
            ) : list.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('order.empty')}
                </TableCell>
              </TableRow>
            ) : (
              list.map((row) => (
                <TableRow key={row.id} hover>
                  <TableCell sx={tdStyle}>{row.orderNo}</TableCell>
                  <TableCell sx={tdStyle}>{t(`order.type.${row.listingType}`, row.listingType)}</TableCell>
                  <TableCell sx={tdStyle}>{row.cryptoAmount} {row.cryptoCurrency}</TableCell>
                  <TableCell sx={tdStyle}>{row.fiatAmount} {row.fiatCurrency}</TableCell>
                  <TableCell sx={tdStyle}>{row.price}</TableCell>
                  <TableCell sx={tdStyle}>{renderStatus(row.status)}</TableCell>
                  <TableCell sx={tdStyle}>{formatDateTime(row.createdAt)}</TableCell>
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
            <IconButton size="small" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              <ChevronLeftIcon fontSize="small" />
            </IconButton>
            <IconButton size="small" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              <ChevronRightIcon fontSize="small" />
            </IconButton>
          </Box>
        </Box>
      </TableContainer>
    </Box>
  )
}

const thStyle = { fontSize: '13px', fontWeight: 600, color: '#333', py: '10px' }
const tdStyle = { fontSize: '13px', color: '#555', py: '10px' }

export default OrderListScreen

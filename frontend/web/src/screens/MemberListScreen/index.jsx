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
} from '@mui/material'
import {
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
} from '@mui/icons-material'
import { membersApi } from 'src/apis/members'

const PAGE_SIZE_OPTIONS = [10, 20, 50]

const MemberListScreen = () => {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState('')
  const [searchKeyword, setSearchKeyword] = useState('')
  const [list, setList] = useState([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [loading, setLoading] = useState(false)

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const res = await membersApi.list({
        keyword: searchKeyword,
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
  }, [searchKeyword, page, pageSize])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleSearch = () => {
    setPage(1)
    setSearchKeyword(keyword)
  }

  const handleReset = () => {
    setKeyword('')
    setSearchKeyword('')
    setPage(1)
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') handleSearch()
  }

  const formatDateTime = (iso) => {
    if (!iso) return '-'
    return new Date(iso).toLocaleString()
  }

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: '#333', mb: '16px' }}>
        {t('page.memberList')}
      </Typography>

      <Paper sx={{ p: '16px', mb: '16px', boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '13px', color: '#666', whiteSpace: 'nowrap' }}>
              {t('member.keyword')}
            </Typography>
            <TextField
              size="small"
              placeholder={t('member.keywordPlaceholder')}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={handleKeyDown}
              sx={{ width: 200, '& .MuiInputBase-input': { fontSize: '13px', py: '6px' } }}
            />
          </Box>

          <Button
            variant="text"
            size="small"
            onClick={handleReset}
            sx={{ color: '#666', fontSize: '13px', minWidth: 'auto' }}
          >
            {t('member.reset')}
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
            {t('member.filter')}
          </Button>
        </Box>
      </Paper>

      <TableContainer component={Paper} sx={{ boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Table size="small">
          <TableHead>
            <TableRow sx={{ bgcolor: '#FAFAFA' }}>
              <TableCell sx={thStyle}>ID</TableCell>
              <TableCell sx={thStyle}>{t('member.column.username')}</TableCell>
              <TableCell sx={thStyle}>{t('member.column.email')}</TableCell>
              <TableCell sx={thStyle}>{t('member.column.createdAt')}</TableCell>
              <TableCell sx={thStyle}>{t('member.column.updatedAt')}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={5} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('member.loading')}
                </TableCell>
              </TableRow>
            ) : list.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} align="center" sx={{ py: '40px', color: '#999' }}>
                  {t('member.empty')}
                </TableCell>
              </TableRow>
            ) : (
              list.map((row) => (
                <TableRow key={row.id} hover>
                  <TableCell sx={tdStyle}>{row.id}</TableCell>
                  <TableCell sx={tdStyle}>{row.username}</TableCell>
                  <TableCell sx={tdStyle}>{row.email || '-'}</TableCell>
                  <TableCell sx={tdStyle}>{formatDateTime(row.createdAt)}</TableCell>
                  <TableCell sx={tdStyle}>{formatDateTime(row.updatedAt)}</TableCell>
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
            {t('member.pagination.total', { total })} {t('member.pagination.page', { current: page, total: totalPages })}
          </Typography>

          <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
            <Typography sx={{ fontSize: '12px', color: '#999' }}>{t('member.pagination.pageSize')}</Typography>
            <Select
              size="small"
              value={pageSize}
              onChange={(e) => { setPageSize(e.target.value); setPage(1) }}
              sx={{ fontSize: '12px', '& .MuiSelect-select': { py: '4px' } }}
            >
              {PAGE_SIZE_OPTIONS.map((n) => (
                <MenuItem key={n} value={n} sx={{ fontSize: '12px' }}>{n}{t('member.pagination.items')}</MenuItem>
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

export default MemberListScreen

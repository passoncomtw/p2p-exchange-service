import { useState, useEffect } from 'react'
import {
  Box,
  Typography,
  Paper,
  Chip,
  Divider,
  Alert,
  CircularProgress,
  Tooltip,
  IconButton,
} from '@mui/material'
import { ContentCopy as CopyIcon, CheckCircle as CheckIcon } from '@mui/icons-material'
import server from 'src/apis'

const BALANCE_DIFF_THRESHOLD = 0.1 // 10%

const InfoRow = ({ label, value, copyable }) => {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: '10px', borderBottom: '1px solid #F5F5F5' }}>
      <Typography sx={{ fontSize: '13px', color: '#888', minWidth: 160 }}>{label}</Typography>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: '4px', flex: 1, justifyContent: 'flex-end' }}>
        <Typography sx={{ fontSize: '13px', color: '#333', fontWeight: 500, textAlign: 'right', wordBreak: 'break-all' }}>
          {value ?? '-'}
        </Typography>
        {copyable && value && (
          <Tooltip title={copied ? '已複製' : '複製'}>
            <IconButton size="small" onClick={handleCopy} sx={{ color: copied ? '#2E7D32' : '#999', p: '2px' }}>
              {copied ? <CheckIcon sx={{ fontSize: 14 }} /> : <CopyIcon sx={{ fontSize: 14 }} />}
            </IconButton>
          </Tooltip>
        )}
      </Box>
    </Box>
  )
}

const StatusChip = ({ enabled }) => (
  <Chip
    label={enabled ? '啟用' : '未啟用'}
    size="small"
    sx={{
      bgcolor: enabled ? '#E8F5E9' : '#F5F5F5',
      color: enabled ? '#2E7D32' : '#999',
      fontWeight: 600,
      fontSize: '12px',
    }}
  />
)

const SectionTitle = ({ children }) => (
  <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#999', mb: '8px', textTransform: 'uppercase', letterSpacing: '0.5px' }}>
    {children}
  </Typography>
)

const formatBalance = (val) => {
  const num = parseFloat(val)
  if (isNaN(num)) return val ?? '-'
  return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })
}

const calcDiffPct = (onChain, virtual) => {
  const a = parseFloat(onChain)
  const b = parseFloat(virtual)
  if (isNaN(a) || isNaN(b) || a === 0) return null
  return Math.abs(a - b) / a
}

const PlatformWalletScreen = () => {
  const [info, setInfo] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    setLoading(true)
    server.get('/backend/platform/wallet-info')
      .then((r) => setInfo(r.data.data))
      .catch(() => setError('載入平台錢包資訊失敗'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: 200 }}>
        <CircularProgress size={28} sx={{ color: '#FFC107' }} />
      </Box>
    )
  }

  if (error) {
    return <Alert severity="error" sx={{ m: 2 }}>{error}</Alert>
  }

  const tron = info?.tron
  const ecpay = info?.ecpay
  const diffPct = tron ? calcDiffPct(tron.onChainBalance, tron.platformVirtualBalance) : null
  const showDiffWarning = diffPct !== null && diffPct > BALANCE_DIFF_THRESHOLD

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: '#333', mb: '16px' }}>
        平台錢包設定（唯讀）
      </Typography>

      {showDiffWarning && (
        <Alert severity="warning" sx={{ mb: '16px', fontSize: '13px' }}>
          鏈上餘額與平台虛擬餘額差距超過 {Math.round(diffPct * 100)}%，請確認是否有未處理的充提記錄。
        </Alert>
      )}

      {/* USDT / Tron 區塊 */}
      <Paper sx={{ p: '20px', mb: '16px', boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '10px', mb: '16px' }}>
          <SectionTitle>USDT / Tron TRC-20</SectionTitle>
          <StatusChip enabled={tron?.enabled} />
        </Box>

        <InfoRow label="網路" value={tron?.network} />
        <InfoRow label="熱錢包地址" value={tron?.hotWalletAddress} copyable />
        <InfoRow label="USDT 合約地址" value={tron?.usdtContractAddress} copyable />
        <InfoRow label="確認區塊數" value={tron?.confirmationBlocks} />

        <Divider sx={{ my: '12px' }} />
        <SectionTitle>餘額資訊</SectionTitle>

        <InfoRow label="鏈上實際餘額" value={tron ? `${formatBalance(tron.onChainBalance)} USDT` : '-'} />
        <InfoRow label="平台虛擬總餘額" value={tron ? `${formatBalance(tron.platformVirtualBalance)} USDT` : '-'} />

        {tron && (
          <Box sx={{ mt: '10px', display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Typography sx={{ fontSize: '12px', color: '#888' }}>餘額差距</Typography>
            <Typography sx={{ fontSize: '13px', fontWeight: 600, color: showDiffWarning ? '#C62828' : '#2E7D32' }}>
              {diffPct !== null ? `${(diffPct * 100).toFixed(2)}%` : '-'}
            </Typography>
            {!showDiffWarning && diffPct !== null && (
              <Chip label="正常" size="small" sx={{ bgcolor: '#E8F5E9', color: '#2E7D32', fontSize: '11px', height: '20px' }} />
            )}
          </Box>
        )}
      </Paper>

      {/* ECPay 區塊 */}
      <Paper sx={{ p: '20px', mb: '16px', boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: '10px', mb: '16px' }}>
          <SectionTitle>TWD / ECPay</SectionTitle>
          <StatusChip enabled={ecpay?.enabled} />
        </Box>

        <InfoRow label="商店代號" value={ecpay?.merchantId} />
        <InfoRow label="API 網址" value={ecpay?.baseUrl} />

        <Box sx={{ mt: '12px', p: '12px', bgcolor: '#FAFAFA', borderRadius: '4px', border: '1px solid #EBEBEB' }}>
          <Typography sx={{ fontSize: '12px', color: '#888', lineHeight: 1.7 }}>
            法幣錢包由平台銀行帳號管理，v0.3.0 不提供鏈上查詢。虛擬餘額總計請查閱後台會員列表。
          </Typography>
        </Box>
      </Paper>

      <Typography sx={{ fontSize: '12px', color: '#CCC', textAlign: 'right', mt: '8px' }}>
        本頁面為唯讀，v0.3.0 不提供編輯功能。
      </Typography>
    </Box>
  )
}

export default PlatformWalletScreen

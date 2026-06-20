import { Box } from '@mui/material'
import { useTranslation } from 'react-i18next'
import { typeColor, statusColor } from 'src/theme'

// 類型標籤（買綠賣紅，實心）
export const TypeTag = ({ type }) => {
  const { t } = useTranslation()
  const color = typeColor(type)
  return (
    <Box
      component="span"
      sx={{
        display: 'inline-block',
        px: '8px',
        py: '2px',
        borderRadius: '4px',
        fontSize: '12px',
        fontWeight: 600,
        color: '#FFFFFF',
        bgcolor: color,
      }}
    >
      {t(`order.type.${type}`)}
    </Box>
  )
}

// 狀態標籤（淡底+語意色，與類型標籤視覺區分）
export const StatusTag = ({ status }) => {
  const { t } = useTranslation()
  const color = statusColor(status)
  return (
    <Box
      component="span"
      sx={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '6px',
        px: '8px',
        py: '2px',
        borderRadius: '4px',
        fontSize: '12px',
        fontWeight: 500,
        color,
        border: `1px solid ${color}`,
        bgcolor: `${color}14`,
      }}
    >
      <Box component="span" sx={{ width: 6, height: 6, borderRadius: '50%', bgcolor: color }} />
      {t(`order.status.${status}`)}
    </Box>
  )
}

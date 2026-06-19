import { useTranslation } from 'react-i18next'
import { Typography, Card, CardContent, Box } from '@mui/material'
import { tokens } from 'src/theme'

const DashboardScreen = () => {
  const { t } = useTranslation()

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: tokens.textPrimary, mb: '16px' }}>
        {t('page.dashboard')}
      </Typography>
      <Card sx={{ boxShadow: 'none', border: `1px solid ${tokens.borderCard}` }}>
        <CardContent>
          <Typography sx={{ fontSize: '13px', color: tokens.textSecondary }}>
            {t('common.welcome')}
          </Typography>
        </CardContent>
      </Card>
    </Box>
  )
}

export default DashboardScreen

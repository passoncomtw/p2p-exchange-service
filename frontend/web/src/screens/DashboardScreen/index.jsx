import { useTranslation } from 'react-i18next'
import { Typography, Card, CardContent, Box } from '@mui/material'

const DashboardScreen = () => {
  const { t } = useTranslation()

  return (
    <Box>
      <Typography sx={{ fontSize: '16px', fontWeight: 600, color: '#333', mb: '16px' }}>
        {t('page.dashboard')}
      </Typography>
      <Card sx={{ boxShadow: 'none', border: '1px solid #EBEBEB' }}>
        <CardContent>
          <Typography sx={{ fontSize: '14px', color: '#666' }}>
            {t('common.welcome')}
          </Typography>
        </CardContent>
      </Card>
    </Box>
  )
}

export default DashboardScreen

import { Suspense } from 'react'
import { CircularProgress, Box } from '@mui/material'

const Loadable = (Component) => (props) => (
  <Suspense
    fallback={
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '200px' }}>
        <CircularProgress sx={{ color: '#FFC107' }} />
      </Box>
    }
  >
    <Component {...props} />
  </Suspense>
)

export default Loadable

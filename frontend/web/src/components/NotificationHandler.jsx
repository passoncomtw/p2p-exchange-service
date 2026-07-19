import { useEffect } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { Alert, Snackbar } from '@mui/material'
import { popNotification } from 'src/slices/notificationSlice'

const SEVERITY_MAP = {
  success: 'success',
  error: 'error',
  warning: 'warning',
  info: 'info',
}

export function NotificationHandler() {
  const dispatch = useDispatch()
  const current = useSelector((s) => s.notification.queue[0])

  const handleClose = (_event, reason) => {
    if (reason === 'clickaway') return
    dispatch(popNotification())
  }

  const autoHide = current?.type === 'success' || current?.type === 'info' ? 4000 : null

  return (
    <Snackbar
      open={Boolean(current)}
      autoHideDuration={autoHide}
      onClose={handleClose}
      anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
    >
      {current ? (
        <Alert
          severity={SEVERITY_MAP[current?.type] ?? 'info'}
          onClose={handleClose}
          sx={{ minWidth: 280, fontSize: '13px' }}
        >
          {current.message}
        </Alert>
      ) : undefined}
    </Snackbar>
  )
}

import 'src/i18n'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider } from 'react-redux'
import { PersistGate } from 'redux-persist/integration/react'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider } from '@mui/material/styles'
import CssBaseline from '@mui/material/CssBaseline'
import store, { persistor } from 'src/store/index'
import theme from 'src/theme'
import App from './App.jsx'
import './index.css'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <BrowserRouter>
          <ThemeProvider theme={theme}>
            <CssBaseline />
            <App />
            <div
              id="aria-live-region"
              aria-live="polite"
              aria-atomic="true"
              style={{
                position: 'absolute',
                left: '-9999px',
                width: '1px',
                height: '1px',
                overflow: 'hidden',
              }}
            />
          </ThemeProvider>
        </BrowserRouter>
      </PersistGate>
    </Provider>
  </StrictMode>,
)

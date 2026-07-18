import { createTheme } from '@mui/material'

export const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: { main: '#f97316' },
    secondary: { main: '#38bdf8' },
    background: { default: '#0b1120', paper: '#111a2e' },
  },
  typography: {
    fontFamily: '"Inter", "Helvetica", "Arial", sans-serif',
    h4: { fontWeight: 700 },
  },
})

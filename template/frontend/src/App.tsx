import { useEffect, useMemo, useState } from 'react'
import { AppBar, Box, Chip, Container, Grid2 as Grid, Toolbar, Typography } from '@mui/material'
import BoltIcon from '@mui/icons-material/Bolt'
import { MetricCard } from './components/MetricCard'
import { UsageChart } from './components/UsageChart'
import { ServiceTable } from './components/ServiceTable'
import { fetchMetrics } from './lib/api'
import { bucketByHour, computeOverview, summarizeByService } from './lib/aggregate'
import { formatCount, formatLatency, formatRate } from './lib/format'
import type { MetricPoint } from './lib/types'

export default function App() {
  const [points, setPoints] = useState<MetricPoint[]>([])
  const [live, setLive] = useState(false)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    let cancelled = false
    fetchMetrics().then((result) => {
      if (cancelled) return
      setPoints(result.points)
      setLive(result.live)
      setLoaded(true)
    })
    return () => {
      cancelled = true
    }
  }, [])

  const overview = useMemo(() => computeOverview(points), [points])
  const summaries = useMemo(() => summarizeByService(points), [points])
  const hourly = useMemo(() => bucketByHour(points), [points])

  return (
    <Box sx={{ minHeight: '100vh' }}>
      <AppBar position="static" color="transparent" elevation={0}>
        <Toolbar>
          <BoltIcon color="primary" sx={{ mr: 1 }} />
          <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 700 }}>
            Anvil Analytics
          </Typography>
          <Chip
            size="small"
            data-testid="data-source"
            label={live ? 'live data' : 'sample data'}
            color={live ? 'success' : 'default'}
            variant="outlined"
          />
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Typography variant="h4" gutterBottom>
          Platform overview
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
          Last 24 hours across all services
        </Typography>

        {loaded && (
          <>
            <Grid container spacing={2} sx={{ mb: 3 }}>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard label="Requests" value={formatCount(overview.totalRequests)} />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard label="Error rate" value={formatRate(overview.overallErrorRate)} />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard label="Worst p95" value={formatLatency(overview.worstP95Ms)} />
              </Grid>
              <Grid size={{ xs: 6, md: 3 }}>
                <MetricCard label="Services" value={String(overview.serviceCount)} />
              </Grid>
            </Grid>

            <UsageChart data={hourly} />
            <ServiceTable summaries={summaries} />
          </>
        )}
      </Container>
    </Box>
  )
}

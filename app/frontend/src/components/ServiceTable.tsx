import {
  Card,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import { formatCount, formatLatency, formatRate } from '../lib/format'
import type { ServiceSummary } from '../lib/types'

export function ServiceTable({ summaries }: { summaries: ServiceSummary[] }) {
  return (
    <Card variant="outlined">
      <TableContainer>
        <Table size="small" data-testid="service-table">
          <TableHead>
            <TableRow>
              <TableCell>Service</TableCell>
              <TableCell align="right">Requests (24h)</TableCell>
              <TableCell align="right">Error rate</TableCell>
              <TableCell align="right">Avg p95</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {summaries.map((s) => (
              <TableRow key={s.service}>
                <TableCell>
                  <Typography variant="body2" fontFamily="monospace">
                    {s.service}
                  </Typography>
                </TableCell>
                <TableCell align="right">{formatCount(s.totalRequests)}</TableCell>
                <TableCell align="right">{formatRate(s.errorRate)}</TableCell>
                <TableCell align="right">{formatLatency(s.avgP95LatencyMs)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Card>
  )
}

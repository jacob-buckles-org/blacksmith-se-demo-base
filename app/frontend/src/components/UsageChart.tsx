import { Card, CardContent, Typography } from '@mui/material'
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

interface HourBucket {
  hour: string
  requests: number
  errors: number
}

export function UsageChart({ data }: { data: HourBucket[] }) {
  return (
    <Card variant="outlined" sx={{ mb: 3 }} data-testid="usage-chart">
      <CardContent>
        <Typography variant="subtitle1" gutterBottom>
          Request volume by hour
        </Typography>
        <ResponsiveContainer width="100%" height={280}>
          <AreaChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
            <XAxis dataKey="hour" tick={{ fontSize: 11 }} tickFormatter={(h: string) => h.slice(11, 16)} />
            <YAxis tick={{ fontSize: 11 }} />
            <Tooltip />
            <Area type="monotone" dataKey="requests" stroke="#f97316" fill="#f9731633" />
            <Area type="monotone" dataKey="errors" stroke="#ef4444" fill="#ef444433" />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  )
}

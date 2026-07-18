import { Card, CardContent, Typography } from '@mui/material'

export function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="overline" color="text.secondary">
          {label}
        </Typography>
        <Typography variant="h4" data-testid={`metric-${label.toLowerCase().replace(/\s+/g, '-')}`}>
          {value}
        </Typography>
      </CardContent>
    </Card>
  )
}

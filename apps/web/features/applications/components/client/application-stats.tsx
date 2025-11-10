'use client'

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@hookie/ui/components/card'
import { Activity, TrendingUp, Webhook } from 'lucide-react'

interface ApplicationStatsProps {
  topicCount?: number
  webhooksToday?: number
  successRate?: string
}

export function ApplicationStats({
  topicCount = 0,
  webhooksToday = 0,
  successRate = '-',
}: ApplicationStatsProps) {
  return (
    <div className="grid md:grid-cols-3 gap-6 mb-8">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Topics</CardTitle>
          <Webhook className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{topicCount}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Webhooks Today</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{webhooksToday}</div>
          <p className="text-xs text-muted-foreground">
            {webhooksToday === 0 ? 'No activity yet' : 'Active today'}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
          <TrendingUp className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{successRate}</div>
          <p className="text-xs text-muted-foreground">
            {successRate === '-' ? 'No data available' : 'Overall success rate'}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

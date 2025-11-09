import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@hookie/ui/components/card'

export function TopicCard() {
  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer">
      <CardHeader>
        <CardTitle>Topic</CardTitle>
        <CardDescription>Description</CardDescription>
      </CardHeader>
    </Card>
  )
}

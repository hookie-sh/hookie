"use client";

import { Badge } from "@hookie/ui/components/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hookie/ui/components/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@hookie/ui/components/tooltip";

interface Request {
  id: string;
  timestamp: string;
  endpoint: string;
  topic: string;
  application: string;
  status: number;
  duration: number;
}

export default function UsagePage() {
  // Fake data for now
  const usageLimit = 10000; // Monthly limit
  const currentUsage = 7234; // Current usage
  const usagePercentage = (currentUsage / usageLimit) * 100;

  const requests: Request[] = [
    {
      id: "req_1",
      timestamp: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
      endpoint: "/webhooks/github",
      topic: "github",
      application: "my-app",
      status: 200,
      duration: 45,
    },
    {
      id: "req_2",
      timestamp: new Date(Date.now() - 1000 * 60 * 10).toISOString(),
      endpoint: "/webhooks/stripe",
      topic: "stripe",
      application: "my-app",
      status: 200,
      duration: 32,
    },
    {
      id: "req_3",
      timestamp: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
      endpoint: "/webhooks/github",
      topic: "github",
      application: "my-app",
      status: 201,
      duration: 78,
    },
    {
      id: "req_4",
      timestamp: new Date(Date.now() - 1000 * 60 * 20).toISOString(),
      endpoint: "/webhooks/stripe",
      topic: "stripe",
      application: "my-app",
      status: 500,
      duration: 156,
    },
    {
      id: "req_5",
      timestamp: new Date(Date.now() - 1000 * 60 * 25).toISOString(),
      endpoint: "/webhooks/some-other-service",
      topic: "some-other-service",
      application: "my-app",
      status: 200,
      duration: 41,
    },
  ];

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const formatDateUTC = (dateString: string) => {
    return new Date(dateString).toUTCString();
  };

  const isSuccessful = (status: number) => {
    return status >= 200 && status < 300;
  };

  const remainingRequests = usageLimit - currentUsage;
  const isNearLimit = usagePercentage >= 80;

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-3xl font-bold mb-2">Usage</h2>
        <p className="text-muted-foreground mb-6">
          Track your account usage and monitor API requests
        </p>

        <div className="border border-border rounded-lg p-6">
          <div className="flex justify-between items-center mb-2">
            <span className="text-sm font-medium">Monthly Requests</span>
            <span className="text-sm text-muted-foreground">
              {currentUsage.toLocaleString()} / {usageLimit.toLocaleString()}
            </span>
          </div>
          <div className="w-full bg-secondary rounded-full h-3 mb-4">
            <div
              className={`h-3 rounded-full transition-all ${
                isNearLimit ? "bg-destructive" : "bg-primary"
              }`}
              style={{ width: `${Math.min(usagePercentage, 100)}%` }}
            />
          </div>
          <div className="flex justify-between items-center text-sm">
            <span className="text-muted-foreground">
              {remainingRequests.toLocaleString()} requests remaining
            </span>
            {isNearLimit && (
              <Badge variant="destructive">Approaching limit</Badge>
            )}
          </div>
        </div>
      </div>

      <div>
        <h3 className="text-xl font-semibold mb-4">Recent Requests</h3>
        {requests.length === 0 ? (
          <div className="border border-border rounded-lg p-8 text-center">
            <p className="text-sm text-muted-foreground">
              No requests processed yet.
            </p>
          </div>
        ) : (
          <TooltipProvider>
            <div className="border border-border rounded-lg overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Status</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead>Destination</TableHead>
                    <TableHead>Topic</TableHead>
                    <TableHead>Application</TableHead>

                    <TableHead className="text-right">Duration (ms)</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {requests.map((request) => (
                    <TableRow key={request.id}>
                      <TableCell>
                        <Badge
                          variant={
                            isSuccessful(request.status)
                              ? "default"
                              : "destructive"
                          }
                        >
                          {isSuccessful(request.status)
                            ? "Successful"
                            : "Failed"}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm">
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="cursor-help">
                              {formatDate(request.timestamp)}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent>
                            <p>UTC: {formatDateUTC(request.timestamp)}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TableCell>
                      <TableCell>
                        <span className="font-mono text-sm">
                          {request.endpoint}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm">{request.topic}</span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm">{request.application}</span>
                      </TableCell>

                      <TableCell className="text-right text-sm">
                        {request.duration}ms
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </TooltipProvider>
        )}
      </div>
    </div>
  );
}

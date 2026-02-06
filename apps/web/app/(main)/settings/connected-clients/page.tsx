"use client";

import { ConnectedClient } from "@/features/connected-clients/db/server";
import { usePresence } from "@/features/connected-clients/hooks/use-presence";
import { fetcher } from "@/utils/api";
import { useAuth } from "@clerk/nextjs";
import { Badge } from "@hookie/ui/components/badge";
import { Button } from "@hookie/ui/components/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hookie/ui/components/table";
import { useMemo, useState } from "react";
import useSWR from "swr";

export default function ConnectedClientsPage() {
  const { userId, orgId } = useAuth();
  const [disconnecting, setDisconnecting] = useState<string | null>(null);
  const onlineMachineIds = usePresence();

  const {
    data: clients,
    error,
    isLoading,
    mutate,
  } = useSWR<ConnectedClient[]>(
    userId ? "/api/connected-clients" : null,
    fetcher,
    {
      revalidateOnFocus: true,
      revalidateOnReconnect: true,
      refreshInterval: 5000, // Refresh every 5 seconds
    }
  );

  // Enrich clients with online status based on connection_count
  const enrichedClients = useMemo(() => {
    if (!clients) return [];
    const enriched = clients.map((client) => ({
      ...client,
      is_online: client.connection_count != null && client.connection_count > 0,
    }));

    // Debug: log the enrichment
    console.log("[ConnectedClientsPage] Clients:", clients.length);
    console.log(
      "[ConnectedClientsPage] Online machine IDs:",
      Array.from(onlineMachineIds)
    );
    console.log(
      "[ConnectedClientsPage] Enriched clients:",
      enriched.map((c) => ({ id: c.id, is_online: c.is_online }))
    );

    return enriched;
  }, [clients, onlineMachineIds]);

  // Calculate connection statistics
  const connectionStats = useMemo(() => {
    if (!clients) return { total: 0, connected: 0, disconnected: 0 };
    const total = clients.length;
    const connected = enrichedClients.filter((c) => c.is_online).length;
    const disconnected = total - connected;
    return { total, connected, disconnected };
  }, [clients, enrichedClients]);

  const handleDisconnect = async (clientId: string) => {
    setDisconnecting(clientId);
    try {
      const response = await fetch(`/api/connected-clients/${clientId}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        throw new Error("Failed to disconnect client");
      }

      await mutate();
    } catch (error) {
      console.error("Error disconnecting client:", error);
      alert("Failed to disconnect client. Please try again.");
    } finally {
      setDisconnecting(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const getStatus = (client: ConnectedClient & { is_online?: boolean }) => {
    // Use connection_count to determine status
    // connection_count > 0 means connected
    if (client.connection_count != null && client.connection_count > 0) {
      return "connected";
    }
    return "disconnected";
  };

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Loading connected clients...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mb-4 p-4 bg-destructive/10 text-destructive rounded-md">
        {error instanceof Error
          ? error.message
          : "Failed to load connected clients"}
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-3xl font-bold mb-2">Connected Clients</h2>
        <p className="text-muted-foreground mb-4">
          Manage CLI clients connected to your account
        </p>
        {clients && clients.length > 0 && (
          <div className="flex gap-4 text-sm">
            <div className="flex items-center gap-2">
              <Badge variant="default">{connectionStats.connected}</Badge>
              <span className="text-muted-foreground">Connected</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="secondary">{connectionStats.disconnected}</Badge>
              <span className="text-muted-foreground">Disconnected</span>
            </div>
            <div className="flex items-center gap-2">
              <Badge variant="outline">{connectionStats.total}</Badge>
              <span className="text-muted-foreground">Total</span>
            </div>
          </div>
        )}
      </div>

      {!enrichedClients || enrichedClients.length === 0 ? (
        <div className="border border-border rounded-lg p-8 text-center">
          <p className="text-sm text-muted-foreground">
            No connected clients. Use the CLI to listen to applications or
            topics.
          </p>
        </div>
      ) : (
        <div className="border border-border rounded-lg overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Connections</TableHead>
                <TableHead>Connected At</TableHead>
                <TableHead>Disconnected At</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {enrichedClients.map((client) => {
                const status = getStatus(client);
                const isConnected = status === "connected";
                // Show first 8 characters of id (mach_<ksuid>) for display
                const idDisplay =
                  client.id.length > 8
                    ? `${client.id.substring(0, 8)}...`
                    : client.id;

                return (
                  <TableRow key={client.id}>
                    <TableCell>
                      <div className="flex flex-col">
                        <span className="font-mono text-sm">{idDisplay}</span>
                        <span className="text-xs text-muted-foreground">
                          {client.id}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={isConnected ? "default" : "secondary"}>
                        {isConnected ? "Connected" : "Disconnected"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-sm">
                      {client.connection_count != null
                        ? client.connection_count
                        : 0}
                    </TableCell>
                    <TableCell className="text-sm">
                      {formatDate(client.connected_at)}
                    </TableCell>
                    <TableCell className="text-sm">
                      {client.disconnected_at
                        ? formatDate(client.disconnected_at)
                        : "-"}
                    </TableCell>
                    <TableCell className="text-right">
                      {isConnected && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDisconnect(client.id)}
                          disabled={disconnecting === client.id}
                        >
                          {disconnecting === client.id
                            ? "Disconnecting..."
                            : "Disconnect"}
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}

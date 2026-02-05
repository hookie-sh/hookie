"use client";

import { supabaseClient } from "@/clients/supabase.client";
import { useAuth } from "@clerk/nextjs";
import { useEffect, useState } from "react";

/**
 * Hook to subscribe to Postgres changes for connected clients
 * Returns a Set of machine IDs that are currently online (connection_count > 0)
 */
export function usePresence() {
  const { userId, orgId, isLoaded } = useAuth();
  const [onlineMachineIds, setOnlineMachineIds] = useState<Set<string>>(
    new Set()
  );

  useEffect(() => {
    if (!isLoaded || !userId) return;

    const channelName = `connected-clients:${userId}:${orgId || "personal"}`;

    // Create channel for Postgres changes
    const channel = supabaseClient
      .channel(channelName)
      .on(
        "postgres_changes",
        {
          event: "*", // Listen to all events (INSERT, UPDATE, DELETE)
          schema: "public",
          table: "connected_clients",
          filter: `user_id=eq.${userId}`,
        },
        (payload) => {
          // Filter by org_id in the callback since Supabase only supports single-column filters
          const record = (payload.new || payload.old) as {
            id: string;
            user_id: string;
            org_id: string; // Empty string for personal, org ID for organizations
            connection_count: number | null;
          };

          // Skip if org_id doesn't match
          const expectedOrgId = orgId || "";
          if (record.org_id !== expectedOrgId) {
            return;
          }

          // Update online machine IDs based on connection_count
          setOnlineMachineIds((prev) => {
            const updated = new Set(prev);

            if (
              payload.eventType === "INSERT" ||
              payload.eventType === "UPDATE"
            ) {
              const newRecord = payload.new as {
                id: string;
                connection_count: number | null;
              };
              const oldRecord = payload.old as
                | {
                    id: string;
                    connection_count: number | null;
                  }
                | undefined;

              // connection_count should always be in the payload now (backend ensures it)
              // But handle the case where it might be missing (fallback to old value)
              let connectionCount = newRecord.connection_count;
              if (
                connectionCount === undefined &&
                payload.eventType === "UPDATE" &&
                oldRecord
              ) {
                connectionCount = oldRecord.connection_count;
              }

              // Online if connection_count > 0
              if (connectionCount != null && connectionCount > 0) {
                updated.add(newRecord.id);
              } else {
                updated.delete(newRecord.id);
              }
            } else if (payload.eventType === "DELETE") {
              const oldRecord = payload.old as { id: string };
              updated.delete(oldRecord.id);
            }

            return updated;
          });
        }
      )
      .subscribe((status) => {
        console.log(`[usePresence] Channel subscription status:`, status);
        if (status === "SUBSCRIBED") {
          // Initial load: fetch current online clients from DB
          const loadInitialState = async () => {
            try {
              let query = supabaseClient
                .from("connected_clients")
                .select("id, connection_count")
                .eq("user_id", userId);

              // org_id is empty string for personal accounts, org ID for organizations
              query = query.eq("org_id", orgId || "");

              const { data, error } = await query;

              if (error) {
                console.error(
                  "[usePresence] Error loading initial state:",
                  error
                );
                return;
              }

              const online = new Set<string>();
              if (data) {
                data.forEach((client) => {
                  // Online if connection_count > 0
                  if (
                    client.connection_count != null &&
                    client.connection_count > 0
                  ) {
                    online.add(client.id);
                  }
                });
              }

              setOnlineMachineIds(online);
            } catch (error) {
              console.error("[usePresence] Error in loadInitialState:", error);
            }
          };

          loadInitialState();
        } else if (status === "CHANNEL_ERROR") {
          console.error(
            `[usePresence] Channel subscription error. Check that NEXT_PUBLIC_SUPABASE_URL and NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY are set correctly.`
          );
        } else if (status === "TIMED_OUT") {
          console.error(
            `[usePresence] Channel subscription timed out. Check your network connection and Supabase configuration.`
          );
        } else if (status === "CLOSED") {
          console.warn(`[usePresence] Channel closed`);
        }
      });

    // Cleanup on unmount
    return () => {
      channel.unsubscribe();
    };
  }, [isLoaded, userId, orgId]);

  return onlineMachineIds;
}

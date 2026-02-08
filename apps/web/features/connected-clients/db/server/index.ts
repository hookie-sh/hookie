import { createSupabaseServerClient } from "@/clients/supabase.server";
import { supabaseServiceClient } from "@/clients/supabase.service";

export interface ConnectedClient {
  id: string; // Part of composite primary key (id, user_id, org_id) - the mach_<ksuid> from CLI
  user_id: string;
  org_id: string; // Empty string for personal, org ID for organization
  connected_at: string;
  disconnected_at: string | null;
  connection_count: number | null; // Deprecated: kept for backward compatibility
  created_at: string;
  updated_at: string;
}

export async function getConnectedClientsByUserId(
  userId: string,
  orgId?: string | null
) {
  const supabase = createSupabaseServerClient();
  // Build query
  let query = supabase
    .from("connected_clients")
    .select("*")
    .eq("user_id", userId)
    .order("connected_at", { ascending: false })
    .limit(100);

  // Always filter by org_id: empty string for personal accounts (null/undefined), org ID for organizations
  query = query.eq("org_id", orgId || "");

  const { data: clients, error: clientsError } = await query;

  if (clientsError) throw clientsError;
  if (!clients || clients.length === 0) return [];

  return clients as ConnectedClient[];
}

export async function disconnectClient(clientId: string) {
  const supabase = createSupabaseServerClient();
  // Update database first - set disconnected_at and connection_count to 0
  const { error } = await supabase
    .from("connected_clients")
    .update({
      disconnected_at: new Date().toISOString(),
      connection_count: 0,
    })
    .eq("id", clientId)
    .is("disconnected_at", null);

  if (error) throw error;

  // Send broadcast message to force disconnect
  // The channel name is the machine_id (clientId)
  const channel = supabaseServiceClient.channel(clientId, {
    config: {
      broadcast: {
        ack: false,
        self: false,
      },
    },
  });

  // Subscribe, send broadcast, then unsubscribe
  await new Promise<void>((resolve, reject) => {
    const timeout = setTimeout(() => {
      channel.unsubscribe();
      reject(new Error("Timeout waiting for broadcast"));
    }, 5000);

    channel.subscribe((status) => {
      if (status === "SUBSCRIBED") {
        // Send broadcast using the correct format
        channel.send({
          type: "broadcast",
          event: "disconnect",
          payload: { machine_id: clientId },
        });
        // Wait a bit for the message to be sent, then unsubscribe
        setTimeout(() => {
          clearTimeout(timeout);
          channel.unsubscribe();
          resolve();
        }, 500);
      } else if (status === "CHANNEL_ERROR" || status === "TIMED_OUT") {
        clearTimeout(timeout);
        channel.unsubscribe();
        // Don't reject - broadcast is best effort, DB update already succeeded
        resolve();
      }
    });
  }).catch((err) => {
    // Log but don't fail - DB update already succeeded
    console.error("Failed to send broadcast:", err);
  });
}

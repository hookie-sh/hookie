
  create table "public"."connected_clients" (
    "user_id" text not null,
    "org_id" text not null default ''::text,
    "connected_at" timestamp with time zone not null default now(),
    "disconnected_at" timestamp with time zone,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now(),
    "id" text not null,
    "connection_count" integer not null default 0
      );


alter table "public"."connected_clients" enable row level security;

CREATE UNIQUE INDEX connected_clients_pkey ON public.connected_clients USING btree (id, user_id, org_id);

CREATE INDEX idx_connected_clients_connected_at ON public.connected_clients USING btree (connected_at DESC);

CREATE INDEX idx_connected_clients_connection_count ON public.connected_clients USING btree (connection_count);

CREATE UNIQUE INDEX idx_connected_clients_id_org_unique ON public.connected_clients USING btree (id, org_id) WHERE (disconnected_at IS NULL);

CREATE INDEX idx_connected_clients_machine_id ON public.connected_clients USING btree (id);

CREATE INDEX idx_connected_clients_user_id ON public.connected_clients USING btree (user_id);

alter table "public"."connected_clients" add constraint "connected_clients_pkey" PRIMARY KEY using index "connected_clients_pkey";

alter table "public"."connected_clients" add constraint "connected_clients_user_id_fkey" FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE not valid;

alter table "public"."connected_clients" validate constraint "connected_clients_user_id_fkey";

set check_function_bodies = off;

CREATE OR REPLACE FUNCTION public.ksuid(prefix text DEFAULT NULL::text)
 RETURNS text
 LANGUAGE plpgsql
 SECURITY DEFINER
 SET search_path TO ''
AS $function$
DECLARE
  chars TEXT[] := ARRAY['0','1','2','3','4','5','6','7','8','9',
                        'A','B','C','D','E','F','G','H','I','J','K','L','M','N','O','P','Q','R','S','T','U','V','W','X','Y','Z',
                        'a','b','c','d','e','f','g','h','i','j','k','l','m','n','o','p','q','r','s','t','u','v','w','x','y','z'];
  timestamp_ms BIGINT;
  timestamp_b62 TEXT := '';
  random_bytes BYTEA;
  random_b62 TEXT := '';
  temp_val BIGINT;
  char_index INT;
  i INT;
  base_id TEXT;
BEGIN
  timestamp_ms := FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000);
  
  -- Generate 7-character timestamp part
  temp_val := timestamp_ms;
  WHILE temp_val > 0 LOOP
    char_index := (temp_val % 62) + 1;
    timestamp_b62 := chars[char_index] || timestamp_b62;
    temp_val := temp_val / 62;
  END LOOP;
  
  -- Generate 10-character random part using pgcrypto
  random_bytes := extensions.gen_random_bytes(10);
  FOR i IN 0..9 LOOP
    char_index := (get_byte(random_bytes, i) % 62) + 1;
    random_b62 := random_b62 || chars[char_index];
  END LOOP;
  
  base_id := timestamp_b62 || random_b62;
  
  IF prefix IS NOT NULL AND prefix != '' THEN
    RETURN prefix || '_' || base_id;
  ELSE
    RETURN base_id;
  END IF;
END;
$function$
;

CREATE OR REPLACE FUNCTION public.touch()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
DECLARE
  touch_column text;
BEGIN
  -- Get the column name from TG_ARGV, default to 'updated_at' if not provided
  touch_column := COALESCE(TG_ARGV[0], 'updated_at');
  
  -- Handle the column update using a CASE statement for common columns
  -- This approach is more reliable than dynamic assignment
  IF touch_column = 'updated_at' THEN
    NEW.updated_at := NOW();
  ELSIF touch_column = 'created_at' THEN
    NEW.created_at := NOW();
  ELSIF touch_column = 'last_active_at' THEN
    NEW.last_active_at := NOW();
  ELSE
    -- For other columns, we need to use dynamic SQL
    -- But PostgreSQL doesn't allow direct dynamic assignment to NEW record
    -- So we'll raise an error for unsupported columns
    RAISE EXCEPTION 'Column % is not supported by touch function. Supported columns: updated_at, created_at, last_active_at', touch_column;
  END IF;
  
  RETURN NEW;
END;
$function$
;

grant delete on table "public"."connected_clients" to "anon";

grant insert on table "public"."connected_clients" to "anon";

grant references on table "public"."connected_clients" to "anon";

grant select on table "public"."connected_clients" to "anon";

grant trigger on table "public"."connected_clients" to "anon";

grant truncate on table "public"."connected_clients" to "anon";

grant update on table "public"."connected_clients" to "anon";

grant delete on table "public"."connected_clients" to "authenticated";

grant insert on table "public"."connected_clients" to "authenticated";

grant references on table "public"."connected_clients" to "authenticated";

grant select on table "public"."connected_clients" to "authenticated";

grant trigger on table "public"."connected_clients" to "authenticated";

grant truncate on table "public"."connected_clients" to "authenticated";

grant update on table "public"."connected_clients" to "authenticated";

grant delete on table "public"."connected_clients" to "postgres";

grant insert on table "public"."connected_clients" to "postgres";

grant references on table "public"."connected_clients" to "postgres";

grant select on table "public"."connected_clients" to "postgres";

grant trigger on table "public"."connected_clients" to "postgres";

grant truncate on table "public"."connected_clients" to "postgres";

grant update on table "public"."connected_clients" to "postgres";

grant delete on table "public"."connected_clients" to "service_role";

grant insert on table "public"."connected_clients" to "service_role";

grant references on table "public"."connected_clients" to "service_role";

grant select on table "public"."connected_clients" to "service_role";

grant trigger on table "public"."connected_clients" to "service_role";

grant truncate on table "public"."connected_clients" to "service_role";

grant update on table "public"."connected_clients" to "service_role";


  create policy "Service role can manage all connected clients"
  on "public"."connected_clients"
  as permissive
  for all
  to service_role
using (true)
with check (true);



  create policy "Users can insert their own connected clients"
  on "public"."connected_clients"
  as permissive
  for insert
  to public
with check ((public.user_id() = user_id));



  create policy "Users can update their own connected clients"
  on "public"."connected_clients"
  as permissive
  for update
  to public
using ((public.user_id() = user_id))
with check ((public.user_id() = user_id));



  create policy "Users can view their own connected clients"
  on "public"."connected_clients"
  as permissive
  for select
  to public
using ((public.user_id() = user_id));


CREATE TRIGGER update_connected_clients_updated_at BEFORE UPDATE ON public.connected_clients FOR EACH ROW EXECUTE FUNCTION public.touch();

CREATE TRIGGER objects_delete_delete_prefix AFTER DELETE ON storage.objects FOR EACH ROW EXECUTE FUNCTION storage.delete_prefix_hierarchy_trigger();

CREATE TRIGGER objects_insert_create_prefix BEFORE INSERT ON storage.objects FOR EACH ROW EXECUTE FUNCTION storage.objects_insert_prefix_trigger();

CREATE TRIGGER objects_update_create_prefix BEFORE UPDATE ON storage.objects FOR EACH ROW WHEN (((new.name <> old.name) OR (new.bucket_id <> old.bucket_id))) EXECUTE FUNCTION storage.objects_update_prefix_trigger();

CREATE TRIGGER prefixes_create_hierarchy BEFORE INSERT ON storage.prefixes FOR EACH ROW WHEN ((pg_trigger_depth() < 1)) EXECUTE FUNCTION storage.prefixes_insert_trigger();

CREATE TRIGGER prefixes_delete_hierarchy AFTER DELETE ON storage.prefixes FOR EACH ROW EXECUTE FUNCTION storage.delete_prefix_hierarchy_trigger();



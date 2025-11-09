
  create table "public"."applications" (
    "id" text not null default public.ksuid('app'::text),
    "name" text not null,
    "description" text,
    "user_id" text,
    "org_id" text,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
      );


alter table "public"."applications" enable row level security;


  create table "public"."topics" (
    "id" text not null default public.ksuid('topic'::text),
    "name" text not null,
    "description" text,
    "application_id" text not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
      );


alter table "public"."topics" enable row level security;


  create table "public"."users" (
    "id" text not null,
    "email" text not null,
    "first_name" text,
    "last_name" text,
    "image_url" text,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now(),
    "last_active_at" timestamp with time zone not null default now()
      );


alter table "public"."users" enable row level security;

CREATE UNIQUE INDEX applications_pkey ON public.applications USING btree (id);

CREATE INDEX idx_applications_created_at ON public.applications USING btree (created_at DESC);

CREATE INDEX idx_applications_org_id ON public.applications USING btree (org_id) WHERE (org_id IS NOT NULL);

CREATE INDEX idx_applications_user_id ON public.applications USING btree (user_id) WHERE (user_id IS NOT NULL);

CREATE INDEX topics_application_id_idx ON public.topics USING btree (application_id);

CREATE UNIQUE INDEX topics_pkey ON public.topics USING btree (id);

CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email);

CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id);

alter table "public"."applications" add constraint "applications_pkey" PRIMARY KEY using index "applications_pkey";

alter table "public"."topics" add constraint "topics_pkey" PRIMARY KEY using index "topics_pkey";

alter table "public"."users" add constraint "users_pkey" PRIMARY KEY using index "users_pkey";

alter table "public"."applications" add constraint "applications_owner_check" CHECK ((((user_id IS NOT NULL) AND (org_id IS NULL)) OR ((user_id IS NULL) AND (org_id IS NOT NULL)))) not valid;

alter table "public"."applications" validate constraint "applications_owner_check";

alter table "public"."applications" add constraint "applications_user_id_fkey" FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE not valid;

alter table "public"."applications" validate constraint "applications_user_id_fkey";

alter table "public"."topics" add constraint "topics_application_id_fkey" FOREIGN KEY (application_id) REFERENCES public.applications(id) ON DELETE CASCADE not valid;

alter table "public"."topics" validate constraint "topics_application_id_fkey";

alter table "public"."users" add constraint "users_email_key" UNIQUE using index "users_email_key";

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

CREATE OR REPLACE FUNCTION public.org_id()
 RETURNS text
 LANGUAGE sql
 STABLE
AS $function$
  select nullif(auth.jwt() -> 'o' ->> 'id', '');
$function$
;

CREATE OR REPLACE FUNCTION public.org_role()
 RETURNS text
 LANGUAGE sql
 STABLE
AS $function$
  select nullif(
    coalesce(
      auth.jwt() -> 'o' ->> 'rol',   -- short claim present in your token
      auth.jwt() ->> 'org_role'      -- fallback if using old/top-level format
    ),
    ''
  );
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

CREATE OR REPLACE FUNCTION public.user_id()
 RETURNS text
 LANGUAGE sql
AS $function$SELECT NULLIF(
    current_setting('request.jwt.claims', true)::json->>'sub',
    ''
)::text;$function$
;

grant delete on table "public"."applications" to "anon";

grant insert on table "public"."applications" to "anon";

grant references on table "public"."applications" to "anon";

grant select on table "public"."applications" to "anon";

grant trigger on table "public"."applications" to "anon";

grant truncate on table "public"."applications" to "anon";

grant update on table "public"."applications" to "anon";

grant delete on table "public"."applications" to "authenticated";

grant insert on table "public"."applications" to "authenticated";

grant references on table "public"."applications" to "authenticated";

grant select on table "public"."applications" to "authenticated";

grant trigger on table "public"."applications" to "authenticated";

grant truncate on table "public"."applications" to "authenticated";

grant update on table "public"."applications" to "authenticated";

grant delete on table "public"."applications" to "service_role";

grant insert on table "public"."applications" to "service_role";

grant references on table "public"."applications" to "service_role";

grant select on table "public"."applications" to "service_role";

grant trigger on table "public"."applications" to "service_role";

grant truncate on table "public"."applications" to "service_role";

grant update on table "public"."applications" to "service_role";

grant delete on table "public"."topics" to "anon";

grant insert on table "public"."topics" to "anon";

grant references on table "public"."topics" to "anon";

grant select on table "public"."topics" to "anon";

grant trigger on table "public"."topics" to "anon";

grant truncate on table "public"."topics" to "anon";

grant update on table "public"."topics" to "anon";

grant delete on table "public"."topics" to "authenticated";

grant insert on table "public"."topics" to "authenticated";

grant references on table "public"."topics" to "authenticated";

grant select on table "public"."topics" to "authenticated";

grant trigger on table "public"."topics" to "authenticated";

grant truncate on table "public"."topics" to "authenticated";

grant update on table "public"."topics" to "authenticated";

grant delete on table "public"."topics" to "service_role";

grant insert on table "public"."topics" to "service_role";

grant references on table "public"."topics" to "service_role";

grant select on table "public"."topics" to "service_role";

grant trigger on table "public"."topics" to "service_role";

grant truncate on table "public"."topics" to "service_role";

grant update on table "public"."topics" to "service_role";

grant delete on table "public"."users" to "anon";

grant insert on table "public"."users" to "anon";

grant references on table "public"."users" to "anon";

grant select on table "public"."users" to "anon";

grant trigger on table "public"."users" to "anon";

grant truncate on table "public"."users" to "anon";

grant update on table "public"."users" to "anon";

grant delete on table "public"."users" to "authenticated";

grant insert on table "public"."users" to "authenticated";

grant references on table "public"."users" to "authenticated";

grant select on table "public"."users" to "authenticated";

grant trigger on table "public"."users" to "authenticated";

grant truncate on table "public"."users" to "authenticated";

grant update on table "public"."users" to "authenticated";

grant delete on table "public"."users" to "service_role";

grant insert on table "public"."users" to "service_role";

grant references on table "public"."users" to "service_role";

grant select on table "public"."users" to "service_role";

grant trigger on table "public"."users" to "service_role";

grant truncate on table "public"."users" to "service_role";

grant update on table "public"."users" to "service_role";


  create policy "Organization members can manage organization applications"
  on "public"."applications"
  as permissive
  for all
  to public
using (((org_id IS NOT NULL) AND (public.org_id() = org_id)))
with check (((org_id IS NOT NULL) AND (public.org_id() = org_id)));



  create policy "Users can manage their own applications"
  on "public"."applications"
  as permissive
  for all
  to public
using (((user_id IS NOT NULL) AND (public.user_id() = user_id)))
with check (((user_id IS NOT NULL) AND (public.user_id() = user_id)));



  create policy "Organization members can manage topics for organization applica"
  on "public"."topics"
  as permissive
  for all
  to public
using ((EXISTS ( SELECT 1
   FROM public.applications
  WHERE ((applications.id = topics.application_id) AND (applications.org_id IS NOT NULL) AND (applications.org_id = public.org_id())))))
with check ((EXISTS ( SELECT 1
   FROM public.applications
  WHERE ((applications.id = topics.application_id) AND (applications.org_id IS NOT NULL) AND (applications.org_id = public.org_id())))));



  create policy "Users can manage topics for their own applications"
  on "public"."topics"
  as permissive
  for all
  to public
using ((EXISTS ( SELECT 1
   FROM public.applications
  WHERE ((applications.id = topics.application_id) AND (applications.user_id IS NOT NULL) AND (applications.user_id = public.user_id())))))
with check ((EXISTS ( SELECT 1
   FROM public.applications
  WHERE ((applications.id = topics.application_id) AND (applications.user_id IS NOT NULL) AND (applications.user_id = public.user_id())))));



  create policy "Users can update their own data"
  on "public"."users"
  as permissive
  for update
  to public
using (((auth.uid())::text = id))
with check (((auth.uid())::text = id));



  create policy "Users can view their own data"
  on "public"."users"
  as permissive
  for select
  to public
using (((auth.uid())::text = id));


CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON public.applications FOR EACH ROW EXECUTE FUNCTION public.touch();

CREATE TRIGGER topics_updated_at BEFORE UPDATE ON public.topics FOR EACH ROW EXECUTE FUNCTION public.touch();

CREATE TRIGGER set_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.touch();



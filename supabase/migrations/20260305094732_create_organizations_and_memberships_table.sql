
  create table "public"."memberships" (
    "id" text not null,
    "organization_id" text not null,
    "user_id" text not null,
    "role" text not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
      );


alter table "public"."memberships" enable row level security;


  create table "public"."organizations" (
    "id" text not null,
    "name" text not null,
    "image_url" text,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
      );


alter table "public"."organizations" enable row level security;

CREATE INDEX idx_memberships_organization_id ON public.memberships USING btree (organization_id);

CREATE INDEX idx_memberships_user_id ON public.memberships USING btree (user_id);

CREATE UNIQUE INDEX memberships_org_user_key ON public.memberships USING btree (organization_id, user_id);

CREATE UNIQUE INDEX memberships_pkey ON public.memberships USING btree (id);

CREATE UNIQUE INDEX organizations_pkey ON public.organizations USING btree (id);

alter table "public"."memberships" add constraint "memberships_pkey" PRIMARY KEY using index "memberships_pkey";

alter table "public"."organizations" add constraint "organizations_pkey" PRIMARY KEY using index "organizations_pkey";

alter table "public"."applications" add constraint "applications_org_id_fkey" FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE CASCADE NOT VALID not valid;

alter table "public"."applications" validate constraint "applications_org_id_fkey";

alter table "public"."memberships" add constraint "memberships_org_user_key" UNIQUE using index "memberships_org_user_key";

alter table "public"."memberships" add constraint "memberships_organization_id_fkey" FOREIGN KEY (organization_id) REFERENCES public.organizations(id) ON DELETE CASCADE not valid;

alter table "public"."memberships" validate constraint "memberships_organization_id_fkey";

alter table "public"."memberships" add constraint "memberships_user_id_fkey" FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE not valid;

alter table "public"."memberships" validate constraint "memberships_user_id_fkey";

grant delete on table "public"."anonymous_topics" to "postgres";

grant insert on table "public"."anonymous_topics" to "postgres";

grant references on table "public"."anonymous_topics" to "postgres";

grant select on table "public"."anonymous_topics" to "postgres";

grant trigger on table "public"."anonymous_topics" to "postgres";

grant truncate on table "public"."anonymous_topics" to "postgres";

grant update on table "public"."anonymous_topics" to "postgres";

grant delete on table "public"."connected_clients" to "postgres";

grant insert on table "public"."connected_clients" to "postgres";

grant references on table "public"."connected_clients" to "postgres";

grant select on table "public"."connected_clients" to "postgres";

grant trigger on table "public"."connected_clients" to "postgres";

grant truncate on table "public"."connected_clients" to "postgres";

grant update on table "public"."connected_clients" to "postgres";

grant delete on table "public"."memberships" to "anon";

grant insert on table "public"."memberships" to "anon";

grant references on table "public"."memberships" to "anon";

grant select on table "public"."memberships" to "anon";

grant trigger on table "public"."memberships" to "anon";

grant truncate on table "public"."memberships" to "anon";

grant update on table "public"."memberships" to "anon";

grant delete on table "public"."memberships" to "authenticated";

grant insert on table "public"."memberships" to "authenticated";

grant references on table "public"."memberships" to "authenticated";

grant select on table "public"."memberships" to "authenticated";

grant trigger on table "public"."memberships" to "authenticated";

grant truncate on table "public"."memberships" to "authenticated";

grant update on table "public"."memberships" to "authenticated";

grant delete on table "public"."memberships" to "postgres";

grant insert on table "public"."memberships" to "postgres";

grant references on table "public"."memberships" to "postgres";

grant select on table "public"."memberships" to "postgres";

grant trigger on table "public"."memberships" to "postgres";

grant truncate on table "public"."memberships" to "postgres";

grant update on table "public"."memberships" to "postgres";

grant delete on table "public"."memberships" to "service_role";

grant insert on table "public"."memberships" to "service_role";

grant references on table "public"."memberships" to "service_role";

grant select on table "public"."memberships" to "service_role";

grant trigger on table "public"."memberships" to "service_role";

grant truncate on table "public"."memberships" to "service_role";

grant update on table "public"."memberships" to "service_role";

grant delete on table "public"."organizations" to "anon";

grant insert on table "public"."organizations" to "anon";

grant references on table "public"."organizations" to "anon";

grant select on table "public"."organizations" to "anon";

grant trigger on table "public"."organizations" to "anon";

grant truncate on table "public"."organizations" to "anon";

grant update on table "public"."organizations" to "anon";

grant delete on table "public"."organizations" to "authenticated";

grant insert on table "public"."organizations" to "authenticated";

grant references on table "public"."organizations" to "authenticated";

grant select on table "public"."organizations" to "authenticated";

grant trigger on table "public"."organizations" to "authenticated";

grant truncate on table "public"."organizations" to "authenticated";

grant update on table "public"."organizations" to "authenticated";

grant delete on table "public"."organizations" to "postgres";

grant insert on table "public"."organizations" to "postgres";

grant references on table "public"."organizations" to "postgres";

grant select on table "public"."organizations" to "postgres";

grant trigger on table "public"."organizations" to "postgres";

grant truncate on table "public"."organizations" to "postgres";

grant update on table "public"."organizations" to "postgres";

grant delete on table "public"."organizations" to "service_role";

grant insert on table "public"."organizations" to "service_role";

grant references on table "public"."organizations" to "service_role";

grant select on table "public"."organizations" to "service_role";

grant trigger on table "public"."organizations" to "service_role";

grant truncate on table "public"."organizations" to "service_role";

grant update on table "public"."organizations" to "service_role";


  create policy "Users can view memberships for their organizations"
  on "public"."memberships"
  as permissive
  for select
  to public
using (((user_id = (auth.uid())::text) OR (EXISTS ( SELECT 1
   FROM public.memberships m
  WHERE ((m.organization_id = memberships.organization_id) AND (m.user_id = (auth.uid())::text))))));



  create policy "Users can view organizations they belong to"
  on "public"."organizations"
  as permissive
  for select
  to public
using ((EXISTS ( SELECT 1
   FROM public.memberships m
  WHERE ((m.organization_id = organizations.id) AND (m.user_id = (auth.uid())::text)))));


CREATE TRIGGER set_memberships_updated_at BEFORE UPDATE ON public.memberships FOR EACH ROW EXECUTE FUNCTION public.touch();

CREATE TRIGGER set_organizations_updated_at BEFORE UPDATE ON public.organizations FOR EACH ROW EXECUTE FUNCTION public.touch();

CREATE TRIGGER objects_delete_delete_prefix AFTER DELETE ON storage.objects FOR EACH ROW EXECUTE FUNCTION storage.delete_prefix_hierarchy_trigger();

CREATE TRIGGER objects_insert_create_prefix BEFORE INSERT ON storage.objects FOR EACH ROW EXECUTE FUNCTION storage.objects_insert_prefix_trigger();

CREATE TRIGGER objects_update_create_prefix BEFORE UPDATE ON storage.objects FOR EACH ROW WHEN (((new.name <> old.name) OR (new.bucket_id <> old.bucket_id))) EXECUTE FUNCTION storage.objects_update_prefix_trigger();

CREATE TRIGGER prefixes_create_hierarchy BEFORE INSERT ON storage.prefixes FOR EACH ROW WHEN ((pg_trigger_depth() < 1)) EXECUTE FUNCTION storage.prefixes_insert_trigger();

CREATE TRIGGER prefixes_delete_hierarchy AFTER DELETE ON storage.prefixes FOR EACH ROW EXECUTE FUNCTION storage.delete_prefix_hierarchy_trigger();



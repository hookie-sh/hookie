
  create table "public"."anonymous_topics" (
    "id" text not null,
    "ip_address" text not null,
    "created_at" timestamp with time zone not null default now(),
    "last_used_at" timestamp with time zone not null default now(),
    "request_count" bigint not null default 0,
    "disabled" boolean not null default false
      );


alter table "public"."anonymous_topics" enable row level security;

CREATE UNIQUE INDEX anonymous_topics_pkey ON public.anonymous_topics USING btree (id);

CREATE INDEX idx_anon_topics_disabled ON public.anonymous_topics USING btree (disabled) WHERE (disabled = true);

CREATE INDEX idx_anon_topics_ip ON public.anonymous_topics USING btree (ip_address);

alter table "public"."anonymous_topics" add constraint "anonymous_topics_pkey" PRIMARY KEY using index "anonymous_topics_pkey";

grant delete on table "public"."anonymous_topics" to "anon";

grant insert on table "public"."anonymous_topics" to "anon";

grant references on table "public"."anonymous_topics" to "anon";

grant select on table "public"."anonymous_topics" to "anon";

grant trigger on table "public"."anonymous_topics" to "anon";

grant truncate on table "public"."anonymous_topics" to "anon";

grant update on table "public"."anonymous_topics" to "anon";

grant delete on table "public"."anonymous_topics" to "authenticated";

grant insert on table "public"."anonymous_topics" to "authenticated";

grant references on table "public"."anonymous_topics" to "authenticated";

grant select on table "public"."anonymous_topics" to "authenticated";

grant trigger on table "public"."anonymous_topics" to "authenticated";

grant truncate on table "public"."anonymous_topics" to "authenticated";

grant update on table "public"."anonymous_topics" to "authenticated";

grant delete on table "public"."anonymous_topics" to "postgres";

grant insert on table "public"."anonymous_topics" to "postgres";

grant references on table "public"."anonymous_topics" to "postgres";

grant select on table "public"."anonymous_topics" to "postgres";

grant trigger on table "public"."anonymous_topics" to "postgres";

grant truncate on table "public"."anonymous_topics" to "postgres";

grant update on table "public"."anonymous_topics" to "postgres";

grant delete on table "public"."anonymous_topics" to "service_role";

grant insert on table "public"."anonymous_topics" to "service_role";

grant references on table "public"."anonymous_topics" to "service_role";

grant select on table "public"."anonymous_topics" to "service_role";

grant trigger on table "public"."anonymous_topics" to "service_role";

grant truncate on table "public"."anonymous_topics" to "service_role";

grant update on table "public"."anonymous_topics" to "service_role";

grant delete on table "public"."connected_clients" to "postgres";

grant insert on table "public"."connected_clients" to "postgres";

grant references on table "public"."connected_clients" to "postgres";

grant select on table "public"."connected_clients" to "postgres";

grant trigger on table "public"."connected_clients" to "postgres";

grant truncate on table "public"."connected_clients" to "postgres";

grant update on table "public"."connected_clients" to "postgres";



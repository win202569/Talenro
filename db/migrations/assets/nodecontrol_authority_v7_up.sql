-- This asset is embedded by the provider-scoped registered Go migration.
-- It is deliberately not a standalone Goose migration.

-- talenro:statement
DO $roles$
DECLARE
  bootstrap_oid oid;
  role_count bigint;
  membership_count bigint;
  setting_count bigint;
  password_count bigint;
BEGIN
  IF session_user IS DISTINCT FROM current_user THEN
    RAISE EXCEPTION 'authority v7 bootstrap requires session_user=current_user' USING ERRCODE = '42501';
  END IF;
  SELECT oid INTO bootstrap_oid
  FROM pg_catalog.pg_roles
  WHERE rolname = current_user AND rolsuper;
  IF bootstrap_oid IS NULL
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_database
       WHERE datname = current_database() AND datdba = bootstrap_oid
     )
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_namespace
       WHERE nspname = 'nodecontrol' AND nspowner = bootstrap_oid
     ) THEN
    RAISE EXCEPTION 'authority v7 bootstrap must be the superuser database and nodecontrol schema owner' USING ERRCODE = '42501';
  END IF;
  SELECT count(*) INTO role_count
  FROM pg_catalog.pg_roles
  WHERE rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  );
  IF role_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role names must all be absent before creation' USING ERRCODE = '42710';
  END IF;

  CREATE ROLE nodecontrol_upgrade_executor
    NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS
    CONNECTION LIMIT -1 PASSWORD NULL;
  CREATE ROLE nodecontrol_migration_downgrader
    NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS
    CONNECTION LIMIT -1 PASSWORD NULL;
  CREATE ROLE nodecontrol_staging_importer
    NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT NOREPLICATION NOBYPASSRLS
    CONNECTION LIMIT -1 PASSWORD NULL;

  SELECT count(*) INTO role_count
  FROM pg_catalog.pg_roles
  WHERE rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  )
    AND NOT rolcanlogin
    AND NOT rolsuper
    AND NOT rolcreatedb
    AND NOT rolcreaterole
    AND rolinherit
    AND NOT rolreplication
    AND NOT rolbypassrls
    AND rolconnlimit = -1
    AND rolvaliduntil IS NULL
    AND rolconfig IS NULL;
  SELECT count(*) INTO setting_count
  FROM pg_catalog.pg_db_role_setting AS setting
  WHERE setting.setrole IN (
    'nodecontrol_upgrade_executor'::regrole,
    'nodecontrol_migration_downgrader'::regrole,
    'nodecontrol_staging_importer'::regrole
  );
  SELECT count(*) INTO password_count
  FROM pg_catalog.pg_authid AS role_catalog
  WHERE role_catalog.rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  )
    AND role_catalog.rolpassword IS NULL;
  SELECT count(*) INTO membership_count
  FROM pg_catalog.pg_auth_members AS membership
  JOIN pg_catalog.pg_roles AS member_role ON member_role.oid = membership.member
  JOIN pg_catalog.pg_roles AS granted_role ON granted_role.oid = membership.roleid
  WHERE member_role.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
     OR granted_role.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    );
  IF role_count <> 3 OR setting_count <> 0 OR password_count <> 3 OR membership_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role creation drifted' USING ERRCODE = '55000';
  END IF;
END
$roles$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_text_array_is_sorted_unique(items text[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
STRICT
SET search_path = pg_catalog, nodecontrol
AS $fn$
  SELECT NOT EXISTS (
    SELECT 1
    FROM generate_subscripts(items, 1) AS s(i)
    WHERE i > 1 AND items[i - 1] COLLATE "C" >= items[i] COLLATE "C"
  )
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_require_role(required_role name)
RETURNS void
LANGUAGE plpgsql
STABLE
SECURITY INVOKER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  bootstrap_oid oid;
  drift_count bigint;
  select_relation_oids oid[];
  lock_relation_oids oid[];
  v7_function_oids oid[];
BEGIN
  IF required_role::text NOT IN ('nodecontrol_upgrade_executor', 'nodecontrol_migration_downgrader', 'nodecontrol_staging_importer')
     OR current_user IS DISTINCT FROM required_role::text THEN
    RAISE EXCEPTION 'nodecontrol v7 role % is required', required_role USING ERRCODE = '42501';
  END IF;

  SELECT oid INTO bootstrap_oid
  FROM pg_catalog.pg_roles
  WHERE rolname = session_user AND rolsuper;
  IF bootstrap_oid IS NULL
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_database
       WHERE datname = current_database() AND datdba = bootstrap_oid
     )
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_namespace
       WHERE nspname = 'nodecontrol' AND nspowner = bootstrap_oid
     ) THEN
    RAISE EXCEPTION 'authority v7 Down requires the exact bootstrap owner' USING ERRCODE = '42501';
  END IF;

  WITH expected(
    role_name, can_login, is_superuser, can_create_database, can_create_role,
    inherits_privileges, can_replicate, bypasses_rls, connection_limit,
    valid_until, role_config
  ) AS (
    VALUES
      ('nodecontrol_upgrade_executor'::name, false, false, false, false, true, false, false, -1, NULL::timestamptz, NULL::text[]),
      ('nodecontrol_migration_downgrader'::name, false, false, false, false, true, false, false, -1, NULL::timestamptz, NULL::text[]),
      ('nodecontrol_staging_importer'::name, false, false, false, false, true, false, false, -1, NULL::timestamptz, NULL::text[])
  ), actual AS (
    SELECT
      role_catalog.rolname, role_catalog.rolcanlogin, role_catalog.rolsuper,
      role_catalog.rolcreatedb, role_catalog.rolcreaterole, role_catalog.rolinherit,
      role_catalog.rolreplication, role_catalog.rolbypassrls, role_catalog.rolconnlimit,
      role_catalog.rolvaliduntil, role_catalog.rolconfig
    FROM pg_catalog.pg_roles AS role_catalog
    WHERE role_catalog.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
  )
  SELECT count(*) INTO drift_count
  FROM (
    (SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
  ) AS difference;
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO drift_count
  FROM pg_catalog.pg_db_role_setting AS setting
  WHERE setting.setrole IN (
    'nodecontrol_upgrade_executor'::regrole,
    'nodecontrol_migration_downgrader'::regrole,
    'nodecontrol_staging_importer'::regrole
  );
  drift_count := drift_count + (
    SELECT count(*)
    FROM pg_catalog.pg_auth_members AS membership
    WHERE membership.member IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
       OR membership.roleid IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
  );
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role setting or membership drift' USING ERRCODE = '55000';
  END IF;

  select_relation_oids := ARRAY[
    'nodecontrol.control_plane_authority_protocol_migration_latches'::regclass::oid,
    'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_applications'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_cancellations'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_intents'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_recovery_applications'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_recovery_intents'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_resolutions'::regclass::oid,
    'nodecontrol.control_plane_authority_epoch_transition_terminal_applications'::regclass::oid,
    'nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass::oid,
    'nodecontrol.control_plane_authority_fresh_restore_requirements'::regclass::oid,
    'nodecontrol.control_plane_authority_indeterminate_source_seals'::regclass::oid,
    'nodecontrol.control_plane_authority_legacy_database_source_retirements'::regclass::oid,
    'nodecontrol.control_plane_authority_legacy_source_seals'::regclass::oid,
    'nodecontrol.control_plane_authority_protocol_activation_completions'::regclass::oid,
    'nodecontrol.control_plane_authority_protocol_activation_releases'::regclass::oid,
    'nodecontrol.control_plane_authority_protocol_activations'::regclass::oid,
    'nodecontrol.control_plane_authority_protocol_upgrade_attempts'::regclass::oid,
    'nodecontrol.control_plane_authority_protocol_upgrade_intents'::regclass::oid,
    'nodecontrol.control_plane_authority_runtime_rebind_results'::regclass::oid,
    'nodecontrol.control_plane_authority_runtime_registration_results'::regclass::oid,
    'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass::oid,
    'nodecontrol.control_plane_authority_staging_import_capability_recovery_applications'::regclass::oid,
    'nodecontrol.control_plane_authority_staging_import_capability_recovery_intents'::regclass::oid,
    'nodecontrol.control_plane_authority_staging_import_capability_revocation_applications'::regclass::oid,
    'nodecontrol.control_plane_authority_fences'::regclass::oid,
    'nodecontrol.control_plane_trust_bundle_high_waters'::regclass::oid,
    'nodecontrol.node_capacity_profiles'::regclass::oid,
    'nodecontrol.node_certificate_issuances'::regclass::oid,
    'nodecontrol.node_certificates'::regclass::oid,
    'nodecontrol.node_desired_states'::regclass::oid,
    'nodecontrol.node_endpoints'::regclass::oid,
    'nodecontrol.node_enrollment_grants'::regclass::oid,
    'nodecontrol.node_failure_domain_membership'::regclass::oid,
    'nodecontrol.node_failure_domains'::regclass::oid,
    'nodecontrol.node_inventory'::regclass::oid,
    'nodecontrol.node_observed_states'::regclass::oid,
    'nodecontrol.node_operator_audit'::regclass::oid,
    'nodecontrol.node_pops'::regclass::oid,
    'nodecontrol.node_process_slots'::regclass::oid,
    'nodecontrol.node_recovery_sessions'::regclass::oid,
    'nodecontrol.node_recovery_states'::regclass::oid,
    'nodecontrol.node_resource_envelopes'::regclass::oid,
    'nodecontrol.node_restore_reauthorization_approvals'::regclass::oid,
    'nodecontrol.node_root_metadata_publish_intents'::regclass::oid,
    'nodecontrol.node_root_metadata_signature_shares'::regclass::oid,
    'nodecontrol.node_security_fault_receipts'::regclass::oid,
    'nodecontrol.node_security_incidents'::regclass::oid,
    'nodecontrol.node_state_signing_intents'::regclass::oid,
    'nodecontrol.node_state_transitions'::regclass::oid
  ];
  lock_relation_oids := select_relation_oids || ARRAY['public.goose_db_version'::regclass::oid];
  v7_function_oids := ARRAY[
    'nodecontrol.v7_text_array_is_sorted_unique(text[])'::regprocedure::oid,
    'nodecontrol.v7_require_role(name)'::regprocedure::oid,
    'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid,
    'nodecontrol.v7_reject_immutable_mutation()'::regprocedure::oid,
    'nodecontrol.v7_source_is_frozen()'::regprocedure::oid,
    'nodecontrol.v7_assert_source_writable()'::regprocedure::oid,
    'nodecontrol.v7_assert_activation_barrier()'::regprocedure::oid,
    'nodecontrol.v7_authority_proof_group_valid(bytea,bytea,bytea,bytea,bytea,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.v7_guard_authority_proof_transition()'::regprocedure::oid,
    'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid
  ];

  WITH expected(grantee, object_oid, object_subid, privilege_type, grantor, is_grantable) AS (
    SELECT 'nodecontrol_migration_downgrader'::regrole::oid, relation.oid, 0,
           'SELECT'::text, relation.relowner, false
    FROM pg_catalog.pg_class AS relation
    WHERE relation.oid = ANY (select_relation_oids)
    UNION ALL
    SELECT 'nodecontrol_migration_downgrader'::regrole::oid, relation.oid, 0,
           'MAINTAIN', relation.relowner, false
    FROM pg_catalog.pg_class AS relation
    WHERE relation.oid = ANY (lock_relation_oids)
    UNION ALL
    SELECT 'nodecontrol_migration_downgrader'::regrole::oid, relation.oid, 0,
           privilege.privilege_type, relation.relowner, false
    FROM (VALUES
      ('nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass::oid, 'INSERT'::text),
      ('nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass::oid, 'DELETE'::text),
      ('nodecontrol.control_plane_authority_protocol_migration_latches'::regclass::oid, 'DELETE'::text)
    ) AS privilege(relation_oid, privilege_type)
    JOIN pg_catalog.pg_class AS relation ON relation.oid = privilege.relation_oid
    UNION ALL
    SELECT 'nodecontrol_staging_importer'::regrole::oid, relation.oid, 0,
           privilege.privilege_type, relation.relowner, false
    FROM (VALUES
      ('nodecontrol.control_plane_authority_staging_import_capabilities'::regclass::oid, 'SELECT'::text),
      ('nodecontrol.control_plane_authority_staging_import_capability_recovery_intents'::regclass::oid, 'SELECT'::text),
      ('nodecontrol.control_plane_authority_staging_import_capability_recovery_applications'::regclass::oid, 'SELECT'::text),
      ('nodecontrol.control_plane_authority_staging_import_capability_revocation_applications'::regclass::oid, 'SELECT'::text),
      ('nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass::oid, 'SELECT'::text),
      ('nodecontrol.node_pops'::regclass::oid, 'INSERT'::text),
      ('nodecontrol.node_failure_domains'::regclass::oid, 'INSERT'::text),
      ('nodecontrol.node_capacity_profiles'::regclass::oid, 'INSERT'::text),
      ('nodecontrol.node_inventory'::regclass::oid, 'INSERT'::text),
      ('nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass::oid, 'INSERT'::text)
    ) AS privilege(relation_oid, privilege_type)
    JOIN pg_catalog.pg_class AS relation ON relation.oid = privilege.relation_oid
    UNION ALL
    SELECT 'nodecontrol_staging_importer'::regrole::oid, relation.oid, attribute.attnum,
           'UPDATE'::text, relation.relowner, false
    FROM pg_catalog.pg_attribute AS attribute
    JOIN pg_catalog.pg_class AS relation ON relation.oid = attribute.attrelid
    WHERE relation.oid = 'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass
      AND attribute.attname = 'body_digest'
      AND NOT attribute.attisdropped
  ), actual AS (
    SELECT acl.grantee, relation.oid, 0, acl.privilege_type, acl.grantor, acl.is_grantable
    FROM pg_catalog.pg_class AS relation
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
    CROSS JOIN LATERAL pg_catalog.aclexplode(
      COALESCE(
        relation.relacl,
        pg_catalog.acldefault(
          CASE WHEN relation.relkind = 'S' THEN 'S'::"char" ELSE 'r'::"char" END,
          relation.relowner
        )
      )
    ) AS acl
    WHERE acl.grantee IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
       OR (namespace.nspname = 'nodecontrol' AND acl.grantee <> relation.relowner)
    UNION ALL
    SELECT acl.grantee, relation.oid, attribute.attnum, acl.privilege_type, acl.grantor, acl.is_grantable
    FROM pg_catalog.pg_attribute AS attribute
    JOIN pg_catalog.pg_class AS relation ON relation.oid = attribute.attrelid
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
    CROSS JOIN LATERAL pg_catalog.aclexplode(attribute.attacl) AS acl
    WHERE attribute.attacl IS NOT NULL
      AND (
        acl.grantee IN (
          'nodecontrol_upgrade_executor'::regrole,
          'nodecontrol_migration_downgrader'::regrole,
          'nodecontrol_staging_importer'::regrole
        )
        OR (namespace.nspname = 'nodecontrol' AND acl.grantee <> relation.relowner)
      )
  )
  SELECT count(*) INTO drift_count
  FROM (
    (SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
  ) AS difference;
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 relation or column ACL drift' USING ERRCODE = '55000';
  END IF;

  WITH expected(grantee, object_oid, privilege_type, grantor, is_grantable) AS (
    SELECT expected_role.role_name::regrole::oid, namespace.oid, 'USAGE'::text,
           namespace.nspowner, false
    FROM (VALUES
      ('nodecontrol_upgrade_executor'::text, 'nodecontrol'::text),
      ('nodecontrol_migration_downgrader'::text, 'nodecontrol'::text),
      ('nodecontrol_migration_downgrader'::text, 'public'::text),
      ('nodecontrol_staging_importer'::text, 'nodecontrol'::text)
    ) AS expected_role(role_name, schema_name)
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.nspname = expected_role.schema_name
  ), actual AS (
    SELECT acl.grantee, namespace.oid, acl.privilege_type, acl.grantor, acl.is_grantable
    FROM pg_catalog.pg_namespace AS namespace
    CROSS JOIN LATERAL pg_catalog.aclexplode(
      COALESCE(namespace.nspacl, pg_catalog.acldefault('n', namespace.nspowner))
    ) AS acl
    WHERE acl.grantee IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
       OR (namespace.nspname = 'nodecontrol' AND acl.grantee <> namespace.nspowner)
  )
  SELECT count(*) INTO drift_count
  FROM (
    (SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
  ) AS difference;
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 schema ACL drift' USING ERRCODE = '55000';
  END IF;

  WITH expected(grantee, object_oid, privilege_type, grantor, is_grantable) AS (
    SELECT expected_acl.grantee, proc.oid, 'EXECUTE'::text, proc.proowner, false
    FROM (VALUES
      (bootstrap_oid, 'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid),
      (bootstrap_oid, 'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid),
      (bootstrap_oid, 'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid),
      (bootstrap_oid, 'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid),
      ('nodecontrol_upgrade_executor'::regrole::oid, 'nodecontrol.v7_require_role(name)'::regprocedure::oid),
      ('nodecontrol_migration_downgrader'::regrole::oid, 'nodecontrol.v7_require_role(name)'::regprocedure::oid),
      ('nodecontrol_staging_importer'::regrole::oid, 'nodecontrol.v7_require_role(name)'::regprocedure::oid)
    ) AS expected_acl(grantee, procedure_oid)
    JOIN pg_catalog.pg_proc AS proc ON proc.oid = expected_acl.procedure_oid
  ), actual AS (
    SELECT acl.grantee, proc.oid, acl.privilege_type, acl.grantor, acl.is_grantable
    FROM pg_catalog.pg_proc AS proc
    CROSS JOIN LATERAL pg_catalog.aclexplode(
      COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
    ) AS acl
    WHERE acl.grantee <> proc.proowner
      AND (
        proc.oid = ANY (v7_function_oids)
        OR acl.grantee IN (
          'nodecontrol_upgrade_executor'::regrole,
          'nodecontrol_migration_downgrader'::regrole,
          'nodecontrol_staging_importer'::regrole
        )
      )
  )
  SELECT count(*) INTO drift_count
  FROM (
    (SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
  ) AS difference;
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 function ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO drift_count
  FROM (
    SELECT acl.grantee
    FROM pg_catalog.pg_database AS database
    CROSS JOIN LATERAL pg_catalog.aclexplode(
      COALESCE(database.datacl, pg_catalog.acldefault('d', database.datdba))
    ) AS acl
    UNION ALL
    SELECT acl.grantee
    FROM pg_catalog.pg_type AS type_catalog
    CROSS JOIN LATERAL pg_catalog.aclexplode(
      COALESCE(type_catalog.typacl, pg_catalog.acldefault('T', type_catalog.typowner))
    ) AS acl
    UNION ALL
    SELECT acl.grantee
    FROM pg_catalog.pg_class AS sequence
    CROSS JOIN LATERAL pg_catalog.aclexplode(
      COALESCE(sequence.relacl, pg_catalog.acldefault('S', sequence.relowner))
    ) AS acl
    WHERE sequence.relkind = 'S'
  ) AS direct_acl
  WHERE direct_acl.grantee IN (
    'nodecontrol_upgrade_executor'::regrole,
    'nodecontrol_migration_downgrader'::regrole,
    'nodecontrol_staging_importer'::regrole
  );
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability database type or sequence ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO drift_count
  FROM (
    VALUES
      ('nodecontrol.v7_text_array_is_sorted_unique(text[])'::regprocedure, session_user::text, 'sql'::text, 'i'::"char", false),
      ('nodecontrol.v7_require_role(name)'::regprocedure, session_user::text, 'plpgsql'::text, 's'::"char", false),
      ('nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure, 'nodecontrol_upgrade_executor'::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.v7_reject_immutable_mutation()'::regprocedure, session_user::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.v7_source_is_frozen()'::regprocedure, session_user::text, 'sql'::text, 'v'::"char", true),
      ('nodecontrol.v7_assert_source_writable()'::regprocedure, session_user::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.v7_assert_activation_barrier()'::regprocedure, session_user::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.v7_authority_proof_group_valid(bytea,bytea,bytea,bytea,bytea,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea)'::regprocedure, session_user::text, 'sql'::text, 'i'::"char", false),
      ('nodecontrol.v7_guard_authority_proof_transition()'::regprocedure, session_user::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure, 'nodecontrol_migration_downgrader'::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure, 'nodecontrol_migration_downgrader'::text, 'plpgsql'::text, 'v'::"char", true),
      ('nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure, 'nodecontrol_staging_importer'::text, 'plpgsql'::text, 'v'::"char", true)
  ) AS expected(procedure_oid, owner_name, language_name, volatility_code, security_definer)
  JOIN pg_catalog.pg_proc AS proc ON proc.oid = expected.procedure_oid
  JOIN pg_catalog.pg_roles AS owner ON owner.oid = proc.proowner
  JOIN pg_catalog.pg_language AS language ON language.oid = proc.prolang
  WHERE owner.rolname IS DISTINCT FROM expected.owner_name
     OR language.lanname IS DISTINCT FROM expected.language_name
     OR proc.provolatile IS DISTINCT FROM expected.volatility_code
     OR proc.prosecdef IS DISTINCT FROM expected.security_definer
     OR proc.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, nodecontrol']::text[];
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 function metadata drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO drift_count
  FROM pg_catalog.pg_default_acl AS default_acl
  LEFT JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = default_acl.defaclnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(default_acl.defaclacl) AS acl
  WHERE default_acl.defaclrole IN (
      'nodecontrol_upgrade_executor'::regrole,
      'nodecontrol_migration_downgrader'::regrole,
      'nodecontrol_staging_importer'::regrole
    )
     OR acl.grantee IN (
      'nodecontrol_upgrade_executor'::regrole,
      'nodecontrol_migration_downgrader'::regrole,
      'nodecontrol_staging_importer'::regrole
    )
     OR (
      default_acl.defaclrole = bootstrap_oid
      AND default_acl.defaclobjtype IN ('r', 'S', 'f', 'T', 'n')
      AND (default_acl.defaclnamespace = 0 OR namespace.nspname = 'nodecontrol')
      AND acl.grantee <> bootstrap_oid
    );
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 default ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO drift_count
  FROM pg_catalog.pg_init_privs AS initial_privilege
  CROSS JOIN LATERAL pg_catalog.aclexplode(initial_privilege.initprivs) AS acl
  WHERE acl.grantee IN (
    'nodecontrol_upgrade_executor'::regrole,
    'nodecontrol_migration_downgrader'::regrole,
    'nodecontrol_staging_importer'::regrole
  );
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 initial privilege drift' USING ERRCODE = '55000';
  END IF;

  WITH expected(dbid, classid, objid, objsubid, refclassid, refobjid, deptype) AS (
    SELECT database.oid, 'pg_catalog.pg_class'::regclass::oid, relation_oid, 0,
           'pg_catalog.pg_authid'::regclass::oid,
           'nodecontrol_migration_downgrader'::regrole::oid, 'a'::"char"
    FROM unnest(lock_relation_oids) AS relation(relation_oid)
    CROSS JOIN pg_catalog.pg_database AS database
    WHERE database.datname = current_database()
    UNION ALL
    SELECT database.oid, 'pg_catalog.pg_class'::regclass::oid, relation_oid, 0,
           'pg_catalog.pg_authid'::regclass::oid,
           'nodecontrol_staging_importer'::regrole::oid, 'a'::"char"
    FROM unnest(ARRAY[
      'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass::oid,
      'nodecontrol.control_plane_authority_staging_import_capability_recovery_intents'::regclass::oid,
      'nodecontrol.control_plane_authority_staging_import_capability_recovery_applications'::regclass::oid,
      'nodecontrol.control_plane_authority_staging_import_capability_revocation_applications'::regclass::oid,
      'nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass::oid,
      'nodecontrol.node_pops'::regclass::oid,
      'nodecontrol.node_failure_domains'::regclass::oid,
      'nodecontrol.node_capacity_profiles'::regclass::oid,
      'nodecontrol.node_inventory'::regclass::oid
    ]) AS relation(relation_oid)
    CROSS JOIN pg_catalog.pg_database AS database
    WHERE database.datname = current_database()
    UNION ALL
    SELECT database.oid, 'pg_catalog.pg_class'::regclass::oid, relation.oid, attribute.attnum,
           'pg_catalog.pg_authid'::regclass::oid,
           'nodecontrol_staging_importer'::regrole::oid, 'a'::"char"
    FROM pg_catalog.pg_attribute AS attribute
    JOIN pg_catalog.pg_class AS relation ON relation.oid = attribute.attrelid
    CROSS JOIN pg_catalog.pg_database AS database
    WHERE relation.oid = 'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass
      AND attribute.attname = 'body_digest'
      AND NOT attribute.attisdropped
      AND database.datname = current_database()
    UNION ALL
    SELECT database.oid, 'pg_catalog.pg_namespace'::regclass::oid, namespace.oid, 0,
           'pg_catalog.pg_authid'::regclass::oid, expected_schema.role_name::regrole::oid, 'a'::"char"
    FROM (VALUES
      ('nodecontrol_upgrade_executor'::text, 'nodecontrol'::text),
      ('nodecontrol_migration_downgrader'::text, 'nodecontrol'::text),
      ('nodecontrol_migration_downgrader'::text, 'public'::text),
      ('nodecontrol_staging_importer'::text, 'nodecontrol'::text)
    ) AS expected_schema(role_name, schema_name)
    JOIN pg_catalog.pg_namespace AS namespace ON namespace.nspname = expected_schema.schema_name
    CROSS JOIN pg_catalog.pg_database AS database
    WHERE database.datname = current_database()
    UNION ALL
    SELECT database.oid, 'pg_catalog.pg_proc'::regclass::oid,
           'nodecontrol.v7_require_role(name)'::regprocedure::oid, 0,
           'pg_catalog.pg_authid'::regclass::oid, expected_helper.role_name::regrole::oid, 'a'::"char"
    FROM (VALUES
      ('nodecontrol_upgrade_executor'::text),
      ('nodecontrol_migration_downgrader'::text),
      ('nodecontrol_staging_importer'::text)
    ) AS expected_helper(role_name)
    CROSS JOIN pg_catalog.pg_database AS database
    WHERE database.datname = current_database()
    UNION ALL
    SELECT database.oid, 'pg_catalog.pg_proc'::regclass::oid,
           expected_owner.procedure_oid, 0,
           'pg_catalog.pg_authid'::regclass::oid, expected_owner.role_name::regrole::oid, 'o'::"char"
    FROM (VALUES
      ('nodecontrol_upgrade_executor'::text, 'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid),
      ('nodecontrol_migration_downgrader'::text, 'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid),
      ('nodecontrol_migration_downgrader'::text, 'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid),
      ('nodecontrol_staging_importer'::text, 'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid)
    ) AS expected_owner(role_name, procedure_oid)
    CROSS JOIN pg_catalog.pg_database AS database
    WHERE database.datname = current_database()
  ), actual AS (
    SELECT dependency.dbid, dependency.classid, dependency.objid, dependency.objsubid,
           dependency.refclassid, dependency.refobjid, dependency.deptype
    FROM pg_catalog.pg_shdepend AS dependency
    WHERE dependency.refclassid = 'pg_catalog.pg_authid'::regclass
      AND dependency.refobjid IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
  )
  SELECT count(*) INTO drift_count
  FROM (
    (SELECT * FROM expected EXCEPT ALL SELECT * FROM actual)
    UNION ALL
    (SELECT * FROM actual EXCEPT ALL SELECT * FROM expected)
  ) AS difference;
  IF drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 shared dependency drift' USING ERRCODE = '55000';
  END IF;
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_acquire_source_freeze_for_seal()
RETURNS void
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
BEGIN
  PERFORM nodecontrol.v7_require_role('nodecontrol_upgrade_executor');
  IF pg_catalog.current_setting('transaction_isolation') IS DISTINCT FROM 'read committed' THEN
    RAISE EXCEPTION 'authority v7 source sealing requires read committed' USING ERRCODE = '25001';
  END IF;
  PERFORM pg_catalog.pg_advisory_xact_lock(
    pg_catalog.hashtextextended('nodecontrol:v7-source-freeze', 0)
  );
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_reject_immutable_mutation()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
BEGIN
  IF TG_OP = 'INSERT'
     AND TG_TABLE_NAME IN (
       'control_plane_authority_legacy_database_source_retirements',
       'control_plane_authority_indeterminate_source_seals',
       'control_plane_authority_legacy_source_seals',
       'control_plane_authority_fresh_restore_requirements'
     ) THEN
    PERFORM pg_catalog.pg_advisory_xact_lock(
      pg_catalog.hashtextextended('nodecontrol:v7-source-freeze',0)
    );
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE'
     AND TG_TABLE_NAME IN ('control_plane_authority_protocol_migration_latches', 'control_plane_authority_protocol_downgrade_authorizations')
     AND current_setting('nodecontrol.v7_down_guard', true) = 'consume' THEN
    RETURN OLD;
  END IF;
  RAISE EXCEPTION 'immutable nodecontrol v7 relation rejects %', TG_OP USING ERRCODE = '55000';
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_assert_source_writable()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
BEGIN
  IF pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'nodecontrol source mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  PERFORM pg_catalog.pg_advisory_xact_lock_shared(
    pg_catalog.hashtextextended('nodecontrol:v7-source-freeze',0)
  );
  IF TG_LEVEL = 'STATEMENT' THEN
    IF TG_OP = 'INSERT'
       AND current_setting('nodecontrol.v7_staging_import', true) = 'verified'
       AND TG_TABLE_NAME IN ('node_pops','node_failure_domains','node_capacity_profiles','node_inventory') THEN
      RETURN NULL;
    END IF;
    IF nodecontrol.v7_source_is_frozen() THEN
      RAISE EXCEPTION 'nodecontrol legacy source is frozen' USING ERRCODE = '55000';
    END IF;
    RETURN NULL;
  END IF;
  IF TG_OP = 'INSERT'
     AND current_setting('nodecontrol.v7_staging_import', true) = 'verified'
	   AND (
	     TG_TABLE_NAME IN ('node_pops','node_failure_domains','node_capacity_profiles')
	     OR (TG_TABLE_NAME = 'node_inventory' AND (pg_catalog.to_jsonb(NEW)->>'identity_state') = 'unauthorized')
	   ) THEN
    RETURN NEW;
  END IF;
  IF nodecontrol.v7_source_is_frozen() THEN
    RAISE EXCEPTION 'nodecontrol legacy source is frozen' USING ERRCODE = '55000';
  END IF;
	IF TG_OP <> 'DELETE'
	   AND TG_TABLE_NAME = 'node_inventory'
	   AND (pg_catalog.to_jsonb(NEW)->>'identity_state') = 'unauthorized' THEN
    RAISE EXCEPTION 'unauthorized inventory is restricted to verified staging import' USING ERRCODE = '42501';
  END IF;
  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_assert_activation_barrier()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  fence_row nodecontrol.control_plane_authority_fences%ROWTYPE;
  owner_count integer;
  matching_owner_count integer;
  commitment_owner_count integer;
  terminal_owner_count integer;
  domain_owner_count integer;
  owner_commitment_digest bytea;
BEGIN
  IF pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'authority proof mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  SELECT * INTO fence_row
  FROM nodecontrol.control_plane_authority_fences
  WHERE operation_id = NEW.operation_id;
  IF NOT FOUND OR fence_row.authority_protocol_profile <> 'claim_v1' THEN
    RAISE EXCEPTION 'claim-v1 deferred fence closure lost its exact fence' USING ERRCODE = '23514';
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM nodecontrol.control_plane_authority_protocol_activations AS activation
    JOIN nodecontrol.control_plane_authority_protocol_upgrade_attempts AS attempt
      ON attempt.body_digest = activation.attempt_digest
     AND attempt.activation_id = activation.activation_id
    JOIN nodecontrol.control_plane_authority_protocol_upgrade_intents AS intent
      ON intent.body_digest = attempt.upgrade_intent_digest
     AND intent.activation_id = activation.activation_id
    JOIN nodecontrol.control_plane_authority_runtime_registration_results AS registration
      ON registration.body_digest = activation.runtime_registration_result_digest
     AND registration.activation_id = activation.activation_id
    LEFT JOIN nodecontrol.control_plane_authority_runtime_rebind_results AS rebind
      ON rebind.body_digest = attempt.latest_runtime_rebind_result_digest_or_null
    JOIN nodecontrol.control_plane_authority_protocol_activation_completions AS completion
      ON completion.activation_id = activation.activation_id
     AND completion.completion_id = attempt.completion_id
    JOIN nodecontrol.control_plane_authority_protocol_activation_releases AS release
      ON release.activation_id = activation.activation_id
     AND release.release_preparation_id = attempt.release_preparation_id
     AND release.open_id = attempt.open_id
    WHERE activation.activation_id = fence_row.protocol_activation_id
      AND activation.mode = 'empty_in_place'
      AND attempt.mode = activation.mode
      AND activation.protocol_profile = 'claim_v1'
      AND activation.deployment_id = attempt.deployment_id
      AND activation.database_identity_digest = intent.database_identity_digest
      AND activation.database_incarnation_attestation_digest = attempt.database_incarnation_attestation_digest
      AND activation.genesis_database_incarnation_registration_digest = attempt.database_incarnation_registration_digest
      AND activation.runtime_registration_result_digest = attempt.runtime_registration_result_digest
      AND activation.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM attempt.latest_runtime_rebind_result_digest_or_null
      AND activation.runtime_rebind_chain_digest = attempt.runtime_rebind_chain_digest
      AND activation.activation_runtime_instance_binding_digest = attempt.runtime_instance_binding_digest
      AND activation.database_legacy_absence_projection_digest = attempt.database_legacy_absence_projection_digest
      AND activation.attempt_database_inventory_digest = attempt.database_inventory_digest
      AND activation.provider_namespace_absence_digest = attempt.provider_namespace_absence_digest
      AND activation.environment_inventory_digest = attempt.environment_inventory_digest
      AND activation.environment_inventory_anchor_set_digest = attempt.environment_inventory_anchor_set_digest
      AND activation.local_runtime_isolation_digest = attempt.local_runtime_isolation_digest
      AND activation.legacy_runtime_shutdown_digest = attempt.legacy_runtime_shutdown_digest
      AND activation.legacy_runtime_shutdown_set_digest = attempt.legacy_runtime_shutdown_set_digest
      AND activation.credential_policy_digest = attempt.credential_policy_digest
      AND activation.epoch_evidence_digest = attempt.epoch_evidence_digest
      AND activation.genesis_epoch_transition_root_digest = attempt.genesis_epoch_transition_root_digest
      AND activation.selected_genesis_epoch = attempt.selected_genesis_epoch
      AND activation.provider_identity_digest = registration.provider_identity_digest
      AND activation.provider_endpoint_identity_digest = registration.provider_endpoint_identity_digest
      AND activation.namespace = registration.namespace
      AND registration.upgrade_intent_digest = intent.body_digest
      AND registration.credential_policy_digest = attempt.credential_policy_digest
      AND registration.database_incarnation_attestation_digest = attempt.database_incarnation_attestation_digest
      AND registration.genesis_database_identity_digest = intent.database_identity_digest
      AND attempt.request_nonce = intent.request_nonce
      AND attempt.deployment_id = intent.observed_deployment_id
      AND attempt.local_runtime_isolation_digest = intent.local_runtime_isolation_digest
      AND intent.created_at <= registration.recorded_at
      AND registration.recorded_at <= attempt.created_at
      AND attempt.created_at <= activation.activated_at
      AND (
        (attempt.latest_runtime_rebind_result_digest_or_null IS NULL
          AND rebind.body_digest IS NULL
          AND attempt.runtime_rebind_chain_digest = registration.runtime_rebind_chain_digest
          AND attempt.runtime_instance_binding_digest = registration.runtime_instance_binding_digest)
        OR
        (attempt.latest_runtime_rebind_result_digest_or_null IS NOT NULL
          AND rebind.body_digest = attempt.latest_runtime_rebind_result_digest_or_null
          AND rebind.activation_id = activation.activation_id
          AND rebind.runtime_registration_result_digest = registration.body_digest
          AND rebind.current_database_identity_digest = activation.database_identity_digest
          AND rebind.current_database_incarnation_registration_digest = attempt.database_incarnation_registration_digest
          AND rebind.current_runtime_rebind_chain_digest = attempt.runtime_rebind_chain_digest
          AND rebind.current_runtime_instance_binding_digest = attempt.runtime_instance_binding_digest
          AND registration.recorded_at <= rebind.recorded_at
          AND rebind.recorded_at <= attempt.created_at)
      )
      AND completion.activation_digest = activation.body_digest
      AND completion.preparation_digest = activation.preparation_digest
      AND completion.provider_completion_phase = 'genesis_completed_pending_release'
      AND completion.current_database_incarnation_registration_digest = activation.genesis_database_incarnation_registration_digest
      AND completion.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM activation.latest_runtime_rebind_result_digest_or_null
      AND completion.runtime_rebind_chain_digest = activation.runtime_rebind_chain_digest
      AND completion.current_runtime_instance_binding_digest = activation.activation_runtime_instance_binding_digest
      AND completion.credential_policy_digest = activation.credential_policy_digest
      AND completion.epoch_evidence_digest = activation.epoch_evidence_digest
      AND completion.genesis_epoch_transition_root_digest = activation.genesis_epoch_transition_root_digest
      AND completion.selected_genesis_epoch = activation.selected_genesis_epoch
      AND release.activation_digest = activation.body_digest
      AND release.completion_digest = completion.body_digest
      AND release.provider_release_phase = 'genesis_release_prepared'
      AND release.current_database_incarnation_registration_digest = completion.current_database_incarnation_registration_digest
      AND release.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM completion.latest_runtime_rebind_result_digest_or_null
      AND release.runtime_rebind_chain_digest = completion.runtime_rebind_chain_digest
      AND release.current_runtime_instance_binding_digest = completion.current_runtime_instance_binding_digest
      AND release.credential_policy_digest = completion.credential_policy_digest
      AND release.epoch_evidence_digest = completion.epoch_evidence_digest
      AND release.genesis_epoch_transition_root_digest = completion.genesis_epoch_transition_root_digest
      AND release.selected_genesis_epoch = completion.selected_genesis_epoch
      AND completion.completed_at >= activation.activated_at
      AND release.released_at >= completion.completed_at
  ) THEN
    RAISE EXCEPTION 'claim-v1 activation release barrier is closed or inexact' USING ERRCODE = '23514';
  END IF;

  WITH owners AS (
    SELECT authority_operation_id AS operation_id,authority_epoch,authority_sequence,'grant_create'::text AS effect_kind,node_id,
           create_authority_effect_commitment_digest AS commitment_digest,
           create_authority_provider_head_digest IS NOT NULL AS terminal
      FROM nodecontrol.node_enrollment_grants
    UNION ALL SELECT claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence,'grant_claim',node_id,
           claim_authority_effect_commitment_digest,claim_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_enrollment_grants
    UNION ALL SELECT authority_operation_id,authority_epoch,authority_sequence,'certificate_activate',node_id,
           activation_authority_effect_commitment_digest,activation_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_certificate_issuances
    UNION ALL SELECT revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence,'certificate_revoke',node_id,
           revoke_authority_effect_commitment_digest,revoke_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_certificates
    UNION ALL SELECT authority_operation_id,authority_epoch,authority_sequence,authority_effect_kind,node_id,
           authority_effect_commitment_digest,authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_state_transitions
    UNION ALL SELECT authority_operation_id,authority_epoch,authority_sequence,'security_incident_open',node_id,
           open_authority_effect_commitment_digest,open_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_security_incidents
    UNION ALL SELECT resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence,'security_incident_resolve',node_id,
           resolve_authority_effect_commitment_digest,resolve_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_security_incidents
    UNION ALL SELECT authority_operation_id,authority_epoch,authority_sequence,'resource_envelope_activate',node_id,
           activation_authority_effect_commitment_digest,activation_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_resource_envelopes
    UNION ALL SELECT authority_operation_id,authority_epoch,authority_sequence,
           CASE signing_kind WHEN 'desired' THEN 'desired_activate' ELSE 'recovery_activate' END,node_id,
           activation_authority_effect_commitment_digest,activation_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_state_signing_intents
    UNION ALL SELECT authority_operation_id,authority_epoch,authority_sequence,
           CASE publish_kind WHEN 'root' THEN 'root_publish' ELSE 'metadata_publish' END,NULL::uuid,
           activation_authority_effect_commitment_digest,activation_authority_provider_head_digest IS NOT NULL
      FROM nodecontrol.node_root_metadata_publish_intents
  ), exact_owners AS (
    SELECT *,
      authority_epoch = fence_row.authority_epoch
      AND authority_sequence = fence_row.authority_sequence
      AND effect_kind = fence_row.effect_kind
      AND (
        (fence_row.scope_kind = 'node'
          AND node_id IS NOT NULL
          AND fence_row.scope_digest = pg_catalog.sha256(
            pg_catalog.convert_to('TALENRO-NODE-AUTHORITY-SCOPE-V1','UTF8')
            || decode('00','hex') || pg_catalog.uuid_send(node_id)))
        OR
        (fence_row.scope_kind = 'global_node_trust'
          AND node_id IS NULL
          AND fence_row.scope_digest = decode('f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1','hex'))
      ) AS exact_match
    FROM owners
    WHERE operation_id = fence_row.operation_id
  )
  SELECT count(*),count(*) FILTER (WHERE exact_match),
         count(*) FILTER (WHERE commitment_digest IS NOT NULL AND NOT terminal),
         count(*) FILTER (WHERE commitment_digest IS NOT NULL AND terminal),
         count(*) FILTER (WHERE commitment_digest IS NULL),
         (array_agg(commitment_digest) FILTER (WHERE commitment_digest IS NOT NULL))[1]
    INTO owner_count,matching_owner_count,commitment_owner_count,terminal_owner_count,
         domain_owner_count,owner_commitment_digest
  FROM exact_owners;

  IF owner_count <> matching_owner_count THEN
    RAISE EXCEPTION 'deferred authority proof owner tuple or scope mismatch' USING ERRCODE = '23514';
  END IF;
  IF fence_row.provider_status = 'committed' AND fence_row.visibility_state = 'active' THEN
    IF owner_count <> 1 OR terminal_owner_count <> 1
       OR owner_commitment_digest IS DISTINCT FROM fence_row.effect_digest THEN
      RAISE EXCEPTION 'deferred committed authority fence requires exactly one terminal proof owner' USING ERRCODE = '23514';
    END IF;
  ELSIF fence_row.provider_status = 'aborted' AND fence_row.visibility_state = 'aborted' THEN
    IF owner_count <> 0 THEN
      RAISE EXCEPTION 'deferred aborted authority fence requires zero proof owners' USING ERRCODE = '23514';
    END IF;
  ELSIF fence_row.provider_status = 'reserved' AND fence_row.effect_digest IS NOT NULL THEN
    IF owner_count <> 1 OR commitment_owner_count <> 1
       OR owner_commitment_digest IS DISTINCT FROM fence_row.effect_digest THEN
      RAISE EXCEPTION 'deferred bound authority fence requires exactly one commitment-only authority proof' USING ERRCODE = '23514';
    END IF;
  ELSIF fence_row.provider_status = 'reserved' AND fence_row.effect_digest IS NULL THEN
    IF fence_row.abort_claimed_at IS NOT NULL AND owner_count <> 0 THEN
      RAISE EXCEPTION 'deferred claimed authority fence requires zero proof owners' USING ERRCODE = '23514';
    END IF;
    IF fence_row.abort_claimed_at IS NULL
       AND NOT (owner_count = 0 OR (owner_count = 1 AND domain_owner_count = 1
         AND fence_row.effect_kind IN ('certificate_activate','desired_activate','recovery_activate','root_publish','metadata_publish'))) THEN
      RAISE EXCEPTION 'deferred unbound authority fence has an invalid domain-prepared owner' USING ERRCODE = '23514';
    END IF;
  ELSE
    RAISE EXCEPTION 'deferred authority fence closure found an illegal lifecycle state' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_authority_proof_group_valid(
  commitment_jcs bytea, commitment_digest bytea,
  provider_head_jcs bytea, provider_head_digest bytea,
  checkpoint_anchor_jcs bytea, checkpoint_anchor_digest bytea,
  effect_reason text, attestation_expires_at timestamptz,
  activation_deadline timestamptz, expected_provider_identity_digest bytea,
  activation_evidence_jcs bytea, activation_evidence_digest bytea,
  effect_resolution_jcs bytea, effect_resolution_digest bytea
)
RETURNS boolean
LANGUAGE sql
IMMUTABLE
SET search_path = pg_catalog, nodecontrol
AS $fn$
  SELECT COALESCE((
    ((commitment_jcs IS NULL) = (commitment_digest IS NULL))
    AND ((provider_head_jcs IS NULL) = (provider_head_digest IS NULL))
    AND ((checkpoint_anchor_jcs IS NULL) = (checkpoint_anchor_digest IS NULL))
    AND ((activation_evidence_jcs IS NULL) = (activation_evidence_digest IS NULL))
    AND ((effect_resolution_jcs IS NULL) = (effect_resolution_digest IS NULL))
    AND (commitment_jcs IS NULL OR octet_length(commitment_jcs) BETWEEN 1 AND 4096)
    AND (provider_head_jcs IS NULL OR octet_length(provider_head_jcs) BETWEEN 1 AND 4096)
    AND (checkpoint_anchor_jcs IS NULL OR octet_length(checkpoint_anchor_jcs) BETWEEN 1 AND 4096)
    AND (activation_evidence_jcs IS NULL OR octet_length(activation_evidence_jcs) BETWEEN 1 AND 4096)
    AND (effect_resolution_jcs IS NULL OR octet_length(effect_resolution_jcs) BETWEEN 1 AND 4096)
    AND (commitment_digest IS NULL OR octet_length(commitment_digest) = 32)
    AND (provider_head_digest IS NULL OR octet_length(provider_head_digest) = 32)
    AND (checkpoint_anchor_digest IS NULL OR octet_length(checkpoint_anchor_digest) = 32)
    AND (activation_evidence_digest IS NULL OR octet_length(activation_evidence_digest) = 32)
    AND (effect_resolution_digest IS NULL OR octet_length(effect_resolution_digest) = 32)
    AND (expected_provider_identity_digest IS NULL OR octet_length(expected_provider_identity_digest) = 32)
    AND (
      -- A legacy row is entirely null.
      (commitment_jcs IS NULL
        AND provider_head_jcs IS NULL
        AND checkpoint_anchor_jcs IS NULL
        AND effect_reason IS NULL
        AND attestation_expires_at IS NULL
        AND activation_deadline IS NULL
        AND expected_provider_identity_digest IS NULL
        AND activation_evidence_jcs IS NULL
        AND effect_resolution_jcs IS NULL)
      OR
      -- A prepared row contains only the immutable commitment pair.
      (commitment_jcs IS NOT NULL
        AND provider_head_jcs IS NULL
        AND checkpoint_anchor_jcs IS NULL
        AND effect_reason IS NULL
        AND attestation_expires_at IS NULL
        AND activation_deadline IS NULL
        AND expected_provider_identity_digest IS NULL
        AND activation_evidence_jcs IS NULL
        AND effect_resolution_jcs IS NULL)
      OR
      -- A terminal row contains Head, evidence, resolution, and a closed reason.
      (commitment_jcs IS NOT NULL
        AND provider_head_jcs IS NOT NULL
        AND activation_evidence_jcs IS NOT NULL
        AND effect_resolution_jcs IS NOT NULL
        AND effect_reason IN ('none','failed','superseded','activation_deadline_expired','validation_rejected')
        AND (
          -- Final-not-applied rows carry neither checkpoint nor external time bounds.
          (checkpoint_anchor_jcs IS NULL
            AND attestation_expires_at IS NULL
            AND activation_deadline IS NULL
            AND expected_provider_identity_digest IS NULL
            AND effect_reason <> 'none')
          OR
          -- A higher-authority terminal carries a checkpoint and no time material.
          (checkpoint_anchor_jcs IS NOT NULL
            AND attestation_expires_at IS NULL
            AND activation_deadline IS NULL
            AND expected_provider_identity_digest IS NULL
            AND effect_reason = 'superseded')
          OR
          -- Rollback-resistant evidence carries all external bounds and no checkpoint.
          (checkpoint_anchor_jcs IS NULL
            AND attestation_expires_at IS NOT NULL
            AND activation_deadline IS NOT NULL
            AND expected_provider_identity_digest IS NOT NULL
            AND effect_reason <> 'superseded'
            AND (effect_reason <> 'activation_deadline_expired' OR activation_deadline < attestation_expires_at))
        ))
    )), false)
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_guard_authority_proof_transition()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  new_row jsonb;
  old_row jsonb;
  row_data jsonb;
  prefixes text[];
  operation_columns text[];
  epoch_columns text[];
  sequence_columns text[];
  expected_kinds text[];
  secondary_preparable boolean[];
  domain_preparable boolean[];
  prefix_name text;
  proof_columns text[];
  allowed_columns text[];
  new_group jsonb;
  old_group jsonb;
  new_nonnull_count integer;
  old_nonnull_count integer;
  group_index integer;
  changed_group_count integer := 0;
  changed_group_index integer := 0;
  new_valid boolean;
  old_valid boolean;
  new_phase text;
  old_phase text;
  new_operation_id uuid;
  old_operation_id uuid;
  new_authority_epoch bigint;
  old_authority_epoch bigint;
  new_authority_sequence bigint;
  old_authority_sequence bigint;
  lock_operation_ids uuid[] := ARRAY[]::uuid[];
  lock_operation_id uuid;
  fence_row nodecontrol.control_plane_authority_fences%ROWTYPE;
  owner_count integer;
  expected_scope_kind text;
  expected_scope_digest bytea;
  node_id_value uuid;
  barrier_valid boolean;
  external_result_update boolean := false;
  effect_reason_value text;
BEGIN
  IF pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'authority proof mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;

  IF TG_OP = 'DELETE' THEN
    old_row := to_jsonb(OLD);
    row_data := old_row;
  ELSE
    new_row := to_jsonb(NEW);
    row_data := new_row;
    IF TG_OP = 'UPDATE' THEN
      old_row := to_jsonb(OLD);
    END IF;
  END IF;

  IF TG_TABLE_NAME = 'node_enrollment_grants' THEN
    prefixes := ARRAY['create_','claim_'];
    operation_columns := ARRAY['authority_operation_id','claim_authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch','claim_authority_epoch'];
    sequence_columns := ARRAY['authority_sequence','claim_authority_sequence'];
    expected_kinds := ARRAY['grant_create','grant_claim'];
    secondary_preparable := ARRAY[false,true];
    domain_preparable := ARRAY[false,false];
  ELSIF TG_TABLE_NAME = 'node_certificate_issuances' THEN
    prefixes := ARRAY['activation_'];
    operation_columns := ARRAY['authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch'];
    sequence_columns := ARRAY['authority_sequence'];
    expected_kinds := ARRAY['certificate_activate'];
    secondary_preparable := ARRAY[false];
    domain_preparable := ARRAY[true];
  ELSIF TG_TABLE_NAME = 'node_certificates' THEN
    prefixes := ARRAY['revoke_'];
    operation_columns := ARRAY['revoke_authority_operation_id'];
    epoch_columns := ARRAY['revoke_authority_epoch'];
    sequence_columns := ARRAY['revoke_authority_sequence'];
    expected_kinds := ARRAY['certificate_revoke'];
    secondary_preparable := ARRAY[true];
    domain_preparable := ARRAY[false];
  ELSIF TG_TABLE_NAME = 'node_state_transitions' THEN
    prefixes := ARRAY[''];
    operation_columns := ARRAY['authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch'];
    sequence_columns := ARRAY['authority_sequence'];
    expected_kinds := ARRAY[row_data->>'authority_effect_kind'];
    secondary_preparable := ARRAY[false];
    domain_preparable := ARRAY[false];
  ELSIF TG_TABLE_NAME = 'node_security_incidents' THEN
    prefixes := ARRAY['open_','resolve_'];
    operation_columns := ARRAY['authority_operation_id','resolution_authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch','resolution_authority_epoch'];
    sequence_columns := ARRAY['authority_sequence','resolution_authority_sequence'];
    expected_kinds := ARRAY['security_incident_open','security_incident_resolve'];
    secondary_preparable := ARRAY[false,true];
    domain_preparable := ARRAY[false,false];
  ELSIF TG_TABLE_NAME = 'node_resource_envelopes' THEN
    prefixes := ARRAY['activation_'];
    operation_columns := ARRAY['authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch'];
    sequence_columns := ARRAY['authority_sequence'];
    expected_kinds := ARRAY['resource_envelope_activate'];
    secondary_preparable := ARRAY[false];
    domain_preparable := ARRAY[false];
  ELSIF TG_TABLE_NAME = 'node_state_signing_intents' THEN
    prefixes := ARRAY['activation_'];
    operation_columns := ARRAY['authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch'];
    sequence_columns := ARRAY['authority_sequence'];
    expected_kinds := ARRAY[
      CASE row_data->>'signing_kind'
        WHEN 'desired' THEN 'desired_activate'
        WHEN 'recovery' THEN 'recovery_activate'
      END
    ];
    secondary_preparable := ARRAY[false];
    domain_preparable := ARRAY[true];
  ELSIF TG_TABLE_NAME = 'node_root_metadata_publish_intents' THEN
    prefixes := ARRAY['activation_'];
    operation_columns := ARRAY['authority_operation_id'];
    epoch_columns := ARRAY['authority_epoch'];
    sequence_columns := ARRAY['authority_sequence'];
    expected_kinds := ARRAY[
      CASE row_data->>'publish_kind'
        WHEN 'root' THEN 'root_publish'
        WHEN 'metadata' THEN 'metadata_publish'
      END
    ];
    secondary_preparable := ARRAY[false];
    domain_preparable := ARRAY[true];
  ELSE
    RAISE EXCEPTION 'unregistered authority proof owner %', TG_TABLE_NAME USING ERRCODE = '23514';
  END IF;

  -- Validate every proof shape and reject multi-group transitions before taking any fence lock.
  FOR group_index IN 1..array_length(prefixes, 1) LOOP
    prefix_name := prefixes[group_index];
    IF TG_OP <> 'DELETE' THEN
      new_group := jsonb_build_array(
        new_row -> (prefix_name || 'authority_effect_commitment_jcs'), new_row -> (prefix_name || 'authority_effect_commitment_digest'),
        new_row -> (prefix_name || 'authority_provider_head_jcs'), new_row -> (prefix_name || 'authority_provider_head_digest'),
        new_row -> (prefix_name || 'authority_checkpoint_anchor_jcs'), new_row -> (prefix_name || 'authority_checkpoint_anchor_digest'),
        new_row -> (prefix_name || 'authority_effect_reason'), new_row -> (prefix_name || 'authority_attestation_expires_at'),
        new_row -> (prefix_name || 'authority_activation_deadline'), new_row -> (prefix_name || 'authority_expected_provider_identity_digest'),
        new_row -> (prefix_name || 'authority_activation_evidence_jcs'), new_row -> (prefix_name || 'authority_activation_evidence_digest'),
        new_row -> (prefix_name || 'authority_effect_resolution_jcs'), new_row -> (prefix_name || 'authority_effect_resolution_digest')
      );
      SELECT count(*) INTO new_nonnull_count
      FROM jsonb_array_elements(new_group) AS item(value)
      WHERE value <> 'null'::jsonb;
      new_valid := nodecontrol.v7_authority_proof_group_valid(
        (new_row->>(prefix_name || 'authority_effect_commitment_jcs'))::bytea,
        (new_row->>(prefix_name || 'authority_effect_commitment_digest'))::bytea,
        (new_row->>(prefix_name || 'authority_provider_head_jcs'))::bytea,
        (new_row->>(prefix_name || 'authority_provider_head_digest'))::bytea,
        (new_row->>(prefix_name || 'authority_checkpoint_anchor_jcs'))::bytea,
        (new_row->>(prefix_name || 'authority_checkpoint_anchor_digest'))::bytea,
        new_row->>(prefix_name || 'authority_effect_reason'),
        (new_row->>(prefix_name || 'authority_attestation_expires_at'))::timestamptz,
        (new_row->>(prefix_name || 'authority_activation_deadline'))::timestamptz,
        (new_row->>(prefix_name || 'authority_expected_provider_identity_digest'))::bytea,
        (new_row->>(prefix_name || 'authority_activation_evidence_jcs'))::bytea,
        (new_row->>(prefix_name || 'authority_activation_evidence_digest'))::bytea,
        (new_row->>(prefix_name || 'authority_effect_resolution_jcs'))::bytea,
        (new_row->>(prefix_name || 'authority_effect_resolution_digest'))::bytea
      );
      IF NOT new_valid THEN
        RAISE EXCEPTION 'authority proof group has an invalid or partial lifecycle shape' USING ERRCODE = '23514';
      END IF;
      new_operation_id := NULLIF(new_row->>operation_columns[group_index], '')::uuid;
      new_authority_epoch := NULLIF(new_row->>epoch_columns[group_index], '')::bigint;
      new_authority_sequence := NULLIF(new_row->>sequence_columns[group_index], '')::bigint;
      IF (new_operation_id IS NULL) <> (new_authority_epoch IS NULL)
         OR (new_operation_id IS NULL) <> (new_authority_sequence IS NULL) THEN
        RAISE EXCEPTION 'authority proof operation tuple is partial' USING ERRCODE = '23514';
      END IF;
    END IF;

    IF TG_OP <> 'INSERT' THEN
      old_group := jsonb_build_array(
        old_row -> (prefix_name || 'authority_effect_commitment_jcs'), old_row -> (prefix_name || 'authority_effect_commitment_digest'),
        old_row -> (prefix_name || 'authority_provider_head_jcs'), old_row -> (prefix_name || 'authority_provider_head_digest'),
        old_row -> (prefix_name || 'authority_checkpoint_anchor_jcs'), old_row -> (prefix_name || 'authority_checkpoint_anchor_digest'),
        old_row -> (prefix_name || 'authority_effect_reason'), old_row -> (prefix_name || 'authority_attestation_expires_at'),
        old_row -> (prefix_name || 'authority_activation_deadline'), old_row -> (prefix_name || 'authority_expected_provider_identity_digest'),
        old_row -> (prefix_name || 'authority_activation_evidence_jcs'), old_row -> (prefix_name || 'authority_activation_evidence_digest'),
        old_row -> (prefix_name || 'authority_effect_resolution_jcs'), old_row -> (prefix_name || 'authority_effect_resolution_digest')
      );
      SELECT count(*) INTO old_nonnull_count
      FROM jsonb_array_elements(old_group) AS item(value)
      WHERE value <> 'null'::jsonb;
      old_valid := nodecontrol.v7_authority_proof_group_valid(
        (old_row->>(prefix_name || 'authority_effect_commitment_jcs'))::bytea,
        (old_row->>(prefix_name || 'authority_effect_commitment_digest'))::bytea,
        (old_row->>(prefix_name || 'authority_provider_head_jcs'))::bytea,
        (old_row->>(prefix_name || 'authority_provider_head_digest'))::bytea,
        (old_row->>(prefix_name || 'authority_checkpoint_anchor_jcs'))::bytea,
        (old_row->>(prefix_name || 'authority_checkpoint_anchor_digest'))::bytea,
        old_row->>(prefix_name || 'authority_effect_reason'),
        (old_row->>(prefix_name || 'authority_attestation_expires_at'))::timestamptz,
        (old_row->>(prefix_name || 'authority_activation_deadline'))::timestamptz,
        (old_row->>(prefix_name || 'authority_expected_provider_identity_digest'))::bytea,
        (old_row->>(prefix_name || 'authority_activation_evidence_jcs'))::bytea,
        (old_row->>(prefix_name || 'authority_activation_evidence_digest'))::bytea,
        (old_row->>(prefix_name || 'authority_effect_resolution_jcs'))::bytea,
        (old_row->>(prefix_name || 'authority_effect_resolution_digest'))::bytea
      );
      IF NOT old_valid THEN
        RAISE EXCEPTION 'stored authority proof group has an invalid lifecycle shape' USING ERRCODE = '23514';
      END IF;
      old_operation_id := NULLIF(old_row->>operation_columns[group_index], '')::uuid;
      old_authority_epoch := NULLIF(old_row->>epoch_columns[group_index], '')::bigint;
      old_authority_sequence := NULLIF(old_row->>sequence_columns[group_index], '')::bigint;
      IF (old_operation_id IS NULL) <> (old_authority_epoch IS NULL)
         OR (old_operation_id IS NULL) <> (old_authority_sequence IS NULL) THEN
        RAISE EXCEPTION 'stored authority proof operation tuple is partial' USING ERRCODE = '23514';
      END IF;
    END IF;

    IF TG_OP = 'UPDATE'
       AND (
         new_group IS DISTINCT FROM old_group
         OR (new_operation_id,new_authority_epoch,new_authority_sequence)
            IS DISTINCT FROM (old_operation_id,old_authority_epoch,old_authority_sequence)
       ) THEN
      changed_group_count := changed_group_count + 1;
      changed_group_index := group_index;
    END IF;
  END LOOP;

  IF changed_group_count > 1 THEN
    RAISE EXCEPTION 'only one authority proof group may advance per row mutation' USING ERRCODE = '23514';
  END IF;

  PERFORM pg_catalog.pg_advisory_xact_lock_shared(
    pg_catalog.hashtextextended('nodecontrol:v7-source-freeze',0)
  );
  IF nodecontrol.v7_source_is_frozen() THEN
    RAISE EXCEPTION 'authority proof owner is frozen after source seal' USING ERRCODE = '23514';
  END IF;

  -- Lock OLD and NEW operation IDs in canonical UUID order.
  FOR group_index IN 1..array_length(prefixes, 1) LOOP
    IF TG_OP <> 'DELETE' THEN
      new_operation_id := NULLIF(new_row->>operation_columns[group_index], '')::uuid;
      IF new_operation_id IS NOT NULL THEN
        lock_operation_ids := array_append(lock_operation_ids, new_operation_id);
      END IF;
    END IF;
    IF TG_OP <> 'INSERT' THEN
      old_operation_id := NULLIF(old_row->>operation_columns[group_index], '')::uuid;
      IF old_operation_id IS NOT NULL THEN
        lock_operation_ids := array_append(lock_operation_ids, old_operation_id);
      END IF;
    END IF;
  END LOOP;
  SELECT array_agg(operation_id ORDER BY operation_id)
    INTO lock_operation_ids
  FROM (
    SELECT DISTINCT operation_id
    FROM unnest(lock_operation_ids) AS locked(operation_id)
  ) AS ordered;
  FOREACH lock_operation_id IN ARRAY COALESCE(lock_operation_ids, ARRAY[]::uuid[]) LOOP
    PERFORM 1
    FROM nodecontrol.control_plane_authority_fences
    WHERE operation_id = lock_operation_id
    FOR UPDATE;
    IF NOT FOUND THEN
      RAISE EXCEPTION 'authority proof owner lacks its exact fence' USING ERRCODE = '23514';
    END IF;
  END LOOP;

  FOR group_index IN 1..array_length(prefixes, 1) LOOP
    prefix_name := prefixes[group_index];
    proof_columns := ARRAY[
      prefix_name || 'authority_effect_commitment_jcs', prefix_name || 'authority_effect_commitment_digest',
      prefix_name || 'authority_provider_head_jcs', prefix_name || 'authority_provider_head_digest',
      prefix_name || 'authority_checkpoint_anchor_jcs', prefix_name || 'authority_checkpoint_anchor_digest',
      prefix_name || 'authority_effect_reason', prefix_name || 'authority_attestation_expires_at',
      prefix_name || 'authority_activation_deadline', prefix_name || 'authority_expected_provider_identity_digest',
      prefix_name || 'authority_activation_evidence_jcs', prefix_name || 'authority_activation_evidence_digest',
      prefix_name || 'authority_effect_resolution_jcs', prefix_name || 'authority_effect_resolution_digest'
    ];

    IF TG_OP <> 'DELETE' THEN
      new_group := jsonb_build_array(
        new_row -> (proof_columns[1]), new_row -> (proof_columns[2]),
        new_row -> (proof_columns[3]), new_row -> (proof_columns[4]),
        new_row -> (proof_columns[5]), new_row -> (proof_columns[6]),
        new_row -> (proof_columns[7]), new_row -> (proof_columns[8]),
        new_row -> (proof_columns[9]), new_row -> (proof_columns[10]),
        new_row -> (proof_columns[11]), new_row -> (proof_columns[12]),
        new_row -> (proof_columns[13]), new_row -> (proof_columns[14])
      );
      SELECT count(*) INTO new_nonnull_count
      FROM jsonb_array_elements(new_group) AS item(value)
      WHERE value <> 'null'::jsonb;
      new_phase := CASE
        WHEN new_nonnull_count = 0 THEN 'legacy'
        WHEN new_nonnull_count = 2
          AND new_row->>(proof_columns[1]) IS NOT NULL
          AND new_row->>(proof_columns[2]) IS NOT NULL THEN 'commitment'
        ELSE 'terminal'
      END;
      new_operation_id := NULLIF(new_row->>operation_columns[group_index], '')::uuid;
      new_authority_epoch := NULLIF(new_row->>epoch_columns[group_index], '')::bigint;
      new_authority_sequence := NULLIF(new_row->>sequence_columns[group_index], '')::bigint;
    END IF;

    IF TG_OP <> 'INSERT' THEN
      old_group := jsonb_build_array(
        old_row -> (proof_columns[1]), old_row -> (proof_columns[2]),
        old_row -> (proof_columns[3]), old_row -> (proof_columns[4]),
        old_row -> (proof_columns[5]), old_row -> (proof_columns[6]),
        old_row -> (proof_columns[7]), old_row -> (proof_columns[8]),
        old_row -> (proof_columns[9]), old_row -> (proof_columns[10]),
        old_row -> (proof_columns[11]), old_row -> (proof_columns[12]),
        old_row -> (proof_columns[13]), old_row -> (proof_columns[14])
      );
      SELECT count(*) INTO old_nonnull_count
      FROM jsonb_array_elements(old_group) AS item(value)
      WHERE value <> 'null'::jsonb;
      old_phase := CASE
        WHEN old_nonnull_count = 0 THEN 'legacy'
        WHEN old_nonnull_count = 2
          AND old_row->>(proof_columns[1]) IS NOT NULL
          AND old_row->>(proof_columns[2]) IS NOT NULL THEN 'commitment'
        ELSE 'terminal'
      END;
      old_operation_id := NULLIF(old_row->>operation_columns[group_index], '')::uuid;
      old_authority_epoch := NULLIF(old_row->>epoch_columns[group_index], '')::bigint;
      old_authority_sequence := NULLIF(old_row->>sequence_columns[group_index], '')::bigint;
    END IF;

    IF TG_OP = 'DELETE' THEN
      new_operation_id := old_operation_id;
      new_authority_epoch := old_authority_epoch;
      new_authority_sequence := old_authority_sequence;
      new_phase := old_phase;
    END IF;

    IF new_operation_id IS NULL THEN
      IF new_phase <> 'legacy' THEN
        RAISE EXCEPTION 'unbound authority proof group must remain empty' USING ERRCODE = '23514';
      END IF;
      IF TG_OP = 'UPDATE' AND old_operation_id IS NOT NULL THEN
        RAISE EXCEPTION 'authority proof operation binding is immutable' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    SELECT * INTO fence_row
    FROM nodecontrol.control_plane_authority_fences
    WHERE operation_id = new_operation_id
    FOR UPDATE;

    IF expected_kinds[group_index] IS NULL
       OR (fence_row.effect_kind,fence_row.authority_epoch,fence_row.authority_sequence)
          IS DISTINCT FROM (expected_kinds[group_index],new_authority_epoch,new_authority_sequence) THEN
      RAISE EXCEPTION 'authority proof/fence effect kind or authority tuple mismatch' USING ERRCODE = '23514';
    END IF;

    IF TG_TABLE_NAME = 'node_root_metadata_publish_intents' THEN
      expected_scope_kind := 'global_node_trust';
      expected_scope_digest := decode('f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1','hex');
    ELSE
      expected_scope_kind := 'node';
      node_id_value := NULLIF(row_data->>'node_id', '')::uuid;
      IF node_id_value IS NULL THEN
        RAISE EXCEPTION 'node authority proof owner lacks its node scope' USING ERRCODE = '23514';
      END IF;
      expected_scope_digest := pg_catalog.sha256(
        pg_catalog.convert_to('TALENRO-NODE-AUTHORITY-SCOPE-V1','UTF8')
        || decode('00','hex') || pg_catalog.uuid_send(node_id_value)
      );
    END IF;
    IF (fence_row.scope_kind,fence_row.scope_digest)
       IS DISTINCT FROM (expected_scope_kind,expected_scope_digest) THEN
      RAISE EXCEPTION 'authority proof/fence scope digest mismatch' USING ERRCODE = '23514';
    END IF;

    IF fence_row.authority_protocol_profile = 'legacy_v6' THEN
      IF TG_OP = 'INSERT' OR TG_OP = 'DELETE'
         OR (TG_OP = 'UPDATE' AND group_index = changed_group_index) THEN
        RAISE EXCEPTION 'legacy_v6 authority proof owners are read-only after v7 activation' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;
    IF fence_row.authority_protocol_profile <> 'claim_v1'
       OR fence_row.protocol_activation_id IS NULL THEN
      RAISE EXCEPTION 'authority proof fence has an unknown protocol profile' USING ERRCODE = '23514';
    END IF;

    SELECT EXISTS (
      SELECT 1
      FROM nodecontrol.control_plane_authority_protocol_activations AS activation
      JOIN nodecontrol.control_plane_authority_protocol_upgrade_attempts AS attempt
        ON attempt.body_digest = activation.attempt_digest
       AND attempt.activation_id = activation.activation_id
      JOIN nodecontrol.control_plane_authority_protocol_upgrade_intents AS intent
        ON intent.body_digest = attempt.upgrade_intent_digest
       AND intent.activation_id = activation.activation_id
      JOIN nodecontrol.control_plane_authority_runtime_registration_results AS registration
        ON registration.body_digest = activation.runtime_registration_result_digest
       AND registration.activation_id = activation.activation_id
      LEFT JOIN nodecontrol.control_plane_authority_runtime_rebind_results AS rebind
        ON rebind.body_digest = attempt.latest_runtime_rebind_result_digest_or_null
      JOIN nodecontrol.control_plane_authority_protocol_activation_completions AS completion
        ON completion.activation_id = activation.activation_id
       AND completion.completion_id = attempt.completion_id
      JOIN nodecontrol.control_plane_authority_protocol_activation_releases AS release
        ON release.activation_id = activation.activation_id
       AND release.release_preparation_id = attempt.release_preparation_id
       AND release.open_id = attempt.open_id
      WHERE activation.activation_id = fence_row.protocol_activation_id
        AND activation.mode = 'empty_in_place'
        AND attempt.mode = activation.mode
        AND activation.protocol_profile = 'claim_v1'
        AND activation.deployment_id = attempt.deployment_id
        AND activation.database_identity_digest = intent.database_identity_digest
        AND activation.database_incarnation_attestation_digest = attempt.database_incarnation_attestation_digest
        AND activation.genesis_database_incarnation_registration_digest = attempt.database_incarnation_registration_digest
        AND activation.runtime_registration_result_digest = attempt.runtime_registration_result_digest
        AND activation.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM attempt.latest_runtime_rebind_result_digest_or_null
        AND activation.runtime_rebind_chain_digest = attempt.runtime_rebind_chain_digest
        AND activation.activation_runtime_instance_binding_digest = attempt.runtime_instance_binding_digest
        AND activation.database_legacy_absence_projection_digest = attempt.database_legacy_absence_projection_digest
        AND activation.attempt_database_inventory_digest = attempt.database_inventory_digest
        AND activation.provider_namespace_absence_digest = attempt.provider_namespace_absence_digest
        AND activation.environment_inventory_digest = attempt.environment_inventory_digest
        AND activation.environment_inventory_anchor_set_digest = attempt.environment_inventory_anchor_set_digest
        AND activation.local_runtime_isolation_digest = attempt.local_runtime_isolation_digest
        AND activation.legacy_runtime_shutdown_digest = attempt.legacy_runtime_shutdown_digest
        AND activation.legacy_runtime_shutdown_set_digest = attempt.legacy_runtime_shutdown_set_digest
        AND activation.credential_policy_digest = attempt.credential_policy_digest
        AND activation.epoch_evidence_digest = attempt.epoch_evidence_digest
        AND activation.genesis_epoch_transition_root_digest = attempt.genesis_epoch_transition_root_digest
        AND activation.selected_genesis_epoch = attempt.selected_genesis_epoch
      AND activation.provider_identity_digest = registration.provider_identity_digest
      AND activation.provider_endpoint_identity_digest = registration.provider_endpoint_identity_digest
      AND activation.namespace = registration.namespace
      AND registration.upgrade_intent_digest = intent.body_digest
      AND registration.credential_policy_digest = attempt.credential_policy_digest
      AND registration.database_incarnation_attestation_digest = attempt.database_incarnation_attestation_digest
      AND registration.genesis_database_identity_digest = intent.database_identity_digest
      AND attempt.request_nonce = intent.request_nonce
      AND attempt.deployment_id = intent.observed_deployment_id
      AND attempt.local_runtime_isolation_digest = intent.local_runtime_isolation_digest
      AND intent.created_at <= registration.recorded_at
      AND registration.recorded_at <= attempt.created_at
      AND attempt.created_at <= activation.activated_at
      AND (
        (attempt.latest_runtime_rebind_result_digest_or_null IS NULL
          AND rebind.body_digest IS NULL
          AND attempt.runtime_rebind_chain_digest = registration.runtime_rebind_chain_digest
          AND attempt.runtime_instance_binding_digest = registration.runtime_instance_binding_digest)
        OR
        (attempt.latest_runtime_rebind_result_digest_or_null IS NOT NULL
          AND rebind.body_digest = attempt.latest_runtime_rebind_result_digest_or_null
          AND rebind.activation_id = activation.activation_id
          AND rebind.runtime_registration_result_digest = registration.body_digest
          AND rebind.current_database_identity_digest = activation.database_identity_digest
          AND rebind.current_database_incarnation_registration_digest = attempt.database_incarnation_registration_digest
          AND rebind.current_runtime_rebind_chain_digest = attempt.runtime_rebind_chain_digest
          AND rebind.current_runtime_instance_binding_digest = attempt.runtime_instance_binding_digest
          AND registration.recorded_at <= rebind.recorded_at
          AND rebind.recorded_at <= attempt.created_at)
      )
      AND completion.activation_digest = activation.body_digest
        AND completion.preparation_digest = activation.preparation_digest
        AND completion.provider_completion_phase = 'genesis_completed_pending_release'
        AND completion.current_database_incarnation_registration_digest = activation.genesis_database_incarnation_registration_digest
        AND completion.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM activation.latest_runtime_rebind_result_digest_or_null
        AND completion.runtime_rebind_chain_digest = activation.runtime_rebind_chain_digest
        AND completion.current_runtime_instance_binding_digest = activation.activation_runtime_instance_binding_digest
        AND completion.credential_policy_digest = activation.credential_policy_digest
        AND completion.epoch_evidence_digest = activation.epoch_evidence_digest
        AND completion.genesis_epoch_transition_root_digest = activation.genesis_epoch_transition_root_digest
        AND completion.selected_genesis_epoch = activation.selected_genesis_epoch
        AND release.activation_digest = activation.body_digest
        AND release.completion_digest = completion.body_digest
        AND release.provider_release_phase = 'genesis_release_prepared'
        AND release.current_database_incarnation_registration_digest = completion.current_database_incarnation_registration_digest
        AND release.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM completion.latest_runtime_rebind_result_digest_or_null
        AND release.runtime_rebind_chain_digest = completion.runtime_rebind_chain_digest
        AND release.current_runtime_instance_binding_digest = completion.current_runtime_instance_binding_digest
        AND release.credential_policy_digest = completion.credential_policy_digest
        AND release.epoch_evidence_digest = completion.epoch_evidence_digest
        AND release.genesis_epoch_transition_root_digest = completion.genesis_epoch_transition_root_digest
        AND release.selected_genesis_epoch = completion.selected_genesis_epoch
        AND completion.completed_at >= activation.activated_at
        AND release.released_at >= completion.completed_at
    ) INTO barrier_valid;
    IF NOT barrier_valid THEN
      RAISE EXCEPTION 'authority proof lacks its exact activation completion release barrier' USING ERRCODE = '23514';
    END IF;

    SELECT count(*) INTO owner_count
    FROM (
      SELECT authority_operation_id AS operation_id FROM nodecontrol.node_enrollment_grants
      UNION ALL SELECT claim_authority_operation_id FROM nodecontrol.node_enrollment_grants
      UNION ALL SELECT authority_operation_id FROM nodecontrol.node_certificate_issuances
      UNION ALL SELECT revoke_authority_operation_id FROM nodecontrol.node_certificates
      UNION ALL SELECT authority_operation_id FROM nodecontrol.node_state_transitions
      UNION ALL SELECT authority_operation_id FROM nodecontrol.node_security_incidents
      UNION ALL SELECT resolution_authority_operation_id FROM nodecontrol.node_security_incidents
      UNION ALL SELECT authority_operation_id FROM nodecontrol.node_resource_envelopes
      UNION ALL SELECT authority_operation_id FROM nodecontrol.node_state_signing_intents
      UNION ALL SELECT authority_operation_id FROM nodecontrol.node_root_metadata_publish_intents
    ) AS owners
    WHERE operation_id = new_operation_id;

    IF TG_OP = 'INSERT' THEN
      IF group_index > 1 THEN
        RAISE EXCEPTION 'secondary authority proof group must begin empty' USING ERRCODE = '23514';
      END IF;
      IF fence_row.provider_status <> 'reserved'
         OR fence_row.visibility_state <> 'fence_pending'
         OR fence_row.effect_digest IS NOT NULL
         OR fence_row.abort_claimed_at IS NOT NULL
         OR fence_row.abort_reason IS NOT NULL
         OR owner_count <> 0 THEN
        RAISE EXCEPTION 'new authority proof owner requires an exact unbound unclaimed reservation' USING ERRCODE = '23514';
      END IF;
      IF new_phase = 'terminal' THEN
        RAISE EXCEPTION 'new authority proof owner cannot begin terminal' USING ERRCODE = '23514';
      END IF;
      IF domain_preparable[group_index] THEN
        IF new_phase <> 'legacy' THEN
          RAISE EXCEPTION 'domain-prepared authority owner must begin with an all-null proof' USING ERRCODE = '23514';
        END IF;
      ELSIF new_phase <> 'commitment' THEN
        RAISE EXCEPTION 'authority proof owner must begin commitment-only' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    IF TG_OP = 'DELETE' THEN
      IF old_phase <> 'terminal'
         OR fence_row.provider_status <> 'committed'
         OR fence_row.visibility_state <> 'active'
         OR fence_row.effect_digest IS DISTINCT FROM
            (old_row->>(proof_columns[2]))::bytea
         OR owner_count <> 1 THEN
        RAISE EXCEPTION 'authority proof DELETE requires its exact terminal fence and owner' USING ERRCODE = '23514';
      END IF;
      IF NOT (old_row ? 'retention_until')
         OR old_row->>'retention_until' IS NULL
         OR (old_row->>'retention_until')::timestamptz > pg_catalog.statement_timestamp() THEN
        RAISE EXCEPTION 'authority proof DELETE requires elapsed owner retention' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    IF old_operation_id IS NULL THEN
      IF NOT secondary_preparable[group_index]
         OR new_phase <> 'commitment'
         OR fence_row.provider_status <> 'reserved'
         OR fence_row.visibility_state <> 'fence_pending'
         OR fence_row.effect_digest IS NOT NULL
         OR fence_row.abort_claimed_at IS NOT NULL
         OR fence_row.abort_reason IS NOT NULL
         OR owner_count <> 0 THEN
        RAISE EXCEPTION 'secondary authority proof preparation requires an exact unbound reservation' USING ERRCODE = '23514';
      END IF;
      allowed_columns := ARRAY[
        operation_columns[group_index],epoch_columns[group_index],sequence_columns[group_index],
        proof_columns[1],proof_columns[2]
      ];
      IF (new_row - allowed_columns) IS DISTINCT FROM (old_row - allowed_columns) THEN
        RAISE EXCEPTION 'secondary authority proof preparation changed unrelated owner state' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    IF (new_operation_id,new_authority_epoch,new_authority_sequence)
       IS DISTINCT FROM (old_operation_id,old_authority_epoch,old_authority_sequence) THEN
      RAISE EXCEPTION 'authority proof operation binding is immutable' USING ERRCODE = '23514';
    END IF;

    IF old_phase = 'legacy' AND domain_preparable[group_index] THEN
      IF new_phase = 'legacy' THEN
        IF fence_row.provider_status <> 'reserved'
           OR fence_row.visibility_state <> 'fence_pending'
           OR fence_row.effect_digest IS NOT NULL
           OR fence_row.abort_claimed_at IS NOT NULL
           OR fence_row.abort_reason IS NOT NULL
           OR owner_count <> 1 THEN
          RAISE EXCEPTION 'domain-prepared external result requires its exact unbound unclaimed reservation' USING ERRCODE = '23514';
        END IF;
        IF changed_group_count <> 0 THEN
          RAISE EXCEPTION 'domain-prepared proof cannot change before commitment' USING ERRCODE = '23514';
        END IF;
        IF new_row IS NOT DISTINCT FROM old_row THEN
          CONTINUE;
        END IF;
        IF TG_TABLE_NAME = 'node_certificate_issuances' THEN
          allowed_columns := ARRAY['serial_bytes','leaf_der','leaf_der_sha256','chain_der','chain_der_sha256','not_before','not_after','updated_at'];
          IF num_nonnulls(
               old_row->>'serial_bytes',old_row->>'leaf_der',old_row->>'leaf_der_sha256',
               old_row->>'chain_der',old_row->>'chain_der_sha256',old_row->>'not_before',old_row->>'not_after'
             ) <> 0
             OR num_nonnulls(
               new_row->>'serial_bytes',new_row->>'leaf_der',new_row->>'leaf_der_sha256',
               new_row->>'chain_der',new_row->>'chain_der_sha256',new_row->>'not_before',new_row->>'not_after'
             ) <> 7 THEN
            RAISE EXCEPTION 'certificate activation external result must advance exactly once' USING ERRCODE = '23514';
          END IF;
        ELSIF TG_TABLE_NAME = 'node_state_signing_intents' THEN
          allowed_columns := ARRAY['signature','signature_verified_at','updated_at'];
          IF num_nonnulls(old_row->>'signature',old_row->>'signature_verified_at') <> 0
             OR num_nonnulls(new_row->>'signature',new_row->>'signature_verified_at') <> 2 THEN
            RAISE EXCEPTION 'state signing external result must advance exactly once' USING ERRCODE = '23514';
          END IF;
        ELSIF TG_TABLE_NAME = 'node_root_metadata_publish_intents' THEN
          allowed_columns := ARRAY['published_envelope','published_envelope_digest','updated_at'];
          IF num_nonnulls(old_row->>'published_envelope',old_row->>'published_envelope_digest') <> 0
             OR num_nonnulls(new_row->>'published_envelope',new_row->>'published_envelope_digest') <> 2 THEN
            RAISE EXCEPTION 'publish external result must advance exactly once' USING ERRCODE = '23514';
          END IF;
        ELSE
          RAISE EXCEPTION 'authority proof domain-prepared result registry is incomplete' USING ERRCODE = '23514';
        END IF;
        IF (new_row - allowed_columns) IS DISTINCT FROM (old_row - allowed_columns) THEN
          RAISE EXCEPTION 'domain-prepared external result changed unrelated owner state' USING ERRCODE = '23514';
        END IF;
        external_result_update := true;
        CONTINUE;
      ELSIF new_phase = 'commitment' THEN
        IF fence_row.provider_status <> 'reserved'
           OR fence_row.visibility_state <> 'fence_pending'
           OR fence_row.effect_digest IS NOT NULL
           OR fence_row.abort_claimed_at IS NOT NULL
           OR fence_row.abort_reason IS NOT NULL
           OR owner_count <> 1 THEN
          RAISE EXCEPTION 'domain-prepared commitment requires its exact unbound reservation' USING ERRCODE = '23514';
        END IF;
        allowed_columns := ARRAY[proof_columns[1],proof_columns[2]];
        IF (new_row - allowed_columns) IS DISTINCT FROM (old_row - allowed_columns) THEN
          RAISE EXCEPTION 'domain-prepared commitment changed external material or owner inputs' USING ERRCODE = '23514';
        END IF;
        CONTINUE;
      END IF;
      RAISE EXCEPTION 'domain-prepared authority proof has an illegal transition' USING ERRCODE = '23514';
    END IF;

    IF old_phase = 'commitment' AND new_phase = 'commitment' THEN
      IF new_group IS DISTINCT FROM old_group THEN
        RAISE EXCEPTION 'authority proof commitment is immutable' USING ERRCODE = '23514';
      END IF;
      IF new_row IS DISTINCT FROM old_row AND changed_group_count = 0 THEN
        RAISE EXCEPTION 'commitment-only authority proof cannot advance business outcome' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    IF old_phase = 'commitment' AND new_phase = 'terminal' THEN
      IF new_row -> (proof_columns[1]) IS DISTINCT FROM old_row -> (proof_columns[1])
         OR new_row -> (proof_columns[2]) IS DISTINCT FROM old_row -> (proof_columns[2]) THEN
        RAISE EXCEPTION 'authority proof commitment is immutable' USING ERRCODE = '23514';
      END IF;
      IF fence_row.provider_status <> 'committed'
         OR fence_row.visibility_state <> 'active'
         OR fence_row.effect_digest IS DISTINCT FROM (new_row->>(proof_columns[2]))::bytea
         OR owner_count <> 1 THEN
        RAISE EXCEPTION 'terminal authority proof requires its exact committed fence digest' USING ERRCODE = '23514';
      END IF;
      allowed_columns := proof_columns;
      effect_reason_value := new_row->>(proof_columns[7]);

      IF TG_TABLE_NAME = 'node_enrollment_grants' AND group_index = 2 THEN
        allowed_columns := allowed_columns || ARRAY[
          'consumed_at','consumption_attempt_id','consumption_request_digest','result_issuance_id',
          'terminal_reason','terminal_at','retention_until'
        ];
        IF effect_reason_value = 'none' THEN
          IF num_nonnulls(
               new_row->>'consumed_at',new_row->>'consumption_attempt_id',
               new_row->>'consumption_request_digest',new_row->>'result_issuance_id'
             ) <> 4
             OR new_row->>'terminal_reason' <> 'consumed'
             OR new_row->>'terminal_at' IS NULL
             OR new_row->>'retention_until' IS NULL THEN
            RAISE EXCEPTION 'grant claim applied proof must atomically record consumption' USING ERRCODE = '23514';
          END IF;
        ELSIF num_nonnulls(
                new_row->>'consumed_at',new_row->>'consumption_attempt_id',
                new_row->>'consumption_request_digest',new_row->>'result_issuance_id',
                new_row->>'terminal_reason',new_row->>'terminal_at',new_row->>'retention_until'
              ) <> 0 THEN
          RAISE EXCEPTION 'grant claim not-applied proof cannot record consumption' USING ERRCODE = '23514';
        END IF;
      ELSIF TG_TABLE_NAME = 'node_certificate_issuances' THEN
        allowed_columns := allowed_columns || ARRAY['status','failure_reason','updated_at','terminal_at','retention_until'];
        IF effect_reason_value = 'none' THEN
          IF new_row->>'status' <> 'active'
             OR new_row->>'failure_reason' IS NOT NULL
             OR new_row->>'terminal_at' IS NULL
             OR new_row->>'retention_until' IS NULL
             OR num_nonnulls(
                  new_row->>'serial_bytes',new_row->>'leaf_der',new_row->>'leaf_der_sha256',
                  new_row->>'chain_der',new_row->>'chain_der_sha256',new_row->>'not_before',new_row->>'not_after'
                ) <> 7 THEN
            RAISE EXCEPTION 'certificate activation applied proof must expose its exact result' USING ERRCODE = '23514';
          END IF;
        ELSIF new_row->>'status' NOT IN ('rejected','superseded','failed')
              OR new_row->>'failure_reason' IS NULL
              OR new_row->>'terminal_at' IS NULL
              OR new_row->>'retention_until' IS NULL THEN
          RAISE EXCEPTION 'certificate activation not-applied proof must remain non-active' USING ERRCODE = '23514';
        END IF;
      ELSIF TG_TABLE_NAME = 'node_certificates' THEN
        allowed_columns := allowed_columns || ARRAY['status','revoked_at','revoke_reason','updated_at'];
        IF effect_reason_value = 'none' THEN
          IF new_row->>'status' <> 'revoked'
             OR new_row->>'revoked_at' IS NULL
             OR new_row->>'revoke_reason' IS NULL THEN
            RAISE EXCEPTION 'certificate revoke applied proof must atomically revoke' USING ERRCODE = '23514';
          END IF;
        ELSIF new_row->>'revoked_at' IS NOT NULL
              OR new_row->>'revoke_reason' IS NOT NULL
              OR new_row->>'status' IS DISTINCT FROM old_row->>'status'
              OR new_row->>'updated_at' IS DISTINCT FROM old_row->>'updated_at' THEN
          RAISE EXCEPTION 'certificate revoke not-applied proof cannot revoke' USING ERRCODE = '23514';
        END IF;
      ELSIF TG_TABLE_NAME = 'node_state_transitions' THEN
        allowed_columns := allowed_columns || ARRAY['authority_effect_disposition'];
        IF (effect_reason_value = 'none' AND new_row->>'authority_effect_disposition' <> 'applied')
           OR (effect_reason_value <> 'none' AND new_row->>'authority_effect_disposition' <> 'not_applied') THEN
          RAISE EXCEPTION 'state transition disposition does not match authority proof outcome' USING ERRCODE = '23514';
        END IF;
      ELSIF TG_TABLE_NAME = 'node_security_incidents' AND group_index = 2 THEN
        allowed_columns := allowed_columns || ARRAY['remediation_digest','resolution_at','status','retention_until'];
        IF effect_reason_value = 'none' THEN
          IF num_nonnulls(new_row->>'remediation_digest',new_row->>'resolution_at') <> 2
             OR new_row->>'status' NOT IN ('resolution_pending_agent_ack','resolved') THEN
            RAISE EXCEPTION 'security incident resolve applied proof must record remediation' USING ERRCODE = '23514';
          END IF;
        ELSIF new_row->>'remediation_digest' IS NOT NULL
              OR new_row->>'resolution_at' IS NOT NULL
              OR new_row->>'status' IS DISTINCT FROM old_row->>'status'
              OR new_row->>'retention_until' IS DISTINCT FROM old_row->>'retention_until' THEN
          RAISE EXCEPTION 'security incident resolve not-applied proof cannot record remediation' USING ERRCODE = '23514';
        END IF;
      ELSIF TG_TABLE_NAME = 'node_state_signing_intents' THEN
        allowed_columns := allowed_columns || ARRAY['status','failure_reason','updated_at','terminal_at'];
        IF effect_reason_value = 'none' THEN
          IF new_row->>'status' <> 'active'
             OR new_row->>'failure_reason' IS NOT NULL
             OR new_row->>'terminal_at' IS NULL
             OR num_nonnulls(new_row->>'signature',new_row->>'signature_verified_at') <> 2 THEN
            RAISE EXCEPTION 'state signing applied proof must expose its exact signature' USING ERRCODE = '23514';
          END IF;
        ELSIF new_row->>'status' NOT IN ('failed','superseded')
              OR new_row->>'failure_reason' IS NULL
              OR new_row->>'terminal_at' IS NULL THEN
          RAISE EXCEPTION 'state signing not-applied proof must remain non-active' USING ERRCODE = '23514';
        END IF;
      ELSIF TG_TABLE_NAME = 'node_root_metadata_publish_intents' THEN
        allowed_columns := allowed_columns || ARRAY['status','failure_reason','updated_at','terminal_at'];
        IF effect_reason_value = 'none' THEN
          IF new_row->>'status' <> 'active'
             OR new_row->>'failure_reason' IS NOT NULL
             OR new_row->>'terminal_at' IS NULL
             OR num_nonnulls(new_row->>'published_envelope',new_row->>'published_envelope_digest') <> 2 THEN
            RAISE EXCEPTION 'publish applied proof must expose its exact envelope' USING ERRCODE = '23514';
          END IF;
        ELSIF new_row->>'status' NOT IN ('failed','superseded')
              OR new_row->>'failure_reason' IS NULL
              OR new_row->>'terminal_at' IS NULL THEN
          RAISE EXCEPTION 'publish not-applied proof must remain non-active' USING ERRCODE = '23514';
        END IF;
      END IF;

      IF (new_row - allowed_columns) IS DISTINCT FROM (old_row - allowed_columns) THEN
        RAISE EXCEPTION 'terminal authority proof changed fields outside its exact outcome' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    IF old_phase = 'terminal' AND new_phase = 'terminal'
       AND new_group IS NOT DISTINCT FROM old_group THEN
      IF new_row IS DISTINCT FROM old_row
         AND TG_TABLE_NAME IN ('node_resource_envelopes','node_state_transitions') THEN
        RAISE EXCEPTION 'immutable authority proof owner inputs cannot change' USING ERRCODE = '23514';
      END IF;
      CONTINUE;
    END IF;

    IF old_phase = 'legacy' AND new_phase = 'legacy'
       AND new_group IS NOT DISTINCT FROM old_group THEN
      CONTINUE;
    END IF;

    RAISE EXCEPTION 'authority proof lifecycle transition is not allowed' USING ERRCODE = '23514';
  END LOOP;

  IF TG_OP = 'UPDATE'
     AND changed_group_count = 0
     AND new_row IS DISTINCT FROM old_row
     AND NOT external_result_update
     AND TG_TABLE_NAME IN ('node_resource_envelopes','node_state_transitions') THEN
    RAISE EXCEPTION 'immutable authority proof owner inputs cannot change' USING ERRCODE = '23514';
  END IF;

  IF TG_OP = 'DELETE' THEN
    RETURN OLD;
  END IF;
  RETURN NEW;
END
$fn$;

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_migration_latches (
  installation_id uuid NOT NULL,
  installation_kind text COLLATE "C" NOT NULL,
  migration_version bigint NOT NULL,
  database_identity_digest bytea NOT NULL,
  up_catalog_digest bytea NOT NULL,
  down_state text COLLATE "C" NOT NULL,
  installed_at timestamp with time zone NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_ml_pk PRIMARY KEY (installation_id),
  CONSTRAINT ncv7_ml_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ml_uq01 UNIQUE (singleton_key),
  CONSTRAINT ncv7_ml_ck01 CHECK (singleton_key),
  CONSTRAINT ncv7_ml_ck02 CHECK (installation_kind IN ('production','disposable_fixture')),
  CONSTRAINT ncv7_ml_ck03 CHECK (migration_version = 7),
  CONSTRAINT ncv7_ml_ck04 CHECK (down_state = 'locked'),
  CONSTRAINT ncv7_ml_ck05 CHECK (octet_length(database_identity_digest) = 32 AND octet_length(up_catalog_digest) = 32 AND octet_length(body_digest) = 32),
  CONSTRAINT ncv7_ml_ck06 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_downgrade_authorizations (
  authorization_id uuid NOT NULL,
  installation_id uuid NOT NULL,
  migration_latch_digest bytea NOT NULL,
  database_identity_digest bytea NOT NULL,
  migration_version bigint NOT NULL,
  current_catalog_digest bytea NOT NULL,
  pristine_downgrade_inventory_digest bytea NOT NULL,
  provider_protocol_downgrade_retirement_set_digest bytea NOT NULL,
  environment_inventory_anchor_set_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_nonce bytea NOT NULL,
  authorization_scope text COLLATE "C" NOT NULL,
  issued_at timestamp with time zone NOT NULL,
  expires_at timestamp with time zone NOT NULL,
  authorization_envelope_jcs bytea NOT NULL,
  pristine_inventory_body_jcs bytea NOT NULL,
  provider_retirement_set_body_jcs bytea NOT NULL,
  provider_retirement_evidence_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_da_pk PRIMARY KEY (authorization_id),
  CONSTRAINT ncv7_da_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_da_uq01 UNIQUE (singleton_key),
  CONSTRAINT ncv7_da_ck01 CHECK (singleton_key),
  CONSTRAINT ncv7_da_ck02 CHECK (migration_version = 7 AND authorization_scope = 'down_00007_only'),
  CONSTRAINT ncv7_da_ck03 CHECK (expires_at > issued_at AND expires_at <= issued_at + interval '2 minutes'),
  CONSTRAINT ncv7_da_ck04 CHECK (octet_length(migration_latch_digest) = 32 AND octet_length(database_identity_digest) = 32 AND octet_length(current_catalog_digest) = 32 AND octet_length(pristine_downgrade_inventory_digest) = 32 AND octet_length(provider_protocol_downgrade_retirement_set_digest) = 32 AND octet_length(environment_inventory_anchor_set_digest) = 32 AND octet_length(transaction_nonce) = 32 AND octet_length(body_digest) = 32),
  CONSTRAINT ncv7_da_ck05 CHECK (
    octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576
    AND octet_length(authorization_envelope_jcs) BETWEEN 1 AND 1048576
    AND octet_length(pristine_inventory_body_jcs) BETWEEN 1 AND 1048576
    AND octet_length(provider_retirement_set_body_jcs) BETWEEN 1 AND 1048576
    AND octet_length(provider_retirement_evidence_jcs) BETWEEN 1 AND 1048576
  ),
  CONSTRAINT ncv7_da_fk01 FOREIGN KEY (installation_id) REFERENCES nodecontrol.control_plane_authority_protocol_migration_latches(installation_id) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX ncv7_da_fk01_ix ON nodecontrol.control_plane_authority_protocol_downgrade_authorizations (installation_id);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_upgrade_intents (
  intent_id uuid NOT NULL,
  installation_id uuid NOT NULL,
  installation_kind text COLLATE "C" NOT NULL,
  activation_id uuid NOT NULL,
  request_nonce bytea NOT NULL,
  credential_policy_update_id_or_null uuid,
  incarnation_registration_id uuid NOT NULL,
  provider_absence_proof_id uuid NOT NULL,
  observed_deployment_id uuid NOT NULL,
  database_identity_digest bytea NOT NULL,
  local_runtime_isolation_digest bytea NOT NULL,
  classification_state text COLLATE "C" NOT NULL,
  created_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_ui_pk PRIMARY KEY (intent_id),
  CONSTRAINT ncv7_ui_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ui_uq01 UNIQUE (activation_id),
  CONSTRAINT ncv7_ui_uq02 UNIQUE (singleton_key),
  CONSTRAINT ncv7_ui_ck01 CHECK (singleton_key),
  CONSTRAINT ncv7_ui_ck02 CHECK (installation_kind = 'production' AND classification_state = 'pending'),
  CONSTRAINT ncv7_ui_ck03 CHECK (octet_length(request_nonce) = 32 AND octet_length(database_identity_digest) = 32 AND octet_length(local_runtime_isolation_digest) = 32 AND octet_length(body_digest) = 32),
  CONSTRAINT ncv7_ui_ck04 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576),
  CONSTRAINT ncv7_ui_fk01 FOREIGN KEY (installation_id) REFERENCES nodecontrol.control_plane_authority_protocol_migration_latches(installation_id) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX ncv7_ui_fk01_ix ON nodecontrol.control_plane_authority_protocol_upgrade_intents (installation_id);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_runtime_registration_results (
  registration_id uuid NOT NULL,
  upgrade_intent_digest bytea NOT NULL,
  activation_id uuid NOT NULL,
  provider_identity_digest bytea NOT NULL,
  provider_endpoint_identity_digest bytea NOT NULL,
  namespace text COLLATE "C" NOT NULL,
  credential_policy_digest bytea NOT NULL,
  database_incarnation_attestation_digest bytea NOT NULL,
  provider_registration_digest bytea NOT NULL,
  genesis_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  runtime_instance_id uuid NOT NULL,
  runtime_instance_generation bigint NOT NULL,
  attestor_runtime_lease_digest bytea NOT NULL,
  runtime_rebind_chain_digest bytea NOT NULL,
  provider_head_digest bytea NOT NULL,
  provider_phase text COLLATE "C" NOT NULL,
  provider_control_sequence bigint NOT NULL,
  database_point bytea NOT NULL,
  recorded_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_rr_pk PRIMARY KEY (registration_id),
  CONSTRAINT ncv7_rr_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_rr_uq01 UNIQUE (activation_id),
  CONSTRAINT ncv7_rr_ck01 CHECK (provider_phase = 'registered_pending_genesis'),
  CONSTRAINT ncv7_rr_ck02 CHECK (runtime_instance_generation > 0 AND provider_control_sequence >= 0),
  CONSTRAINT ncv7_rr_ck03 CHECK (octet_length(database_point) BETWEEN 1 AND 4096 AND octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_runtime_rebind_results (
  rebind_id uuid NOT NULL,
  activation_id uuid NOT NULL,
  runtime_registration_result_digest bytea NOT NULL,
  previous_runtime_rebind_result_digest_or_null bytea,
  provider_rebind_request_digest bytea NOT NULL,
  provider_rebind_digest bytea NOT NULL,
  authorized_database_authority_head_digest bytea NOT NULL,
  database_authority_rebind_gap_attestation_digest_or_null bytea,
  database_authority_post_recovery_rebind_attestation_digest_or_null bytea,
  authorized_database_point bytea NOT NULL,
  previous_database_identity_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  previous_database_timeline_lineage_chain_digest bytea NOT NULL,
  current_database_timeline_lineage_chain_digest bytea NOT NULL,
  database_timeline_lineage_attestation_digest_or_null bytea,
  previous_database_incarnation_registration_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  previous_runtime_rebind_chain_digest bytea NOT NULL,
  current_runtime_rebind_chain_digest bytea NOT NULL,
  previous_runtime_instance_binding_digest bytea NOT NULL,
  current_runtime_instance_binding_digest bytea NOT NULL,
  current_runtime_instance_id uuid NOT NULL,
  current_runtime_instance_generation bigint NOT NULL,
  current_attestor_runtime_lease_digest bytea NOT NULL,
  revoked_previous_serving_lease_digest_or_null bytea,
  current_serving_lease_digest_or_null bytea,
  cancelled_epoch_transition_intent_digest_or_null bytea,
  cancelled_epoch_transition_application_digest_or_null bytea,
  cancelled_epoch_transition_resolution_digest_or_null bytea,
  cancelled_provider_pending_transition_digest_or_null bytea,
  cancellation_id_or_null uuid,
  previous_epoch_transition_terminal_chain_digest_or_null bytea,
  resulting_epoch_transition_terminal_chain_digest_or_null bytea,
  epoch_transition_recovery_state_or_null text COLLATE "C",
  epoch_transition_recovery_request_digest_or_null bytea,
  deferred_runtime_rebind_suffix_id_or_null uuid,
  deferred_runtime_rebind_ordinal_or_null bigint,
  epoch_transition_recovery_prefix_decision_digest_or_null bytea,
  epoch_transition_recovery_prefix_decision_commit_challenge_digest_or_null bytea,
  epoch_transition_recovery_prefix_decision_commit_attestation_digest_or_null bytea,
  epoch_transition_recovery_application_digest_or_null bytea,
  epoch_transition_recovery_application_commit_challenge_digest_or_null bytea,
  epoch_transition_recovery_application_commit_attestation_digest_or_null bytea,
  epoch_transition_recovery_application_archive_inspect_response_digest_or_null bytea,
  epoch_transition_recovery_resume_phase_or_null text COLLATE "C",
  previous_epoch_transition_recovery_consumption_history_digest_or_null bytea,
  result_epoch_transition_recovery_consumption_history_digest_or_null bytea,
  database_result_application jsonb NOT NULL,
  preserved_staging_exclusion_id_or_null uuid,
  preserved_staging_exclusion_acquire_request_digest_or_null bytea,
  preserved_staging_exclusion_digest_or_null bytea,
  preserved_staging_import_capability_recovery_intent_digest_or_null bytea,
  preserved_staging_recovery_commit_evidence_set_digest_or_null bytea,
  result_staging_exclusion_state_or_null text COLLATE "C",
  provider_phase text COLLATE "C" NOT NULL,
  provider_head_digest bytea NOT NULL,
  provider_control_sequence bigint NOT NULL,
  recorded_database_point bytea NOT NULL,
  recorded_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_rb_pk PRIMARY KEY (rebind_id),
  CONSTRAINT ncv7_rb_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_rb_ck01 CHECK (current_runtime_instance_generation > 0 AND provider_control_sequence >= 0),
  CONSTRAINT ncv7_rb_ck02 CHECK (octet_length(authorized_database_point) BETWEEN 1 AND 4096 AND octet_length(recorded_database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_rb_ck03 CHECK (deferred_runtime_rebind_ordinal_or_null IS NULL OR deferred_runtime_rebind_ordinal_or_null BETWEEN 1 AND 64),
  CONSTRAINT ncv7_rb_ck04 CHECK ((previous_runtime_rebind_result_digest_or_null IS NULL) OR octet_length(previous_runtime_rebind_result_digest_or_null) = 32),
  CONSTRAINT ncv7_rb_ck05 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);
CREATE UNIQUE INDEX ncv7_rb_uq01 ON nodecontrol.control_plane_authority_runtime_rebind_results (body_digest);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_upgrade_attempts (
  attempt_id uuid NOT NULL,
  upgrade_intent_digest bytea NOT NULL,
  activation_id uuid NOT NULL,
  preparation_id uuid NOT NULL,
  completion_id uuid NOT NULL,
  release_preparation_id uuid NOT NULL,
  open_id uuid NOT NULL,
  mode text COLLATE "C" NOT NULL,
  deployment_id uuid NOT NULL,
  request_nonce bytea NOT NULL,
  environment_inventory_digest bytea NOT NULL,
  environment_inventory_anchor_set_digest bytea NOT NULL,
  local_runtime_isolation_digest bytea NOT NULL,
  legacy_runtime_shutdown_digest bytea NOT NULL,
  legacy_runtime_shutdown_set_digest bytea NOT NULL,
  credential_policy_digest bytea NOT NULL,
  database_legacy_absence_projection_digest bytea NOT NULL,
  attempt_database_observation_digest bytea NOT NULL,
  provider_namespace_absence_digest bytea NOT NULL,
  database_inventory_digest bytea NOT NULL,
  database_incarnation_attestation_digest bytea NOT NULL,
  database_incarnation_registration_digest bytea NOT NULL,
  runtime_registration_result_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  epoch_evidence_digest bytea NOT NULL,
  genesis_epoch_transition_root_digest bytea NOT NULL,
  selected_genesis_epoch bigint NOT NULL,
  created_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_ua_pk PRIMARY KEY (attempt_id),
  CONSTRAINT ncv7_ua_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ua_uq01 UNIQUE (activation_id),
  CONSTRAINT ncv7_ua_uq02 UNIQUE (preparation_id),
  CONSTRAINT ncv7_ua_uq03 UNIQUE (completion_id),
  CONSTRAINT ncv7_ua_uq04 UNIQUE (release_preparation_id),
  CONSTRAINT ncv7_ua_uq05 UNIQUE (open_id),
  CONSTRAINT ncv7_ua_uq06 UNIQUE (singleton_key),
  CONSTRAINT ncv7_ua_ck01 CHECK (singleton_key),
  CONSTRAINT ncv7_ua_ck02 CHECK (mode IN ('empty_in_place','fresh_restore_target')),
  CONSTRAINT ncv7_ua_ck03 CHECK (selected_genesis_epoch > 0),
  CONSTRAINT ncv7_ua_ck04 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_intents (
  transition_id uuid NOT NULL,
  resolution_id uuid NOT NULL,
  cancellation_id uuid NOT NULL,
  terminal_application_id uuid NOT NULL,
  activation_id uuid NOT NULL,
  previous_epoch bigint NOT NULL,
  next_epoch bigint NOT NULL,
  previous_transition_digest bytea NOT NULL,
  previous_epoch_transition_resolution_digest_or_null bytea,
  previous_epoch_transition_terminal_application_digest_or_null bytea,
  previous_epoch_transition_terminal_chain_digest bytea NOT NULL,
  reason text COLLATE "C" NOT NULL,
  expected_provider_head_digest bytea NOT NULL,
  pre_transition_database_authority_head_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  created_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_ei_pk PRIMARY KEY (transition_id),
  CONSTRAINT ncv7_ei_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ei_uq01 UNIQUE (resolution_id),
  CONSTRAINT ncv7_ei_uq02 UNIQUE (cancellation_id),
  CONSTRAINT ncv7_ei_uq03 UNIQUE (terminal_application_id),
  CONSTRAINT ncv7_ei_ck01 CHECK (reason IN ('scheduled_authority_rotation','incident_recovery')),
  CONSTRAINT ncv7_ei_ck02 CHECK (previous_epoch > 0 AND next_epoch > previous_epoch),
  CONSTRAINT ncv7_ei_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_applications (
  transition_id uuid NOT NULL,
  resolution_id uuid NOT NULL,
  terminal_application_id uuid NOT NULL,
  transition_intent_digest bytea NOT NULL,
  provider_transition_request_digest bytea NOT NULL,
  provider_transition_digest bytea NOT NULL,
  previous_epoch bigint NOT NULL,
  next_epoch bigint NOT NULL,
  previous_epoch_transition_chain_digest bytea NOT NULL,
  next_epoch_transition_chain_digest bytea NOT NULL,
  previous_epoch_transition_terminal_chain_digest bytea NOT NULL,
  expected_provider_head_digest bytea NOT NULL,
  pending_provider_head_digest bytea NOT NULL,
  provider_control_sequence bigint NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  runtime_rebind_chain_digest bytea NOT NULL,
  serving_lease_digest bytea NOT NULL,
  provider_prepared_at timestamp with time zone NOT NULL,
  recorded_database_point bytea NOT NULL,
  recorded_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_ea_pk PRIMARY KEY (transition_id),
  CONSTRAINT ncv7_ea_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ea_uq01 UNIQUE (resolution_id),
  CONSTRAINT ncv7_ea_ck01 CHECK (previous_epoch > 0 AND next_epoch > previous_epoch AND provider_control_sequence >= 0),
  CONSTRAINT ncv7_ea_ck02 CHECK (octet_length(recorded_database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_ea_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_resolutions (
  resolution_id uuid NOT NULL,
  transition_id uuid NOT NULL,
  terminal_application_id uuid NOT NULL,
  transition_intent_digest bytea NOT NULL,
  transition_application_digest bytea NOT NULL,
  previous_epoch_transition_resolution_digest_or_null bytea,
  previous_epoch_transition_terminal_application_digest_or_null bytea,
  previous_epoch_transition_terminal_chain_digest bytea NOT NULL,
  resolved_epoch bigint NOT NULL,
  resolved_epoch_transition_chain_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  status text COLLATE "C" NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  resolved_database_point bytea NOT NULL,
  resolved_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_er_pk PRIMARY KEY (resolution_id),
  CONSTRAINT ncv7_er_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_er_uq01 UNIQUE (transition_id),
  CONSTRAINT ncv7_er_uq02 UNIQUE (terminal_application_id),
  CONSTRAINT ncv7_er_ck01 CHECK (status = 'approved_for_provider_resolution'),
  CONSTRAINT ncv7_er_ck02 CHECK (resolved_epoch > 0 AND octet_length(resolved_database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_er_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_cancellations (
  cancellation_id uuid NOT NULL,
  terminal_application_id uuid NOT NULL,
  transition_id uuid NOT NULL,
  resolution_id uuid NOT NULL,
  transition_intent_digest bytea NOT NULL,
  transition_application_digest_or_null bytea,
  transition_resolution_digest_or_null bytea,
  provider_pending_transition_digest_or_null bytea,
  provider_rebind_digest bytea NOT NULL,
  runtime_rebind_result_digest bytea NOT NULL,
  cancellation_kind text COLLATE "C" NOT NULL,
  previous_effective_epoch_transition_resolution_digest_or_null bytea,
  preserved_epoch bigint NOT NULL,
  preserved_epoch_transition_chain_digest bytea NOT NULL,
  previous_epoch_transition_terminal_chain_digest bytea NOT NULL,
  resulting_epoch_transition_terminal_chain_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  status text COLLATE "C" NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  cancelled_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_ec_pk PRIMARY KEY (cancellation_id),
  CONSTRAINT ncv7_ec_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ec_uq01 UNIQUE (transition_id),
  CONSTRAINT ncv7_ec_uq02 UNIQUE (terminal_application_id),
  CONSTRAINT ncv7_ec_ck01 CHECK (cancellation_kind IN ('preparation_absent_rebind','pending_provider_rebind')),
  CONSTRAINT ncv7_ec_ck02 CHECK (status = 'cancelled_without_epoch_advance'),
  CONSTRAINT ncv7_ec_ck03 CHECK ((transition_application_digest_or_null IS NULL) = (transition_resolution_digest_or_null IS NULL)),
  CONSTRAINT ncv7_ec_ck04 CHECK (preserved_epoch > 0 AND octet_length(database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_ec_ck05 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_terminal_applications (
  terminal_application_id uuid NOT NULL,
  transition_id uuid NOT NULL,
  resolution_id uuid NOT NULL,
  transition_intent_digest bytea NOT NULL,
  transition_application_digest_or_null bytea,
  transition_resolution_digest_or_null bytea,
  outcome text COLLATE "C" NOT NULL,
  provider_resolution_receipt_digest_or_null bytea,
  epoch_transition_cancellation_digest_or_null bytea,
  runtime_rebind_result_digest_or_null bytea,
  previous_epoch_transition_terminal_application_digest_or_null bytea,
  previous_epoch_transition_terminal_chain_digest bytea NOT NULL,
  terminal_epoch_transition_terminal_chain_digest bytea NOT NULL,
  previous_effective_epoch_transition_resolution_digest_or_null bytea,
  effective_epoch_transition_resolution_digest_or_null bytea,
  effective_epoch bigint NOT NULL,
  effective_epoch_transition_chain_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  applied_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_et_pk PRIMARY KEY (terminal_application_id),
  CONSTRAINT ncv7_et_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_et_uq01 UNIQUE (transition_id),
  CONSTRAINT ncv7_et_uq02 UNIQUE (resolution_id),
  CONSTRAINT ncv7_et_ck01 CHECK (outcome IN ('resolved','cancelled_by_rebind')),
  CONSTRAINT ncv7_et_ck02 CHECK (
    (outcome = 'resolved' AND transition_application_digest_or_null IS NOT NULL AND transition_resolution_digest_or_null IS NOT NULL AND provider_resolution_receipt_digest_or_null IS NOT NULL AND epoch_transition_cancellation_digest_or_null IS NULL AND runtime_rebind_result_digest_or_null IS NULL)
    OR
    (outcome = 'cancelled_by_rebind' AND provider_resolution_receipt_digest_or_null IS NULL AND epoch_transition_cancellation_digest_or_null IS NOT NULL AND runtime_rebind_result_digest_or_null IS NOT NULL AND ((transition_application_digest_or_null IS NULL AND transition_resolution_digest_or_null IS NULL) OR (transition_application_digest_or_null IS NOT NULL AND transition_resolution_digest_or_null IS NOT NULL)))
  ),
  CONSTRAINT ncv7_et_ck03 CHECK (effective_epoch > 0 AND octet_length(database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_et_ck04 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_intents (
  recovery_intent_id uuid NOT NULL,
  recovery_id uuid NOT NULL,
  recovery_application_id uuid NOT NULL,
  transition_id uuid NOT NULL,
  resolution_id uuid NOT NULL,
  cancellation_id uuid NOT NULL,
  terminal_application_id uuid NOT NULL,
  activation_id uuid NOT NULL,
  expected_terminal_outcome text COLLATE "C" NOT NULL,
  provider_terminal_inspect_digest bytea NOT NULL,
  provider_terminal_rebind_digest_or_null bytea,
  provider_terminal_runtime_rebind_chain_digest bytea NOT NULL,
  deferred_runtime_rebind_suffix_id uuid NOT NULL,
  expected_provider_head_digest bytea NOT NULL,
  expected_provider_current_epoch bigint NOT NULL,
  expected_provider_epoch_transition_chain_digest bytea NOT NULL,
  expected_provider_epoch_transition_terminal_chain_digest bytea NOT NULL,
  observed_transition_intent_row_count bigint NOT NULL,
  observed_transition_intent_digest_or_null bytea,
  observed_transition_application_row_count bigint NOT NULL,
  observed_transition_application_digest_or_null bytea,
  observed_transition_resolution_row_count bigint NOT NULL,
  observed_transition_resolution_digest_or_null bytea,
  observed_transition_cancellation_row_count bigint NOT NULL,
  observed_terminal_application_row_count bigint NOT NULL,
  pre_intent_database_authority_head_digest bytea NOT NULL,
  database_rebind_gap_anchor_result_digest_or_null bytea,
  database_rebind_gap_anchor_runtime_chain_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  transaction_nonce bytea NOT NULL,
  recovery_reason text COLLATE "C" NOT NULL,
  created_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_ri_pk PRIMARY KEY (recovery_intent_id),
  CONSTRAINT ncv7_ri_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ri_uq01 UNIQUE (recovery_id),
  CONSTRAINT ncv7_ri_uq02 UNIQUE (recovery_application_id),
  CONSTRAINT ncv7_ri_uq03 UNIQUE (transition_id),
  CONSTRAINT ncv7_ri_ck01 CHECK (expected_terminal_outcome IN ('resolved','cancelled_by_rebind')),
  CONSTRAINT ncv7_ri_ck02 CHECK (observed_transition_intent_row_count IN (0,1) AND observed_transition_application_row_count IN (0,1) AND observed_transition_resolution_row_count IN (0,1) AND observed_transition_cancellation_row_count = 0 AND observed_terminal_application_row_count = 0),
  CONSTRAINT ncv7_ri_ck03 CHECK ((observed_transition_intent_row_count = 0) = (observed_transition_intent_digest_or_null IS NULL) AND (observed_transition_application_row_count = 0) = (observed_transition_application_digest_or_null IS NULL) AND (observed_transition_resolution_row_count = 0) = (observed_transition_resolution_digest_or_null IS NULL)),
  CONSTRAINT ncv7_ri_ck04 CHECK (observed_transition_application_row_count = observed_transition_resolution_row_count),
  CONSTRAINT ncv7_ri_ck05 CHECK (recovery_reason = 'pitr_epoch_db_preimage_missing'),
  CONSTRAINT ncv7_ri_ck06 CHECK (expected_provider_current_epoch > 0 AND octet_length(database_point) BETWEEN 1 AND 4096 AND octet_length(transaction_nonce) = 32),
  CONSTRAINT ncv7_ri_ck07 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions (
  decision_id uuid NOT NULL,
  recovery_prefix_key_digest bytea NOT NULL,
  activation_id uuid NOT NULL,
  recovery_id uuid NOT NULL,
  recovery_intent_digest bytea NOT NULL,
  recovery_transcript_id bytea NOT NULL,
  provider_epoch_transition_recovery_digest bytea NOT NULL,
  deferred_runtime_rebind_suffix_id uuid NOT NULL,
  deferred_runtime_rebind_count bigint NOT NULL,
  deferred_runtime_rebind_tail_digest_or_null bytea,
  expected_provider_head_digest bytea NOT NULL,
  expected_provider_recovery_request_digest bytea NOT NULL,
  decision_kind text COLLATE "C" NOT NULL,
  next_deferred_runtime_rebind_ordinal_or_null bigint,
  pre_decision_database_authority_head_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  transaction_nonce bytea NOT NULL,
  decided_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_pd_pk PRIMARY KEY (decision_id),
  CONSTRAINT ncv7_pd_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_pd_uq01 UNIQUE (recovery_prefix_key_digest),
  CONSTRAINT ncv7_pd_ck01 CHECK (octet_length(recovery_prefix_key_digest) = 32),
  CONSTRAINT ncv7_pd_ck02 CHECK (decision_kind IN ('apply','replacement_rebind')),
  CONSTRAINT ncv7_pd_ck03 CHECK (deferred_runtime_rebind_count BETWEEN 0 AND 64 AND ((deferred_runtime_rebind_count = 0) = (deferred_runtime_rebind_tail_digest_or_null IS NULL))),
  CONSTRAINT ncv7_pd_ck04 CHECK ((decision_kind = 'apply' AND next_deferred_runtime_rebind_ordinal_or_null IS NULL) OR (decision_kind = 'replacement_rebind' AND deferred_runtime_rebind_count < 64 AND next_deferred_runtime_rebind_ordinal_or_null = deferred_runtime_rebind_count + 1)),
  CONSTRAINT ncv7_pd_ck05 CHECK (octet_length(database_point) BETWEEN 1 AND 4096 AND octet_length(transaction_nonce) = 32),
  CONSTRAINT ncv7_pd_ck06 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_applications (
  recovery_application_id uuid NOT NULL,
  recovery_intent_digest bytea NOT NULL,
  recovery_transcript_id bytea NOT NULL,
  provider_epoch_transition_recovery_digest bytea NOT NULL,
  recovery_prefix_key_digest bytea NOT NULL,
  epoch_transition_recovery_prefix_decision_digest bytea NOT NULL,
  transition_id uuid NOT NULL,
  resolution_id uuid NOT NULL,
  cancellation_id uuid NOT NULL,
  terminal_application_id uuid NOT NULL,
  terminal_outcome text COLLATE "C" NOT NULL,
  deferred_runtime_rebind_suffix_id uuid NOT NULL,
  deferred_runtime_rebind_count bigint NOT NULL,
  deferred_runtime_rebind_tail_digest_or_null bytea,
  provider_tail_runtime_rebind_chain_digest bytea NOT NULL,
  database_tail_runtime_rebind_result_digest_or_null bytea,
  database_tail_runtime_rebind_chain_digest bytea NOT NULL,
  post_rehydration_database_identity_digest bytea NOT NULL,
  post_rehydration_database_timeline_lineage_chain_digest bytea NOT NULL,
  source_database_commit_anchor_schema text COLLATE "C" NOT NULL,
  source_database_commit_anchor_digest bytea NOT NULL,
  pre_rehydration_database_authority_head_digest bytea NOT NULL,
  rehydrated_row_count bigint NOT NULL,
  rehydrated_rows jsonb NOT NULL,
  post_rehydration_database_authority_head_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  rehydrated_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_ra_pk PRIMARY KEY (recovery_application_id),
  CONSTRAINT ncv7_ra_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ra_uq01 UNIQUE (recovery_intent_digest),
  CONSTRAINT ncv7_ra_uq02 UNIQUE (recovery_prefix_key_digest),
  CONSTRAINT ncv7_ra_ck01 CHECK (terminal_outcome IN ('resolved','cancelled_by_rebind')),
  CONSTRAINT ncv7_ra_ck02 CHECK (octet_length(recovery_prefix_key_digest) = 32),
  CONSTRAINT ncv7_ra_ck03 CHECK (deferred_runtime_rebind_count BETWEEN 0 AND 64 AND ((deferred_runtime_rebind_count = 0) = (deferred_runtime_rebind_tail_digest_or_null IS NULL))),
  CONSTRAINT ncv7_ra_ck04 CHECK (rehydrated_row_count >= 0 AND jsonb_typeof(rehydrated_rows) = 'array' AND jsonb_array_length(rehydrated_rows) = rehydrated_row_count),
  CONSTRAINT ncv7_ra_ck05 CHECK (octet_length(database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_ra_ck06 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_legacy_database_source_retirements (
  database_retirement_id uuid NOT NULL,
  retirement_plan_id uuid NOT NULL,
  source_retirement_authorization_digest bytea NOT NULL,
  source_membership_retirement_digest bytea NOT NULL,
  upgrade_intent_digest bytea NOT NULL,
  activation_id uuid NOT NULL,
  environment_inventory_digest bytea NOT NULL,
  environment_inventory_anchor_set_digest bytea NOT NULL,
  environment_record_digest bytea NOT NULL,
  deployment_id uuid NOT NULL,
  database_identity_digest bytea NOT NULL,
  local_runtime_isolation_digest bytea NOT NULL,
  legacy_runtime_shutdown_digest bytea NOT NULL,
  legacy_source_inventory_digest bytea NOT NULL,
  normalized_catalog_digest bytea NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  phase text COLLATE "C" NOT NULL,
  retired_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_dr_pk PRIMARY KEY (database_retirement_id),
  CONSTRAINT ncv7_dr_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_dr_uq01 UNIQUE (activation_id),
  CONSTRAINT ncv7_dr_uq02 UNIQUE (singleton_key),
  CONSTRAINT ncv7_dr_ck01 CHECK (singleton_key AND phase = 'source_database_retired'),
  CONSTRAINT ncv7_dr_ck02 CHECK (octet_length(database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_dr_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_indeterminate_source_seals (
  source_seal_id uuid NOT NULL,
  upgrade_intent_digest bytea NOT NULL,
  activation_id uuid NOT NULL,
  request_nonce bytea NOT NULL,
  classification text COLLATE "C" NOT NULL,
  reason text COLLATE "C" NOT NULL,
  failed_checks text[] NOT NULL,
  local_runtime_isolation_digest bytea NOT NULL,
  observed_deployment_id uuid NOT NULL,
  observed_database_identity_digest bytea NOT NULL,
  migration_version bigint NOT NULL,
  normalized_catalog_digest_or_null bytea,
  local_locked_table_inventory_digest_or_null bytea,
  environment_inventory_digest_or_null bytea,
  environment_inventory_anchor_set_digest_or_null bytea,
  observed_provider_identity_digest_or_null bytea,
  observed_provider_namespace_or_null text COLLATE "C",
  observed_provider_head_digest_or_null bytea,
  missing_external_evidence text[] NOT NULL,
  sealed_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_is_pk PRIMARY KEY (source_seal_id),
  CONSTRAINT ncv7_is_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_is_uq01 UNIQUE (singleton_key),
  CONSTRAINT ncv7_is_ck01 CHECK (singleton_key AND classification = 'indeterminate' AND migration_version = 7),
  CONSTRAINT ncv7_is_ck02 CHECK (cardinality(failed_checks) > 0 AND nodecontrol.v7_text_array_is_sorted_unique(failed_checks)),
  CONSTRAINT ncv7_is_ck03 CHECK (nodecontrol.v7_text_array_is_sorted_unique(missing_external_evidence)),
  CONSTRAINT ncv7_is_ck04 CHECK (reason IN ('dependency_unavailable','dependency_timeout','environment_inventory_missing','environment_inventory_incomplete','environment_inventory_stale','environment_inventory_replayed','environment_inventory_signature_invalid','provider_identity_mismatch','provider_inspect_unavailable','provider_credential_policy_invalid','external_shutdown_incomplete','nonce_conflict','catalog_unknown','legacy_epoch_indeterminate')),
  CONSTRAINT ncv7_is_ck05 CHECK (octet_length(request_nonce) = 32),
  CONSTRAINT ncv7_is_ck06 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_legacy_source_seals (
  source_seal_id uuid NOT NULL,
  upgrade_intent_digest bytea NOT NULL,
  activation_id uuid NOT NULL,
  classification text COLLATE "C" NOT NULL,
  indeterminate_source_seal_digest_or_null bytea,
  environment_inventory_digest bytea NOT NULL,
  environment_inventory_anchor_set_digest bytea NOT NULL,
  local_runtime_isolation_digest bytea NOT NULL,
  legacy_runtime_shutdown_digest bytea NOT NULL,
  legacy_runtime_shutdown_set_digest bytea NOT NULL,
  source_membership_retirement_digest bytea NOT NULL,
  legacy_source_retirement_set_digest bytea NOT NULL,
  deployment_id uuid NOT NULL,
  database_identity_digest bytea NOT NULL,
  normalized_catalog_digest bytea NOT NULL,
  legacy_source_inventory_digest bytea NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  provider_identity_digest bytea NOT NULL,
  provider_endpoint_identity_digest bytea NOT NULL,
  provider_namespace text COLLATE "C" NOT NULL,
  provider_profile text COLLATE "C" NOT NULL,
  provider_head_digest bytea NOT NULL,
  provider_history_high_water bigint NOT NULL,
  credential_policy_digest bytea NOT NULL,
  database_route_closed_digest bytea NOT NULL,
  deployment_capability_revocation_digest bytea NOT NULL,
  legacy_epoch_source_set_digest bytea NOT NULL,
  legacy_epoch_maximum bigint NOT NULL,
  legacy_epoch_maximum_evidence_digest bytea NOT NULL,
  sealed_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_ls_pk PRIMARY KEY (source_seal_id),
  CONSTRAINT ncv7_ls_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_ls_uq01 UNIQUE (singleton_key),
  CONSTRAINT ncv7_ls_ck01 CHECK (singleton_key AND classification = 'legacy_or_nonempty'),
  CONSTRAINT ncv7_ls_ck02 CHECK (octet_length(provider_profile) BETWEEN 1 AND 128 AND provider_history_high_water >= 0 AND legacy_epoch_maximum >= 0),
  CONSTRAINT ncv7_ls_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_fresh_restore_requirements (
  requirement_id uuid NOT NULL,
  upgrade_intent_digest bytea NOT NULL,
  source_seal_kind text COLLATE "C" NOT NULL,
  source_seal_digest bytea NOT NULL,
  classification text COLLATE "C" NOT NULL,
  reason text COLLATE "C" NOT NULL,
  failed_checks text[] NOT NULL,
  missing_external_evidence text[] NOT NULL,
  source_deployment_id_or_null uuid,
  source_database_identity_digest bytea NOT NULL,
  source_provider_identity_digest_or_null bytea,
  source_provider_namespace_or_null text COLLATE "C",
  environment_inventory_digest_or_null bytea,
  environment_inventory_anchor_set_digest_or_null bytea,
  local_runtime_isolation_digest bytea NOT NULL,
  legacy_runtime_shutdown_digest_or_null bytea,
  legacy_runtime_shutdown_set_digest_or_null bytea,
  required_new_database_identity boolean NOT NULL,
  required_new_deployment_id boolean NOT NULL,
  required_new_provider_namespace boolean NOT NULL,
  required_non_exportable_incarnation boolean NOT NULL,
  blocked_capabilities text[] NOT NULL,
  created_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_fr_pk PRIMARY KEY (requirement_id),
  CONSTRAINT ncv7_fr_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_fr_uq01 UNIQUE (singleton_key),
  CONSTRAINT ncv7_fr_ck01 CHECK (singleton_key AND source_seal_kind IN ('indeterminate','legacy_complete') AND classification IN ('indeterminate','legacy_or_nonempty')),
  CONSTRAINT ncv7_fr_ck02 CHECK (required_new_database_identity AND required_new_deployment_id AND required_new_provider_namespace AND required_non_exportable_incarnation),
  CONSTRAINT ncv7_fr_ck03 CHECK (blocked_capabilities = ARRAY['c12_root_metadata_signer_rotation','node_authority_restore_reenrollment','node_operator_server_ca_rotation','operator_authorizer_change','trust_bundle_publish']::text[]),
  CONSTRAINT ncv7_fr_ck04 CHECK (nodecontrol.v7_text_array_is_sorted_unique(failed_checks) AND nodecontrol.v7_text_array_is_sorted_unique(missing_external_evidence)),
  CONSTRAINT ncv7_fr_ck05 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_activations (
  activation_id uuid NOT NULL,
  mode text COLLATE "C" NOT NULL,
  attempt_digest bytea NOT NULL,
  deployment_id uuid NOT NULL,
  database_identity_digest bytea NOT NULL,
  database_incarnation_attestation_digest bytea NOT NULL,
  genesis_database_incarnation_registration_digest bytea NOT NULL,
  runtime_registration_result_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  activation_runtime_instance_binding_digest bytea NOT NULL,
  database_legacy_absence_projection_digest bytea NOT NULL,
  attempt_database_inventory_digest bytea NOT NULL,
  activation_database_observation_digest bytea NOT NULL,
  provider_namespace_absence_digest bytea NOT NULL,
  environment_inventory_digest bytea NOT NULL,
  environment_inventory_anchor_set_digest bytea NOT NULL,
  local_runtime_isolation_digest bytea NOT NULL,
  legacy_runtime_shutdown_digest bytea NOT NULL,
  legacy_runtime_shutdown_set_digest bytea NOT NULL,
  credential_policy_digest bytea NOT NULL,
  epoch_evidence_digest bytea NOT NULL,
  genesis_epoch_transition_root_digest bytea NOT NULL,
  selected_genesis_epoch bigint NOT NULL,
  provider_identity_digest bytea NOT NULL,
  provider_endpoint_identity_digest bytea NOT NULL,
  namespace text COLLATE "C" NOT NULL,
  protocol_profile text COLLATE "C" NOT NULL,
  preparation_digest bytea NOT NULL,
  activated_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_pa_pk PRIMARY KEY (activation_id),
  CONSTRAINT ncv7_pa_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_pa_uq01 UNIQUE (singleton_key),
  CONSTRAINT ncv7_pa_ck01 CHECK (singleton_key AND mode IN ('empty_in_place','fresh_restore_target') AND protocol_profile = 'claim_v1'),
  CONSTRAINT ncv7_pa_ck02 CHECK (selected_genesis_epoch > 0),
  CONSTRAINT ncv7_pa_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_activation_completions (
  completion_id uuid NOT NULL,
  activation_id uuid NOT NULL,
  activation_digest bytea NOT NULL,
  preparation_digest bytea NOT NULL,
  provider_completion_digest bytea NOT NULL,
  provider_completion_phase text COLLATE "C" NOT NULL,
  database_activation_attestation_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  current_runtime_instance_binding_digest bytea NOT NULL,
  credential_policy_digest bytea NOT NULL,
  epoch_evidence_digest bytea NOT NULL,
  genesis_epoch_transition_root_digest bytea NOT NULL,
  selected_genesis_epoch bigint NOT NULL,
  completed_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_pc_pk PRIMARY KEY (completion_id),
  CONSTRAINT ncv7_pc_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_pc_uq01 UNIQUE (activation_id),
  CONSTRAINT ncv7_pc_uq02 UNIQUE (singleton_key),
  CONSTRAINT ncv7_pc_ck01 CHECK (singleton_key AND provider_completion_phase = 'genesis_completed_pending_release' AND selected_genesis_epoch > 0),
  CONSTRAINT ncv7_pc_ck02 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_protocol_activation_releases (
  release_preparation_id uuid NOT NULL,
  open_id uuid NOT NULL,
  activation_id uuid NOT NULL,
  activation_digest bytea NOT NULL,
  completion_digest bytea NOT NULL,
  provider_release_preparation_digest bytea NOT NULL,
  provider_release_phase text COLLATE "C" NOT NULL,
  database_completion_attestation_digest bytea NOT NULL,
  open_nonce bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  current_runtime_instance_binding_digest bytea NOT NULL,
  credential_policy_digest bytea NOT NULL,
  epoch_evidence_digest bytea NOT NULL,
  genesis_epoch_transition_root_digest bytea NOT NULL,
  selected_genesis_epoch bigint NOT NULL,
  released_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  singleton_key boolean NOT NULL DEFAULT true,
  CONSTRAINT ncv7_pr_pk PRIMARY KEY (release_preparation_id),
  CONSTRAINT ncv7_pr_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_pr_uq01 UNIQUE (open_id),
  CONSTRAINT ncv7_pr_uq02 UNIQUE (activation_id),
  CONSTRAINT ncv7_pr_uq03 UNIQUE (singleton_key),
  CONSTRAINT ncv7_pr_ck01 CHECK (singleton_key AND provider_release_phase = 'genesis_release_prepared' AND selected_genesis_epoch > 0),
  CONSTRAINT ncv7_pr_ck02 CHECK (octet_length(open_nonce) = 32),
  CONSTRAINT ncv7_pr_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_staging_import_capabilities (
  capability_id uuid NOT NULL,
  single_use_apply_id uuid NOT NULL,
  manifest_digest bytea NOT NULL,
  target_activation_id uuid NOT NULL,
  target_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  target_database_incarnation_registration_digest bytea NOT NULL,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  planned_staging_exclusion_id uuid NOT NULL,
  expected_pre_acquire_provider_head_digest bytea NOT NULL,
  pre_acquire_database_incarnation_proof_digest bytea NOT NULL,
  provider_phase text COLLATE "C" NOT NULL,
  provider_serving_lease_absent_digest bytea NOT NULL,
  database_route_closed_digest bytea NOT NULL,
  pre_import_inventory_digest bytea NOT NULL,
  allowed_object_set_digest bytea NOT NULL,
  expected_post_import_inventory_digest bytea NOT NULL,
  transaction_nonce bytea NOT NULL,
  issued_at timestamp with time zone NOT NULL,
  expires_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_sc_pk PRIMARY KEY (capability_id),
  CONSTRAINT ncv7_sc_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_sc_uq01 UNIQUE (single_use_apply_id),
  CONSTRAINT ncv7_sc_uq02 UNIQUE (planned_staging_exclusion_id),
  CONSTRAINT ncv7_sc_ck01 CHECK (provider_phase = 'fresh_v7_staging_closed'),
  CONSTRAINT ncv7_sc_ck02 CHECK (expires_at > issued_at AND expires_at <= issued_at + interval '5 minutes'),
  CONSTRAINT ncv7_sc_ck03 CHECK (octet_length(transaction_nonce) = 32),
  CONSTRAINT ncv7_sc_ck04 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_staging_import_capability_revocation_applications (
  revocation_application_id uuid NOT NULL,
  staging_import_capability_revocation_digest bytea NOT NULL,
  staging_import_capability_digest bytea NOT NULL,
  staging_import_capability_recovery_intent_digest_or_null bytea,
  staging_import_capability_recovery_application_digest_or_null bytea,
  capability_id uuid NOT NULL,
  single_use_apply_id uuid NOT NULL,
  manifest_digest bytea NOT NULL,
  target_activation_id uuid NOT NULL,
  target_database_incarnation_registration_digest bytea NOT NULL,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  current_database_timeline_lineage_chain_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  current_runtime_rebind_chain_digest bytea NOT NULL,
  current_runtime_instance_binding_digest bytea NOT NULL,
  import_application_absence_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  applied_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_sv_pk PRIMARY KEY (revocation_application_id),
  CONSTRAINT ncv7_sv_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_sv_uq01 UNIQUE (capability_id),
  CONSTRAINT ncv7_sv_uq02 UNIQUE (single_use_apply_id),
  CONSTRAINT ncv7_sv_ck01 CHECK ((staging_import_capability_recovery_intent_digest_or_null IS NULL) = (staging_import_capability_recovery_application_digest_or_null IS NULL)),
  CONSTRAINT ncv7_sv_ck02 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_staging_import_capability_recovery_intents (
  recovery_intent_id uuid NOT NULL,
  recovery_id uuid NOT NULL,
  recovery_application_id uuid NOT NULL,
  revocation_id uuid NOT NULL,
  revocation_application_id uuid NOT NULL,
  terminal_transition_id uuid NOT NULL,
  target_activation_id uuid NOT NULL,
  capability_id uuid NOT NULL,
  single_use_apply_id uuid NOT NULL,
  staging_import_capability_digest bytea NOT NULL,
  capability_registration_commit_attestation_digest bytea NOT NULL,
  staging_exclusion_id uuid NOT NULL,
  staging_exclusion_acquire_request_digest bytea NOT NULL,
  staging_exclusion_digest bytea NOT NULL,
  expected_current_held_provider_head_digest bytea NOT NULL,
  expected_staging_exclusion_state text COLLATE "C" NOT NULL,
  observed_capability_row_count bigint NOT NULL,
  observed_exact_capability_digest_or_null bytea,
  observed_import_application_row_count bigint NOT NULL,
  observed_revocation_application_row_count bigint NOT NULL,
  observed_recovery_application_row_count bigint NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  pre_intent_database_authority_head_digest bytea NOT NULL,
  recovery_reason text COLLATE "C" NOT NULL,
  recovery_scope text COLLATE "C" NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  transaction_nonce bytea NOT NULL,
  created_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_si_pk PRIMARY KEY (recovery_intent_id),
  CONSTRAINT ncv7_si_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_si_uq01 UNIQUE (recovery_id),
  CONSTRAINT ncv7_si_uq02 UNIQUE (recovery_application_id),
  CONSTRAINT ncv7_si_uq03 UNIQUE (revocation_id),
  CONSTRAINT ncv7_si_uq04 UNIQUE (revocation_application_id),
  CONSTRAINT ncv7_si_uq05 UNIQUE (terminal_transition_id),
  CONSTRAINT ncv7_si_uq06 UNIQUE (capability_id),
  CONSTRAINT ncv7_si_ck01 CHECK (expected_staging_exclusion_state IN ('held','held_recovery_required')),
  CONSTRAINT ncv7_si_ck02 CHECK (observed_capability_row_count IN (0,1) AND ((observed_capability_row_count = 0) = (observed_exact_capability_digest_or_null IS NULL))),
  CONSTRAINT ncv7_si_ck03 CHECK (observed_import_application_row_count = 0 AND observed_revocation_application_row_count = 0 AND observed_recovery_application_row_count = 0),
  CONSTRAINT ncv7_si_ck04 CHECK (recovery_reason IN ('pitr_capability_row_missing','pitr_exact_capability_runtime_stale') AND recovery_scope = 'rehydrate_for_revocation_only'),
  CONSTRAINT ncv7_si_ck05 CHECK (octet_length(database_point) BETWEEN 1 AND 4096 AND octet_length(transaction_nonce) = 32),
  CONSTRAINT ncv7_si_ck06 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_staging_import_capability_recovery_applications (
  recovery_application_id uuid NOT NULL,
  provider_staging_import_capability_recovery_digest bytea NOT NULL,
  staging_import_capability_recovery_intent_digest bytea NOT NULL,
  recovery_id uuid NOT NULL,
  capability_id uuid NOT NULL,
  single_use_apply_id uuid NOT NULL,
  staging_import_capability_digest bytea NOT NULL,
  staging_exclusion_id uuid NOT NULL,
  staging_exclusion_acquire_request_digest bytea NOT NULL,
  staging_exclusion_digest bytea NOT NULL,
  current_held_provider_head_digest bytea NOT NULL,
  observed_capability_row_count bigint NOT NULL,
  materialized_capability_row_count bigint NOT NULL,
  capability_materialization_kind text COLLATE "C" NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  current_database_incarnation_registration_digest bytea NOT NULL,
  latest_runtime_rebind_result_digest_or_null bytea,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  recovery_scope text COLLATE "C" NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  database_point bytea NOT NULL,
  recovered_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_sa_pk PRIMARY KEY (recovery_application_id),
  CONSTRAINT ncv7_sa_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_sa_uq01 UNIQUE (recovery_id),
  CONSTRAINT ncv7_sa_uq02 UNIQUE (capability_id),
  CONSTRAINT ncv7_sa_ck01 CHECK (observed_capability_row_count IN (0,1) AND materialized_capability_row_count = 1),
  CONSTRAINT ncv7_sa_ck02 CHECK (capability_materialization_kind IN ('inserted_from_provider_bundle','exact_existing_reused')),
  CONSTRAINT ncv7_sa_ck03 CHECK (recovery_scope = 'rehydrate_for_revocation_only'),
  CONSTRAINT ncv7_sa_ck04 CHECK (octet_length(database_point) BETWEEN 1 AND 4096),
  CONSTRAINT ncv7_sa_ck05 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE TABLE nodecontrol.control_plane_authority_fresh_restore_import_applications (
  single_use_apply_id uuid NOT NULL,
  staging_import_capability_digest bytea NOT NULL,
  staging_import_capability_recovery_intent_digest_or_null bytea,
  staging_import_capability_recovery_application_digest_or_null bytea,
  manifest_digest bytea NOT NULL,
  target_activation_id uuid NOT NULL,
  current_database_identity_digest bytea NOT NULL,
  database_timeline_lineage_chain_digest bytea NOT NULL,
  target_database_incarnation_registration_digest bytea NOT NULL,
  runtime_rebind_chain_digest bytea NOT NULL,
  runtime_instance_binding_digest bytea NOT NULL,
  staging_exclusion_lease_digest bytea NOT NULL,
  acquisition_locked_provider_head_digest bytea NOT NULL,
  database_route_closed_digest bytea NOT NULL,
  pre_import_inventory_digest bytea NOT NULL,
  post_import_inventory_digest bytea NOT NULL,
  imported_object_count bigint NOT NULL,
  complete_node_set_digest bytea NOT NULL,
  forbidden_state_zero_digest bytea NOT NULL,
  database_transaction_id numeric(20,0) NOT NULL,
  transaction_snapshot_digest bytea NOT NULL,
  applied_at timestamp with time zone NOT NULL,
  canonical_evidence_bundle_jcs bytea NOT NULL,
  canonical_body_jcs bytea NOT NULL,
  body_digest bytea NOT NULL,
  CONSTRAINT ncv7_fi_pk PRIMARY KEY (single_use_apply_id),
  CONSTRAINT ncv7_fi_body_uq UNIQUE (body_digest),
  CONSTRAINT ncv7_fi_uq01 UNIQUE (staging_import_capability_digest),
  CONSTRAINT ncv7_fi_ck01 CHECK (staging_import_capability_recovery_intent_digest_or_null IS NULL AND staging_import_capability_recovery_application_digest_or_null IS NULL),
  CONSTRAINT ncv7_fi_ck02 CHECK (imported_object_count >= 0),
  CONSTRAINT ncv7_fi_ck03 CHECK (octet_length(canonical_body_jcs) BETWEEN 1 AND 1048576 AND octet_length(canonical_evidence_bundle_jcs) BETWEEN 1 AND 1048576 AND octet_length(body_digest) = 32)
);

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_source_is_frozen()
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
  SELECT EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_legacy_database_source_retirements)
      OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_indeterminate_source_seals)
      OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_legacy_source_seals)
      OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_fresh_restore_requirements)
$fn$;

-- talenro:statement
ALTER TABLE nodecontrol.control_plane_authority_fences
  DROP CONSTRAINT control_plane_authority_fences_abort_reason_pair,
  ADD COLUMN authority_protocol_profile text COLLATE "C" NOT NULL DEFAULT 'legacy_v6',
  ADD COLUMN abort_claimed_at timestamp with time zone,
  ADD COLUMN protocol_activation_id uuid;
ALTER TABLE nodecontrol.control_plane_authority_fences
  ALTER COLUMN authority_protocol_profile DROP DEFAULT,
  ADD CONSTRAINT ncv7_fence_profile_ck CHECK (authority_protocol_profile IN ('legacy_v6','claim_v1')),
  ADD CONSTRAINT ncv7_fence_claim_ck CHECK (
    (authority_protocol_profile = 'legacy_v6' AND abort_claimed_at IS NULL AND protocol_activation_id IS NULL AND ((provider_status = 'aborted') = (abort_reason IS NOT NULL)))
    OR
    (authority_protocol_profile = 'claim_v1'
      AND protocol_activation_id IS NOT NULL
      AND ((abort_claimed_at IS NULL AND abort_reason IS NULL) OR (abort_claimed_at IS NOT NULL AND abort_reason IS NOT NULL))
      AND (provider_status <> 'aborted' OR abort_claimed_at IS NOT NULL)
      AND (abort_claimed_at IS NULL OR effect_digest IS NULL)
      AND (abort_claimed_at IS NULL OR abort_claimed_at >= reserved_at)
      AND (terminal_at IS NULL OR abort_claimed_at IS NULL OR terminal_at >= abort_claimed_at))
  ),
  ADD CONSTRAINT ncv7_fence_activation_fk FOREIGN KEY (protocol_activation_id) REFERENCES nodecontrol.control_plane_authority_protocol_activations(activation_id) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_fence_activation_fk_ix ON nodecontrol.control_plane_authority_fences (protocol_activation_id) WHERE protocol_activation_id IS NOT NULL;

-- talenro:statement
CREATE OR REPLACE FUNCTION nodecontrol.enforce_authority_fence_update()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  owner_count integer := 0;
  commitment_owner_count integer := 0;
  terminal_owner_count integer := 0;
  domain_owner_count integer := 0;
  owner_commitment_digest bytea;
BEGIN
  IF pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'authority fence mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  PERFORM pg_catalog.pg_advisory_xact_lock_shared(
    pg_catalog.hashtextextended('nodecontrol:v7-source-freeze',0)
  );
  IF nodecontrol.v7_source_is_frozen() THEN
    RAISE EXCEPTION 'authority fence is frozen after source seal' USING ERRCODE = '23514';
  END IF;

  IF TG_OP = 'DELETE' THEN
    RAISE EXCEPTION 'authority fence rows are append-only under v7' USING ERRCODE = '23514';
  END IF;

  IF TG_OP = 'INSERT' THEN
    IF NEW.authority_protocol_profile = 'legacy_v6' THEN
      RAISE EXCEPTION 'new legacy_v6 authority fences are forbidden after v7 activation' USING ERRCODE = '23514';
    END IF;
    IF NEW.authority_protocol_profile <> 'claim_v1'
       OR NEW.effect_kind NOT IN (
         'grant_create','grant_claim','certificate_activate','certificate_revoke',
         'identity_epoch_advance','operator_transition','security_incident_open',
         'security_incident_resolve','resource_envelope_activate','desired_activate',
         'recovery_activate','root_publish','metadata_publish'
       )
       OR NEW.provider_status <> 'reserved'
       OR NEW.visibility_state <> 'fence_pending'
       OR num_nonnulls(NEW.effect_digest,NEW.provider_receipt_digest,NEW.db_system_id,
                       NEW.db_timeline,NEW.required_lsn,NEW.abort_reason,NEW.abort_claimed_at,
                       NEW.effect_bound_at,NEW.terminal_at) <> 0
       OR NEW.protocol_activation_id IS NULL THEN
      RAISE EXCEPTION 'new claim-v1 authority fence must begin as an exact unbound reservation' USING ERRCODE = '23514';
    END IF;
    IF NEW.scope_kind = 'global_node_trust'
       AND NEW.scope_digest <> decode('f9911591db5e19d9c4eff72a30b4fd0325416e63d8fb00f8fbd81f0860f6b0a1','hex') THEN
      RAISE EXCEPTION 'authority fence scope digest mismatch' USING ERRCODE = '23514';
    END IF;
  ELSE
    IF NEW IS NOT DISTINCT FROM OLD THEN
      RETURN NEW;
    END IF;
    IF NEW.authority_protocol_profile = 'legacy_v6' OR OLD.authority_protocol_profile = 'legacy_v6' THEN
      RAISE EXCEPTION 'legacy_v6 authority fences are read-only after v7 activation' USING ERRCODE = '23514';
    END IF;
    IF NEW.operation_id IS DISTINCT FROM OLD.operation_id
       OR NEW.effect_kind IS DISTINCT FROM OLD.effect_kind
       OR NEW.scope_kind IS DISTINCT FROM OLD.scope_kind
       OR NEW.scope_digest IS DISTINCT FROM OLD.scope_digest
       OR NEW.provider_reservation_digest IS DISTINCT FROM OLD.provider_reservation_digest
       OR NEW.authority_epoch IS DISTINCT FROM OLD.authority_epoch
       OR NEW.authority_sequence IS DISTINCT FROM OLD.authority_sequence
       OR NEW.reserved_at IS DISTINCT FROM OLD.reserved_at
       OR NEW.authority_protocol_profile IS DISTINCT FROM OLD.authority_protocol_profile
       OR NEW.protocol_activation_id IS DISTINCT FROM OLD.protocol_activation_id THEN
      RAISE EXCEPTION 'authority fence reservation identity is immutable' USING ERRCODE = '23514';
    END IF;
    IF OLD.provider_status <> 'reserved' OR OLD.terminal_at IS NOT NULL THEN
      RAISE EXCEPTION 'terminal authority fence rows are immutable' USING ERRCODE = '23514';
    END IF;
  END IF;

  IF NEW.authority_protocol_profile <> 'claim_v1' THEN
    RAISE EXCEPTION 'authority fence has an unknown protocol profile' USING ERRCODE = '23514';
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM nodecontrol.control_plane_authority_protocol_activations AS activation
    JOIN nodecontrol.control_plane_authority_protocol_upgrade_attempts AS attempt
      ON attempt.body_digest = activation.attempt_digest
     AND attempt.activation_id = activation.activation_id
    JOIN nodecontrol.control_plane_authority_protocol_upgrade_intents AS intent
      ON intent.body_digest = attempt.upgrade_intent_digest
     AND intent.activation_id = activation.activation_id
    JOIN nodecontrol.control_plane_authority_runtime_registration_results AS registration
      ON registration.body_digest = activation.runtime_registration_result_digest
     AND registration.activation_id = activation.activation_id
    LEFT JOIN nodecontrol.control_plane_authority_runtime_rebind_results AS rebind
      ON rebind.body_digest = attempt.latest_runtime_rebind_result_digest_or_null
    JOIN nodecontrol.control_plane_authority_protocol_activation_completions AS completion
      ON completion.activation_id = activation.activation_id
     AND completion.completion_id = attempt.completion_id
    JOIN nodecontrol.control_plane_authority_protocol_activation_releases AS release
      ON release.activation_id = activation.activation_id
     AND release.release_preparation_id = attempt.release_preparation_id
     AND release.open_id = attempt.open_id
    WHERE activation.activation_id = NEW.protocol_activation_id
      AND activation.mode = 'empty_in_place'
      AND attempt.mode = activation.mode
      AND activation.protocol_profile = 'claim_v1'
      AND activation.deployment_id = attempt.deployment_id
      AND activation.database_identity_digest = intent.database_identity_digest
      AND activation.database_incarnation_attestation_digest = attempt.database_incarnation_attestation_digest
      AND activation.genesis_database_incarnation_registration_digest = attempt.database_incarnation_registration_digest
      AND activation.runtime_registration_result_digest = attempt.runtime_registration_result_digest
      AND activation.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM attempt.latest_runtime_rebind_result_digest_or_null
      AND activation.runtime_rebind_chain_digest = attempt.runtime_rebind_chain_digest
      AND activation.activation_runtime_instance_binding_digest = attempt.runtime_instance_binding_digest
      AND activation.database_legacy_absence_projection_digest = attempt.database_legacy_absence_projection_digest
      AND activation.attempt_database_inventory_digest = attempt.database_inventory_digest
      AND activation.provider_namespace_absence_digest = attempt.provider_namespace_absence_digest
      AND activation.environment_inventory_digest = attempt.environment_inventory_digest
      AND activation.environment_inventory_anchor_set_digest = attempt.environment_inventory_anchor_set_digest
      AND activation.local_runtime_isolation_digest = attempt.local_runtime_isolation_digest
      AND activation.legacy_runtime_shutdown_digest = attempt.legacy_runtime_shutdown_digest
      AND activation.legacy_runtime_shutdown_set_digest = attempt.legacy_runtime_shutdown_set_digest
      AND activation.credential_policy_digest = attempt.credential_policy_digest
      AND activation.epoch_evidence_digest = attempt.epoch_evidence_digest
      AND activation.genesis_epoch_transition_root_digest = attempt.genesis_epoch_transition_root_digest
      AND activation.selected_genesis_epoch = attempt.selected_genesis_epoch
      AND activation.provider_identity_digest = registration.provider_identity_digest
      AND activation.provider_endpoint_identity_digest = registration.provider_endpoint_identity_digest
      AND activation.namespace = registration.namespace
      AND registration.upgrade_intent_digest = intent.body_digest
      AND registration.credential_policy_digest = attempt.credential_policy_digest
      AND registration.database_incarnation_attestation_digest = attempt.database_incarnation_attestation_digest
      AND registration.genesis_database_identity_digest = intent.database_identity_digest
      AND attempt.request_nonce = intent.request_nonce
      AND attempt.deployment_id = intent.observed_deployment_id
      AND attempt.local_runtime_isolation_digest = intent.local_runtime_isolation_digest
      AND intent.created_at <= registration.recorded_at
      AND registration.recorded_at <= attempt.created_at
      AND attempt.created_at <= activation.activated_at
      AND (
        (attempt.latest_runtime_rebind_result_digest_or_null IS NULL
          AND rebind.body_digest IS NULL
          AND attempt.runtime_rebind_chain_digest = registration.runtime_rebind_chain_digest
          AND attempt.runtime_instance_binding_digest = registration.runtime_instance_binding_digest)
        OR
        (attempt.latest_runtime_rebind_result_digest_or_null IS NOT NULL
          AND rebind.body_digest = attempt.latest_runtime_rebind_result_digest_or_null
          AND rebind.activation_id = activation.activation_id
          AND rebind.runtime_registration_result_digest = registration.body_digest
          AND rebind.current_database_identity_digest = activation.database_identity_digest
          AND rebind.current_database_incarnation_registration_digest = attempt.database_incarnation_registration_digest
          AND rebind.current_runtime_rebind_chain_digest = attempt.runtime_rebind_chain_digest
          AND rebind.current_runtime_instance_binding_digest = attempt.runtime_instance_binding_digest
          AND registration.recorded_at <= rebind.recorded_at
          AND rebind.recorded_at <= attempt.created_at)
      )
      AND completion.activation_digest = activation.body_digest
      AND completion.preparation_digest = activation.preparation_digest
      AND completion.provider_completion_phase = 'genesis_completed_pending_release'
      AND completion.current_database_incarnation_registration_digest = activation.genesis_database_incarnation_registration_digest
      AND completion.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM activation.latest_runtime_rebind_result_digest_or_null
      AND completion.runtime_rebind_chain_digest = activation.runtime_rebind_chain_digest
      AND completion.current_runtime_instance_binding_digest = activation.activation_runtime_instance_binding_digest
      AND completion.credential_policy_digest = activation.credential_policy_digest
      AND completion.epoch_evidence_digest = activation.epoch_evidence_digest
      AND completion.genesis_epoch_transition_root_digest = activation.genesis_epoch_transition_root_digest
      AND completion.selected_genesis_epoch = activation.selected_genesis_epoch
      AND release.activation_digest = activation.body_digest
      AND release.completion_digest = completion.body_digest
      AND release.provider_release_phase = 'genesis_release_prepared'
      AND release.current_database_incarnation_registration_digest = completion.current_database_incarnation_registration_digest
      AND release.latest_runtime_rebind_result_digest_or_null IS NOT DISTINCT FROM completion.latest_runtime_rebind_result_digest_or_null
      AND release.runtime_rebind_chain_digest = completion.runtime_rebind_chain_digest
      AND release.current_runtime_instance_binding_digest = completion.current_runtime_instance_binding_digest
      AND release.credential_policy_digest = completion.credential_policy_digest
      AND release.epoch_evidence_digest = completion.epoch_evidence_digest
      AND release.genesis_epoch_transition_root_digest = completion.genesis_epoch_transition_root_digest
      AND release.selected_genesis_epoch = completion.selected_genesis_epoch
      AND completion.completed_at >= activation.activated_at
      AND release.released_at >= completion.completed_at
  ) THEN
    RAISE EXCEPTION 'claim-v1 authority fence lacks its exact activation completion release barrier' USING ERRCODE = '23514';
  END IF;

  IF TG_OP = 'INSERT' THEN
    RETURN NEW;
  END IF;

  WITH owners AS (
    SELECT authority_operation_id AS operation_id, create_authority_effect_commitment_digest AS commitment_digest,
           create_authority_provider_head_digest IS NOT NULL AS terminal
      FROM nodecontrol.node_enrollment_grants
    UNION ALL SELECT claim_authority_operation_id, claim_authority_effect_commitment_digest,
           claim_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_enrollment_grants
    UNION ALL SELECT authority_operation_id, activation_authority_effect_commitment_digest,
           activation_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_certificate_issuances
    UNION ALL SELECT revoke_authority_operation_id, revoke_authority_effect_commitment_digest,
           revoke_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_certificates
    UNION ALL SELECT authority_operation_id, authority_effect_commitment_digest,
           authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_state_transitions
    UNION ALL SELECT authority_operation_id, open_authority_effect_commitment_digest,
           open_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_security_incidents
    UNION ALL SELECT resolution_authority_operation_id, resolve_authority_effect_commitment_digest,
           resolve_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_security_incidents
    UNION ALL SELECT authority_operation_id, activation_authority_effect_commitment_digest,
           activation_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_resource_envelopes
    UNION ALL SELECT authority_operation_id, activation_authority_effect_commitment_digest,
           activation_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_state_signing_intents
    UNION ALL SELECT authority_operation_id, activation_authority_effect_commitment_digest,
           activation_authority_provider_head_digest IS NOT NULL FROM nodecontrol.node_root_metadata_publish_intents
  )
  SELECT count(*),
         count(*) FILTER (WHERE commitment_digest IS NOT NULL AND NOT terminal),
         count(*) FILTER (WHERE commitment_digest IS NOT NULL AND terminal),
         count(*) FILTER (WHERE commitment_digest IS NULL),
         (array_agg(commitment_digest) FILTER (WHERE commitment_digest IS NOT NULL))[1]
    INTO owner_count,commitment_owner_count,terminal_owner_count,domain_owner_count,owner_commitment_digest
  FROM owners
  WHERE operation_id = NEW.operation_id;

  IF OLD.effect_digest IS NULL AND NEW.effect_digest IS NOT NULL AND NEW.provider_status <> 'reserved' THEN
    RAISE EXCEPTION 'authority fence Bind and Finalize cannot occur in one proof transition' USING ERRCODE = '23514';
  END IF;

  IF OLD.abort_reason IS NULL AND NEW.abort_reason IS NOT NULL THEN
    IF owner_count <> 0 THEN
      RAISE EXCEPTION 'authority fence claim or abort requires no proof owner' USING ERRCODE = '23514';
    END IF;
    IF NEW.abort_claimed_at IS NULL OR NEW.abort_claimed_at < NEW.reserved_at
       OR (to_jsonb(NEW) - ARRAY['abort_reason','abort_claimed_at'])
          IS DISTINCT FROM (to_jsonb(OLD) - ARRAY['abort_reason','abort_claimed_at']) THEN
      RAISE EXCEPTION 'authority fence abort claim has an invalid transition' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;

  IF OLD.effect_digest IS NULL AND NEW.effect_digest IS NOT NULL THEN
    IF OLD.abort_reason IS NOT NULL OR NEW.abort_reason IS NOT NULL
       OR owner_count <> 1 OR commitment_owner_count <> 1 OR terminal_owner_count <> 0 THEN
      RAISE EXCEPTION 'authority fence Bind requires exactly one commitment-only owner' USING ERRCODE = '23514';
    END IF;
    IF NEW.effect_digest IS DISTINCT FROM owner_commitment_digest THEN
      RAISE EXCEPTION 'authority fence effect digest does not match owner commitment' USING ERRCODE = '23514';
    END IF;
    IF (to_jsonb(NEW) - ARRAY['effect_digest','db_system_id','db_timeline','required_lsn','effect_bound_at'])
       IS DISTINCT FROM (to_jsonb(OLD) - ARRAY['effect_digest','db_system_id','db_timeline','required_lsn','effect_bound_at']) THEN
      RAISE EXCEPTION 'authority fence Bind changed fields outside the database binding' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;

  IF OLD.effect_digest IS NOT NULL AND NEW.provider_status = 'committed' THEN
    IF owner_count <> 1 OR commitment_owner_count <> 1 OR terminal_owner_count <> 0
       OR NEW.effect_digest IS DISTINCT FROM owner_commitment_digest THEN
      RAISE EXCEPTION 'authority fence Finalize requires its exact commitment-only owner' USING ERRCODE = '23514';
    END IF;
    IF NEW.visibility_state <> 'active' OR NEW.provider_receipt_digest IS NULL OR NEW.terminal_at IS NULL
       OR (to_jsonb(NEW) - ARRAY['provider_status','provider_receipt_digest','visibility_state','terminal_at'])
          IS DISTINCT FROM (to_jsonb(OLD) - ARRAY['provider_status','provider_receipt_digest','visibility_state','terminal_at']) THEN
      RAISE EXCEPTION 'authority fence Finalize changed fields outside its receipt transition' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;

  IF OLD.abort_reason IS NOT NULL AND NEW.provider_status = 'aborted' THEN
    IF owner_count <> 0 THEN
      RAISE EXCEPTION 'authority fence claim or abort requires no proof owner' USING ERRCODE = '23514';
    END IF;
    IF NEW.abort_reason IS DISTINCT FROM OLD.abort_reason
       OR NEW.abort_claimed_at IS DISTINCT FROM OLD.abort_claimed_at
       OR NEW.visibility_state <> 'aborted' OR NEW.provider_receipt_digest IS NULL OR NEW.terminal_at IS NULL
       OR NEW.terminal_at < NEW.abort_claimed_at
       OR (to_jsonb(NEW) - ARRAY['provider_status','provider_receipt_digest','visibility_state','terminal_at'])
          IS DISTINCT FROM (to_jsonb(OLD) - ARRAY['provider_status','provider_receipt_digest','visibility_state','terminal_at']) THEN
      RAISE EXCEPTION 'authority fence Abort changed fields outside its claimed receipt transition' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;

  RAISE EXCEPTION 'illegal authority fence v7 lifecycle transition' USING ERRCODE = '23514';
END
$fn$;

-- talenro:statement
ALTER TABLE nodecontrol.node_inventory
  DROP CONSTRAINT node_inventory_identity_state_enum,
  DROP CONSTRAINT node_inventory_identity_lineage_state,
  DROP CONSTRAINT node_inventory_quarantine_state,
  ADD CONSTRAINT ncv7_inventory_identity_state_ck CHECK (identity_state IN ('never_enrolled','active','recovery_pending','recovery_limited','revoked','unauthorized')),
  ADD CONSTRAINT ncv7_inventory_identity_lineage_ck CHECK (
    (identity_state = 'never_enrolled' AND identity_epoch = 0 AND lineage_id IS NULL)
    OR (identity_state = 'unauthorized' AND identity_epoch = 0 AND lineage_id IS NULL)
    OR (identity_state NOT IN ('never_enrolled','unauthorized') AND identity_epoch > 0 AND lineage_id IS NOT NULL)
  ),
  ADD CONSTRAINT ncv7_inventory_quarantine_ck CHECK (
    (identity_state = 'unauthorized' AND operator_state = 'disabled' AND security_state = 'quarantined' AND identity_epoch = 0 AND lineage_id IS NULL AND resume_operator_state IS NULL AND pending_operator_transition IS NULL AND pending_transition_signing_id IS NULL AND active_desired_generation IS NULL AND active_recovery_generation IS NULL AND active_root_publish_id IS NULL AND active_root_version IS NULL AND active_metadata_publish_id IS NULL AND active_metadata_version IS NULL AND last_authority_operation_id IS NULL AND last_authority_epoch IS NULL AND last_authority_sequence IS NULL AND next_desired_generation = 1 AND next_recovery_generation = 1)
    OR
    (identity_state <> 'unauthorized' AND ((security_state = 'normal' AND resume_operator_state IS NULL) OR (security_state = 'quarantined' AND operator_state = 'disabled' AND resume_operator_state IS NOT NULL)))
  );

-- talenro:statement
DO $source_guards$
DECLARE
  relation_name text;
BEGIN
  FOREACH relation_name IN ARRAY ARRAY[
    'node_pops','node_failure_domains','node_capacity_profiles','node_inventory',
    'node_failure_domain_membership','node_endpoints','node_process_slots',
    'node_resource_envelopes','node_certificate_issuances','node_enrollment_grants',
    'node_certificates','node_security_incidents','node_security_fault_receipts',
    'node_recovery_sessions','node_restore_reauthorization_approvals',
    'node_state_signing_intents','node_root_metadata_publish_intents',
    'node_root_metadata_signature_shares','node_desired_states','node_recovery_states',
    'node_observed_states','node_operator_audit','node_state_transitions',
    'control_plane_trust_bundle_high_waters'
  ] LOOP
    EXECUTE format(
      'CREATE TRIGGER ncv7_00_source_lock BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.%I FOR EACH STATEMENT EXECUTE FUNCTION nodecontrol.v7_assert_source_writable()',
      relation_name
    );
    EXECUTE format(
      'CREATE TRIGGER ncv7_01_source_guard BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.%I FOR EACH ROW EXECUTE FUNCTION nodecontrol.v7_assert_source_writable()',
      relation_name
    );
  END LOOP;
END
$source_guards$;

-- talenro:statement
CREATE TRIGGER ncv7_00_fence_source_lock
BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.control_plane_authority_fences
FOR EACH STATEMENT EXECUTE FUNCTION nodecontrol.v7_assert_source_writable();

-- talenro:statement
ALTER TABLE nodecontrol.node_enrollment_grants
  ADD COLUMN create_authority_effect_commitment_jcs bytea,
  ADD COLUMN create_authority_effect_commitment_digest bytea,
  ADD COLUMN create_authority_provider_head_jcs bytea,
  ADD COLUMN create_authority_provider_head_digest bytea,
  ADD COLUMN create_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN create_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN create_authority_effect_reason text COLLATE "C",
  ADD COLUMN create_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN create_authority_activation_deadline timestamp with time zone,
  ADD COLUMN create_authority_expected_provider_identity_digest bytea,
  ADD COLUMN create_authority_activation_evidence_jcs bytea,
  ADD COLUMN create_authority_activation_evidence_digest bytea,
  ADD COLUMN create_authority_effect_resolution_jcs bytea,
  ADD COLUMN create_authority_effect_resolution_digest bytea,
  ADD COLUMN claim_authority_effect_commitment_jcs bytea,
  ADD COLUMN claim_authority_effect_commitment_digest bytea,
  ADD COLUMN claim_authority_provider_head_jcs bytea,
  ADD COLUMN claim_authority_provider_head_digest bytea,
  ADD COLUMN claim_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN claim_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN claim_authority_effect_reason text COLLATE "C",
  ADD COLUMN claim_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN claim_authority_activation_deadline timestamp with time zone,
  ADD COLUMN claim_authority_expected_provider_identity_digest bytea,
  ADD COLUMN claim_authority_activation_evidence_jcs bytea,
  ADD COLUMN claim_authority_activation_evidence_digest bytea,
  ADD COLUMN claim_authority_effect_resolution_jcs bytea,
  ADD COLUMN claim_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_grant_create_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(create_authority_effect_commitment_jcs,create_authority_effect_commitment_digest,create_authority_provider_head_jcs,create_authority_provider_head_digest,create_authority_checkpoint_anchor_jcs,create_authority_checkpoint_anchor_digest,create_authority_effect_reason,create_authority_attestation_expires_at,create_authority_activation_deadline,create_authority_expected_provider_identity_digest,create_authority_activation_evidence_jcs,create_authority_activation_evidence_digest,create_authority_effect_resolution_jcs,create_authority_effect_resolution_digest)),
  ADD CONSTRAINT ncv7_grant_claim_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(claim_authority_effect_commitment_jcs,claim_authority_effect_commitment_digest,claim_authority_provider_head_jcs,claim_authority_provider_head_digest,claim_authority_checkpoint_anchor_jcs,claim_authority_checkpoint_anchor_digest,claim_authority_effect_reason,claim_authority_attestation_expires_at,claim_authority_activation_deadline,claim_authority_expected_provider_identity_digest,claim_authority_activation_evidence_jcs,claim_authority_activation_evidence_digest,claim_authority_effect_resolution_jcs,claim_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_certificate_issuances
  ADD COLUMN activation_authority_effect_commitment_jcs bytea,
  ADD COLUMN activation_authority_effect_commitment_digest bytea,
  ADD COLUMN activation_authority_provider_head_jcs bytea,
  ADD COLUMN activation_authority_provider_head_digest bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN activation_authority_effect_reason text COLLATE "C",
  ADD COLUMN activation_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN activation_authority_activation_deadline timestamp with time zone,
  ADD COLUMN activation_authority_expected_provider_identity_digest bytea,
  ADD COLUMN activation_authority_activation_evidence_jcs bytea,
  ADD COLUMN activation_authority_activation_evidence_digest bytea,
  ADD COLUMN activation_authority_effect_resolution_jcs bytea,
  ADD COLUMN activation_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_issuance_activation_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(activation_authority_effect_commitment_jcs,activation_authority_effect_commitment_digest,activation_authority_provider_head_jcs,activation_authority_provider_head_digest,activation_authority_checkpoint_anchor_jcs,activation_authority_checkpoint_anchor_digest,activation_authority_effect_reason,activation_authority_attestation_expires_at,activation_authority_activation_deadline,activation_authority_expected_provider_identity_digest,activation_authority_activation_evidence_jcs,activation_authority_activation_evidence_digest,activation_authority_effect_resolution_jcs,activation_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_certificates
  ADD COLUMN revoke_authority_effect_commitment_jcs bytea,
  ADD COLUMN revoke_authority_effect_commitment_digest bytea,
  ADD COLUMN revoke_authority_provider_head_jcs bytea,
  ADD COLUMN revoke_authority_provider_head_digest bytea,
  ADD COLUMN revoke_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN revoke_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN revoke_authority_effect_reason text COLLATE "C",
  ADD COLUMN revoke_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN revoke_authority_activation_deadline timestamp with time zone,
  ADD COLUMN revoke_authority_expected_provider_identity_digest bytea,
  ADD COLUMN revoke_authority_activation_evidence_jcs bytea,
  ADD COLUMN revoke_authority_activation_evidence_digest bytea,
  ADD COLUMN revoke_authority_effect_resolution_jcs bytea,
  ADD COLUMN revoke_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_certificate_revoke_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(revoke_authority_effect_commitment_jcs,revoke_authority_effect_commitment_digest,revoke_authority_provider_head_jcs,revoke_authority_provider_head_digest,revoke_authority_checkpoint_anchor_jcs,revoke_authority_checkpoint_anchor_digest,revoke_authority_effect_reason,revoke_authority_attestation_expires_at,revoke_authority_activation_deadline,revoke_authority_expected_provider_identity_digest,revoke_authority_activation_evidence_jcs,revoke_authority_activation_evidence_digest,revoke_authority_effect_resolution_jcs,revoke_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_state_transitions
  ADD COLUMN authority_effect_kind text COLLATE "C",
  ADD COLUMN authority_effect_disposition text COLLATE "C",
  ADD COLUMN authority_effect_commitment_jcs bytea,
  ADD COLUMN authority_effect_commitment_digest bytea,
  ADD COLUMN authority_provider_head_jcs bytea,
  ADD COLUMN authority_provider_head_digest bytea,
  ADD COLUMN authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN authority_checkpoint_anchor_digest bytea,
  ADD COLUMN authority_effect_reason text COLLATE "C",
  ADD COLUMN authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN authority_activation_deadline timestamp with time zone,
  ADD COLUMN authority_expected_provider_identity_digest bytea,
  ADD COLUMN authority_activation_evidence_jcs bytea,
  ADD COLUMN authority_activation_evidence_digest bytea,
  ADD COLUMN authority_effect_resolution_jcs bytea,
  ADD COLUMN authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_transition_kind_ck CHECK (authority_effect_kind IS NULL OR authority_effect_kind IN ('identity_epoch_advance','operator_transition')),
  ADD CONSTRAINT ncv7_transition_disposition_ck CHECK (
    (authority_operation_id IS NULL AND authority_effect_kind IS NULL AND authority_effect_disposition IS NULL)
    OR
    (authority_operation_id IS NOT NULL AND authority_effect_kind IS NULL AND authority_effect_disposition IS NULL)
    OR
    (authority_operation_id IS NOT NULL AND authority_effect_kind IS NOT NULL
      AND (authority_effect_disposition IS NULL OR authority_effect_disposition IN ('applied','not_applied')))
  ),
  ADD CONSTRAINT ncv7_transition_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(authority_effect_commitment_jcs,authority_effect_commitment_digest,authority_provider_head_jcs,authority_provider_head_digest,authority_checkpoint_anchor_jcs,authority_checkpoint_anchor_digest,authority_effect_reason,authority_attestation_expires_at,authority_activation_deadline,authority_expected_provider_identity_digest,authority_activation_evidence_jcs,authority_activation_evidence_digest,authority_effect_resolution_jcs,authority_effect_resolution_digest)),
  ADD CONSTRAINT ncv7_transition_operation_uq UNIQUE (authority_operation_id);

-- talenro:statement
ALTER TABLE nodecontrol.node_security_incidents
  ADD COLUMN open_authority_effect_commitment_jcs bytea,
  ADD COLUMN open_authority_effect_commitment_digest bytea,
  ADD COLUMN open_authority_provider_head_jcs bytea,
  ADD COLUMN open_authority_provider_head_digest bytea,
  ADD COLUMN open_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN open_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN open_authority_effect_reason text COLLATE "C",
  ADD COLUMN open_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN open_authority_activation_deadline timestamp with time zone,
  ADD COLUMN open_authority_expected_provider_identity_digest bytea,
  ADD COLUMN open_authority_activation_evidence_jcs bytea,
  ADD COLUMN open_authority_activation_evidence_digest bytea,
  ADD COLUMN open_authority_effect_resolution_jcs bytea,
  ADD COLUMN open_authority_effect_resolution_digest bytea,
  ADD COLUMN resolve_authority_effect_commitment_jcs bytea,
  ADD COLUMN resolve_authority_effect_commitment_digest bytea,
  ADD COLUMN resolve_authority_provider_head_jcs bytea,
  ADD COLUMN resolve_authority_provider_head_digest bytea,
  ADD COLUMN resolve_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN resolve_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN resolve_authority_effect_reason text COLLATE "C",
  ADD COLUMN resolve_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN resolve_authority_activation_deadline timestamp with time zone,
  ADD COLUMN resolve_authority_expected_provider_identity_digest bytea,
  ADD COLUMN resolve_authority_activation_evidence_jcs bytea,
  ADD COLUMN resolve_authority_activation_evidence_digest bytea,
  ADD COLUMN resolve_authority_effect_resolution_jcs bytea,
  ADD COLUMN resolve_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_incident_open_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(open_authority_effect_commitment_jcs,open_authority_effect_commitment_digest,open_authority_provider_head_jcs,open_authority_provider_head_digest,open_authority_checkpoint_anchor_jcs,open_authority_checkpoint_anchor_digest,open_authority_effect_reason,open_authority_attestation_expires_at,open_authority_activation_deadline,open_authority_expected_provider_identity_digest,open_authority_activation_evidence_jcs,open_authority_activation_evidence_digest,open_authority_effect_resolution_jcs,open_authority_effect_resolution_digest)),
  ADD CONSTRAINT ncv7_incident_resolve_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(resolve_authority_effect_commitment_jcs,resolve_authority_effect_commitment_digest,resolve_authority_provider_head_jcs,resolve_authority_provider_head_digest,resolve_authority_checkpoint_anchor_jcs,resolve_authority_checkpoint_anchor_digest,resolve_authority_effect_reason,resolve_authority_attestation_expires_at,resolve_authority_activation_deadline,resolve_authority_expected_provider_identity_digest,resolve_authority_activation_evidence_jcs,resolve_authority_activation_evidence_digest,resolve_authority_effect_resolution_jcs,resolve_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_resource_envelopes
  ADD COLUMN activation_authority_effect_commitment_jcs bytea,
  ADD COLUMN activation_authority_effect_commitment_digest bytea,
  ADD COLUMN activation_authority_provider_head_jcs bytea,
  ADD COLUMN activation_authority_provider_head_digest bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN activation_authority_effect_reason text COLLATE "C",
  ADD COLUMN activation_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN activation_authority_activation_deadline timestamp with time zone,
  ADD COLUMN activation_authority_expected_provider_identity_digest bytea,
  ADD COLUMN activation_authority_activation_evidence_jcs bytea,
  ADD COLUMN activation_authority_activation_evidence_digest bytea,
  ADD COLUMN activation_authority_effect_resolution_jcs bytea,
  ADD COLUMN activation_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_resource_activation_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(activation_authority_effect_commitment_jcs,activation_authority_effect_commitment_digest,activation_authority_provider_head_jcs,activation_authority_provider_head_digest,activation_authority_checkpoint_anchor_jcs,activation_authority_checkpoint_anchor_digest,activation_authority_effect_reason,activation_authority_attestation_expires_at,activation_authority_activation_deadline,activation_authority_expected_provider_identity_digest,activation_authority_activation_evidence_jcs,activation_authority_activation_evidence_digest,activation_authority_effect_resolution_jcs,activation_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_state_signing_intents
  ADD COLUMN activation_authority_effect_commitment_jcs bytea,
  ADD COLUMN activation_authority_effect_commitment_digest bytea,
  ADD COLUMN activation_authority_provider_head_jcs bytea,
  ADD COLUMN activation_authority_provider_head_digest bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN activation_authority_effect_reason text COLLATE "C",
  ADD COLUMN activation_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN activation_authority_activation_deadline timestamp with time zone,
  ADD COLUMN activation_authority_expected_provider_identity_digest bytea,
  ADD COLUMN activation_authority_activation_evidence_jcs bytea,
  ADD COLUMN activation_authority_activation_evidence_digest bytea,
  ADD COLUMN activation_authority_effect_resolution_jcs bytea,
  ADD COLUMN activation_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_signing_activation_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(activation_authority_effect_commitment_jcs,activation_authority_effect_commitment_digest,activation_authority_provider_head_jcs,activation_authority_provider_head_digest,activation_authority_checkpoint_anchor_jcs,activation_authority_checkpoint_anchor_digest,activation_authority_effect_reason,activation_authority_attestation_expires_at,activation_authority_activation_deadline,activation_authority_expected_provider_identity_digest,activation_authority_activation_evidence_jcs,activation_authority_activation_evidence_digest,activation_authority_effect_resolution_jcs,activation_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_root_metadata_publish_intents
  ADD COLUMN activation_authority_effect_commitment_jcs bytea,
  ADD COLUMN activation_authority_effect_commitment_digest bytea,
  ADD COLUMN activation_authority_provider_head_jcs bytea,
  ADD COLUMN activation_authority_provider_head_digest bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_jcs bytea,
  ADD COLUMN activation_authority_checkpoint_anchor_digest bytea,
  ADD COLUMN activation_authority_effect_reason text COLLATE "C",
  ADD COLUMN activation_authority_attestation_expires_at timestamp with time zone,
  ADD COLUMN activation_authority_activation_deadline timestamp with time zone,
  ADD COLUMN activation_authority_expected_provider_identity_digest bytea,
  ADD COLUMN activation_authority_activation_evidence_jcs bytea,
  ADD COLUMN activation_authority_activation_evidence_digest bytea,
  ADD COLUMN activation_authority_effect_resolution_jcs bytea,
  ADD COLUMN activation_authority_effect_resolution_digest bytea,
  ADD CONSTRAINT ncv7_publish_activation_proof_ck CHECK (nodecontrol.v7_authority_proof_group_valid(activation_authority_effect_commitment_jcs,activation_authority_effect_commitment_digest,activation_authority_provider_head_jcs,activation_authority_provider_head_digest,activation_authority_checkpoint_anchor_jcs,activation_authority_checkpoint_anchor_digest,activation_authority_effect_reason,activation_authority_attestation_expires_at,activation_authority_activation_deadline,activation_authority_expected_provider_identity_digest,activation_authority_activation_evidence_jcs,activation_authority_activation_evidence_digest,activation_authority_effect_resolution_jcs,activation_authority_effect_resolution_digest));

-- talenro:statement
ALTER TABLE nodecontrol.node_enrollment_grants
  DROP CONSTRAINT node_enrollment_grants_consumption_all_or_none,
  ADD CONSTRAINT ncv7_grant_claim_authority_tuple_ck CHECK (num_nonnulls(claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence) IN (0,3)),
  ADD CONSTRAINT ncv7_grant_consumption_result_ck CHECK (num_nonnulls(consumed_at,consumption_attempt_id,consumption_request_digest,result_issuance_id) IN (0,4)),
  ADD CONSTRAINT ncv7_grant_claim_lifecycle_ck CHECK (
    (num_nonnulls(claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence) = 0
      AND num_nonnulls(consumed_at,consumption_attempt_id,consumption_request_digest,result_issuance_id) = 0)
    OR
    (num_nonnulls(claim_authority_operation_id,claim_authority_epoch,claim_authority_sequence) = 3
      AND num_nonnulls(consumed_at,consumption_attempt_id,consumption_request_digest,result_issuance_id) IN (0,4))
  );

-- talenro:statement
ALTER TABLE nodecontrol.node_certificates
  DROP CONSTRAINT node_certificates_revoke_all_or_none,
  ADD CONSTRAINT ncv7_certificate_revoke_authority_tuple_ck CHECK (num_nonnulls(revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence) IN (0,3)),
  ADD CONSTRAINT ncv7_certificate_revoke_result_ck CHECK (num_nonnulls(revoked_at,revoke_reason) IN (0,2)),
  ADD CONSTRAINT ncv7_certificate_revoke_lifecycle_ck CHECK (
    (num_nonnulls(revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence) = 0
      AND num_nonnulls(revoked_at,revoke_reason) = 0)
    OR
    (num_nonnulls(revoke_authority_operation_id,revoke_authority_epoch,revoke_authority_sequence) = 3
      AND num_nonnulls(revoked_at,revoke_reason) IN (0,2))
  );

-- talenro:statement
ALTER TABLE nodecontrol.node_security_incidents
  DROP CONSTRAINT node_security_incidents_resolution_all_or_none,
  ADD CONSTRAINT ncv7_incident_resolution_authority_tuple_ck CHECK (num_nonnulls(resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence) IN (0,3)),
  ADD CONSTRAINT ncv7_incident_resolution_result_ck CHECK (num_nonnulls(remediation_digest,resolution_at) IN (0,2)),
  ADD CONSTRAINT ncv7_incident_resolution_lifecycle_ck CHECK (
    (num_nonnulls(resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence) = 0
      AND num_nonnulls(remediation_digest,resolution_at) = 0)
    OR
    (num_nonnulls(resolution_authority_operation_id,resolution_authority_epoch,resolution_authority_sequence) = 3
      AND num_nonnulls(remediation_digest,resolution_at) IN (0,2))
  );

-- talenro:statement
CREATE OR REPLACE FUNCTION nodecontrol.enforce_security_incident_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $fn$
DECLARE
  open_count integer;
  opening_fence nodecontrol.control_plane_authority_fences%ROWTYPE;
BEGIN
  IF TG_OP <> 'DELETE' AND pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'nodecontrol security incident mutation requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  IF TG_OP = 'INSERT' THEN
    PERFORM 1 FROM nodecontrol.node_inventory WHERE node_id = NEW.node_id FOR UPDATE;
    SELECT count(*) INTO open_count
    FROM nodecontrol.node_security_incidents
    WHERE node_id = NEW.node_id
      AND status IN ('open','resolution_pending_agent_ack','overflow');
    IF open_count >= 16 THEN
      RAISE EXCEPTION 'node security incident cap reached' USING ERRCODE = '23514';
    END IF;
    IF (NEW.fault_subtype = 'incident_overflow' AND NEW.status <> 'overflow')
       OR (NEW.fault_subtype <> 'incident_overflow' AND NEW.status <> 'open') THEN
      RAISE EXCEPTION 'security incident must start in its nonterminal opening status' USING ERRCODE = '23514';
    END IF;
    SELECT * INTO opening_fence
    FROM nodecontrol.control_plane_authority_fences
    WHERE (operation_id,authority_epoch,authority_sequence) =
          (NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence)
    FOR SHARE;
    IF NOT FOUND
       OR opening_fence.authority_protocol_profile <> 'claim_v1'
       OR opening_fence.provider_status <> 'reserved'
       OR opening_fence.visibility_state <> 'fence_pending'
       OR opening_fence.effect_digest IS NOT NULL
       OR opening_fence.abort_claimed_at IS NOT NULL
       OR opening_fence.abort_reason IS NOT NULL
       OR opening_fence.effect_kind <> 'security_incident_open'
       OR opening_fence.scope_kind <> 'node'
       OR opening_fence.scope_digest <> pg_catalog.sha256(
            pg_catalog.convert_to('TALENRO-NODE-AUTHORITY-SCOPE-V1','UTF8')
            || decode('00','hex') || pg_catalog.uuid_send(NEW.node_id)
          )
       OR num_nonnulls(
            NEW.open_authority_effect_commitment_jcs,NEW.open_authority_effect_commitment_digest,
            NEW.open_authority_provider_head_jcs,NEW.open_authority_provider_head_digest,
            NEW.open_authority_checkpoint_anchor_jcs,NEW.open_authority_checkpoint_anchor_digest,
            NEW.open_authority_effect_reason,NEW.open_authority_attestation_expires_at,
            NEW.open_authority_activation_deadline,NEW.open_authority_expected_provider_identity_digest,
            NEW.open_authority_activation_evidence_jcs,NEW.open_authority_activation_evidence_digest,
            NEW.open_authority_effect_resolution_jcs,NEW.open_authority_effect_resolution_digest
          ) <> 2
       OR NEW.open_authority_effect_commitment_jcs IS NULL
       OR NEW.open_authority_effect_commitment_digest IS NULL THEN
      RAISE EXCEPTION 'security incident opening requires its exact reserved claim-v1 prepared proof' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.status <> 'resolved' OR OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'security incident retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  PERFORM 1 FROM nodecontrol.node_inventory WHERE node_id = NEW.node_id FOR UPDATE;
  IF NOT FOUND THEN
    RAISE EXCEPTION 'unknown node for security incident' USING ERRCODE = '23503';
  END IF;
  IF OLD.status = 'resolved' THEN
    RAISE EXCEPTION 'resolved security incident is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.incident_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,
      NEW.node_id,NEW.identity_epoch,NEW.fault_subtype,NEW.subtype_slot,NEW.first_evidence_digest,
      NEW.trust_context_digest,NEW.first_occurred_at)
     IS DISTINCT FROM
     (OLD.incident_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,
      OLD.node_id,OLD.identity_epoch,OLD.fault_subtype,OLD.subtype_slot,OLD.first_evidence_digest,
      OLD.trust_context_digest,OLD.first_occurred_at) THEN
    RAISE EXCEPTION 'security incident identity is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.occurrence_count < OLD.occurrence_count OR NEW.last_occurred_at < OLD.last_occurred_at THEN
    RAISE EXCEPTION 'security incident evidence cannot move backward' USING ERRCODE = '23514';
  END IF;
  IF OLD.resolution_at IS NOT NULL
     AND (NEW.resolution_authority_operation_id,NEW.resolution_authority_epoch,
          NEW.resolution_authority_sequence,NEW.remediation_digest,NEW.resolution_at)
         IS DISTINCT FROM
         (OLD.resolution_authority_operation_id,OLD.resolution_authority_epoch,
          OLD.resolution_authority_sequence,OLD.remediation_digest,OLD.resolution_at) THEN
    RAISE EXCEPTION 'security incident resolution binding is immutable' USING ERRCODE = '23514';
  END IF;
  IF NEW.resolution_at IS NOT NULL AND NOT EXISTS (
    SELECT 1 FROM nodecontrol.control_plane_authority_fences fence
    WHERE (fence.operation_id,fence.authority_epoch,fence.authority_sequence) =
          (NEW.resolution_authority_operation_id,NEW.resolution_authority_epoch,NEW.resolution_authority_sequence)
      AND fence.provider_status = 'committed'
      AND fence.visibility_state = 'active'
      AND fence.effect_kind = 'security_incident_resolve'
      AND fence.scope_kind = 'node'
  ) THEN
    RAISE EXCEPTION 'security incident resolution requires its committed node fence' USING ERRCODE = '23514';
  END IF;
  IF NEW.status = 'resolved'
     AND EXISTS (
       SELECT 1 FROM nodecontrol.node_security_fault_receipts
       WHERE incident_id = OLD.incident_id AND binding_status = 'active'
     ) THEN
    RAISE EXCEPTION 'active security-fault bindings prevent direct resolution' USING ERRCODE = '23514';
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND NOT (
    (OLD.status IN ('open','overflow') AND NEW.status IN ('resolution_pending_agent_ack','resolved')) OR
    (OLD.status = 'resolution_pending_agent_ack' AND NEW.status = 'resolved')
  ) THEN
    RAISE EXCEPTION 'illegal security incident status transition' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$fn$;

-- talenro:statement
CREATE OR REPLACE FUNCTION nodecontrol.enforce_enrollment_grant_workflow()
RETURNS trigger
LANGUAGE plpgsql
AS $fn$
BEGIN
  IF TG_OP = 'INSERT' THEN
    IF num_nonnulls(NEW.consumed_at,NEW.consumption_attempt_id,NEW.consumption_request_digest,
                    NEW.claim_authority_operation_id,NEW.claim_authority_epoch,NEW.claim_authority_sequence,
                    NEW.result_issuance_id,NEW.expired_at,NEW.invalidated_at,NEW.terminal_reason,
                    NEW.terminal_at,NEW.retention_until) <> 0 THEN
      RAISE EXCEPTION 'enrollment grants must start live and unterminated' USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_OP = 'DELETE' THEN
    IF OLD.retention_until IS NULL OR OLD.retention_until > transaction_timestamp() THEN
      RAISE EXCEPTION 'enrollment grant retention has not elapsed' USING ERRCODE = '23514';
    END IF;
    RETURN OLD;
  END IF;
  IF num_nonnulls(OLD.consumed_at,OLD.consumption_attempt_id,OLD.consumption_request_digest,
                  OLD.result_issuance_id,OLD.expired_at,OLD.invalidated_at,OLD.terminal_reason,
                  OLD.terminal_at,OLD.retention_until) > 0 THEN
    RAISE EXCEPTION 'terminal enrollment grant is immutable' USING ERRCODE = '23514';
  END IF;
  IF (NEW.grant_id,NEW.authority_operation_id,NEW.authority_epoch,NEW.authority_sequence,NEW.node_id,
      NEW.identity_epoch,NEW.token_digest,NEW.csr_digest,NEW.idempotency_digest,NEW.created_at,NEW.expires_at)
     IS DISTINCT FROM
     (OLD.grant_id,OLD.authority_operation_id,OLD.authority_epoch,OLD.authority_sequence,OLD.node_id,
      OLD.identity_epoch,OLD.token_digest,OLD.csr_digest,OLD.idempotency_digest,OLD.created_at,OLD.expires_at) THEN
    RAISE EXCEPTION 'enrollment grant inputs are immutable' USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END
$fn$;

-- talenro:statement
ALTER TABLE nodecontrol.control_plane_authority_runtime_registration_results
  ADD CONSTRAINT ncv7_rr_fk01 FOREIGN KEY (upgrade_intent_digest) REFERENCES nodecontrol.control_plane_authority_protocol_upgrade_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_rr_fk01_ix ON nodecontrol.control_plane_authority_runtime_registration_results (upgrade_intent_digest);
ALTER TABLE nodecontrol.control_plane_authority_runtime_rebind_results
  ADD CONSTRAINT ncv7_rb_fk01 FOREIGN KEY (runtime_registration_result_digest) REFERENCES nodecontrol.control_plane_authority_runtime_registration_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_rb_fk02 FOREIGN KEY (previous_runtime_rebind_result_digest_or_null) REFERENCES nodecontrol.control_plane_authority_runtime_rebind_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_rb_fk01_ix ON nodecontrol.control_plane_authority_runtime_rebind_results (runtime_registration_result_digest);
CREATE INDEX ncv7_rb_fk02_ix ON nodecontrol.control_plane_authority_runtime_rebind_results (previous_runtime_rebind_result_digest_or_null) WHERE previous_runtime_rebind_result_digest_or_null IS NOT NULL;
ALTER TABLE nodecontrol.control_plane_authority_protocol_upgrade_attempts
  ADD CONSTRAINT ncv7_ua_fk01 FOREIGN KEY (upgrade_intent_digest) REFERENCES nodecontrol.control_plane_authority_protocol_upgrade_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_ua_fk02 FOREIGN KEY (runtime_registration_result_digest) REFERENCES nodecontrol.control_plane_authority_runtime_registration_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_ua_fk03 FOREIGN KEY (latest_runtime_rebind_result_digest_or_null) REFERENCES nodecontrol.control_plane_authority_runtime_rebind_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_ua_fk01_ix ON nodecontrol.control_plane_authority_protocol_upgrade_attempts (upgrade_intent_digest);
CREATE INDEX ncv7_ua_fk02_ix ON nodecontrol.control_plane_authority_protocol_upgrade_attempts (runtime_registration_result_digest);
CREATE INDEX ncv7_ua_fk03_ix ON nodecontrol.control_plane_authority_protocol_upgrade_attempts (latest_runtime_rebind_result_digest_or_null) WHERE latest_runtime_rebind_result_digest_or_null IS NOT NULL;
ALTER TABLE nodecontrol.control_plane_authority_protocol_activations
  ADD CONSTRAINT ncv7_pa_fk01 FOREIGN KEY (attempt_digest) REFERENCES nodecontrol.control_plane_authority_protocol_upgrade_attempts(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_pa_fk02 FOREIGN KEY (runtime_registration_result_digest) REFERENCES nodecontrol.control_plane_authority_runtime_registration_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_pa_fk03 FOREIGN KEY (latest_runtime_rebind_result_digest_or_null) REFERENCES nodecontrol.control_plane_authority_runtime_rebind_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_pa_fk01_ix ON nodecontrol.control_plane_authority_protocol_activations (attempt_digest);
CREATE INDEX ncv7_pa_fk02_ix ON nodecontrol.control_plane_authority_protocol_activations (runtime_registration_result_digest);
CREATE INDEX ncv7_pa_fk03_ix ON nodecontrol.control_plane_authority_protocol_activations (latest_runtime_rebind_result_digest_or_null) WHERE latest_runtime_rebind_result_digest_or_null IS NOT NULL;
ALTER TABLE nodecontrol.control_plane_authority_protocol_activation_completions
  ADD CONSTRAINT ncv7_pc_fk01 FOREIGN KEY (activation_id) REFERENCES nodecontrol.control_plane_authority_protocol_activations(activation_id) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_pc_fk01_ix ON nodecontrol.control_plane_authority_protocol_activation_completions (activation_id);
ALTER TABLE nodecontrol.control_plane_authority_protocol_activation_releases
  ADD CONSTRAINT ncv7_pr_fk01 FOREIGN KEY (activation_id) REFERENCES nodecontrol.control_plane_authority_protocol_activations(activation_id) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_pr_fk02 FOREIGN KEY (completion_digest) REFERENCES nodecontrol.control_plane_authority_protocol_activation_completions(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_pr_fk01_ix ON nodecontrol.control_plane_authority_protocol_activation_releases (activation_id);
CREATE INDEX ncv7_pr_fk02_ix ON nodecontrol.control_plane_authority_protocol_activation_releases (completion_digest);
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_intents
  ADD CONSTRAINT ncv7_ei_fk01 FOREIGN KEY (activation_id) REFERENCES nodecontrol.control_plane_authority_protocol_activations(activation_id) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_ei_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_intents (activation_id);
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_applications
  ADD CONSTRAINT ncv7_ea_fk01 FOREIGN KEY (transition_intent_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_ea_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_applications (transition_intent_digest);
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_resolutions
  ADD CONSTRAINT ncv7_er_fk01 FOREIGN KEY (transition_intent_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_er_fk02 FOREIGN KEY (transition_application_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_applications(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_er_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_resolutions (transition_intent_digest);
CREATE INDEX ncv7_er_fk02_ix ON nodecontrol.control_plane_authority_epoch_transition_resolutions (transition_application_digest);
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_cancellations
  ADD CONSTRAINT ncv7_ec_fk01 FOREIGN KEY (transition_intent_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_ec_fk02 FOREIGN KEY (runtime_rebind_result_digest) REFERENCES nodecontrol.control_plane_authority_runtime_rebind_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_ec_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_cancellations (transition_intent_digest);
CREATE INDEX ncv7_ec_fk02_ix ON nodecontrol.control_plane_authority_epoch_transition_cancellations (runtime_rebind_result_digest);
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_terminal_applications
  ADD CONSTRAINT ncv7_et_fk01 FOREIGN KEY (transition_intent_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_et_fk02 FOREIGN KEY (epoch_transition_cancellation_digest_or_null) REFERENCES nodecontrol.control_plane_authority_epoch_transition_cancellations(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_et_fk03 FOREIGN KEY (runtime_rebind_result_digest_or_null) REFERENCES nodecontrol.control_plane_authority_runtime_rebind_results(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_et_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_terminal_applications (transition_intent_digest);
CREATE INDEX ncv7_et_fk02_ix ON nodecontrol.control_plane_authority_epoch_transition_terminal_applications (epoch_transition_cancellation_digest_or_null) WHERE epoch_transition_cancellation_digest_or_null IS NOT NULL;
CREATE INDEX ncv7_et_fk03_ix ON nodecontrol.control_plane_authority_epoch_transition_terminal_applications (runtime_rebind_result_digest_or_null) WHERE runtime_rebind_result_digest_or_null IS NOT NULL;
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions
  ADD CONSTRAINT ncv7_pd_fk01 FOREIGN KEY (recovery_intent_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_recovery_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_pd_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions (recovery_intent_digest);
ALTER TABLE nodecontrol.control_plane_authority_epoch_transition_recovery_applications
  ADD CONSTRAINT ncv7_ra_fk01 FOREIGN KEY (recovery_intent_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_recovery_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_ra_fk02 FOREIGN KEY (epoch_transition_recovery_prefix_decision_digest) REFERENCES nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_ra_fk01_ix ON nodecontrol.control_plane_authority_epoch_transition_recovery_applications (recovery_intent_digest);
CREATE INDEX ncv7_ra_fk02_ix ON nodecontrol.control_plane_authority_epoch_transition_recovery_applications (epoch_transition_recovery_prefix_decision_digest);
ALTER TABLE nodecontrol.control_plane_authority_staging_import_capability_recovery_applications
  ADD CONSTRAINT ncv7_sa_fk01 FOREIGN KEY (staging_import_capability_recovery_intent_digest) REFERENCES nodecontrol.control_plane_authority_staging_import_capability_recovery_intents(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION,
  ADD CONSTRAINT ncv7_sa_fk02 FOREIGN KEY (staging_import_capability_digest) REFERENCES nodecontrol.control_plane_authority_staging_import_capabilities(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_sa_fk01_ix ON nodecontrol.control_plane_authority_staging_import_capability_recovery_applications (staging_import_capability_recovery_intent_digest);
CREATE INDEX ncv7_sa_fk02_ix ON nodecontrol.control_plane_authority_staging_import_capability_recovery_applications (staging_import_capability_digest);
ALTER TABLE nodecontrol.control_plane_authority_fresh_restore_import_applications
  ADD CONSTRAINT ncv7_fi_fk01 FOREIGN KEY (staging_import_capability_digest) REFERENCES nodecontrol.control_plane_authority_staging_import_capabilities(body_digest) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX ncv7_fi_fk01_ix ON nodecontrol.control_plane_authority_fresh_restore_import_applications (staging_import_capability_digest);

-- talenro:statement
DO $triggers$
DECLARE
  relation_name text;
  relation_code text;
  names text[] := ARRAY[
    'control_plane_authority_protocol_migration_latches','control_plane_authority_protocol_downgrade_authorizations','control_plane_authority_protocol_upgrade_intents','control_plane_authority_runtime_registration_results','control_plane_authority_runtime_rebind_results','control_plane_authority_protocol_upgrade_attempts','control_plane_authority_epoch_transition_intents','control_plane_authority_epoch_transition_applications','control_plane_authority_epoch_transition_resolutions','control_plane_authority_epoch_transition_cancellations','control_plane_authority_epoch_transition_terminal_applications','control_plane_authority_epoch_transition_recovery_intents','control_plane_authority_epoch_transition_recovery_prefix_decisions','control_plane_authority_epoch_transition_recovery_applications','control_plane_authority_legacy_database_source_retirements','control_plane_authority_indeterminate_source_seals','control_plane_authority_legacy_source_seals','control_plane_authority_fresh_restore_requirements','control_plane_authority_protocol_activations','control_plane_authority_protocol_activation_completions','control_plane_authority_protocol_activation_releases','control_plane_authority_staging_import_capabilities','control_plane_authority_staging_import_capability_revocation_applications','control_plane_authority_staging_import_capability_recovery_intents','control_plane_authority_staging_import_capability_recovery_applications','control_plane_authority_fresh_restore_import_applications'
  ];
  codes text[] := ARRAY['ml','da','ui','rr','rb','ua','ei','ea','er','ec','et','ri','pd','ra','dr','is','ls','fr','pa','pc','pr','sc','sv','si','sa','fi'];
  i integer;
BEGIN
  FOR i IN 1..array_length(names, 1) LOOP
    relation_name := names[i];
    relation_code := codes[i];
    IF relation_name IN (
      'control_plane_authority_legacy_database_source_retirements',
      'control_plane_authority_indeterminate_source_seals',
      'control_plane_authority_legacy_source_seals',
      'control_plane_authority_fresh_restore_requirements'
    ) THEN
      EXECUTE format('CREATE TRIGGER %I BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.%I FOR EACH ROW EXECUTE FUNCTION nodecontrol.v7_reject_immutable_mutation()', 'ncv7_' || relation_code || '_immutable_ud', relation_name);
    ELSE
      EXECUTE format('CREATE TRIGGER %I BEFORE UPDATE OR DELETE ON nodecontrol.%I FOR EACH ROW EXECUTE FUNCTION nodecontrol.v7_reject_immutable_mutation()', 'ncv7_' || relation_code || '_immutable_ud', relation_name);
    END IF;
    EXECUTE format('CREATE TRIGGER %I BEFORE TRUNCATE ON nodecontrol.%I FOR EACH STATEMENT EXECUTE FUNCTION nodecontrol.v7_reject_immutable_mutation()', 'ncv7_' || relation_code || '_truncate', relation_name);
  END LOOP;
END
$triggers$;

-- talenro:statement
DROP TRIGGER node_resource_envelopes_immutable ON nodecontrol.node_resource_envelopes;
CREATE TRIGGER node_resource_envelopes_immutable
BEFORE DELETE ON nodecontrol.node_resource_envelopes
FOR EACH ROW EXECUTE FUNCTION nodecontrol.reject_row_mutation();

-- talenro:statement
DROP TRIGGER node_state_transitions_immutable ON nodecontrol.node_state_transitions;
-- The v7 proof guard below is the sole DELETE authority: it requires an exact
-- terminal fence/owner pair and elapsed retention.  Retaining the v6
-- reject_row_mutation trigger here would make that authorized path unreachable.

-- talenro:statement
DO $proof_triggers$
DECLARE
  relation_name text;
BEGIN
  FOREACH relation_name IN ARRAY ARRAY['node_enrollment_grants','node_certificate_issuances','node_certificates','node_state_transitions','node_security_incidents','node_resource_envelopes','node_state_signing_intents','node_root_metadata_publish_intents'] LOOP
    EXECUTE format('CREATE TRIGGER %I BEFORE INSERT OR UPDATE OR DELETE ON nodecontrol.%I FOR EACH ROW EXECUTE FUNCTION nodecontrol.v7_guard_authority_proof_transition()', 'ncv7_' || relation_name || '_proof_guard', relation_name);
  END LOOP;
END
$proof_triggers$;

-- talenro:statement
CREATE CONSTRAINT TRIGGER ncv7_fence_owner_closure
AFTER UPDATE ON nodecontrol.control_plane_authority_fences
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION nodecontrol.v7_assert_activation_barrier();

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_insert_downgrade_authorization(
  p_authorization_id uuid,
  p_installation_id uuid,
  p_migration_latch_digest bytea,
  p_database_identity_digest bytea,
  p_migration_version bigint,
  p_current_catalog_digest bytea,
  p_pristine_inventory_digest bytea,
  p_provider_retirement_set_digest bytea,
  p_environment_anchor_set_digest bytea,
  p_database_transaction_id numeric,
  p_transaction_nonce bytea,
  p_authorization_scope text,
  p_issued_at timestamptz,
  p_expires_at timestamptz,
  p_authorization_body_jcs bytea,
  p_authorization_envelope_jcs bytea,
  p_authorization_digest bytea,
  p_pristine_inventory_body_jcs bytea,
  p_provider_retirement_set_body_jcs bytea,
  p_provider_retirement_evidence_jcs bytea
)
RETURNS TABLE (
  pristine_downgrade_inventory_digest bytea,
  migration_latch_digest bytea,
  downgrade_authorization_digest bytea,
  provider_protocol_downgrade_retirement_set_digest bytea,
  database_transaction_id numeric,
  transaction_nonce bytea,
  migration_latch_count bigint,
  downgrade_authorization_count bigint,
  non_control_protocol_row_count bigint,
  stage text,
  observed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  v_forbidden_count bigint;
  v_table_name text;
  v_classification text;
  v_row_count bigint;
  v_empty_digest bytea;
  v_non_control_count bigint := 0;
  v_registry_tables text[] := ARRAY[]::text[];
  v_select_tables text[];
  v_lock_tables text[];
  v_select_oids oid[];
  v_lock_oids oid[];
  v_bootstrap_oid oid;
  v_acl_count bigint;
  v_drift_count bigint;
  v_authorization jsonb;
  v_pristine jsonb;
  v_retirement_set jsonb;
  v_authorization_envelope jsonb;
  v_retirement_evidence jsonb;
  v_inventory jsonb := '[]'::jsonb;
  v_evidence_item jsonb;
  v_authorization_raw json;
  v_pristine_raw json;
  v_retirement_set_raw json;
  v_authorization_envelope_raw json;
  v_retirement_evidence_raw json;
  v_inventory_jcs text := '[';
  v_retirements_jcs text;
  v_retired_members_jcs text;
  v_evidence_envelope jsonb;
  v_evidence_body jsonb;
  v_evidence_schema text;
  v_evidence_digest bytea;
  v_evidence_body_jcs text;
  v_membership_request_body_jcs text;
  v_provider_request_body_jcs text;
  v_evidence_envelope_jcs text;
  v_evidence_item_jcs text;
  v_evidence_array_jcs text := '[';
  v_evidence_index bigint := 0;
  v_previous_evidence_digest bytea;
  v_previous_evidence_schema text;
  v_expected_signer_role text;
  v_provider_pair text;
  v_provider_namespace text;
  v_provider_namespace_key text;
  v_provider_namespace_count bigint;
  v_members_jcs text;
  v_namespace_states_jcs text;
  v_environment_record_digests_jcs text;
  v_database_identity_digests_jcs text;
  v_member_set_digest bytea;
  v_membership_count bigint := 0;
  v_validation_now timestamptz;
  v_membership_retired_at timestamptz;
  v_admin_issued_at timestamptz;
  v_admin_expires_at timestamptz;
  v_history_observed_at timestamptz;
  v_history_expires_at timestamptz;
  v_final_retired_at timestamptz;
  v_admin_evidence jsonb := '{}'::jsonb;
  v_namespace_evidence jsonb := '{}'::jsonb;
  v_final_evidence jsonb := '{}'::jsonb;
BEGIN
  v_validation_now := clock_timestamp();
  PERFORM nodecontrol.v7_require_role('nodecontrol_migration_downgrader');
  LOCK TABLE
    nodecontrol.control_plane_authority_protocol_migration_latches,
    nodecontrol.control_plane_authority_protocol_downgrade_authorizations,
    nodecontrol.control_plane_authority_epoch_transition_applications,
    nodecontrol.control_plane_authority_epoch_transition_cancellations,
    nodecontrol.control_plane_authority_epoch_transition_intents,
    nodecontrol.control_plane_authority_epoch_transition_recovery_applications,
    nodecontrol.control_plane_authority_epoch_transition_recovery_intents,
    nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions,
    nodecontrol.control_plane_authority_epoch_transition_resolutions,
    nodecontrol.control_plane_authority_epoch_transition_terminal_applications,
    nodecontrol.control_plane_authority_fresh_restore_import_applications,
    nodecontrol.control_plane_authority_fresh_restore_requirements,
    nodecontrol.control_plane_authority_indeterminate_source_seals,
    nodecontrol.control_plane_authority_legacy_database_source_retirements,
    nodecontrol.control_plane_authority_legacy_source_seals,
    nodecontrol.control_plane_authority_protocol_activation_completions,
    nodecontrol.control_plane_authority_protocol_activation_releases,
    nodecontrol.control_plane_authority_protocol_activations,
    nodecontrol.control_plane_authority_protocol_upgrade_attempts,
    nodecontrol.control_plane_authority_protocol_upgrade_intents,
    nodecontrol.control_plane_authority_runtime_rebind_results,
    nodecontrol.control_plane_authority_runtime_registration_results,
    nodecontrol.control_plane_authority_staging_import_capabilities,
    nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
    nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
    nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
    nodecontrol.control_plane_authority_fences,
    nodecontrol.control_plane_trust_bundle_high_waters,
    nodecontrol.node_capacity_profiles,
    nodecontrol.node_certificate_issuances,
    nodecontrol.node_certificates,
    nodecontrol.node_desired_states,
    nodecontrol.node_endpoints,
    nodecontrol.node_enrollment_grants,
    nodecontrol.node_failure_domain_membership,
    nodecontrol.node_failure_domains,
    nodecontrol.node_inventory,
    nodecontrol.node_observed_states,
    nodecontrol.node_operator_audit,
    nodecontrol.node_pops,
    nodecontrol.node_process_slots,
    nodecontrol.node_recovery_sessions,
    nodecontrol.node_recovery_states,
    nodecontrol.node_resource_envelopes,
    nodecontrol.node_restore_reauthorization_approvals,
    nodecontrol.node_root_metadata_publish_intents,
    nodecontrol.node_root_metadata_signature_shares,
    nodecontrol.node_security_fault_receipts,
    nodecontrol.node_security_incidents,
    nodecontrol.node_state_signing_intents,
    nodecontrol.node_state_transitions,
    public.goose_db_version IN ACCESS EXCLUSIVE MODE;

  PERFORM nodecontrol.v7_require_role('nodecontrol_migration_downgrader');

  FOR v_table_name, v_classification IN
    SELECT table_name, classification
    FROM (VALUES
      ('nodecontrol.control_plane_authority_epoch_transition_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_cancellations', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_recovery_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_recovery_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_resolutions', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_terminal_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_fences', 'base_v6'),
      ('nodecontrol.control_plane_authority_fresh_restore_import_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_fresh_restore_requirements', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_indeterminate_source_seals', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_legacy_database_source_retirements', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_legacy_source_seals', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_activation_completions', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_activation_releases', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_activations', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_upgrade_attempts', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_upgrade_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_runtime_rebind_results', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_runtime_registration_results', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capabilities', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capability_recovery_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capability_recovery_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capability_revocation_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_trust_bundle_high_waters', 'base_v6'),
      ('nodecontrol.node_capacity_profiles', 'base_v6'),
      ('nodecontrol.node_certificate_issuances', 'base_v6'),
      ('nodecontrol.node_certificates', 'base_v6'),
      ('nodecontrol.node_desired_states', 'base_v6'),
      ('nodecontrol.node_endpoints', 'base_v6'),
      ('nodecontrol.node_enrollment_grants', 'base_v6'),
      ('nodecontrol.node_failure_domain_membership', 'base_v6'),
      ('nodecontrol.node_failure_domains', 'base_v6'),
      ('nodecontrol.node_inventory', 'base_v6'),
      ('nodecontrol.node_observed_states', 'base_v6'),
      ('nodecontrol.node_operator_audit', 'base_v6'),
      ('nodecontrol.node_pops', 'base_v6'),
      ('nodecontrol.node_process_slots', 'base_v6'),
      ('nodecontrol.node_recovery_sessions', 'base_v6'),
      ('nodecontrol.node_recovery_states', 'base_v6'),
      ('nodecontrol.node_resource_envelopes', 'base_v6'),
      ('nodecontrol.node_restore_reauthorization_approvals', 'base_v6'),
      ('nodecontrol.node_root_metadata_publish_intents', 'base_v6'),
      ('nodecontrol.node_root_metadata_signature_shares', 'base_v6'),
      ('nodecontrol.node_security_fault_receipts', 'base_v6'),
      ('nodecontrol.node_security_incidents', 'base_v6'),
      ('nodecontrol.node_state_signing_intents', 'base_v6'),
      ('nodecontrol.node_state_transitions', 'base_v6')
    ) AS registry(table_name, classification)
  LOOP
    EXECUTE format('SELECT count(*) FROM %s', v_table_name) INTO v_row_count;
    IF v_row_count <> 0 THEN
      RAISE EXCEPTION 'authority v7 pristine registry relation % is nonempty', v_table_name USING ERRCODE = '55000';
    END IF;
    v_registry_tables := array_append(v_registry_tables, v_table_name);
    v_non_control_count := v_non_control_count + v_row_count;
    v_empty_digest := sha256(
      convert_to('talenro.c12.pristine-downgrade-empty-table.v1', 'UTF8')
      || decode('00', 'hex')
      || convert_to(
        format(
          '{"classification":"%s","row_count":"0","table_name":"%s"}',
          v_classification,
          v_table_name
        ),
        'UTF8'
      )
    );
    IF octet_length(v_empty_digest) <> 32 THEN
      RAISE EXCEPTION 'authority v7 pristine registry digest failure' USING ERRCODE = '55000';
    END IF;
    v_inventory := v_inventory || jsonb_build_array(jsonb_build_object(
      'table_name', v_table_name,
      'classification', v_classification,
      'row_count', '0',
      'content_digest', encode(v_empty_digest, 'hex')
    ));
    v_inventory_jcs := v_inventory_jcs
      || CASE WHEN array_length(v_registry_tables, 1) > 1 THEN ',' ELSE '' END
      || format(
        '{"classification":%s,"content_digest":%s,"row_count":"0","table_name":%s}',
        to_json(v_classification)::text,
        to_json(encode(v_empty_digest, 'hex'))::text,
        to_json(v_table_name)::text
      );
  END LOOP;
  v_inventory_jcs := v_inventory_jcs || ']';
  v_select_tables := ARRAY[
    'nodecontrol.control_plane_authority_protocol_migration_latches',
    'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'
  ] || v_registry_tables;
  v_lock_tables := v_select_tables || ARRAY['public.goose_db_version'];
  SELECT array_agg(table_name::regclass::oid ORDER BY ordinal)
  INTO v_select_oids
  FROM unnest(v_select_tables) WITH ORDINALITY AS selected(table_name, ordinal);
  SELECT array_agg(table_name::regclass::oid ORDER BY ordinal)
  INTO v_lock_oids
  FROM unnest(v_lock_tables) WITH ORDINALITY AS locked(table_name, ordinal);
  SELECT oid INTO v_bootstrap_oid
  FROM pg_catalog.pg_roles
  WHERE rolname = session_user AND rolsuper;
  IF v_bootstrap_oid IS NULL
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_database
       WHERE datname = current_database() AND datdba = v_bootstrap_oid
     )
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_namespace
       WHERE nspname = 'nodecontrol' AND nspowner = v_bootstrap_oid
     ) THEN
    RAISE EXCEPTION 'authority v7 Down requires the exact bootstrap owner' USING ERRCODE = '42501';
  END IF;

  SELECT count(*) INTO v_acl_count
  FROM pg_catalog.pg_roles
  WHERE rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  )
    AND NOT rolcanlogin
    AND NOT rolsuper
    AND NOT rolcreatedb
    AND NOT rolcreaterole
    AND NOT rolreplication
    AND NOT rolbypassrls
    AND rolconfig IS NULL;
  SELECT count(*) INTO v_drift_count
  FROM pg_catalog.pg_auth_members AS membership
  WHERE membership.member IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    )
     OR membership.roleid IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    );
  IF v_acl_count <> 3 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_acl_count
  FROM pg_catalog.pg_proc AS restricted_function
  JOIN pg_catalog.pg_roles AS owner_role ON owner_role.oid = restricted_function.proowner
  WHERE (restricted_function.oid =
      'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure
      AND owner_role.rolname = 'nodecontrol_upgrade_executor')
     OR (restricted_function.oid IN (
      'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
      'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure
    ) AND owner_role.rolname = 'nodecontrol_migration_downgrader')
     OR (restricted_function.oid =
      'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure
      AND owner_role.rolname = 'nodecontrol_staging_importer');
  IF v_acl_count <> 4 THEN
    RAISE EXCEPTION 'authority v7 restricted function owner drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_drift_count
  FROM pg_catalog.pg_class AS relation
  WHERE relation.relowner IN (
    SELECT oid FROM pg_catalog.pg_roles
    WHERE rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
  );
  v_drift_count := v_drift_count + (
    SELECT count(*) FROM pg_catalog.pg_namespace AS namespace
    WHERE namespace.nspowner IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    )
  ) + (
    SELECT count(*) FROM pg_catalog.pg_database AS database
    WHERE database.datdba IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    )
  ) + (
    SELECT count(*) FROM pg_catalog.pg_proc AS proc
    JOIN pg_catalog.pg_roles AS owner ON owner.oid = proc.proowner
    WHERE owner.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
      AND proc.oid <> ALL (ARRAY[
        'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid,
        'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
        'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
        'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid
      ])
  );
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role foreign ownership drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*) FILTER (WHERE acl.grantee = v_bootstrap_oid AND acl.privilege_type = 'EXECUTE' AND NOT acl.is_grantable),
    count(*) FILTER (
      WHERE acl.privilege_type <> 'EXECUTE'
         OR acl.is_grantable
         OR acl.grantee NOT IN (proc.proowner, v_bootstrap_oid)
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_proc AS proc
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
  ) AS acl
  WHERE proc.oid = ANY (ARRAY[
    'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid,
    'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid
  ]);
  IF v_acl_count <> 4 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 entry function ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*) FILTER (
      WHERE grantee.rolname IN ('nodecontrol_upgrade_executor', 'nodecontrol_migration_downgrader', 'nodecontrol_staging_importer')
        AND acl.privilege_type = 'EXECUTE'
        AND NOT acl.is_grantable
    ),
    count(*) FILTER (
      WHERE acl.privilege_type <> 'EXECUTE'
         OR acl.is_grantable
         OR acl.grantee NOT IN (
           proc.proowner,
           'nodecontrol_upgrade_executor'::regrole::oid,
           'nodecontrol_migration_downgrader'::regrole::oid,
           'nodecontrol_staging_importer'::regrole::oid
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_proc AS proc
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
  ) AS acl
  LEFT JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = acl.grantee
  WHERE proc.oid = 'nodecontrol.v7_require_role(name)'::regprocedure;
  IF v_acl_count <> 3 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 role helper ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR NOT (
           (acl.privilege_type = 'SELECT' AND relation.oid = ANY (v_select_oids))
           OR (acl.privilege_type = 'MAINTAIN' AND relation.oid = ANY (v_lock_oids))
           OR (acl.privilege_type = 'INSERT' AND relation.oid =
             'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass)
           OR (acl.privilege_type = 'DELETE' AND relation.oid IN (
             'nodecontrol.control_plane_authority_protocol_migration_latches'::regclass,
             'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass
           ))
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_class AS relation
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(relation.relacl, pg_catalog.acldefault('r', relation.relowner))
  ) AS acl
  WHERE acl.grantee = 'nodecontrol_migration_downgrader'::regrole;
  IF v_acl_count <> 106 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrader table ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR NOT (
           (acl.privilege_type = 'SELECT' AND relation.oid IN (
             'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass,
             'nodecontrol.control_plane_authority_staging_import_capability_recovery_intents'::regclass,
             'nodecontrol.control_plane_authority_staging_import_capability_recovery_applications'::regclass,
             'nodecontrol.control_plane_authority_staging_import_capability_revocation_applications'::regclass,
             'nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass
           ))
           OR (acl.privilege_type = 'INSERT' AND relation.oid IN (
             'nodecontrol.node_pops'::regclass,
             'nodecontrol.node_failure_domains'::regclass,
             'nodecontrol.node_capacity_profiles'::regclass,
             'nodecontrol.node_inventory'::regclass,
             'nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass
           ))
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_class AS relation
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(relation.relacl, pg_catalog.acldefault('r', relation.relowner))
  ) AS acl
  WHERE acl.grantee = 'nodecontrol_staging_importer'::regrole;
  IF v_acl_count <> 10 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 staging table ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR acl.privilege_type <> 'UPDATE'
         OR relation.oid <> 'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass
         OR attribute.attname <> 'body_digest'
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_attribute AS attribute
  JOIN pg_catalog.pg_class AS relation ON relation.oid = attribute.attrelid
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(attribute.attacl) AS acl
  WHERE attribute.attacl IS NOT NULL
    AND acl.grantee = 'nodecontrol_staging_importer'::regrole;
  IF v_acl_count <> 1 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 staging column ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR acl.privilege_type <> 'USAGE'
         OR NOT (
           (grantee.rolname = 'nodecontrol_upgrade_executor' AND namespace.nspname = 'nodecontrol')
           OR (grantee.rolname = 'nodecontrol_migration_downgrader' AND namespace.nspname IN ('nodecontrol', 'public'))
           OR (grantee.rolname = 'nodecontrol_staging_importer' AND namespace.nspname = 'nodecontrol')
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_namespace AS namespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(namespace.nspacl, pg_catalog.acldefault('n', namespace.nspowner))
  ) AS acl
  JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = acl.grantee
  WHERE grantee.rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  );
  IF v_acl_count <> 4 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability schema ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_drift_count
  FROM pg_catalog.pg_proc AS proc
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = proc.pronamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
  ) AS acl
  JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = acl.grantee
  WHERE grantee.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
    AND NOT (
      acl.privilege_type = 'EXECUTE'
      AND NOT acl.is_grantable
      AND (
        (grantee.rolname = 'nodecontrol_upgrade_executor'
          AND proc.oid IN (
            'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure,
            'nodecontrol.v7_require_role(name)'::regprocedure
          ))
        OR (grantee.rolname = 'nodecontrol_migration_downgrader'
          AND proc.oid IN (
            'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
            'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
            'nodecontrol.v7_require_role(name)'::regprocedure
          ))
        OR (grantee.rolname = 'nodecontrol_staging_importer'
          AND proc.oid IN (
            'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure,
            'nodecontrol.v7_require_role(name)'::regprocedure
          ))
      )
    );
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability function ACL drift' USING ERRCODE = '55000';
  END IF;
  SELECT count(*), count(*) FILTER (WHERE NOT (
    dependency.dbid = (SELECT oid FROM pg_catalog.pg_database WHERE datname = current_database())
    AND dependency.classid = 'pg_catalog.pg_proc'::regclass
    AND dependency.objsubid = 0
    AND (
      (role_catalog.rolname = 'nodecontrol_upgrade_executor'
        AND dependency.objid = 'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure)
      OR (role_catalog.rolname = 'nodecontrol_migration_downgrader' AND dependency.objid IN (
        'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
        'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure
      ))
      OR (role_catalog.rolname = 'nodecontrol_staging_importer' AND dependency.objid = 'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure)
    )
  )) INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_shdepend AS dependency
  JOIN pg_catalog.pg_roles AS role_catalog ON role_catalog.oid = dependency.refobjid
  WHERE dependency.refclassid = 'pg_catalog.pg_authid'::regclass
    AND dependency.deptype = 'o'
    AND role_catalog.rolname IN ('nodecontrol_upgrade_executor','nodecontrol_migration_downgrader','nodecontrol_staging_importer');
  IF v_acl_count <> 4 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability foreign ownership drift' USING ERRCODE = '55000';
  END IF;
  SELECT count(*) INTO v_drift_count
  FROM (
    VALUES
      ('nodecontrol.v7_text_array_is_sorted_unique(text[])'::regprocedure, session_user::text, 'sql', 'i', false),
      ('nodecontrol.v7_reject_immutable_mutation()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_require_role(name)'::regprocedure, session_user::text, 'plpgsql', 's', false),
      ('nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure, 'nodecontrol_upgrade_executor', 'plpgsql', 'v', true),
      ('nodecontrol.v7_source_is_frozen()'::regprocedure, session_user::text, 'sql', 'v', true),
      ('nodecontrol.v7_assert_source_writable()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_assert_activation_barrier()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_authority_proof_group_valid(bytea,bytea,bytea,bytea,bytea,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea)'::regprocedure, session_user::text, 'sql', 'i', false),
      ('nodecontrol.v7_guard_authority_proof_transition()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure, 'nodecontrol_migration_downgrader', 'plpgsql', 'v', true),
      ('nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure, 'nodecontrol_migration_downgrader', 'plpgsql', 'v', true),
      ('nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure, 'nodecontrol_staging_importer', 'plpgsql', 'v', true)
  ) AS expected(proc_oid, owner_name, language_name, volatility_code, security_definer)
  JOIN pg_catalog.pg_proc AS proc ON proc.oid = expected.proc_oid
  JOIN pg_catalog.pg_roles AS owner ON owner.oid = proc.proowner
  JOIN pg_catalog.pg_language AS language ON language.oid = proc.prolang
  WHERE owner.rolname IS DISTINCT FROM expected.owner_name
     OR language.lanname IS DISTINCT FROM expected.language_name
     OR proc.provolatile IS DISTINCT FROM expected.volatility_code::"char"
     OR proc.prosecdef IS DISTINCT FROM expected.security_definer
     OR proc.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, nodecontrol']::text[];
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 function metadata drift' USING ERRCODE = '55000';
  END IF;
  SELECT count(*) INTO v_drift_count
  FROM (
    SELECT acl.grantee
    FROM pg_catalog.pg_database AS database
    CROSS JOIN LATERAL pg_catalog.aclexplode(COALESCE(database.datacl, pg_catalog.acldefault('d', database.datdba))) AS acl
    UNION ALL
    SELECT acl.grantee
    FROM pg_catalog.pg_type AS type_catalog
    CROSS JOIN LATERAL pg_catalog.aclexplode(COALESCE(type_catalog.typacl, pg_catalog.acldefault('T', type_catalog.typowner))) AS acl
    UNION ALL
    SELECT acl.grantee
    FROM pg_catalog.pg_class AS sequence
    CROSS JOIN LATERAL pg_catalog.aclexplode(COALESCE(sequence.relacl, pg_catalog.acldefault('S', sequence.relowner))) AS acl
    WHERE sequence.relkind = 'S'
  ) AS direct_acl
  JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = direct_acl.grantee
  WHERE grantee.rolname IN ('nodecontrol_upgrade_executor','nodecontrol_migration_downgrader','nodecontrol_staging_importer');
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability database/type/sequence ACL drift' USING ERRCODE = '55000';
  END IF;
  SELECT count(*) INTO v_drift_count
  FROM (
    SELECT acl.grantee, acl.grantor, relation.relowner AS owner_oid
    FROM pg_catalog.pg_class AS relation
    CROSS JOIN LATERAL pg_catalog.aclexplode(COALESCE(relation.relacl, pg_catalog.acldefault(CASE WHEN relation.relkind = 'S' THEN 'S'::"char" ELSE 'r'::"char" END, relation.relowner))) AS acl
    UNION ALL
    SELECT acl.grantee, acl.grantor, relation.relowner
    FROM pg_catalog.pg_attribute AS attribute
    JOIN pg_catalog.pg_class AS relation ON relation.oid = attribute.attrelid
    CROSS JOIN LATERAL pg_catalog.aclexplode(attribute.attacl) AS acl
    WHERE attribute.attacl IS NOT NULL
    UNION ALL
    SELECT acl.grantee, acl.grantor, namespace.nspowner
    FROM pg_catalog.pg_namespace AS namespace
    CROSS JOIN LATERAL pg_catalog.aclexplode(COALESCE(namespace.nspacl, pg_catalog.acldefault('n', namespace.nspowner))) AS acl
    UNION ALL
    SELECT acl.grantee, acl.grantor, proc.proowner
    FROM pg_catalog.pg_proc AS proc
    CROSS JOIN LATERAL pg_catalog.aclexplode(COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))) AS acl
  ) AS granted
  JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = granted.grantee
  WHERE grantee.rolname IN ('nodecontrol_upgrade_executor','nodecontrol_migration_downgrader','nodecontrol_staging_importer')
    AND granted.grantor <> granted.owner_oid;
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability ACL grantor drift' USING ERRCODE = '55000';
  END IF;
  IF p_authorization_id IS NULL
     OR p_installation_id IS NULL
     OR p_migration_latch_digest IS NULL
     OR p_database_identity_digest IS NULL
     OR p_migration_version IS NULL
     OR p_current_catalog_digest IS NULL
     OR p_pristine_inventory_digest IS NULL
     OR p_provider_retirement_set_digest IS NULL
     OR p_environment_anchor_set_digest IS NULL
     OR p_database_transaction_id IS NULL
     OR p_transaction_nonce IS NULL
     OR p_authorization_scope IS NULL
     OR p_issued_at IS NULL
     OR p_expires_at IS NULL
     OR p_authorization_body_jcs IS NULL
     OR p_authorization_envelope_jcs IS NULL
     OR p_authorization_digest IS NULL
     OR p_pristine_inventory_body_jcs IS NULL
     OR p_provider_retirement_set_body_jcs IS NULL
     OR p_provider_retirement_evidence_jcs IS NULL
     OR octet_length(p_migration_latch_digest) IS DISTINCT FROM 32
     OR octet_length(p_database_identity_digest) IS DISTINCT FROM 32
     OR octet_length(p_current_catalog_digest) IS DISTINCT FROM 32
     OR octet_length(p_pristine_inventory_digest) IS DISTINCT FROM 32
     OR octet_length(p_provider_retirement_set_digest) IS DISTINCT FROM 32
     OR octet_length(p_environment_anchor_set_digest) IS DISTINCT FROM 32
     OR octet_length(p_transaction_nonce) IS DISTINCT FROM 32
     OR octet_length(p_authorization_digest) IS DISTINCT FROM 32
     OR octet_length(p_authorization_body_jcs) NOT BETWEEN 1 AND 1048576
     OR octet_length(p_authorization_envelope_jcs) NOT BETWEEN 1 AND 1048576
     OR octet_length(p_pristine_inventory_body_jcs) NOT BETWEEN 1 AND 1048576
     OR octet_length(p_provider_retirement_set_body_jcs) NOT BETWEEN 1 AND 1048576
     OR octet_length(p_provider_retirement_evidence_jcs) NOT BETWEEN 1 AND 1048576
     OR p_database_transaction_id IS DISTINCT FROM txid_current()::numeric
     OR p_migration_version IS DISTINCT FROM 7
     OR p_authorization_scope IS DISTINCT FROM 'down_00007_only'
     OR p_expires_at <= p_issued_at
     OR p_expires_at > p_issued_at + interval '2 minutes'
     OR p_issued_at > clock_timestamp()
     OR clock_timestamp() >= p_expires_at THEN
    RAISE EXCEPTION 'invalid authority v7 downgrade authorization scalar binding' USING ERRCODE = '22023';
  END IF;

  BEGIN
    v_authorization_raw := convert_from(p_authorization_body_jcs, 'UTF8')::json;
    v_pristine_raw := convert_from(p_pristine_inventory_body_jcs, 'UTF8')::json;
    v_retirement_set_raw := convert_from(p_provider_retirement_set_body_jcs, 'UTF8')::json;
    v_authorization_envelope_raw := convert_from(p_authorization_envelope_jcs, 'UTF8')::json;
    v_retirement_evidence_raw := convert_from(p_provider_retirement_evidence_jcs, 'UTF8')::json;
    v_authorization := convert_from(p_authorization_body_jcs, 'UTF8')::jsonb;
    v_pristine := convert_from(p_pristine_inventory_body_jcs, 'UTF8')::jsonb;
    v_retirement_set := convert_from(p_provider_retirement_set_body_jcs, 'UTF8')::jsonb;
    v_authorization_envelope := convert_from(p_authorization_envelope_jcs, 'UTF8')::jsonb;
    v_retirement_evidence := convert_from(p_provider_retirement_evidence_jcs, 'UTF8')::jsonb;
  EXCEPTION WHEN OTHERS THEN
    RAISE EXCEPTION 'invalid authority v7 canonical JSON preimage' USING ERRCODE = '22023';
  END;

  IF jsonb_typeof(v_authorization) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_authorization)) <> 14
     OR NOT (v_authorization ?& ARRAY[
       'authorization_id',
       'installation_id',
       'migration_latch_digest',
       'database_identity_digest',
       'migration_version',
       'current_catalog_digest',
       'pristine_downgrade_inventory_digest',
       'provider_protocol_downgrade_retirement_set_digest',
       'environment_inventory_anchor_set_digest',
       'database_transaction_id',
       'transaction_nonce',
       'authorization_scope',
       'issued_at',
       'expires_at'
     ])
     OR EXISTS (
       SELECT 1 FROM jsonb_each(v_authorization) AS field(name, value)
       WHERE jsonb_typeof(value) <> 'string'
     )
     OR convert_from(p_authorization_body_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"authorization_id":%s,"authorization_scope":%s,"current_catalog_digest":%s,"database_identity_digest":%s,"database_transaction_id":%s,"environment_inventory_anchor_set_digest":%s,"expires_at":%s,"installation_id":%s,"issued_at":%s,"migration_latch_digest":%s,"migration_version":%s,"pristine_downgrade_inventory_digest":%s,"provider_protocol_downgrade_retirement_set_digest":%s,"transaction_nonce":%s}',
       to_json(v_authorization->>'authorization_id')::text,
       to_json(v_authorization->>'authorization_scope')::text,
       to_json(v_authorization->>'current_catalog_digest')::text,
       to_json(v_authorization->>'database_identity_digest')::text,
       to_json(v_authorization->>'database_transaction_id')::text,
       to_json(v_authorization->>'environment_inventory_anchor_set_digest')::text,
       to_json(v_authorization->>'expires_at')::text,
       to_json(v_authorization->>'installation_id')::text,
       to_json(v_authorization->>'issued_at')::text,
       to_json(v_authorization->>'migration_latch_digest')::text,
       to_json(v_authorization->>'migration_version')::text,
       to_json(v_authorization->>'pristine_downgrade_inventory_digest')::text,
       to_json(v_authorization->>'provider_protocol_downgrade_retirement_set_digest')::text,
       to_json(v_authorization->>'transaction_nonce')::text
     )
     OR sha256(
       convert_to('talenro.c12.authority-protocol-downgrade-authorization.v1', 'UTF8')
       || decode('00', 'hex')
       || p_authorization_body_jcs
     ) IS DISTINCT FROM p_authorization_digest
     OR v_authorization->>'authorization_id' IS DISTINCT FROM p_authorization_id::text
     OR v_authorization->>'installation_id' IS DISTINCT FROM p_installation_id::text
     OR v_authorization->>'migration_latch_digest' IS DISTINCT FROM encode(p_migration_latch_digest, 'hex')
     OR v_authorization->>'database_identity_digest' IS DISTINCT FROM encode(p_database_identity_digest, 'hex')
     OR v_authorization->>'migration_version' IS DISTINCT FROM p_migration_version::text
     OR v_authorization->>'current_catalog_digest' IS DISTINCT FROM encode(p_current_catalog_digest, 'hex')
     OR v_authorization->>'pristine_downgrade_inventory_digest' IS DISTINCT FROM encode(p_pristine_inventory_digest, 'hex')
     OR v_authorization->>'provider_protocol_downgrade_retirement_set_digest' IS DISTINCT FROM encode(p_provider_retirement_set_digest, 'hex')
     OR v_authorization->>'environment_inventory_anchor_set_digest' IS DISTINCT FROM encode(p_environment_anchor_set_digest, 'hex')
     OR v_authorization->>'database_transaction_id' IS DISTINCT FROM p_database_transaction_id::text
     OR v_authorization->>'transaction_nonce' IS DISTINCT FROM encode(p_transaction_nonce, 'hex')
     OR v_authorization->>'authorization_scope' IS DISTINCT FROM p_authorization_scope
     OR (v_authorization->>'issued_at')::timestamptz IS DISTINCT FROM p_issued_at
     OR (v_authorization->>'expires_at')::timestamptz IS DISTINCT FROM p_expires_at THEN
    RAISE EXCEPTION 'authority v7 authorization body cross-binding mismatch' USING ERRCODE = '22023';
  END IF;

  IF jsonb_typeof(v_pristine) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_pristine)) <> 17
     OR NOT (v_pristine ?& ARRAY[
       'installation_id',
       'installation_kind',
       'database_identity_digest',
       'migration_latch_digest',
       'current_catalog_digest',
       'manifest_id',
       'manifest_digest',
       'stable_table_count',
       'stable_table_inventory',
       'control_table_count',
       'non_control_protocol_row_count',
       'migration_latch_count',
       'downgrade_authorization_count',
       'database_transaction_id',
       'transaction_nonce',
       'stage',
       'observed_at'
     ])
     OR jsonb_typeof(v_pristine->'stable_table_inventory') <> 'array'
     OR jsonb_array_length(v_pristine->'stable_table_inventory') <> 49
     OR v_pristine->'stable_table_inventory' IS DISTINCT FROM v_inventory
     OR convert_from(p_pristine_inventory_body_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"control_table_count":%s,"current_catalog_digest":%s,"database_identity_digest":%s,"database_transaction_id":%s,"downgrade_authorization_count":%s,"installation_id":%s,"installation_kind":%s,"manifest_digest":%s,"manifest_id":%s,"migration_latch_count":%s,"migration_latch_digest":%s,"non_control_protocol_row_count":%s,"observed_at":%s,"stable_table_count":%s,"stable_table_inventory":%s,"stage":%s,"transaction_nonce":%s}',
       to_json(v_pristine->>'control_table_count')::text,
       to_json(v_pristine->>'current_catalog_digest')::text,
       to_json(v_pristine->>'database_identity_digest')::text,
       to_json(v_pristine->>'database_transaction_id')::text,
       to_json(v_pristine->>'downgrade_authorization_count')::text,
       to_json(v_pristine->>'installation_id')::text,
       to_json(v_pristine->>'installation_kind')::text,
       to_json(v_pristine->>'manifest_digest')::text,
       to_json(v_pristine->>'manifest_id')::text,
       to_json(v_pristine->>'migration_latch_count')::text,
       to_json(v_pristine->>'migration_latch_digest')::text,
       to_json(v_pristine->>'non_control_protocol_row_count')::text,
       to_json(v_pristine->>'observed_at')::text,
       to_json(v_pristine->>'stable_table_count')::text,
       v_inventory_jcs,
       to_json(v_pristine->>'stage')::text,
       to_json(v_pristine->>'transaction_nonce')::text
     )
     OR sha256(
       convert_to('talenro.c12.pristine-downgrade-inventory.v1', 'UTF8')
       || decode('00', 'hex')
       || p_pristine_inventory_body_jcs
     ) IS DISTINCT FROM p_pristine_inventory_digest
     OR v_pristine->>'installation_id' IS DISTINCT FROM p_installation_id::text
     OR v_pristine->>'installation_kind' IS DISTINCT FROM 'disposable_fixture'
     OR v_pristine->>'database_identity_digest' IS DISTINCT FROM encode(p_database_identity_digest, 'hex')
     OR v_pristine->>'migration_latch_digest' IS DISTINCT FROM encode(p_migration_latch_digest, 'hex')
     OR v_pristine->>'current_catalog_digest' IS DISTINCT FROM encode(p_current_catalog_digest, 'hex')
     OR octet_length(decode(v_pristine->>'manifest_digest', 'hex')) IS DISTINCT FROM 32
     OR (v_pristine->>'manifest_id')::uuid IS NULL
     OR v_pristine->>'stable_table_count' IS DISTINCT FROM '49'
     OR v_pristine->>'control_table_count' IS DISTINCT FROM '2'
     OR v_pristine->>'non_control_protocol_row_count' IS DISTINCT FROM '0'
     OR v_pristine->>'migration_latch_count' IS DISTINCT FROM '1'
     OR v_pristine->>'downgrade_authorization_count' IS DISTINCT FROM '0'
     OR v_pristine->>'database_transaction_id' IS DISTINCT FROM p_database_transaction_id::text
     OR v_pristine->>'transaction_nonce' IS DISTINCT FROM encode(p_transaction_nonce, 'hex')
     OR v_pristine->>'stage' IS DISTINCT FROM 'pre_authorization'
     OR (v_pristine->>'observed_at')::timestamptz IS DISTINCT FROM transaction_timestamp() THEN
    RAISE EXCEPTION 'authority v7 pristine inventory cross-binding mismatch' USING ERRCODE = '22023';
  END IF;

  SELECT '[' || COALESCE(string_agg(format(
    '{"provider_endpoint_identity_digest":%s,"provider_identity_digest":%s,"provider_protocol_downgrade_retirement_digest":%s}',
    to_json(value->>'provider_endpoint_identity_digest')::text,
    to_json(value->>'provider_identity_digest')::text,
    to_json(value->>'provider_protocol_downgrade_retirement_digest')::text
  ), ',' ORDER BY ordinal), '') || ']'
  INTO v_retirements_jcs
  FROM jsonb_array_elements(v_retirement_set->'retirements') WITH ORDINALITY AS retirement(value, ordinal);
  SELECT '[' || COALESCE(string_agg(format(
    '{"database_identity_digest":%s,"database_name":%s,"database_oid":%s,"deployment_id":%s,"environment_attestation_digest":%s,"environment_instance_generation":%s,"environment_record_digest":%s,"postgres_system_id":%s,"provider_endpoint_identity_digest":%s,"provider_identity_digest":%s,"provider_namespace":%s,"provider_profile":%s,"timeline":%s}',
    to_json(value->>'database_identity_digest')::text,
    to_json(value->>'database_name')::text,
    to_json(value->>'database_oid')::text,
    to_json(value->>'deployment_id')::text,
    to_json(value->>'environment_attestation_digest')::text,
    to_json(value->>'environment_instance_generation')::text,
    to_json(value->>'environment_record_digest')::text,
    to_json(value->>'postgres_system_id')::text,
    to_json(value->>'provider_endpoint_identity_digest')::text,
    to_json(value->>'provider_identity_digest')::text,
    to_json(value->>'provider_namespace')::text,
    to_json(value->>'provider_profile')::text,
    to_json(value->>'timeline')::text
  ), ',' ORDER BY ordinal), '') || ']'
  INTO v_retired_members_jcs
  FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal);

  IF jsonb_typeof(v_retirement_set) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_retirement_set)) <> 10
     OR NOT (v_retirement_set ?& ARRAY[
       'installation_id',
       'release_scope',
       'environment_inventory_digest',
       'environment_inventory_anchor_set_digest',
       'environment_inventory_membership_retirement_digest',
       'provider_count',
       'retirements',
       'retired_member_count',
       'retired_members',
       'created_at'
     ])
     OR jsonb_typeof(v_retirement_set->'retirements') <> 'array'
     OR jsonb_typeof(v_retirement_set->'retired_members') <> 'array'
     OR convert_from(p_provider_retirement_set_body_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"created_at":%s,"environment_inventory_anchor_set_digest":%s,"environment_inventory_digest":%s,"environment_inventory_membership_retirement_digest":%s,"installation_id":%s,"provider_count":%s,"release_scope":%s,"retired_member_count":%s,"retired_members":%s,"retirements":%s}',
       to_json(v_retirement_set->>'created_at')::text,
       to_json(v_retirement_set->>'environment_inventory_anchor_set_digest')::text,
       to_json(v_retirement_set->>'environment_inventory_digest')::text,
       to_json(v_retirement_set->>'environment_inventory_membership_retirement_digest')::text,
       to_json(v_retirement_set->>'installation_id')::text,
       to_json(v_retirement_set->>'provider_count')::text,
       to_json(v_retirement_set->>'release_scope')::text,
       to_json(v_retirement_set->>'retired_member_count')::text,
       v_retired_members_jcs,
       v_retirements_jcs
     )
     OR jsonb_array_length(v_retirement_set->'retirements') = 0
     OR jsonb_array_length(v_retirement_set->'retired_members') = 0
     OR v_retirement_set->>'installation_id' !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
     OR (v_retirement_set->>'installation_id')::uuid::text IS DISTINCT FROM v_retirement_set->>'installation_id'
     OR (v_retirement_set->>'installation_id')::uuid = '00000000-0000-0000-0000-000000000000'::uuid
     OR v_retirement_set->>'provider_count' !~ '^[1-9][0-9]*$'
     OR (v_retirement_set->>'provider_count')::numeric > 9223372036854775807
     OR v_retirement_set->>'retired_member_count' !~ '^[1-9][0-9]*$'
     OR (v_retirement_set->>'retired_member_count')::numeric > 9223372036854775807
     OR v_retirement_set->>'provider_count' IS DISTINCT FROM jsonb_array_length(v_retirement_set->'retirements')::text
     OR v_retirement_set->>'retired_member_count' IS DISTINCT FROM jsonb_array_length(v_retirement_set->'retired_members')::text
     OR v_retirement_set->>'installation_id' IS DISTINCT FROM p_installation_id::text
     OR v_retirement_set->>'environment_inventory_anchor_set_digest' IS DISTINCT FROM encode(p_environment_anchor_set_digest, 'hex')
     OR v_retirement_set->>'environment_inventory_digest' !~ '^[0-9a-f]{64}$'
     OR v_retirement_set->>'environment_inventory_digest' = repeat('0', 64)
     OR v_retirement_set->>'environment_inventory_anchor_set_digest' !~ '^[0-9a-f]{64}$'
     OR v_retirement_set->>'environment_inventory_anchor_set_digest' = repeat('0', 64)
     OR v_retirement_set->>'environment_inventory_membership_retirement_digest' !~ '^[0-9a-f]{64}$'
     OR v_retirement_set->>'environment_inventory_membership_retirement_digest' = repeat('0', 64)
     OR length(v_retirement_set->>'release_scope') NOT BETWEEN 1 AND 128
     OR v_retirement_set->>'release_scope' ~ '[^ -~]'
     OR (v_retirement_set->>'created_at')::timestamptz > clock_timestamp()
     OR sha256(
       convert_to('talenro.c12.provider-protocol-downgrade-retirement-set.v1', 'UTF8')
       || decode('00', 'hex')
       || p_provider_retirement_set_body_jcs
     ) IS DISTINCT FROM p_provider_retirement_set_digest
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
       WHERE jsonb_typeof(value) <> 'object'
          OR (SELECT count(*) FROM jsonb_object_keys(value)) <> 3
          OR NOT (value ?& ARRAY[
            'provider_identity_digest',
            'provider_endpoint_identity_digest',
            'provider_protocol_downgrade_retirement_digest'
          ])
          OR value->>'provider_identity_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'provider_identity_digest' = repeat('0', 64)
          OR value->>'provider_endpoint_identity_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'provider_endpoint_identity_digest' = repeat('0', 64)
          OR value->>'provider_protocol_downgrade_retirement_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'provider_protocol_downgrade_retirement_digest' = repeat('0', 64)
     )
     OR EXISTS (
       SELECT 1
       FROM (
         SELECT value,
                lag(value) OVER (ORDER BY ordinal) AS previous
         FROM jsonb_array_elements(v_retirement_set->'retirements') WITH ORDINALITY AS retirement(value, ordinal)
       ) AS ordered
       WHERE previous IS NOT NULL
         AND ROW(decode(previous->>'provider_identity_digest','hex'), decode(previous->>'provider_endpoint_identity_digest','hex'))
             >= ROW(decode(value->>'provider_identity_digest','hex'), decode(value->>'provider_endpoint_identity_digest','hex'))
     )
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_set->'retired_members') AS member(value)
       WHERE jsonb_typeof(value) <> 'object'
          OR (SELECT count(*) FROM jsonb_object_keys(value)) <> 13
          OR NOT (value ?& ARRAY[
            'environment_record_digest',
            'environment_attestation_digest',
            'deployment_id',
            'postgres_system_id',
            'timeline',
            'database_oid',
            'database_name',
            'database_identity_digest',
            'environment_instance_generation',
            'provider_identity_digest',
            'provider_endpoint_identity_digest',
            'provider_namespace',
            'provider_profile'
          ])
          OR value->>'environment_record_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'environment_record_digest' = repeat('0', 64)
          OR value->>'environment_attestation_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'environment_attestation_digest' = repeat('0', 64)
          OR value->>'database_identity_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'database_identity_digest' = repeat('0', 64)
          OR value->>'provider_identity_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'provider_identity_digest' = repeat('0', 64)
          OR value->>'provider_endpoint_identity_digest' !~ '^[0-9a-f]{64}$'
          OR value->>'provider_endpoint_identity_digest' = repeat('0', 64)
          OR value->>'deployment_id' !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
          OR (value->>'deployment_id')::uuid::text IS DISTINCT FROM value->>'deployment_id'
          OR (value->>'deployment_id')::uuid = '00000000-0000-0000-0000-000000000000'::uuid
          OR value->>'postgres_system_id' !~ '^[1-9][0-9]*$'
          OR value->>'timeline' !~ '^[1-9][0-9]*$'
          OR (value->>'timeline')::numeric > 4294967295
          OR value->>'database_oid' !~ '^[1-9][0-9]*$'
          OR (value->>'database_oid')::numeric > 4294967295
          OR value->>'environment_instance_generation' !~ '^[1-9][0-9]*$'
          OR (value->>'environment_instance_generation')::numeric > 9223372036854775807
          OR length(value->>'database_name') NOT BETWEEN 1 AND 63
          OR value->>'database_name' ~ '[^ -~]'
          OR length(value->>'provider_namespace') NOT BETWEEN 1 AND 128
          OR value->>'provider_namespace' ~ '[^ -~]'
          OR length(value->>'provider_profile') NOT BETWEEN 1 AND 128
          OR value->>'provider_profile' ~ '[^ -~]'
          OR NOT EXISTS (
            SELECT 1
            FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
            WHERE retirement.value->>'provider_identity_digest' = member.value->>'provider_identity_digest'
              AND retirement.value->>'provider_endpoint_identity_digest' = member.value->>'provider_endpoint_identity_digest'
          )
     )
     OR EXISTS (
       SELECT 1
       FROM (
         SELECT value,
                lag(value) OVER (ORDER BY ordinal) AS previous
         FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
       ) AS ordered
       WHERE previous IS NOT NULL
        AND ROW(
          (previous->>'deployment_id') COLLATE "C",
          (previous->>'postgres_system_id')::numeric,
           (previous->>'database_oid')::numeric,
          (previous->>'database_name') COLLATE "C",
           decode(previous->>'provider_identity_digest','hex'),
           (previous->>'provider_namespace') COLLATE "C"
        ) >= ROW(
          (value->>'deployment_id') COLLATE "C",
          (value->>'postgres_system_id')::numeric,
           (value->>'database_oid')::numeric,
          (value->>'database_name') COLLATE "C",
           decode(value->>'provider_identity_digest','hex'),
           (value->>'provider_namespace') COLLATE "C"
         )
     )
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
       WHERE NOT EXISTS (
         SELECT 1
         FROM jsonb_array_elements(v_retirement_set->'retired_members') AS member(value)
         WHERE member.value->>'provider_identity_digest' = retirement.value->>'provider_identity_digest'
           AND member.value->>'provider_endpoint_identity_digest' = retirement.value->>'provider_endpoint_identity_digest'
       )
     ) THEN
    RAISE EXCEPTION 'authority v7 provider retirement-set cross-binding mismatch' USING ERRCODE = '22023';
  END IF;

  IF jsonb_typeof(v_authorization_envelope) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_authorization_envelope)) <> 9
     OR NOT (v_authorization_envelope ?& ARRAY[
       'schema',
       'body',
       'body_digest',
       'signer_role',
       'signer_key_id',
       'signature_policy_version',
       'trust_root_digest',
       'signature_algorithm',
       'signature'
     ])
     OR convert_from(p_authorization_envelope_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"body":%s,"body_digest":%s,"schema":%s,"signature":%s,"signature_algorithm":%s,"signature_policy_version":%s,"signer_key_id":%s,"signer_role":%s,"trust_root_digest":%s}',
       convert_from(p_authorization_body_jcs, 'UTF8'),
       to_json(v_authorization_envelope->>'body_digest')::text,
       to_json(v_authorization_envelope->>'schema')::text,
       to_json(v_authorization_envelope->>'signature')::text,
       to_json(v_authorization_envelope->>'signature_algorithm')::text,
       to_json(v_authorization_envelope->>'signature_policy_version')::text,
       to_json(v_authorization_envelope->>'signer_key_id')::text,
       to_json(v_authorization_envelope->>'signer_role')::text,
       to_json(v_authorization_envelope->>'trust_root_digest')::text
     )
     OR v_authorization_envelope->>'schema' IS DISTINCT FROM 'authority-protocol-downgrade-authorization.v1'
     OR v_authorization_envelope->'body' IS DISTINCT FROM v_authorization
     OR v_authorization_envelope->>'body_digest' IS DISTINCT FROM encode(p_authorization_digest, 'hex')
     OR v_authorization_envelope->>'signer_role' IS DISTINCT FROM 'authority_protocol_downgrade_authorizer'
     OR v_authorization_envelope->>'signature_policy_version' !~ '^[1-9][0-9]*$'
     OR v_authorization_envelope->>'signature_algorithm' NOT IN ('ed25519', 'ecdsa-p256-sha256')
     OR v_authorization_envelope->>'trust_root_digest' !~ '^[0-9a-f]{64}$'
     OR v_authorization_envelope->>'trust_root_digest' = repeat('0', 64)
     OR length(v_authorization_envelope->>'signer_key_id') NOT BETWEEN 1 AND 128
     OR v_authorization_envelope->>'signer_key_id' ~ '[^ -~]'
     OR v_authorization_envelope->>'signature' !~ '^[A-Za-z0-9_-]{85}[AQgw]$' THEN
    RAISE EXCEPTION 'authority v7 authorization envelope cross-binding mismatch' USING ERRCODE = '22023';
  END IF;

  SELECT count(*) INTO v_provider_namespace_count
  FROM (
    SELECT DISTINCT
      member.value->>'provider_identity_digest' AS provider_identity_digest,
      member.value->>'provider_endpoint_identity_digest' AS provider_endpoint_identity_digest,
      member.value->>'provider_namespace' AS provider_namespace
    FROM jsonb_array_elements(v_retirement_set->'retired_members') AS member(value)
  ) AS provider_namespace_registry;

  IF jsonb_typeof(v_retirement_evidence) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_retirement_evidence)) <> 4
     OR NOT (v_retirement_evidence ?& ARRAY[
       'message_schema',
       'message_body_digest',
       'evidence_count',
       'evidence'
     ])
     OR v_retirement_evidence->>'message_schema' IS DISTINCT FROM 'provider-protocol-downgrade-retirement-set.v1'
     OR v_retirement_evidence->>'message_body_digest' IS DISTINCT FROM encode(p_provider_retirement_set_digest, 'hex')
     OR jsonb_typeof(v_retirement_evidence->'evidence') <> 'array'
     OR v_retirement_evidence->>'evidence_count' IS DISTINCT FROM jsonb_array_length(v_retirement_evidence->'evidence')::text
     OR jsonb_array_length(v_retirement_evidence->'evidence')
          IS DISTINCT FROM (1 + 2 * jsonb_array_length(v_retirement_set->'retirements') + v_provider_namespace_count)
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
       WHERE jsonb_typeof(value) <> 'object'
          OR (SELECT count(*) FROM jsonb_object_keys(value)) <> 5
          OR NOT (value ?& ARRAY[
            'evidence_kind',
            'schema',
            'body_digest',
            'canonical_body_or_null',
            'canonical_envelope_or_null'
          ])
          OR value->>'evidence_kind' IS DISTINCT FROM 'external_signed_envelope'
          OR value->>'schema' NOT IN (
            'environment-inventory-membership-retirement.v1',
            'provider-protocol-downgrade-retirement-authorization.v1',
            'provider-protocol-history-zero-projection.v1',
            'provider-protocol-downgrade-retirement.v1'
          )
          OR octet_length(decode(value->>'body_digest', 'hex')) IS DISTINCT FROM 32
          OR jsonb_typeof(value->'canonical_body_or_null') <> 'null'
          OR jsonb_typeof(value->'canonical_envelope_or_null') <> 'object'
          OR value->'canonical_envelope_or_null'->>'schema' IS DISTINCT FROM value->>'schema'
          OR value->'canonical_envelope_or_null'->>'body_digest' IS DISTINCT FROM value->>'body_digest'
     )
     OR (
       SELECT count(*)
       FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
       WHERE value->>'schema' = 'environment-inventory-membership-retirement.v1'
         AND value->>'body_digest' = v_retirement_set->>'environment_inventory_membership_retirement_digest'
     ) <> 1
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
       WHERE (
         SELECT count(*)
         FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
         WHERE evidence.value->>'schema' = 'provider-protocol-downgrade-retirement.v1'
           AND evidence.value->>'body_digest' = retirement.value->>'provider_protocol_downgrade_retirement_digest'
       ) <> 1
     ) THEN
    RAISE EXCEPTION 'authority v7 retirement evidence cross-binding mismatch' USING ERRCODE = '22023';
  END IF;

  v_evidence_index := 0;
  v_previous_evidence_digest := NULL;
  v_previous_evidence_schema := NULL;
  FOR v_evidence_item IN
    SELECT evidence.value
    FROM jsonb_array_elements(v_retirement_evidence->'evidence') WITH ORDINALITY AS evidence(value, ordinal)
    ORDER BY evidence.ordinal
  LOOP
    v_evidence_index := v_evidence_index + 1;
    v_evidence_schema := v_evidence_item->>'schema';
    v_evidence_digest := decode(v_evidence_item->>'body_digest', 'hex');
    v_evidence_envelope := v_evidence_item->'canonical_envelope_or_null';
    v_evidence_body := v_evidence_envelope->'body';
    v_expected_signer_role := CASE v_evidence_schema
      WHEN 'environment-inventory-membership-retirement.v1' THEN 'release_deployment_operator'
      WHEN 'provider-protocol-downgrade-retirement-authorization.v1' THEN 'provider_downgrade_retirement_admin'
      WHEN 'provider-protocol-history-zero-projection.v1' THEN 'claim_v1_provider_history_auditor'
      WHEN 'provider-protocol-downgrade-retirement.v1' THEN 'claim_v1_provider'
      ELSE NULL
    END;
    IF v_expected_signer_role IS NULL
       OR (v_previous_evidence_digest IS NOT NULL AND (
         v_previous_evidence_digest > v_evidence_digest
         OR (v_previous_evidence_digest = v_evidence_digest AND v_previous_evidence_schema COLLATE "C" >= v_evidence_schema COLLATE "C")
       ))
       OR jsonb_typeof(v_evidence_envelope) <> 'object'
       OR (SELECT count(*) FROM jsonb_object_keys(v_evidence_envelope)) <> 9
       OR NOT (v_evidence_envelope ?& ARRAY[
         'schema',
         'body',
         'body_digest',
         'signer_role',
         'signer_key_id',
         'signature_policy_version',
         'trust_root_digest',
         'signature_algorithm',
         'signature'
       ])
       OR jsonb_typeof(v_evidence_body) <> 'object'
       OR v_evidence_envelope->>'schema' IS DISTINCT FROM v_evidence_schema
       OR v_evidence_envelope->>'body_digest' IS DISTINCT FROM v_evidence_item->>'body_digest'
       OR v_evidence_envelope->>'signer_role' IS DISTINCT FROM v_expected_signer_role
       OR v_evidence_envelope->>'signature_policy_version' !~ '^[1-9][0-9]*$'
       OR v_evidence_envelope->>'signature_algorithm' NOT IN ('ed25519', 'ecdsa-p256-sha256')
       OR v_evidence_envelope->>'trust_root_digest' !~ '^[0-9a-f]{64}$'
       OR v_evidence_envelope->>'trust_root_digest' = repeat('0', 64)
       OR length(v_evidence_envelope->>'signer_key_id') NOT BETWEEN 1 AND 128
       OR v_evidence_envelope->>'signer_key_id' ~ '[^ -~]'
       OR v_evidence_envelope->>'signature' !~ '^[A-Za-z0-9_-]{85}[AQgw]$' THEN
      RAISE EXCEPTION 'authority v7 nested retirement evidence envelope mismatch' USING ERRCODE = '22023';
    END IF;

    CASE v_evidence_schema
      WHEN 'environment-inventory-membership-retirement.v1' THEN
        v_membership_retired_at := (v_evidence_body->>'retired_at')::timestamptz;
        SELECT '[' || COALESCE(string_agg(to_json(member.value->>'environment_record_digest')::text, ',' ORDER BY member.ordinal), '') || ']'
        INTO v_environment_record_digests_jcs
        FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal);
        v_member_set_digest := sha256(
          convert_to('talenro.c12.environment-inventory-member-set.v1', 'UTF8')
          || decode('00', 'hex')
          || convert_to(v_environment_record_digests_jcs, 'UTF8')
        );
        IF (SELECT count(*) FROM jsonb_object_keys(v_evidence_body)) <> 19
           OR NOT (v_evidence_body ?& ARRAY[
             'membership_retirement_id','installation_id','installation_kind','migration_latch_digest','database_identity_digest',
             'release_scope','final_environment_inventory_digest','final_inventory_sequence','final_environment_inventory_anchor_set_digest',
             'environment_member_set_digest','environment_count','authorization_nonce','request_nonce','issued_at','expires_at',
             'request_digest','membership_retirement_tombstone_id','phase','retired_at'
           ])
           OR EXISTS (SELECT 1 FROM jsonb_each(v_evidence_body) AS field(name, value) WHERE jsonb_typeof(value) <> 'string')
           OR v_evidence_body->>'installation_id' IS DISTINCT FROM p_installation_id::text
           OR v_evidence_body->>'installation_kind' IS DISTINCT FROM 'disposable_fixture'
           OR v_evidence_body->>'migration_latch_digest' IS DISTINCT FROM encode(p_migration_latch_digest, 'hex')
           OR v_evidence_body->>'database_identity_digest' IS DISTINCT FROM encode(p_database_identity_digest, 'hex')
           OR v_evidence_body->>'release_scope' IS DISTINCT FROM v_retirement_set->>'release_scope'
           OR v_evidence_body->>'final_environment_inventory_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_digest'
           OR v_evidence_body->>'final_environment_inventory_anchor_set_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_anchor_set_digest'
           OR v_evidence_body->>'environment_member_set_digest' IS DISTINCT FROM encode(v_member_set_digest, 'hex')
           OR v_evidence_body->>'environment_count' IS DISTINCT FROM v_retirement_set->>'retired_member_count'
           OR v_evidence_body->>'phase' IS DISTINCT FROM 'inventory_membership_retired'
           OR v_evidence_body->>'final_inventory_sequence' !~ '^[1-9][0-9]*$'
           OR (v_evidence_body->>'final_inventory_sequence')::numeric > 18446744073709551615
           OR v_evidence_body->>'authorization_nonce' !~ '^[0-9a-f]{64}$'
           OR v_evidence_body->>'request_nonce' !~ '^[0-9a-f]{64}$'
           OR (v_evidence_body->>'membership_retirement_id')::uuid::text IS DISTINCT FROM v_evidence_body->>'membership_retirement_id'
           OR (v_evidence_body->>'membership_retirement_tombstone_id')::uuid::text IS DISTINCT FROM v_evidence_body->>'membership_retirement_tombstone_id'
           OR (v_evidence_body->>'expires_at')::timestamptz <= (v_evidence_body->>'issued_at')::timestamptz
           OR (v_evidence_body->>'expires_at')::timestamptz > (v_evidence_body->>'issued_at')::timestamptz + interval '5 minutes'
           OR (v_evidence_body->>'issued_at')::timestamptz > v_membership_retired_at
           OR v_membership_retired_at >= (v_evidence_body->>'expires_at')::timestamptz
           OR v_membership_retired_at > v_validation_now THEN
          RAISE EXCEPTION 'authority v7 membership-retirement evidence body mismatch' USING ERRCODE = '22023';
        END IF;
        v_membership_request_body_jcs := format(
          '{"authorization_nonce":%s,"database_identity_digest":%s,"environment_count":%s,"environment_member_set_digest":%s,"expires_at":%s,"final_environment_inventory_anchor_set_digest":%s,"final_environment_inventory_digest":%s,"final_inventory_sequence":%s,"installation_id":%s,"installation_kind":%s,"issued_at":%s,"membership_retirement_id":%s,"migration_latch_digest":%s,"release_scope":%s,"request_nonce":%s}',
          to_json(v_evidence_body->>'authorization_nonce')::text,
          to_json(v_evidence_body->>'database_identity_digest')::text,
          to_json(v_evidence_body->>'environment_count')::text,
          to_json(v_evidence_body->>'environment_member_set_digest')::text,
          to_json(v_evidence_body->>'expires_at')::text,
          to_json(v_evidence_body->>'final_environment_inventory_anchor_set_digest')::text,
          to_json(v_evidence_body->>'final_environment_inventory_digest')::text,
          to_json(v_evidence_body->>'final_inventory_sequence')::text,
          to_json(v_evidence_body->>'installation_id')::text,
          to_json(v_evidence_body->>'installation_kind')::text,
          to_json(v_evidence_body->>'issued_at')::text,
          to_json(v_evidence_body->>'membership_retirement_id')::text,
          to_json(v_evidence_body->>'migration_latch_digest')::text,
          to_json(v_evidence_body->>'release_scope')::text,
          to_json(v_evidence_body->>'request_nonce')::text
        );
        IF v_evidence_body->>'request_digest' IS DISTINCT FROM encode(sha256(
          convert_to('talenro.c12.retire-environment-inventory-membership-request.v1', 'UTF8')
          || decode('00', 'hex')
          || convert_to(v_membership_request_body_jcs, 'UTF8')
        ), 'hex') THEN
          RAISE EXCEPTION 'authority v7 membership-retirement request digest binding mismatch' USING ERRCODE = '22023';
        END IF;
        v_membership_count := v_membership_count + 1;
        SELECT '{' || string_agg(to_json(field.name)::text || ':' || field.value::text, ',' ORDER BY field.name COLLATE "C") || '}'
        INTO v_evidence_body_jcs
        FROM jsonb_each(v_evidence_body) AS field(name, value);

      WHEN 'provider-protocol-downgrade-retirement-authorization.v1' THEN
        v_admin_issued_at := (v_evidence_body->>'issued_at')::timestamptz;
        v_admin_expires_at := (v_evidence_body->>'expires_at')::timestamptz;
        v_provider_pair := (v_evidence_body->>'provider_identity_digest') || ':' || (v_evidence_body->>'provider_endpoint_identity_digest');
        SELECT '[' || COALESCE(string_agg(format(
          '{"database_identity_digest":%s,"database_name":%s,"database_oid":%s,"deployment_id":%s,"environment_attestation_digest":%s,"environment_instance_generation":%s,"environment_record_digest":%s,"postgres_system_id":%s,"provider_endpoint_identity_digest":%s,"provider_identity_digest":%s,"provider_namespace":%s,"provider_profile":%s,"timeline":%s}',
          to_json(value->>'database_identity_digest')::text,to_json(value->>'database_name')::text,to_json(value->>'database_oid')::text,
          to_json(value->>'deployment_id')::text,to_json(value->>'environment_attestation_digest')::text,to_json(value->>'environment_instance_generation')::text,
          to_json(value->>'environment_record_digest')::text,to_json(value->>'postgres_system_id')::text,to_json(value->>'provider_endpoint_identity_digest')::text,
          to_json(value->>'provider_identity_digest')::text,to_json(value->>'provider_namespace')::text,to_json(value->>'provider_profile')::text,to_json(value->>'timeline')::text
        ), ',' ORDER BY ordinal), '') || ']'
        INTO v_members_jcs
        FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
        WHERE value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
          AND value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest';
        v_member_set_digest := sha256(
          convert_to('talenro.c12.provider-protocol-retirement-member-set.v1', 'UTF8')
          || decode('00', 'hex')
          || convert_to(v_members_jcs, 'UTF8')
        );
        IF (SELECT count(*) FROM jsonb_object_keys(v_evidence_body)) <> 18
           OR NOT (v_evidence_body ?& ARRAY[
             'authorization_id','retirement_id','installation_id','installation_kind','migration_latch_digest','database_identity_digest',
             'release_scope','environment_inventory_digest','environment_inventory_anchor_set_digest',
             'environment_inventory_membership_retirement_digest','provider_identity_digest','provider_endpoint_identity_digest',
             'retirement_member_set_digest','member_count','authorization_scope','authorization_nonce','issued_at','expires_at'
           ])
           OR EXISTS (SELECT 1 FROM jsonb_each(v_evidence_body) AS field(name, value) WHERE jsonb_typeof(value) <> 'string')
           OR v_admin_evidence ? v_provider_pair
           OR NOT EXISTS (
             SELECT 1 FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
             WHERE value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
               AND value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
           )
           OR v_members_jcs = '[]'
           OR v_evidence_body->>'installation_id' IS DISTINCT FROM p_installation_id::text
           OR v_evidence_body->>'installation_kind' IS DISTINCT FROM 'disposable_fixture'
           OR v_evidence_body->>'migration_latch_digest' IS DISTINCT FROM encode(p_migration_latch_digest, 'hex')
           OR v_evidence_body->>'database_identity_digest' IS DISTINCT FROM encode(p_database_identity_digest, 'hex')
           OR v_evidence_body->>'release_scope' IS DISTINCT FROM v_retirement_set->>'release_scope'
           OR v_evidence_body->>'environment_inventory_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_digest'
           OR v_evidence_body->>'environment_inventory_anchor_set_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_anchor_set_digest'
           OR v_evidence_body->>'environment_inventory_membership_retirement_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_membership_retirement_digest'
           OR v_evidence_body->>'retirement_member_set_digest' IS DISTINCT FROM encode(v_member_set_digest, 'hex')
           OR v_evidence_body->>'member_count' IS DISTINCT FROM jsonb_array_length(convert_from(convert_to(v_members_jcs,'UTF8'),'UTF8')::jsonb)::text
           OR v_evidence_body->>'authorization_scope' IS DISTINCT FROM 'permanent_disposable_namespace_retirement'
           OR v_evidence_body->>'authorization_nonce' !~ '^[0-9a-f]{64}$'
           OR (v_evidence_body->>'authorization_id')::uuid::text IS DISTINCT FROM v_evidence_body->>'authorization_id'
           OR (v_evidence_body->>'retirement_id')::uuid::text IS DISTINCT FROM v_evidence_body->>'retirement_id'
           OR v_admin_expires_at <= v_admin_issued_at
           OR v_admin_expires_at > v_admin_issued_at + interval '5 minutes' THEN
          RAISE EXCEPTION 'authority v7 provider-retirement admin evidence body mismatch' USING ERRCODE = '22023';
        END IF;
        v_admin_evidence := jsonb_set(v_admin_evidence, ARRAY[v_provider_pair], jsonb_build_object(
          'digest', v_evidence_item->>'body_digest',
          'retirement_id', v_evidence_body->>'retirement_id',
          'issued_at', v_evidence_body->>'issued_at',
          'expires_at', v_evidence_body->>'expires_at'
        ), true);
        SELECT '{' || string_agg(to_json(field.name)::text || ':' || field.value::text, ',' ORDER BY field.name COLLATE "C") || '}'
        INTO v_evidence_body_jcs
        FROM jsonb_each(v_evidence_body) AS field(name, value);

      WHEN 'provider-protocol-history-zero-projection.v1' THEN
        v_history_observed_at := (v_evidence_body->>'observed_at')::timestamptz;
        v_history_expires_at := (v_evidence_body->>'expires_at')::timestamptz;
        v_provider_pair := (v_evidence_body->>'provider_identity_digest') || ':' || (v_evidence_body->>'provider_endpoint_identity_digest');
        v_provider_namespace := v_evidence_body->>'namespace';
        v_provider_namespace_key := v_provider_pair || ':' || encode(convert_to(v_provider_namespace, 'UTF8'), 'hex');
        SELECT count(*) INTO v_row_count
        FROM jsonb_array_elements(v_retirement_set->'retired_members') AS member(value)
        WHERE value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
          AND value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
          AND value->>'provider_namespace' = v_provider_namespace;
        IF (SELECT count(*) FROM jsonb_object_keys(v_evidence_body)) <> 32
           OR NOT (v_evidence_body ?& ARRAY[
             'provider_identity_digest','provider_endpoint_identity_digest','namespace','observed_profiles','unknown_profile_count',
             'unknown_record_count','legacy_reservation_count','legacy_terminal_count','legacy_epoch_transition_count','legacy_cutover_count',
             'claim_v1_credential_policy_count','claim_v1_accepted_key_mutation_count','incarnation_registration_count',
             'genesis_preparation_count','genesis_completion_count','genesis_release_preparation_count','genesis_open_count',
             'claim_v1_reservation_count','claim_v1_terminal_count','claim_v1_epoch_transition_count','claim_v1_epoch_recovery_count',
             'serving_lease_event_count','runtime_rebind_count','timeline_lineage_event_count','staging_exclusion_count',
             'staging_recovery_count','source_retirement_count','genesis_authorizing_control_count','other_mutation_count',
             'provider_history_high_water','observed_at','expires_at'
           ])
           OR v_namespace_evidence ? v_provider_namespace_key
           OR v_row_count = 0
           OR v_evidence_body->>'namespace' IS DISTINCT FROM v_provider_namespace
           OR v_evidence_body->'observed_profiles' IS DISTINCT FROM '["claim_v1","legacy_v6"]'::jsonb
           OR EXISTS (
             SELECT 1 FROM jsonb_each(v_evidence_body) AS field(name, value)
             WHERE (name LIKE '%\_count' ESCAPE '\' OR name = 'provider_history_high_water')
               AND (jsonb_typeof(value) <> 'string' OR value #>> '{}' IS DISTINCT FROM '0')
           )
           OR EXISTS (
             SELECT 1 FROM jsonb_each(v_evidence_body) AS field(name, value)
             WHERE name <> 'observed_profiles' AND jsonb_typeof(value) <> 'string'
           )
           OR v_history_expires_at <= v_history_observed_at
           OR v_history_expires_at > v_history_observed_at + interval '2 minutes' THEN
          RAISE EXCEPTION 'authority v7 provider-history-zero evidence body mismatch' USING ERRCODE = '22023';
        END IF;
        v_namespace_evidence := jsonb_set(v_namespace_evidence, ARRAY[v_provider_namespace_key], jsonb_build_object(
          'digest', v_evidence_item->>'body_digest',
          'provider_pair', v_provider_pair,
          'namespace', v_provider_namespace,
          'observed_at', v_evidence_body->>'observed_at',
          'expires_at', v_evidence_body->>'expires_at'
        ), true);
        SELECT '{' || string_agg(
          to_json(field.name)::text || ':' || CASE WHEN field.name = 'observed_profiles' THEN '["claim_v1","legacy_v6"]' ELSE field.value::text END,
          ',' ORDER BY field.name COLLATE "C"
        ) || '}'
        INTO v_evidence_body_jcs
        FROM jsonb_each(v_evidence_body) AS field(name, value);

      WHEN 'provider-protocol-downgrade-retirement.v1' THEN
        v_final_retired_at := (v_evidence_body->>'retired_at')::timestamptz;
        v_provider_pair := (v_evidence_body->>'provider_identity_digest') || ':' || (v_evidence_body->>'provider_endpoint_identity_digest');
        SELECT '[' || COALESCE(string_agg(format(
          '{"database_identity_digest":%s,"database_name":%s,"database_oid":%s,"deployment_id":%s,"environment_attestation_digest":%s,"environment_instance_generation":%s,"environment_record_digest":%s,"postgres_system_id":%s,"provider_endpoint_identity_digest":%s,"provider_identity_digest":%s,"provider_namespace":%s,"provider_profile":%s,"timeline":%s}',
          to_json(value->>'database_identity_digest')::text,to_json(value->>'database_name')::text,to_json(value->>'database_oid')::text,
          to_json(value->>'deployment_id')::text,to_json(value->>'environment_attestation_digest')::text,to_json(value->>'environment_instance_generation')::text,
          to_json(value->>'environment_record_digest')::text,to_json(value->>'postgres_system_id')::text,to_json(value->>'provider_endpoint_identity_digest')::text,
          to_json(value->>'provider_identity_digest')::text,to_json(value->>'provider_namespace')::text,to_json(value->>'provider_profile')::text,to_json(value->>'timeline')::text
        ), ',' ORDER BY ordinal), '') || ']'
        INTO v_members_jcs
        FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
        WHERE value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
          AND value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest';
        SELECT count(DISTINCT member.value->>'provider_namespace')
        INTO v_row_count
        FROM jsonb_array_elements(v_retirement_set->'retired_members') AS member(value)
        WHERE value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
          AND value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest';
        v_member_set_digest := sha256(
          convert_to('talenro.c12.provider-protocol-retirement-member-set.v1', 'UTF8')
          || decode('00', 'hex')
          || convert_to(v_members_jcs, 'UTF8')
        );
        IF (SELECT count(*) FROM jsonb_object_keys(v_evidence_body)) <> 19
           OR NOT (v_evidence_body ?& ARRAY[
             'retirement_id','retirement_nonce','provider_downgrade_retirement_authorization_digest',
             'environment_inventory_membership_retirement_digest','installation_id','release_scope','environment_inventory_digest',
             'environment_inventory_anchor_set_digest','provider_identity_digest','provider_endpoint_identity_digest','member_count',
             'members','retirement_member_set_digest','request_nonce','request_digest','provider_control_sequence','namespace_states','phase','retired_at'
           ])
           OR v_final_evidence ? v_provider_pair
           OR v_row_count = 0
           OR v_members_jcs = '[]'
           OR v_evidence_body->'members' IS DISTINCT FROM convert_from(convert_to(v_members_jcs,'UTF8'),'UTF8')::jsonb
           OR v_evidence_body->>'member_count' IS DISTINCT FROM jsonb_array_length(v_evidence_body->'members')::text
           OR v_evidence_body->>'retirement_member_set_digest' IS DISTINCT FROM encode(v_member_set_digest, 'hex')
           OR v_evidence_body->>'environment_inventory_membership_retirement_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_membership_retirement_digest'
           OR v_evidence_body->>'installation_id' IS DISTINCT FROM p_installation_id::text
           OR v_evidence_body->>'release_scope' IS DISTINCT FROM v_retirement_set->>'release_scope'
           OR v_evidence_body->>'environment_inventory_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_digest'
           OR v_evidence_body->>'environment_inventory_anchor_set_digest' IS DISTINCT FROM v_retirement_set->>'environment_inventory_anchor_set_digest'
           OR v_evidence_body->>'phase' IS DISTINCT FROM 'down_retired'
           OR v_final_retired_at > v_validation_now
           OR v_evidence_body->>'provider_control_sequence' !~ '^(0|[1-9][0-9]*)$'
           OR (v_evidence_body->>'retirement_id')::uuid::text IS DISTINCT FROM v_evidence_body->>'retirement_id'
           OR jsonb_typeof(v_evidence_body->'namespace_states') <> 'array'
           OR jsonb_array_length(v_evidence_body->'namespace_states') <> v_row_count
           OR EXISTS (
             SELECT 1
             FROM jsonb_array_elements(v_evidence_body->'namespace_states') WITH ORDINALITY AS namespace_state(value, ordinal)
             WHERE jsonb_typeof(namespace_state.value) <> 'object'
                OR (SELECT count(*) FROM jsonb_object_keys(namespace_state.value)) <> 7
                OR NOT (namespace_state.value ?& ARRAY[
                  'namespace','environment_record_digests','database_identity_digests','history_zero_projection_digest',
                  'pre_retirement_control_sequence','provider_history_high_water','retirement_tombstone_id'
                ])
                OR namespace_state.value->>'history_zero_projection_digest' !~ '^[0-9a-f]{64}$'
                OR namespace_state.value->>'history_zero_projection_digest' = repeat('0', 64)
                OR namespace_state.value->>'provider_history_high_water' IS DISTINCT FROM '0'
                OR namespace_state.value->>'pre_retirement_control_sequence' !~ '^(0|[1-9][0-9]*)$'
                OR (namespace_state.value->>'retirement_tombstone_id')::uuid::text
                   IS DISTINCT FROM namespace_state.value->>'retirement_tombstone_id'
                OR NOT EXISTS (
                  SELECT 1
                  FROM jsonb_array_elements(v_retirement_set->'retired_members') AS member(value)
                  WHERE member.value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
                    AND member.value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
                    AND member.value->>'provider_namespace' = namespace_state.value->>'namespace'
                )
                OR namespace_state.value->'environment_record_digests' IS DISTINCT FROM (
                  SELECT COALESCE(jsonb_agg(to_jsonb(member.value->>'environment_record_digest') ORDER BY member.ordinal), '[]'::jsonb)
                  FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
                  WHERE member.value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
                    AND member.value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
                    AND member.value->>'provider_namespace' = namespace_state.value->>'namespace'
                )
                OR namespace_state.value->'database_identity_digests' IS DISTINCT FROM (
                  SELECT COALESCE(jsonb_agg(to_jsonb(member.value->>'database_identity_digest') ORDER BY member.ordinal), '[]'::jsonb)
                  FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
                  WHERE member.value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
                    AND member.value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
                    AND member.value->>'provider_namespace' = namespace_state.value->>'namespace'
                )
           )
           OR EXISTS (
             SELECT 1
             FROM (
               SELECT namespace_state.value,
                      lag(namespace_state.value->>'namespace') OVER (ORDER BY namespace_state.ordinal) AS previous_namespace
               FROM jsonb_array_elements(v_evidence_body->'namespace_states') WITH ORDINALITY AS namespace_state(value, ordinal)
             ) AS ordered_namespace
             WHERE ordered_namespace.previous_namespace IS NOT NULL
               AND ordered_namespace.previous_namespace COLLATE "C"
                   >= ordered_namespace.value->>'namespace' COLLATE "C"
           )
           OR NOT EXISTS (
             SELECT 1 FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
             WHERE value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
               AND value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
               AND value->>'provider_protocol_downgrade_retirement_digest' = v_evidence_item->>'body_digest'
           ) THEN
          RAISE EXCEPTION 'authority v7 final provider-retirement evidence body mismatch' USING ERRCODE = '22023';
        END IF;
        v_provider_request_body_jcs := format(
          '{"environment_inventory_anchor_set_digest":%s,"environment_inventory_digest":%s,"environment_inventory_membership_retirement_digest":%s,"installation_id":%s,"member_count":%s,"members":%s,"provider_downgrade_retirement_authorization_digest":%s,"provider_endpoint_identity_digest":%s,"provider_identity_digest":%s,"release_scope":%s,"request_nonce":%s,"retirement_id":%s,"retirement_member_set_digest":%s,"retirement_nonce":%s}',
          to_json(v_evidence_body->>'environment_inventory_anchor_set_digest')::text,
          to_json(v_evidence_body->>'environment_inventory_digest')::text,
          to_json(v_evidence_body->>'environment_inventory_membership_retirement_digest')::text,
          to_json(v_evidence_body->>'installation_id')::text,
          to_json(v_evidence_body->>'member_count')::text,
          v_members_jcs,
          to_json(v_evidence_body->>'provider_downgrade_retirement_authorization_digest')::text,
          to_json(v_evidence_body->>'provider_endpoint_identity_digest')::text,
          to_json(v_evidence_body->>'provider_identity_digest')::text,
          to_json(v_evidence_body->>'release_scope')::text,
          to_json(v_evidence_body->>'request_nonce')::text,
          to_json(v_evidence_body->>'retirement_id')::text,
          to_json(v_evidence_body->>'retirement_member_set_digest')::text,
          to_json(v_evidence_body->>'retirement_nonce')::text
        );
        IF v_evidence_body->>'request_digest' IS DISTINCT FROM encode(sha256(
          convert_to('talenro.c12.retire-provider-protocol-for-downgrade-request.v1', 'UTF8')
          || decode('00', 'hex')
          || convert_to(v_provider_request_body_jcs, 'UTF8')
        ), 'hex') THEN
          RAISE EXCEPTION 'authority v7 provider-retirement request digest binding mismatch' USING ERRCODE = '22023';
        END IF;
        SELECT '[' || COALESCE(string_agg(format(
          '{"database_identity_digests":%s,"environment_record_digests":%s,"history_zero_projection_digest":%s,"namespace":%s,"pre_retirement_control_sequence":%s,"provider_history_high_water":%s,"retirement_tombstone_id":%s}',
          (SELECT '[' || COALESCE(string_agg(to_json(member.value->>'database_identity_digest')::text, ',' ORDER BY member.ordinal), '') || ']'
           FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
           WHERE member.value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
             AND member.value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
             AND member.value->>'provider_namespace' = namespace_state.value->>'namespace'),
          (SELECT '[' || COALESCE(string_agg(to_json(member.value->>'environment_record_digest')::text, ',' ORDER BY member.ordinal), '') || ']'
           FROM jsonb_array_elements(v_retirement_set->'retired_members') WITH ORDINALITY AS member(value, ordinal)
           WHERE member.value->>'provider_identity_digest' = v_evidence_body->>'provider_identity_digest'
             AND member.value->>'provider_endpoint_identity_digest' = v_evidence_body->>'provider_endpoint_identity_digest'
             AND member.value->>'provider_namespace' = namespace_state.value->>'namespace'),
          to_json(namespace_state.value->>'history_zero_projection_digest')::text,
          to_json(namespace_state.value->>'namespace')::text,
          to_json(namespace_state.value->>'pre_retirement_control_sequence')::text,
          to_json(namespace_state.value->>'provider_history_high_water')::text,
          to_json(namespace_state.value->>'retirement_tombstone_id')::text
        ), ',' ORDER BY namespace_state.ordinal), '') || ']'
        INTO v_namespace_states_jcs
        FROM jsonb_array_elements(v_evidence_body->'namespace_states') WITH ORDINALITY AS namespace_state(value, ordinal);
        v_final_evidence := jsonb_set(v_final_evidence, ARRAY[v_provider_pair], jsonb_build_object(
          'digest', v_evidence_item->>'body_digest',
          'authorization_digest', v_evidence_body->>'provider_downgrade_retirement_authorization_digest',
          'retirement_id', v_evidence_body->>'retirement_id',
          'namespace_states', v_evidence_body->'namespace_states',
          'retired_at', v_evidence_body->>'retired_at'
        ), true);
        SELECT '{' || string_agg(
          to_json(field.name)::text || ':' || CASE field.name
            WHEN 'members' THEN v_members_jcs
            WHEN 'namespace_states' THEN v_namespace_states_jcs
            ELSE field.value::text
          END,
          ',' ORDER BY field.name COLLATE "C"
        ) || '}'
        INTO v_evidence_body_jcs
        FROM jsonb_each(v_evidence_body) AS field(name, value);
    END CASE;

    IF EXISTS (
         SELECT 1
         FROM jsonb_each(v_evidence_body) AS field(name, value)
         WHERE name LIKE '%\_digest' ESCAPE '\'
           AND (
             jsonb_typeof(value) <> 'string'
             OR value #>> '{}' !~ '^[0-9a-f]{64}$'
             OR value #>> '{}' = repeat('0', 64)
           )
       )
       OR EXISTS (
         SELECT 1
         FROM jsonb_each(v_evidence_body) AS field(name, value)
         WHERE name LIKE '%\_nonce' ESCAPE '\'
           AND (
             jsonb_typeof(value) <> 'string'
             OR value #>> '{}' !~ '^[0-9a-f]{64}$'
             OR value #>> '{}' = repeat('0', 64)
           )
       )
       OR EXISTS (
         SELECT 1
         FROM jsonb_each(v_evidence_body) AS field(name, value)
         WHERE name LIKE '%\_id' ESCAPE '\'
           AND CASE
             WHEN jsonb_typeof(value) <> 'string' OR value #>> '{}' !~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN true
             ELSE (value #>> '{}')::uuid::text IS DISTINCT FROM value #>> '{}'
               OR (value #>> '{}')::uuid = '00000000-0000-0000-0000-000000000000'::uuid
           END
       )
       OR sha256(
         convert_to('talenro.c12.' || v_evidence_schema, 'UTF8')
         || decode('00', 'hex')
         || convert_to(v_evidence_body_jcs, 'UTF8')
       ) IS DISTINCT FROM v_evidence_digest THEN
      RAISE EXCEPTION 'authority v7 nested retirement evidence body digest mismatch' USING ERRCODE = '22023';
    END IF;

    v_evidence_envelope_jcs := format(
      '{"body":%s,"body_digest":%s,"schema":%s,"signature":%s,"signature_algorithm":%s,"signature_policy_version":%s,"signer_key_id":%s,"signer_role":%s,"trust_root_digest":%s}',
      v_evidence_body_jcs,
      to_json(v_evidence_envelope->>'body_digest')::text,
      to_json(v_evidence_envelope->>'schema')::text,
      to_json(v_evidence_envelope->>'signature')::text,
      to_json(v_evidence_envelope->>'signature_algorithm')::text,
      to_json(v_evidence_envelope->>'signature_policy_version')::text,
      to_json(v_evidence_envelope->>'signer_key_id')::text,
      to_json(v_evidence_envelope->>'signer_role')::text,
      to_json(v_evidence_envelope->>'trust_root_digest')::text
    );
    v_evidence_item_jcs := format(
      '{"body_digest":%s,"canonical_body_or_null":null,"canonical_envelope_or_null":%s,"evidence_kind":"external_signed_envelope","schema":%s}',
      to_json(v_evidence_item->>'body_digest')::text,
      v_evidence_envelope_jcs,
      to_json(v_evidence_item->>'schema')::text
    );
    IF v_evidence_index > 1 THEN
      v_evidence_array_jcs := v_evidence_array_jcs || ',';
    END IF;
    v_evidence_array_jcs := v_evidence_array_jcs || v_evidence_item_jcs;
    v_previous_evidence_digest := v_evidence_digest;
    v_previous_evidence_schema := v_evidence_schema;
  END LOOP;
  v_evidence_array_jcs := v_evidence_array_jcs || ']';
  IF (SELECT count(*) FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
      WHERE value->>'schema' = 'environment-inventory-membership-retirement.v1') <> 1
     OR (SELECT count(*) FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
         WHERE value->>'schema' = 'provider-protocol-downgrade-retirement-authorization.v1')
        <> jsonb_array_length(v_retirement_set->'retirements')
     OR (SELECT count(*) FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
         WHERE value->>'schema' = 'provider-protocol-history-zero-projection.v1')
        <> v_provider_namespace_count
     OR (SELECT count(*) FROM jsonb_array_elements(v_retirement_evidence->'evidence') AS evidence(value)
         WHERE value->>'schema' = 'provider-protocol-downgrade-retirement.v1')
        <> jsonb_array_length(v_retirement_set->'retirements')
     OR v_membership_count <> 1
     OR (SELECT count(*) FROM jsonb_object_keys(v_admin_evidence)) <> jsonb_array_length(v_retirement_set->'retirements')
     OR (SELECT count(*) FROM jsonb_object_keys(v_namespace_evidence)) <> v_provider_namespace_count
     OR (SELECT count(*) FROM jsonb_object_keys(v_final_evidence)) <> jsonb_array_length(v_retirement_set->'retirements')
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
       CROSS JOIN LATERAL (
        SELECT (value->>'provider_identity_digest') || ':' || (value->>'provider_endpoint_identity_digest') AS pair
       ) AS expected
       WHERE NOT (v_admin_evidence ? expected.pair)
          OR NOT (v_final_evidence ? expected.pair)
          OR (v_final_evidence -> expected.pair) ->> 'digest' IS DISTINCT FROM value->>'provider_protocol_downgrade_retirement_digest'
          OR (v_final_evidence -> expected.pair) ->> 'authorization_digest' IS DISTINCT FROM (v_admin_evidence -> expected.pair) ->> 'digest'
          OR (v_final_evidence -> expected.pair) ->> 'retirement_id' IS DISTINCT FROM (v_admin_evidence -> expected.pair) ->> 'retirement_id'
          OR ((v_admin_evidence -> expected.pair) ->> 'issued_at')::timestamptz
             > ((v_final_evidence -> expected.pair) ->> 'retired_at')::timestamptz
          OR ((v_final_evidence -> expected.pair) ->> 'retired_at')::timestamptz
             >= ((v_admin_evidence -> expected.pair) ->> 'expires_at')::timestamptz
     )
     OR EXISTS (
       SELECT 1
       FROM jsonb_array_elements(v_retirement_set->'retirements') AS retirement(value)
       CROSS JOIN LATERAL (
        SELECT (value->>'provider_identity_digest') || ':' || (value->>'provider_endpoint_identity_digest') AS pair
       ) AS expected_pair
       CROSS JOIN LATERAL jsonb_array_elements(
         (v_final_evidence -> expected_pair.pair) -> 'namespace_states'
       ) AS namespace_state(value)
       CROSS JOIN LATERAL (
         SELECT expected_pair.pair || ':' || encode(convert_to(namespace_state.value->>'namespace', 'UTF8'), 'hex') AS namespace_key
       ) AS expected_namespace
       WHERE NOT (v_namespace_evidence ? expected_namespace.namespace_key)
          OR (v_namespace_evidence -> expected_namespace.namespace_key) ->> 'provider_pair'
             IS DISTINCT FROM expected_pair.pair
          OR (v_namespace_evidence -> expected_namespace.namespace_key) ->> 'namespace'
             IS DISTINCT FROM namespace_state.value->>'namespace'
          OR (v_namespace_evidence -> expected_namespace.namespace_key) ->> 'digest'
             IS DISTINCT FROM namespace_state.value->>'history_zero_projection_digest'
          OR ((v_namespace_evidence -> expected_namespace.namespace_key) ->> 'observed_at')::timestamptz
             > ((v_final_evidence -> expected_pair.pair) ->> 'retired_at')::timestamptz
          OR ((v_final_evidence -> expected_pair.pair) ->> 'retired_at')::timestamptz
             >= ((v_namespace_evidence -> expected_namespace.namespace_key) ->> 'expires_at')::timestamptz
     )
     OR convert_from(p_provider_retirement_evidence_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"evidence":%s,"evidence_count":%s,"message_body_digest":%s,"message_schema":%s}',
       v_evidence_array_jcs,
       to_json(v_retirement_evidence->>'evidence_count')::text,
       to_json(v_retirement_evidence->>'message_body_digest')::text,
       to_json(v_retirement_evidence->>'message_schema')::text
     ) THEN
    RAISE EXCEPTION 'authority v7 retirement evidence schema multiset mismatch' USING ERRCODE = '22023';
  END IF;
  IF (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_migration_latches
      WHERE installation_id = p_installation_id
        AND body_digest = p_migration_latch_digest
        AND database_identity_digest = p_database_identity_digest
        AND up_catalog_digest = p_current_catalog_digest
        AND installation_kind = 'disposable_fixture'
        AND migration_version = p_migration_version
        AND down_state = 'locked') <> 1
     OR (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations) <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade control cardinality mismatch' USING ERRCODE = '55000';
  END IF;
  SELECT
      (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_upgrade_intents)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_runtime_registration_results)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_runtime_rebind_results)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_upgrade_attempts)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_intents)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_applications)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_resolutions)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_cancellations)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_terminal_applications)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_recovery_intents)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_epoch_transition_recovery_applications)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_legacy_database_source_retirements)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_indeterminate_source_seals)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_legacy_source_seals)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_fresh_restore_requirements)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_activations)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_activation_completions)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_activation_releases)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_staging_import_capabilities)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_staging_import_capability_revocation_applications)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_staging_import_capability_recovery_intents)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_staging_import_capability_recovery_applications)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_fresh_restore_import_applications)
    + (SELECT count(*) FROM nodecontrol.control_plane_authority_fences WHERE authority_protocol_profile = 'claim_v1')
  INTO v_forbidden_count;
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade catalog is not pristine' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM nodecontrol.control_plane_authority_fences
  WHERE authority_protocol_profile = 'claim_v1'
     OR (
       authority_protocol_profile = 'legacy_v6'
       AND abort_claimed_at IS NULL
       AND protocol_activation_id IS NULL
       AND (
         (provider_status = 'reserved' AND visibility_state = 'fence_pending')
         OR (provider_status = 'committed' AND visibility_state = 'active')
         OR (provider_status = 'aborted' AND visibility_state = 'aborted')
       )
     );
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade fence profile is not pristine' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM (
    SELECT
      create_authority_effect_commitment_jcs AS c01, create_authority_effect_commitment_digest AS c02,
      create_authority_provider_head_jcs AS c03, create_authority_provider_head_digest AS c04,
      create_authority_checkpoint_anchor_jcs AS c05, create_authority_checkpoint_anchor_digest AS c06,
      create_authority_effect_reason AS c07, create_authority_attestation_expires_at AS c08,
      create_authority_activation_deadline AS c09, create_authority_expected_provider_identity_digest AS c10,
      create_authority_activation_evidence_jcs AS c11, create_authority_activation_evidence_digest AS c12,
      create_authority_effect_resolution_jcs AS c13, create_authority_effect_resolution_digest AS c14
    FROM nodecontrol.node_enrollment_grants
    UNION ALL
    SELECT
      claim_authority_effect_commitment_jcs, claim_authority_effect_commitment_digest,
      claim_authority_provider_head_jcs, claim_authority_provider_head_digest,
      claim_authority_checkpoint_anchor_jcs, claim_authority_checkpoint_anchor_digest,
      claim_authority_effect_reason, claim_authority_attestation_expires_at,
      claim_authority_activation_deadline, claim_authority_expected_provider_identity_digest,
      claim_authority_activation_evidence_jcs, claim_authority_activation_evidence_digest,
      claim_authority_effect_resolution_jcs, claim_authority_effect_resolution_digest
    FROM nodecontrol.node_enrollment_grants
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_certificate_issuances
    UNION ALL
    SELECT
      revoke_authority_effect_commitment_jcs, revoke_authority_effect_commitment_digest,
      revoke_authority_provider_head_jcs, revoke_authority_provider_head_digest,
      revoke_authority_checkpoint_anchor_jcs, revoke_authority_checkpoint_anchor_digest,
      revoke_authority_effect_reason, revoke_authority_attestation_expires_at,
      revoke_authority_activation_deadline, revoke_authority_expected_provider_identity_digest,
      revoke_authority_activation_evidence_jcs, revoke_authority_activation_evidence_digest,
      revoke_authority_effect_resolution_jcs, revoke_authority_effect_resolution_digest
    FROM nodecontrol.node_certificates
    UNION ALL
    SELECT
      authority_effect_commitment_jcs, authority_effect_commitment_digest,
      authority_provider_head_jcs, authority_provider_head_digest,
      authority_checkpoint_anchor_jcs, authority_checkpoint_anchor_digest,
      authority_effect_reason, authority_attestation_expires_at,
      authority_activation_deadline, authority_expected_provider_identity_digest,
      authority_activation_evidence_jcs, authority_activation_evidence_digest,
      authority_effect_resolution_jcs, authority_effect_resolution_digest
    FROM nodecontrol.node_state_transitions
    UNION ALL
    SELECT
      open_authority_effect_commitment_jcs, open_authority_effect_commitment_digest,
      open_authority_provider_head_jcs, open_authority_provider_head_digest,
      open_authority_checkpoint_anchor_jcs, open_authority_checkpoint_anchor_digest,
      open_authority_effect_reason, open_authority_attestation_expires_at,
      open_authority_activation_deadline, open_authority_expected_provider_identity_digest,
      open_authority_activation_evidence_jcs, open_authority_activation_evidence_digest,
      open_authority_effect_resolution_jcs, open_authority_effect_resolution_digest
    FROM nodecontrol.node_security_incidents
    UNION ALL
    SELECT
      resolve_authority_effect_commitment_jcs, resolve_authority_effect_commitment_digest,
      resolve_authority_provider_head_jcs, resolve_authority_provider_head_digest,
      resolve_authority_checkpoint_anchor_jcs, resolve_authority_checkpoint_anchor_digest,
      resolve_authority_effect_reason, resolve_authority_attestation_expires_at,
      resolve_authority_activation_deadline, resolve_authority_expected_provider_identity_digest,
      resolve_authority_activation_evidence_jcs, resolve_authority_activation_evidence_digest,
      resolve_authority_effect_resolution_jcs, resolve_authority_effect_resolution_digest
    FROM nodecontrol.node_security_incidents
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_resource_envelopes
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_state_signing_intents
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_root_metadata_publish_intents
  ) AS proof_group
  WHERE pg_catalog.num_nonnulls(
    proof_group.c01, proof_group.c02, proof_group.c03, proof_group.c04,
    proof_group.c05, proof_group.c06, proof_group.c07, proof_group.c08,
    proof_group.c09, proof_group.c10, proof_group.c11, proof_group.c12,
    proof_group.c13, proof_group.c14
  ) <> 0;
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade proof fields are not pristine' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM nodecontrol.node_inventory
  WHERE identity_state = 'unauthorized';
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade unauthorized inventory is not pristine' USING ERRCODE = '55000';
  END IF;

  INSERT INTO nodecontrol.control_plane_authority_protocol_downgrade_authorizations (
    authorization_id, installation_id, migration_latch_digest, database_identity_digest,
    migration_version, current_catalog_digest, pristine_downgrade_inventory_digest,
    provider_protocol_downgrade_retirement_set_digest, environment_inventory_anchor_set_digest,
    database_transaction_id, transaction_nonce, authorization_scope, issued_at, expires_at,
    authorization_envelope_jcs, pristine_inventory_body_jcs, provider_retirement_set_body_jcs, provider_retirement_evidence_jcs,
    canonical_body_jcs, body_digest
  ) VALUES (
    p_authorization_id, p_installation_id, p_migration_latch_digest, p_database_identity_digest,
    p_migration_version, p_current_catalog_digest, p_pristine_inventory_digest,
    p_provider_retirement_set_digest, p_environment_anchor_set_digest,
    p_database_transaction_id, p_transaction_nonce, p_authorization_scope, p_issued_at, p_expires_at,
    p_authorization_envelope_jcs, p_pristine_inventory_body_jcs, p_provider_retirement_set_body_jcs, p_provider_retirement_evidence_jcs,
    p_authorization_body_jcs, p_authorization_digest
  );
  RETURN QUERY SELECT
    p_pristine_inventory_digest,
    p_migration_latch_digest,
    p_authorization_digest,
    p_provider_retirement_set_digest,
    p_database_transaction_id,
    p_transaction_nonce,
    1::bigint,
    1::bigint,
    v_non_control_count,
    'authorization_inserted'::text,
    transaction_timestamp();
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.v7_consume_down_guard(
  p_installation_id uuid,
  p_migration_latch_digest bytea,
  p_authorization_digest bytea,
  p_transaction_nonce bytea,
  p_authorized_state_body_jcs bytea,
  p_authorized_state_digest bytea
)
RETURNS TABLE (
  pristine_downgrade_inventory_digest bytea,
  authorized_state_digest bytea,
  consumed_migration_latch_digest bytea,
  consumed_downgrade_authorization_digest bytea,
  database_transaction_id numeric,
  transaction_nonce bytea,
  post_consume_migration_latch_count bigint,
  post_consume_downgrade_authorization_count bigint,
  non_control_protocol_row_count bigint,
  stage text,
  consumed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  v_forbidden_count bigint;
  v_txid numeric;
  v_deleted_authorization bigint;
  v_deleted_latch bigint;
  v_pristine_inventory_digest bytea;
  v_table_name text;
  v_classification text;
  v_row_count bigint;
  v_empty_digest bytea;
  v_non_control_count bigint := 0;
  v_registry_tables text[] := ARRAY[]::text[];
  v_select_tables text[];
  v_lock_tables text[];
  v_select_oids oid[];
  v_lock_oids oid[];
  v_bootstrap_oid oid;
  v_acl_count bigint;
  v_drift_count bigint;
  v_authorized_state jsonb;
  v_pristine jsonb;
  v_inventory jsonb := '[]'::jsonb;
  v_authorized_state_raw json;
  v_pristine_raw json;
  v_inventory_jcs text := '[';
  v_pristine_inventory_body_jcs bytea;
  v_provider_retirement_set_digest bytea;
  v_authorization_id uuid;
  v_database_identity_digest bytea;
  v_current_catalog_digest bytea;
  v_migration_version bigint;
  v_authorization_scope text;
  v_issued_at timestamptz;
  v_expires_at timestamptz;
  v_post_latch_count bigint;
  v_post_authorization_count bigint;
  v_post_non_control_count bigint;
  v_candidate_count bigint;
BEGIN
  PERFORM nodecontrol.v7_require_role('nodecontrol_migration_downgrader');
  LOCK TABLE
    nodecontrol.control_plane_authority_protocol_migration_latches,
    nodecontrol.control_plane_authority_protocol_downgrade_authorizations,
    nodecontrol.control_plane_authority_epoch_transition_applications,
    nodecontrol.control_plane_authority_epoch_transition_cancellations,
    nodecontrol.control_plane_authority_epoch_transition_intents,
    nodecontrol.control_plane_authority_epoch_transition_recovery_applications,
    nodecontrol.control_plane_authority_epoch_transition_recovery_intents,
    nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions,
    nodecontrol.control_plane_authority_epoch_transition_resolutions,
    nodecontrol.control_plane_authority_epoch_transition_terminal_applications,
    nodecontrol.control_plane_authority_fresh_restore_import_applications,
    nodecontrol.control_plane_authority_fresh_restore_requirements,
    nodecontrol.control_plane_authority_indeterminate_source_seals,
    nodecontrol.control_plane_authority_legacy_database_source_retirements,
    nodecontrol.control_plane_authority_legacy_source_seals,
    nodecontrol.control_plane_authority_protocol_activation_completions,
    nodecontrol.control_plane_authority_protocol_activation_releases,
    nodecontrol.control_plane_authority_protocol_activations,
    nodecontrol.control_plane_authority_protocol_upgrade_attempts,
    nodecontrol.control_plane_authority_protocol_upgrade_intents,
    nodecontrol.control_plane_authority_runtime_rebind_results,
    nodecontrol.control_plane_authority_runtime_registration_results,
    nodecontrol.control_plane_authority_staging_import_capabilities,
    nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
    nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
    nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
    nodecontrol.control_plane_authority_fences,
    nodecontrol.control_plane_trust_bundle_high_waters,
    nodecontrol.node_capacity_profiles,
    nodecontrol.node_certificate_issuances,
    nodecontrol.node_certificates,
    nodecontrol.node_desired_states,
    nodecontrol.node_endpoints,
    nodecontrol.node_enrollment_grants,
    nodecontrol.node_failure_domain_membership,
    nodecontrol.node_failure_domains,
    nodecontrol.node_inventory,
    nodecontrol.node_observed_states,
    nodecontrol.node_operator_audit,
    nodecontrol.node_pops,
    nodecontrol.node_process_slots,
    nodecontrol.node_recovery_sessions,
    nodecontrol.node_recovery_states,
    nodecontrol.node_resource_envelopes,
    nodecontrol.node_restore_reauthorization_approvals,
    nodecontrol.node_root_metadata_publish_intents,
    nodecontrol.node_root_metadata_signature_shares,
    nodecontrol.node_security_fault_receipts,
    nodecontrol.node_security_incidents,
    nodecontrol.node_state_signing_intents,
    nodecontrol.node_state_transitions,
    public.goose_db_version IN ACCESS EXCLUSIVE MODE;

  PERFORM nodecontrol.v7_require_role('nodecontrol_migration_downgrader');

  FOR v_table_name, v_classification IN
    SELECT table_name, classification
    FROM (VALUES
      ('nodecontrol.control_plane_authority_epoch_transition_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_cancellations', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_recovery_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_recovery_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_resolutions', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_epoch_transition_terminal_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_fences', 'base_v6'),
      ('nodecontrol.control_plane_authority_fresh_restore_import_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_fresh_restore_requirements', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_indeterminate_source_seals', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_legacy_database_source_retirements', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_legacy_source_seals', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_activation_completions', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_activation_releases', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_activations', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_upgrade_attempts', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_protocol_upgrade_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_runtime_rebind_results', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_runtime_registration_results', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capabilities', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capability_recovery_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capability_recovery_intents', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_authority_staging_import_capability_revocation_applications', 'authority_v7_non_control'),
      ('nodecontrol.control_plane_trust_bundle_high_waters', 'base_v6'),
      ('nodecontrol.node_capacity_profiles', 'base_v6'),
      ('nodecontrol.node_certificate_issuances', 'base_v6'),
      ('nodecontrol.node_certificates', 'base_v6'),
      ('nodecontrol.node_desired_states', 'base_v6'),
      ('nodecontrol.node_endpoints', 'base_v6'),
      ('nodecontrol.node_enrollment_grants', 'base_v6'),
      ('nodecontrol.node_failure_domain_membership', 'base_v6'),
      ('nodecontrol.node_failure_domains', 'base_v6'),
      ('nodecontrol.node_inventory', 'base_v6'),
      ('nodecontrol.node_observed_states', 'base_v6'),
      ('nodecontrol.node_operator_audit', 'base_v6'),
      ('nodecontrol.node_pops', 'base_v6'),
      ('nodecontrol.node_process_slots', 'base_v6'),
      ('nodecontrol.node_recovery_sessions', 'base_v6'),
      ('nodecontrol.node_recovery_states', 'base_v6'),
      ('nodecontrol.node_resource_envelopes', 'base_v6'),
      ('nodecontrol.node_restore_reauthorization_approvals', 'base_v6'),
      ('nodecontrol.node_root_metadata_publish_intents', 'base_v6'),
      ('nodecontrol.node_root_metadata_signature_shares', 'base_v6'),
      ('nodecontrol.node_security_fault_receipts', 'base_v6'),
      ('nodecontrol.node_security_incidents', 'base_v6'),
      ('nodecontrol.node_state_signing_intents', 'base_v6'),
      ('nodecontrol.node_state_transitions', 'base_v6')
    ) AS registry(table_name, classification)
  LOOP
    EXECUTE format('SELECT count(*) FROM %s', v_table_name) INTO v_row_count;
    IF v_row_count <> 0 THEN
      RAISE EXCEPTION 'authority v7 pristine registry relation % is nonempty', v_table_name USING ERRCODE = '55000';
    END IF;
    v_registry_tables := array_append(v_registry_tables, v_table_name);
    v_non_control_count := v_non_control_count + v_row_count;
    v_empty_digest := sha256(
      convert_to('talenro.c12.pristine-downgrade-empty-table.v1', 'UTF8')
      || decode('00', 'hex')
      || convert_to(
        format(
          '{"classification":"%s","row_count":"0","table_name":"%s"}',
          v_classification,
          v_table_name
        ),
        'UTF8'
      )
    );
    IF octet_length(v_empty_digest) <> 32 THEN
      RAISE EXCEPTION 'authority v7 pristine registry digest failure' USING ERRCODE = '55000';
    END IF;
    v_inventory := v_inventory || jsonb_build_array(jsonb_build_object(
      'table_name', v_table_name,
      'classification', v_classification,
      'row_count', '0',
      'content_digest', encode(v_empty_digest, 'hex')
    ));
    v_inventory_jcs := v_inventory_jcs
      || CASE WHEN array_length(v_registry_tables, 1) > 1 THEN ',' ELSE '' END
      || format(
        '{"classification":%s,"content_digest":%s,"row_count":"0","table_name":%s}',
        to_json(v_classification)::text,
        to_json(encode(v_empty_digest, 'hex'))::text,
        to_json(v_table_name)::text
      );
  END LOOP;
  v_inventory_jcs := v_inventory_jcs || ']';
  v_select_tables := ARRAY[
    'nodecontrol.control_plane_authority_protocol_migration_latches',
    'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'
  ] || v_registry_tables;
  v_lock_tables := v_select_tables || ARRAY['public.goose_db_version'];
  SELECT array_agg(table_name::regclass::oid ORDER BY ordinal)
  INTO v_select_oids
  FROM unnest(v_select_tables) WITH ORDINALITY AS selected(table_name, ordinal);
  SELECT array_agg(table_name::regclass::oid ORDER BY ordinal)
  INTO v_lock_oids
  FROM unnest(v_lock_tables) WITH ORDINALITY AS locked(table_name, ordinal);
  SELECT oid INTO v_bootstrap_oid
  FROM pg_catalog.pg_roles
  WHERE rolname = session_user AND rolsuper;
  IF v_bootstrap_oid IS NULL
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_database
       WHERE datname = current_database() AND datdba = v_bootstrap_oid
     )
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_namespace
       WHERE nspname = 'nodecontrol' AND nspowner = v_bootstrap_oid
     ) THEN
    RAISE EXCEPTION 'authority v7 Down requires the exact bootstrap owner' USING ERRCODE = '42501';
  END IF;

  SELECT count(*) INTO v_acl_count
  FROM pg_catalog.pg_roles
  WHERE rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  )
    AND NOT rolcanlogin
    AND NOT rolsuper
    AND NOT rolcreatedb
    AND NOT rolcreaterole
    AND NOT rolreplication
    AND NOT rolbypassrls
    AND rolconfig IS NULL;
  SELECT count(*) INTO v_drift_count
  FROM pg_catalog.pg_auth_members AS membership
  WHERE membership.member IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    )
     OR membership.roleid IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    );
  IF v_acl_count <> 3 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_acl_count
  FROM pg_catalog.pg_proc AS restricted_function
  JOIN pg_catalog.pg_roles AS owner_role ON owner_role.oid = restricted_function.proowner
  WHERE (restricted_function.oid =
      'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure
      AND owner_role.rolname = 'nodecontrol_upgrade_executor')
     OR (restricted_function.oid IN (
      'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
      'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure
    ) AND owner_role.rolname = 'nodecontrol_migration_downgrader')
     OR (restricted_function.oid =
      'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure
      AND owner_role.rolname = 'nodecontrol_staging_importer');
  IF v_acl_count <> 4 THEN
    RAISE EXCEPTION 'authority v7 restricted function owner drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_drift_count
  FROM pg_catalog.pg_class AS relation
  WHERE relation.relowner IN (
    SELECT oid FROM pg_catalog.pg_roles
    WHERE rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
  );
  v_drift_count := v_drift_count + (
    SELECT count(*) FROM pg_catalog.pg_namespace AS namespace
    WHERE namespace.nspowner IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    )
  ) + (
    SELECT count(*) FROM pg_catalog.pg_database AS database
    WHERE database.datdba IN (
      SELECT oid FROM pg_catalog.pg_roles
      WHERE rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
      )
    )
  ) + (
    SELECT count(*) FROM pg_catalog.pg_proc AS proc
    JOIN pg_catalog.pg_roles AS owner ON owner.oid = proc.proowner
    WHERE owner.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
      AND proc.oid <> ALL (ARRAY[
        'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid,
        'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
        'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
        'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid
      ])
  );
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability role foreign ownership drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*) FILTER (WHERE acl.grantee = v_bootstrap_oid AND acl.privilege_type = 'EXECUTE' AND NOT acl.is_grantable),
    count(*) FILTER (
      WHERE acl.privilege_type <> 'EXECUTE'
         OR acl.is_grantable
         OR acl.grantee NOT IN (proc.proowner, v_bootstrap_oid)
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_proc AS proc
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
  ) AS acl
  WHERE proc.oid = ANY (ARRAY[
    'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure::oid,
    'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure::oid,
    'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure::oid
  ]);
  IF v_acl_count <> 4 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 entry function ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*) FILTER (
      WHERE grantee.rolname IN ('nodecontrol_upgrade_executor', 'nodecontrol_migration_downgrader', 'nodecontrol_staging_importer')
        AND acl.privilege_type = 'EXECUTE'
        AND NOT acl.is_grantable
    ),
    count(*) FILTER (
      WHERE acl.privilege_type <> 'EXECUTE'
         OR acl.is_grantable
         OR acl.grantee NOT IN (
           proc.proowner,
           'nodecontrol_upgrade_executor'::regrole::oid,
           'nodecontrol_migration_downgrader'::regrole::oid,
           'nodecontrol_staging_importer'::regrole::oid
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_proc AS proc
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
  ) AS acl
  LEFT JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = acl.grantee
  WHERE proc.oid = 'nodecontrol.v7_require_role(name)'::regprocedure;
  IF v_acl_count <> 3 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 role helper ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR NOT (
           (acl.privilege_type = 'SELECT' AND relation.oid = ANY (v_select_oids))
           OR (acl.privilege_type = 'MAINTAIN' AND relation.oid = ANY (v_lock_oids))
           OR (acl.privilege_type = 'INSERT' AND relation.oid =
             'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass)
           OR (acl.privilege_type = 'DELETE' AND relation.oid IN (
             'nodecontrol.control_plane_authority_protocol_migration_latches'::regclass,
             'nodecontrol.control_plane_authority_protocol_downgrade_authorizations'::regclass
           ))
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_class AS relation
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(relation.relacl, pg_catalog.acldefault('r', relation.relowner))
  ) AS acl
  WHERE acl.grantee = 'nodecontrol_migration_downgrader'::regrole;
  IF v_acl_count <> 106 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrader table ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR NOT (
           (acl.privilege_type = 'SELECT' AND relation.oid IN (
             'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass,
             'nodecontrol.control_plane_authority_staging_import_capability_recovery_intents'::regclass,
             'nodecontrol.control_plane_authority_staging_import_capability_recovery_applications'::regclass,
             'nodecontrol.control_plane_authority_staging_import_capability_revocation_applications'::regclass,
             'nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass
           ))
           OR (acl.privilege_type = 'INSERT' AND relation.oid IN (
             'nodecontrol.node_pops'::regclass,
             'nodecontrol.node_failure_domains'::regclass,
             'nodecontrol.node_capacity_profiles'::regclass,
             'nodecontrol.node_inventory'::regclass,
             'nodecontrol.control_plane_authority_fresh_restore_import_applications'::regclass
           ))
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_class AS relation
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(relation.relacl, pg_catalog.acldefault('r', relation.relowner))
  ) AS acl
  WHERE acl.grantee = 'nodecontrol_staging_importer'::regrole;
  IF v_acl_count <> 10 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 staging table ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR acl.privilege_type <> 'UPDATE'
         OR relation.oid <> 'nodecontrol.control_plane_authority_staging_import_capabilities'::regclass
         OR attribute.attname <> 'body_digest'
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_attribute AS attribute
  JOIN pg_catalog.pg_class AS relation ON relation.oid = attribute.attrelid
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(attribute.attacl) AS acl
  WHERE attribute.attacl IS NOT NULL
    AND acl.grantee = 'nodecontrol_staging_importer'::regrole;
  IF v_acl_count <> 1 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 staging column ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT
    count(*),
    count(*) FILTER (
      WHERE acl.is_grantable
         OR acl.privilege_type <> 'USAGE'
         OR NOT (
           (grantee.rolname = 'nodecontrol_upgrade_executor' AND namespace.nspname = 'nodecontrol')
           OR (grantee.rolname = 'nodecontrol_migration_downgrader' AND namespace.nspname IN ('nodecontrol', 'public'))
           OR (grantee.rolname = 'nodecontrol_staging_importer' AND namespace.nspname = 'nodecontrol')
         )
    )
  INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_namespace AS namespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(namespace.nspacl, pg_catalog.acldefault('n', namespace.nspowner))
  ) AS acl
  JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = acl.grantee
  WHERE grantee.rolname IN (
    'nodecontrol_upgrade_executor',
    'nodecontrol_migration_downgrader',
    'nodecontrol_staging_importer'
  );
  IF v_acl_count <> 4 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability schema ACL drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_drift_count
  FROM pg_catalog.pg_proc AS proc
  JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = proc.pronamespace
  CROSS JOIN LATERAL pg_catalog.aclexplode(
    COALESCE(proc.proacl, pg_catalog.acldefault('f', proc.proowner))
  ) AS acl
  JOIN pg_catalog.pg_roles AS grantee ON grantee.oid = acl.grantee
  WHERE grantee.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
    AND NOT (
      acl.privilege_type = 'EXECUTE'
      AND NOT acl.is_grantable
      AND (
        (grantee.rolname = 'nodecontrol_upgrade_executor'
          AND proc.oid IN (
            'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure,
            'nodecontrol.v7_require_role(name)'::regprocedure
          ))
        OR (grantee.rolname = 'nodecontrol_migration_downgrader'
          AND proc.oid IN (
            'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
            'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
            'nodecontrol.v7_require_role(name)'::regprocedure
          ))
        OR (grantee.rolname = 'nodecontrol_staging_importer'
          AND proc.oid IN (
            'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure,
            'nodecontrol.v7_require_role(name)'::regprocedure
          ))
      )
    );
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability function ACL drift' USING ERRCODE = '55000';
  END IF;
  SELECT count(*), count(*) FILTER (WHERE NOT (
    dependency.dbid = (SELECT oid FROM pg_catalog.pg_database WHERE datname = current_database())
    AND dependency.classid = 'pg_catalog.pg_proc'::regclass
    AND dependency.objsubid = 0
    AND (
      (role_catalog.rolname = 'nodecontrol_upgrade_executor'
        AND dependency.objid = 'nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure)
      OR (role_catalog.rolname = 'nodecontrol_migration_downgrader' AND dependency.objid IN (
        'nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure,
        'nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure
      ))
      OR (role_catalog.rolname = 'nodecontrol_staging_importer' AND dependency.objid = 'nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure)
    )
  )) INTO v_acl_count, v_drift_count
  FROM pg_catalog.pg_shdepend AS dependency
  JOIN pg_catalog.pg_roles AS role_catalog ON role_catalog.oid = dependency.refobjid
  WHERE dependency.refclassid = 'pg_catalog.pg_authid'::regclass
    AND dependency.deptype = 'o'
    AND role_catalog.rolname IN ('nodecontrol_upgrade_executor','nodecontrol_migration_downgrader','nodecontrol_staging_importer');
  IF v_acl_count <> 4 OR v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 capability foreign ownership drift' USING ERRCODE = '55000';
  END IF;
  SELECT count(*) INTO v_drift_count
  FROM (
    VALUES
      ('nodecontrol.v7_text_array_is_sorted_unique(text[])'::regprocedure, session_user::text, 'sql', 'i', false),
      ('nodecontrol.v7_reject_immutable_mutation()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_require_role(name)'::regprocedure, session_user::text, 'plpgsql', 's', false),
      ('nodecontrol.v7_acquire_source_freeze_for_seal()'::regprocedure, 'nodecontrol_upgrade_executor', 'plpgsql', 'v', true),
      ('nodecontrol.v7_source_is_frozen()'::regprocedure, session_user::text, 'sql', 'v', true),
      ('nodecontrol.v7_assert_source_writable()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_assert_activation_barrier()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_authority_proof_group_valid(bytea,bytea,bytea,bytea,bytea,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea)'::regprocedure, session_user::text, 'sql', 'i', false),
      ('nodecontrol.v7_guard_authority_proof_transition()'::regprocedure, session_user::text, 'plpgsql', 'v', true),
      ('nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea)'::regprocedure, 'nodecontrol_migration_downgrader', 'plpgsql', 'v', true),
      ('nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea)'::regprocedure, 'nodecontrol_migration_downgrader', 'plpgsql', 'v', true),
      ('nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb)'::regprocedure, 'nodecontrol_staging_importer', 'plpgsql', 'v', true)
  ) AS expected(proc_oid, owner_name, language_name, volatility_code, security_definer)
  JOIN pg_catalog.pg_proc AS proc ON proc.oid = expected.proc_oid
  JOIN pg_catalog.pg_roles AS owner ON owner.oid = proc.proowner
  JOIN pg_catalog.pg_language AS language ON language.oid = proc.prolang
  WHERE owner.rolname IS DISTINCT FROM expected.owner_name
     OR language.lanname IS DISTINCT FROM expected.language_name
     OR proc.provolatile IS DISTINCT FROM expected.volatility_code::"char"
     OR proc.prosecdef IS DISTINCT FROM expected.security_definer
     OR proc.proconfig IS DISTINCT FROM ARRAY['search_path=pg_catalog, nodecontrol']::text[];
  IF v_drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 function metadata drift' USING ERRCODE = '55000';
  END IF;
    v_txid := txid_current();

    SELECT count(*)
    INTO v_drift_count
    FROM (
        SELECT acl.grantee
        FROM pg_catalog.pg_database AS database
        CROSS JOIN LATERAL pg_catalog.aclexplode(
            COALESCE(
                database.datacl,
                pg_catalog.acldefault('d'::"char", database.datdba)
            )
        ) AS acl
        WHERE database.datname = pg_catalog.current_database()

        UNION ALL

        SELECT acl.grantee
        FROM pg_catalog.pg_type AS type
        CROSS JOIN LATERAL pg_catalog.aclexplode(
            COALESCE(
                type.typacl,
                pg_catalog.acldefault('T'::"char", type.typowner)
            )
        ) AS acl
        WHERE type.typnamespace = 'nodecontrol'::pg_catalog.regnamespace

        UNION ALL

        SELECT acl.grantee
        FROM pg_catalog.pg_class AS relation
        CROSS JOIN LATERAL pg_catalog.aclexplode(
            COALESCE(
                relation.relacl,
                pg_catalog.acldefault('S'::"char", relation.relowner)
            )
        ) AS acl
        WHERE relation.relnamespace = 'nodecontrol'::pg_catalog.regnamespace
          AND relation.relkind = 'S'
    ) AS direct_acl
    JOIN pg_catalog.pg_roles AS grantee
      ON grantee.oid = direct_acl.grantee
    WHERE grantee.rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
    );

    IF v_drift_count <> 0 THEN
        RAISE EXCEPTION 'authority v7 capability database/type/sequence ACL drift'
            USING ERRCODE = '55000';
    END IF;

    SELECT count(*)
    INTO v_drift_count
    FROM (
        SELECT relation.relowner AS owner_oid, acl.grantor, acl.grantee
        FROM pg_catalog.pg_class AS relation
        CROSS JOIN LATERAL pg_catalog.aclexplode(
            COALESCE(
                relation.relacl,
                pg_catalog.acldefault(
                    CASE
                        WHEN relation.relkind = 'S' THEN 'S'::"char"
                        ELSE 'r'::"char"
                    END,
                    relation.relowner
                )
            )
        ) AS acl
        WHERE relation.relnamespace = 'nodecontrol'::pg_catalog.regnamespace

        UNION ALL

        SELECT relation.relowner AS owner_oid, acl.grantor, acl.grantee
        FROM pg_catalog.pg_class AS relation
        JOIN pg_catalog.pg_attribute AS attribute
          ON attribute.attrelid = relation.oid
         AND attribute.attnum > 0
         AND NOT attribute.attisdropped
        CROSS JOIN LATERAL pg_catalog.aclexplode(attribute.attacl) AS acl
        WHERE attribute.attacl IS NOT NULL
          AND relation.relnamespace = 'nodecontrol'::pg_catalog.regnamespace

        UNION ALL

        SELECT namespace.nspowner AS owner_oid, acl.grantor, acl.grantee
        FROM pg_catalog.pg_namespace AS namespace
        CROSS JOIN LATERAL pg_catalog.aclexplode(
            COALESCE(
                namespace.nspacl,
                pg_catalog.acldefault('n'::"char", namespace.nspowner)
            )
        ) AS acl
        WHERE namespace.oid = 'nodecontrol'::pg_catalog.regnamespace

        UNION ALL

        SELECT procedure.proowner AS owner_oid, acl.grantor, acl.grantee
        FROM pg_catalog.pg_proc AS procedure
        CROSS JOIN LATERAL pg_catalog.aclexplode(
            COALESCE(
                procedure.proacl,
                pg_catalog.acldefault('f'::"char", procedure.proowner)
            )
        ) AS acl
        WHERE procedure.pronamespace = 'nodecontrol'::pg_catalog.regnamespace
    ) AS granted
    JOIN pg_catalog.pg_roles AS grantee
      ON grantee.oid = granted.grantee
    WHERE grantee.rolname IN (
        'nodecontrol_upgrade_executor',
        'nodecontrol_migration_downgrader',
        'nodecontrol_staging_importer'
    )
      AND granted.grantor <> granted.owner_oid;

    IF v_drift_count <> 0 THEN
        RAISE EXCEPTION 'authority v7 capability ACL grantor drift'
            USING ERRCODE = '55000';
    END IF;
  IF p_installation_id IS NULL
     OR p_migration_latch_digest IS NULL
     OR p_authorization_digest IS NULL
     OR p_transaction_nonce IS NULL
     OR p_authorized_state_body_jcs IS NULL
     OR p_authorized_state_digest IS NULL
     OR octet_length(p_migration_latch_digest) IS DISTINCT FROM 32
     OR octet_length(p_authorization_digest) IS DISTINCT FROM 32
     OR octet_length(p_transaction_nonce) IS DISTINCT FROM 32
     OR octet_length(p_authorized_state_digest) IS DISTINCT FROM 32
     OR octet_length(p_authorized_state_body_jcs) NOT BETWEEN 1 AND 1048576 THEN
    RAISE EXCEPTION 'invalid authority v7 authorized-state scalar binding' USING ERRCODE = '22023';
  END IF;

  SELECT
    downgrade_authorization.authorization_id,
    downgrade_authorization.pristine_downgrade_inventory_digest,
    downgrade_authorization.pristine_inventory_body_jcs,
    downgrade_authorization.provider_protocol_downgrade_retirement_set_digest,
    downgrade_authorization.database_identity_digest,
    downgrade_authorization.current_catalog_digest,
    downgrade_authorization.migration_version,
    downgrade_authorization.authorization_scope,
    downgrade_authorization.issued_at,
    downgrade_authorization.expires_at
  INTO
    v_authorization_id,
    v_pristine_inventory_digest,
    v_pristine_inventory_body_jcs,
    v_provider_retirement_set_digest,
    v_database_identity_digest,
    v_current_catalog_digest,
    v_migration_version,
    v_authorization_scope,
    v_issued_at,
    v_expires_at
  FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations AS downgrade_authorization
  JOIN nodecontrol.control_plane_authority_protocol_migration_latches AS latch
    ON latch.installation_id = downgrade_authorization.installation_id
  WHERE downgrade_authorization.installation_id = p_installation_id
    AND downgrade_authorization.body_digest = p_authorization_digest
    AND downgrade_authorization.migration_latch_digest = p_migration_latch_digest
    AND downgrade_authorization.transaction_nonce = p_transaction_nonce
    AND downgrade_authorization.database_transaction_id = v_txid
    AND downgrade_authorization.migration_version = 7
    AND downgrade_authorization.authorization_scope = 'down_00007_only'
    AND downgrade_authorization.issued_at <= clock_timestamp()
    AND downgrade_authorization.expires_at > clock_timestamp()
    AND latch.installation_id = p_installation_id
    AND latch.body_digest = p_migration_latch_digest
    AND latch.installation_kind = 'disposable_fixture'
    AND latch.database_identity_digest = downgrade_authorization.database_identity_digest
    AND latch.up_catalog_digest = downgrade_authorization.current_catalog_digest
    AND latch.migration_version = downgrade_authorization.migration_version
    AND latch.down_state = 'locked';
  IF v_authorization_id IS NULL THEN
    RAISE EXCEPTION 'authority v7 downgrade guard binding mismatch' USING ERRCODE = '55000';
  END IF;

  BEGIN
    v_authorized_state_raw := convert_from(p_authorized_state_body_jcs, 'UTF8')::json;
    v_pristine_raw := convert_from(v_pristine_inventory_body_jcs, 'UTF8')::json;
    v_authorized_state := convert_from(p_authorized_state_body_jcs, 'UTF8')::jsonb;
    v_pristine := convert_from(v_pristine_inventory_body_jcs, 'UTF8')::jsonb;
  EXCEPTION WHEN OTHERS THEN
    RAISE EXCEPTION 'invalid authority v7 authorized-state JSON preimage' USING ERRCODE = '22023';
  END;
  IF jsonb_typeof(v_authorized_state) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_authorized_state)) <> 11
     OR NOT (v_authorized_state ?& ARRAY[
       'pristine_downgrade_inventory_digest',
       'migration_latch_digest',
       'downgrade_authorization_digest',
       'provider_protocol_downgrade_retirement_set_digest',
       'database_transaction_id',
       'transaction_nonce',
       'migration_latch_count',
       'downgrade_authorization_count',
       'non_control_protocol_row_count',
       'stage',
       'observed_at'
     ])
     OR EXISTS (
       SELECT 1 FROM jsonb_each(v_authorized_state) AS field(name, value)
       WHERE jsonb_typeof(value) <> 'string'
     )
     OR convert_from(p_authorized_state_body_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"database_transaction_id":%s,"downgrade_authorization_count":%s,"downgrade_authorization_digest":%s,"migration_latch_count":%s,"migration_latch_digest":%s,"non_control_protocol_row_count":%s,"observed_at":%s,"pristine_downgrade_inventory_digest":%s,"provider_protocol_downgrade_retirement_set_digest":%s,"stage":%s,"transaction_nonce":%s}',
       to_json(v_authorized_state->>'database_transaction_id')::text,
       to_json(v_authorized_state->>'downgrade_authorization_count')::text,
       to_json(v_authorized_state->>'downgrade_authorization_digest')::text,
       to_json(v_authorized_state->>'migration_latch_count')::text,
       to_json(v_authorized_state->>'migration_latch_digest')::text,
       to_json(v_authorized_state->>'non_control_protocol_row_count')::text,
       to_json(v_authorized_state->>'observed_at')::text,
       to_json(v_authorized_state->>'pristine_downgrade_inventory_digest')::text,
       to_json(v_authorized_state->>'provider_protocol_downgrade_retirement_set_digest')::text,
       to_json(v_authorized_state->>'stage')::text,
       to_json(v_authorized_state->>'transaction_nonce')::text
     )
     OR sha256(
       convert_to('talenro.c12.pristine-downgrade-authorized-state.v1', 'UTF8')
       || decode('00', 'hex')
       || p_authorized_state_body_jcs
     ) IS DISTINCT FROM p_authorized_state_digest
     OR v_authorized_state->>'pristine_downgrade_inventory_digest' IS DISTINCT FROM encode(v_pristine_inventory_digest, 'hex')
     OR v_authorized_state->>'migration_latch_digest' IS DISTINCT FROM encode(p_migration_latch_digest, 'hex')
     OR v_authorized_state->>'downgrade_authorization_digest' IS DISTINCT FROM encode(p_authorization_digest, 'hex')
     OR v_authorized_state->>'provider_protocol_downgrade_retirement_set_digest' IS DISTINCT FROM encode(v_provider_retirement_set_digest, 'hex')
     OR v_authorized_state->>'database_transaction_id' IS DISTINCT FROM v_txid::text
     OR v_authorized_state->>'transaction_nonce' IS DISTINCT FROM encode(p_transaction_nonce, 'hex')
     OR v_authorized_state->>'migration_latch_count' IS DISTINCT FROM '1'
     OR v_authorized_state->>'downgrade_authorization_count' IS DISTINCT FROM '1'
     OR v_authorized_state->>'non_control_protocol_row_count' IS DISTINCT FROM '0'
     OR v_authorized_state->>'stage' IS DISTINCT FROM 'authorization_inserted'
     OR (v_authorized_state->>'observed_at')::timestamptz IS DISTINCT FROM transaction_timestamp() THEN
    RAISE EXCEPTION 'authority v7 authorized-state cross-binding mismatch' USING ERRCODE = '22023';
  END IF;
  IF jsonb_typeof(v_pristine) <> 'object'
     OR (SELECT count(*) FROM jsonb_object_keys(v_pristine)) <> 17
     OR NOT (v_pristine ?& ARRAY[
       'installation_id',
       'installation_kind',
       'database_identity_digest',
       'migration_latch_digest',
       'current_catalog_digest',
       'manifest_id',
       'manifest_digest',
       'stable_table_count',
       'stable_table_inventory',
       'control_table_count',
       'non_control_protocol_row_count',
       'migration_latch_count',
       'downgrade_authorization_count',
       'database_transaction_id',
       'transaction_nonce',
       'stage',
       'observed_at'
     ])
     OR jsonb_typeof(v_pristine->'stable_table_inventory') <> 'array'
     OR jsonb_array_length(v_pristine->'stable_table_inventory') <> 49
     OR v_pristine->'stable_table_inventory' IS DISTINCT FROM v_inventory
     OR convert_from(v_pristine_inventory_body_jcs, 'UTF8') IS DISTINCT FROM format(
       '{"control_table_count":%s,"current_catalog_digest":%s,"database_identity_digest":%s,"database_transaction_id":%s,"downgrade_authorization_count":%s,"installation_id":%s,"installation_kind":%s,"manifest_digest":%s,"manifest_id":%s,"migration_latch_count":%s,"migration_latch_digest":%s,"non_control_protocol_row_count":%s,"observed_at":%s,"stable_table_count":%s,"stable_table_inventory":%s,"stage":%s,"transaction_nonce":%s}',
       to_json(v_pristine->>'control_table_count')::text,
       to_json(v_pristine->>'current_catalog_digest')::text,
       to_json(v_pristine->>'database_identity_digest')::text,
       to_json(v_pristine->>'database_transaction_id')::text,
       to_json(v_pristine->>'downgrade_authorization_count')::text,
       to_json(v_pristine->>'installation_id')::text,
       to_json(v_pristine->>'installation_kind')::text,
       to_json(v_pristine->>'manifest_digest')::text,
       to_json(v_pristine->>'manifest_id')::text,
       to_json(v_pristine->>'migration_latch_count')::text,
       to_json(v_pristine->>'migration_latch_digest')::text,
       to_json(v_pristine->>'non_control_protocol_row_count')::text,
       to_json(v_pristine->>'observed_at')::text,
       to_json(v_pristine->>'stable_table_count')::text,
       v_inventory_jcs,
       to_json(v_pristine->>'stage')::text,
       to_json(v_pristine->>'transaction_nonce')::text
     )
     OR sha256(
       convert_to('talenro.c12.pristine-downgrade-inventory.v1', 'UTF8')
       || decode('00', 'hex')
       || v_pristine_inventory_body_jcs
     ) IS DISTINCT FROM v_pristine_inventory_digest
     OR v_pristine->>'installation_id' IS DISTINCT FROM p_installation_id::text
     OR v_pristine->>'installation_kind' IS DISTINCT FROM 'disposable_fixture'
     OR v_pristine->>'database_identity_digest' IS DISTINCT FROM encode(v_database_identity_digest, 'hex')
     OR v_pristine->>'migration_latch_digest' IS DISTINCT FROM encode(p_migration_latch_digest, 'hex')
     OR v_pristine->>'current_catalog_digest' IS DISTINCT FROM encode(v_current_catalog_digest, 'hex')
     OR octet_length(decode(v_pristine->>'manifest_digest', 'hex')) IS DISTINCT FROM 32
     OR (v_pristine->>'manifest_id')::uuid IS NULL
     OR v_pristine->>'stable_table_count' IS DISTINCT FROM '49'
     OR v_pristine->>'control_table_count' IS DISTINCT FROM '2'
     OR v_pristine->>'non_control_protocol_row_count' IS DISTINCT FROM '0'
     OR v_pristine->>'migration_latch_count' IS DISTINCT FROM '1'
     OR v_pristine->>'downgrade_authorization_count' IS DISTINCT FROM '0'
     OR v_pristine->>'database_transaction_id' IS DISTINCT FROM v_txid::text
     OR v_pristine->>'transaction_nonce' IS DISTINCT FROM encode(p_transaction_nonce, 'hex')
     OR v_pristine->>'stage' IS DISTINCT FROM 'pre_authorization'
     OR (v_pristine->>'observed_at')::timestamptz IS DISTINCT FROM transaction_timestamp()
     OR v_migration_version IS DISTINCT FROM 7
     OR v_authorization_scope IS DISTINCT FROM 'down_00007_only'
     OR v_issued_at > clock_timestamp()
     OR v_expires_at <= clock_timestamp() THEN
    RAISE EXCEPTION 'authority v7 stored pristine inventory drift' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM nodecontrol.control_plane_authority_fences
  WHERE authority_protocol_profile = 'claim_v1'
     OR (
       authority_protocol_profile = 'legacy_v6'
       AND abort_claimed_at IS NULL
       AND protocol_activation_id IS NULL
       AND (
         (provider_status = 'reserved' AND visibility_state = 'fence_pending')
         OR (provider_status = 'committed' AND visibility_state = 'active')
         OR (provider_status = 'aborted' AND visibility_state = 'aborted')
       )
     );
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade fence profile is not pristine' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM (
    SELECT
      create_authority_effect_commitment_jcs AS c01, create_authority_effect_commitment_digest AS c02,
      create_authority_provider_head_jcs AS c03, create_authority_provider_head_digest AS c04,
      create_authority_checkpoint_anchor_jcs AS c05, create_authority_checkpoint_anchor_digest AS c06,
      create_authority_effect_reason AS c07, create_authority_attestation_expires_at AS c08,
      create_authority_activation_deadline AS c09, create_authority_expected_provider_identity_digest AS c10,
      create_authority_activation_evidence_jcs AS c11, create_authority_activation_evidence_digest AS c12,
      create_authority_effect_resolution_jcs AS c13, create_authority_effect_resolution_digest AS c14
    FROM nodecontrol.node_enrollment_grants
    UNION ALL
    SELECT
      claim_authority_effect_commitment_jcs, claim_authority_effect_commitment_digest,
      claim_authority_provider_head_jcs, claim_authority_provider_head_digest,
      claim_authority_checkpoint_anchor_jcs, claim_authority_checkpoint_anchor_digest,
      claim_authority_effect_reason, claim_authority_attestation_expires_at,
      claim_authority_activation_deadline, claim_authority_expected_provider_identity_digest,
      claim_authority_activation_evidence_jcs, claim_authority_activation_evidence_digest,
      claim_authority_effect_resolution_jcs, claim_authority_effect_resolution_digest
    FROM nodecontrol.node_enrollment_grants
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_certificate_issuances
    UNION ALL
    SELECT
      revoke_authority_effect_commitment_jcs, revoke_authority_effect_commitment_digest,
      revoke_authority_provider_head_jcs, revoke_authority_provider_head_digest,
      revoke_authority_checkpoint_anchor_jcs, revoke_authority_checkpoint_anchor_digest,
      revoke_authority_effect_reason, revoke_authority_attestation_expires_at,
      revoke_authority_activation_deadline, revoke_authority_expected_provider_identity_digest,
      revoke_authority_activation_evidence_jcs, revoke_authority_activation_evidence_digest,
      revoke_authority_effect_resolution_jcs, revoke_authority_effect_resolution_digest
    FROM nodecontrol.node_certificates
    UNION ALL
    SELECT
      authority_effect_commitment_jcs, authority_effect_commitment_digest,
      authority_provider_head_jcs, authority_provider_head_digest,
      authority_checkpoint_anchor_jcs, authority_checkpoint_anchor_digest,
      authority_effect_reason, authority_attestation_expires_at,
      authority_activation_deadline, authority_expected_provider_identity_digest,
      authority_activation_evidence_jcs, authority_activation_evidence_digest,
      authority_effect_resolution_jcs, authority_effect_resolution_digest
    FROM nodecontrol.node_state_transitions
    UNION ALL
    SELECT
      open_authority_effect_commitment_jcs, open_authority_effect_commitment_digest,
      open_authority_provider_head_jcs, open_authority_provider_head_digest,
      open_authority_checkpoint_anchor_jcs, open_authority_checkpoint_anchor_digest,
      open_authority_effect_reason, open_authority_attestation_expires_at,
      open_authority_activation_deadline, open_authority_expected_provider_identity_digest,
      open_authority_activation_evidence_jcs, open_authority_activation_evidence_digest,
      open_authority_effect_resolution_jcs, open_authority_effect_resolution_digest
    FROM nodecontrol.node_security_incidents
    UNION ALL
    SELECT
      resolve_authority_effect_commitment_jcs, resolve_authority_effect_commitment_digest,
      resolve_authority_provider_head_jcs, resolve_authority_provider_head_digest,
      resolve_authority_checkpoint_anchor_jcs, resolve_authority_checkpoint_anchor_digest,
      resolve_authority_effect_reason, resolve_authority_attestation_expires_at,
      resolve_authority_activation_deadline, resolve_authority_expected_provider_identity_digest,
      resolve_authority_activation_evidence_jcs, resolve_authority_activation_evidence_digest,
      resolve_authority_effect_resolution_jcs, resolve_authority_effect_resolution_digest
    FROM nodecontrol.node_security_incidents
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_resource_envelopes
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_state_signing_intents
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_root_metadata_publish_intents
  ) AS proof_group
  WHERE pg_catalog.num_nonnulls(
    proof_group.c01, proof_group.c02, proof_group.c03, proof_group.c04,
    proof_group.c05, proof_group.c06, proof_group.c07, proof_group.c08,
    proof_group.c09, proof_group.c10, proof_group.c11, proof_group.c12,
    proof_group.c13, proof_group.c14
  ) <> 0;
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade proof fields are not pristine' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM nodecontrol.node_inventory
  WHERE identity_state = 'unauthorized';
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade unauthorized inventory is not pristine' USING ERRCODE = '55000';
  END IF;

  PERFORM set_config('nodecontrol.v7_down_guard', 'consume', true);
  WITH down_candidate AS MATERIALIZED (
    SELECT
      downgrade_authorization.authorization_id,
      downgrade_authorization.installation_id,
      downgrade_authorization.body_digest AS authorization_digest,
      latch.body_digest AS latch_digest
    FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations AS downgrade_authorization
    JOIN nodecontrol.control_plane_authority_protocol_migration_latches AS latch
      ON latch.installation_id = downgrade_authorization.installation_id
    WHERE downgrade_authorization.authorization_id = v_authorization_id
      AND downgrade_authorization.installation_id = p_installation_id
      AND downgrade_authorization.body_digest = p_authorization_digest
      AND downgrade_authorization.migration_latch_digest = p_migration_latch_digest
      AND downgrade_authorization.database_identity_digest = v_database_identity_digest
      AND downgrade_authorization.current_catalog_digest = v_current_catalog_digest
      AND downgrade_authorization.pristine_downgrade_inventory_digest = v_pristine_inventory_digest
      AND downgrade_authorization.provider_protocol_downgrade_retirement_set_digest = v_provider_retirement_set_digest
      AND downgrade_authorization.transaction_nonce = p_transaction_nonce
      AND downgrade_authorization.database_transaction_id = v_txid
      AND downgrade_authorization.migration_version = 7
      AND downgrade_authorization.authorization_scope = 'down_00007_only'
      AND downgrade_authorization.issued_at = v_issued_at
      AND downgrade_authorization.expires_at = v_expires_at
      AND downgrade_authorization.issued_at <= clock_timestamp()
      AND downgrade_authorization.expires_at > clock_timestamp()
      AND latch.body_digest = p_migration_latch_digest
      AND latch.installation_kind = 'disposable_fixture'
      AND latch.database_identity_digest = v_database_identity_digest
      AND latch.up_catalog_digest = v_current_catalog_digest
      AND latch.migration_version = 7
      AND latch.down_state = 'locked'
      AND (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations) = 1
      AND (SELECT count(*) FROM nodecontrol.control_plane_authority_protocol_migration_latches) = 1
  ), deleted_authorization AS MATERIALIZED (
    DELETE FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations AS downgrade_authorization
    USING down_candidate AS candidate
    WHERE downgrade_authorization.authorization_id = candidate.authorization_id
      AND downgrade_authorization.body_digest = candidate.authorization_digest
    RETURNING downgrade_authorization.authorization_id
  ), deleted_latch AS MATERIALIZED (
    DELETE FROM nodecontrol.control_plane_authority_protocol_migration_latches AS latch
    USING down_candidate AS candidate, deleted_authorization
    WHERE latch.installation_id = candidate.installation_id
      AND latch.body_digest = candidate.latch_digest
    RETURNING latch.installation_id
  )
  SELECT
    (SELECT count(*) FROM down_candidate),
    (SELECT count(*) FROM deleted_authorization),
    (SELECT count(*) FROM deleted_latch)
  INTO v_candidate_count, v_deleted_authorization, v_deleted_latch;
  IF v_candidate_count <> 1 OR v_deleted_authorization <> 1 OR v_deleted_latch <> 1 THEN
    RAISE EXCEPTION 'authority v7 downgrade guard consumption mismatch' USING ERRCODE = '55000';
  END IF;

  SELECT count(*) INTO v_post_latch_count
  FROM nodecontrol.control_plane_authority_protocol_migration_latches AS latch;
  SELECT count(*) INTO v_post_authorization_count
  FROM nodecontrol.control_plane_authority_protocol_downgrade_authorizations AS downgrade_authorization;
  v_post_non_control_count := 0;
  FOREACH v_table_name IN ARRAY v_registry_tables LOOP
    EXECUTE format('SELECT count(*) FROM %s', v_table_name) INTO v_row_count;
    v_post_non_control_count := v_post_non_control_count + v_row_count;
  END LOOP;
  IF v_post_latch_count <> 0
     OR v_post_authorization_count <> 0
     OR v_post_non_control_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 post-consume 0/0/0 rescan mismatch' USING ERRCODE = '55000';
  END IF;
  SELECT count(*)
  INTO v_forbidden_count
  FROM nodecontrol.control_plane_authority_fences
  WHERE authority_protocol_profile = 'claim_v1'
     OR (
       authority_protocol_profile = 'legacy_v6'
       AND abort_claimed_at IS NULL
       AND protocol_activation_id IS NULL
       AND (
         (provider_status = 'reserved' AND visibility_state = 'fence_pending')
         OR (provider_status = 'committed' AND visibility_state = 'active')
         OR (provider_status = 'aborted' AND visibility_state = 'aborted')
       )
     );
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade fence profile is not pristine' USING ERRCODE = '55000';
  END IF;
  SELECT count(*)
  INTO v_forbidden_count
  FROM (
    SELECT
      create_authority_effect_commitment_jcs AS c01, create_authority_effect_commitment_digest AS c02,
      create_authority_provider_head_jcs AS c03, create_authority_provider_head_digest AS c04,
      create_authority_checkpoint_anchor_jcs AS c05, create_authority_checkpoint_anchor_digest AS c06,
      create_authority_effect_reason AS c07, create_authority_attestation_expires_at AS c08,
      create_authority_activation_deadline AS c09, create_authority_expected_provider_identity_digest AS c10,
      create_authority_activation_evidence_jcs AS c11, create_authority_activation_evidence_digest AS c12,
      create_authority_effect_resolution_jcs AS c13, create_authority_effect_resolution_digest AS c14
    FROM nodecontrol.node_enrollment_grants
    UNION ALL
    SELECT
      claim_authority_effect_commitment_jcs, claim_authority_effect_commitment_digest,
      claim_authority_provider_head_jcs, claim_authority_provider_head_digest,
      claim_authority_checkpoint_anchor_jcs, claim_authority_checkpoint_anchor_digest,
      claim_authority_effect_reason, claim_authority_attestation_expires_at,
      claim_authority_activation_deadline, claim_authority_expected_provider_identity_digest,
      claim_authority_activation_evidence_jcs, claim_authority_activation_evidence_digest,
      claim_authority_effect_resolution_jcs, claim_authority_effect_resolution_digest
    FROM nodecontrol.node_enrollment_grants
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_certificate_issuances
    UNION ALL
    SELECT
      revoke_authority_effect_commitment_jcs, revoke_authority_effect_commitment_digest,
      revoke_authority_provider_head_jcs, revoke_authority_provider_head_digest,
      revoke_authority_checkpoint_anchor_jcs, revoke_authority_checkpoint_anchor_digest,
      revoke_authority_effect_reason, revoke_authority_attestation_expires_at,
      revoke_authority_activation_deadline, revoke_authority_expected_provider_identity_digest,
      revoke_authority_activation_evidence_jcs, revoke_authority_activation_evidence_digest,
      revoke_authority_effect_resolution_jcs, revoke_authority_effect_resolution_digest
    FROM nodecontrol.node_certificates
    UNION ALL
    SELECT
      authority_effect_commitment_jcs, authority_effect_commitment_digest,
      authority_provider_head_jcs, authority_provider_head_digest,
      authority_checkpoint_anchor_jcs, authority_checkpoint_anchor_digest,
      authority_effect_reason, authority_attestation_expires_at,
      authority_activation_deadline, authority_expected_provider_identity_digest,
      authority_activation_evidence_jcs, authority_activation_evidence_digest,
      authority_effect_resolution_jcs, authority_effect_resolution_digest
    FROM nodecontrol.node_state_transitions
    UNION ALL
    SELECT
      open_authority_effect_commitment_jcs, open_authority_effect_commitment_digest,
      open_authority_provider_head_jcs, open_authority_provider_head_digest,
      open_authority_checkpoint_anchor_jcs, open_authority_checkpoint_anchor_digest,
      open_authority_effect_reason, open_authority_attestation_expires_at,
      open_authority_activation_deadline, open_authority_expected_provider_identity_digest,
      open_authority_activation_evidence_jcs, open_authority_activation_evidence_digest,
      open_authority_effect_resolution_jcs, open_authority_effect_resolution_digest
    FROM nodecontrol.node_security_incidents
    UNION ALL
    SELECT
      resolve_authority_effect_commitment_jcs, resolve_authority_effect_commitment_digest,
      resolve_authority_provider_head_jcs, resolve_authority_provider_head_digest,
      resolve_authority_checkpoint_anchor_jcs, resolve_authority_checkpoint_anchor_digest,
      resolve_authority_effect_reason, resolve_authority_attestation_expires_at,
      resolve_authority_activation_deadline, resolve_authority_expected_provider_identity_digest,
      resolve_authority_activation_evidence_jcs, resolve_authority_activation_evidence_digest,
      resolve_authority_effect_resolution_jcs, resolve_authority_effect_resolution_digest
    FROM nodecontrol.node_security_incidents
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_resource_envelopes
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_state_signing_intents
    UNION ALL
    SELECT
      activation_authority_effect_commitment_jcs, activation_authority_effect_commitment_digest,
      activation_authority_provider_head_jcs, activation_authority_provider_head_digest,
      activation_authority_checkpoint_anchor_jcs, activation_authority_checkpoint_anchor_digest,
      activation_authority_effect_reason, activation_authority_attestation_expires_at,
      activation_authority_activation_deadline, activation_authority_expected_provider_identity_digest,
      activation_authority_activation_evidence_jcs, activation_authority_activation_evidence_digest,
      activation_authority_effect_resolution_jcs, activation_authority_effect_resolution_digest
    FROM nodecontrol.node_root_metadata_publish_intents
  ) AS proof_group
  WHERE pg_catalog.num_nonnulls(
    proof_group.c01, proof_group.c02, proof_group.c03, proof_group.c04,
    proof_group.c05, proof_group.c06, proof_group.c07, proof_group.c08,
    proof_group.c09, proof_group.c10, proof_group.c11, proof_group.c12,
    proof_group.c13, proof_group.c14
  ) <> 0;
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade proof fields are not pristine' USING ERRCODE = '55000';
  END IF;

  SELECT count(*)
  INTO v_forbidden_count
  FROM nodecontrol.node_inventory
  WHERE identity_state = 'unauthorized';
  IF v_forbidden_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 downgrade unauthorized inventory is not pristine' USING ERRCODE = '55000';
  END IF;

  RETURN QUERY SELECT
    v_pristine_inventory_digest,
    p_authorized_state_digest,
    p_migration_latch_digest,
    p_authorization_digest,
    v_txid,
    p_transaction_nonce,
    v_post_latch_count,
    v_post_authorization_count,
    v_post_non_control_count,
    'guard_consumed'::text,
    clock_timestamp();
END
$fn$;

-- talenro:statement
CREATE FUNCTION nodecontrol.begin_staging_import(
  capability_digest bytea,
  manifest_digest bytea,
  exclusion_lease_digest bytea,
  acquisition_head_digest bytea,
  route_closed_digest bytea,
  projection_body_jcs bytea,
  projection_digest bytea,
  projection_objects jsonb
)
RETURNS SETOF nodecontrol.control_plane_authority_fresh_restore_import_applications
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, nodecontrol
AS $fn$
DECLARE
  capability nodecontrol.control_plane_authority_staging_import_capabilities%ROWTYPE;
  projection jsonb;
  application_body bytea;
  application_digest bytea;
  transaction_snapshot_digest bytea;
  imported_count bigint;
  complete_node_set_digest bytea;
  forbidden_state_zero_digest bytea;
  applied_time timestamp with time zone := transaction_timestamp();
BEGIN
  PERFORM nodecontrol.v7_require_role('nodecontrol_staging_importer');
  IF pg_catalog.current_setting('transaction_isolation') <> 'read committed' THEN
    RAISE EXCEPTION 'staging import requires read committed isolation' USING ERRCODE = '25001';
  END IF;
  IF octet_length(capability_digest) <> 32 OR octet_length(manifest_digest) <> 32
     OR octet_length(exclusion_lease_digest) <> 32 OR octet_length(acquisition_head_digest) <> 32
     OR octet_length(route_closed_digest) <> 32 OR octet_length(projection_digest) <> 32
     OR octet_length(projection_body_jcs) NOT BETWEEN 1 AND 1048576
     OR jsonb_typeof(projection_objects) <> 'array' THEN
    RETURN;
  END IF;
  PERFORM pg_catalog.pg_advisory_xact_lock_shared(
    pg_catalog.hashtextextended('nodecontrol:v7-source-freeze',0)
  );
  PERFORM pg_advisory_xact_lock(1313031735, 1);
  PERFORM pg_advisory_xact_lock(1313031735, 2);
  PERFORM pg_advisory_xact_lock(1313031735, 3);
  SELECT * INTO capability
  FROM nodecontrol.control_plane_authority_staging_import_capabilities
  WHERE body_digest = capability_digest
  FOR UPDATE;
  IF NOT FOUND
     OR capability.manifest_digest <> begin_staging_import.manifest_digest
     OR capability.database_route_closed_digest <> route_closed_digest
     OR capability.expires_at <= clock_timestamp()
     OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_staging_import_capability_recovery_intents WHERE capability_id = capability.capability_id)
     OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_staging_import_capability_recovery_applications WHERE capability_id = capability.capability_id)
     OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_staging_import_capability_revocation_applications WHERE capability_id = capability.capability_id)
     OR EXISTS (SELECT 1 FROM nodecontrol.control_plane_authority_fresh_restore_import_applications WHERE single_use_apply_id = capability.single_use_apply_id) THEN
    RETURN;
  END IF;
  IF EXISTS (
    SELECT 1
    FROM jsonb_array_elements(projection_objects) AS element(value)
    WHERE jsonb_typeof(value) <> 'object'
       OR EXISTS (SELECT 1 FROM jsonb_object_keys(value) AS key(name) WHERE name NOT IN ('object_type','canonical_key','body'))
       OR NOT (value ? 'object_type' AND value ? 'canonical_key' AND value ? 'body')
       OR jsonb_typeof(value->'object_type') <> 'string'
       OR jsonb_typeof(value->'canonical_key') <> 'string'
       OR jsonb_typeof(value->'body') <> 'object'
       OR value->>'object_type' NOT IN ('pop','failure_domain_definition','capacity_profile_definition','node_reconstruction_seed')
  ) THEN
    RETURN;
  END IF;
  IF EXISTS (
    SELECT 1
    FROM (
      SELECT value,
             lag(value->>'object_type') OVER (ORDER BY ordinal) AS previous_type,
             lag(value->>'canonical_key') OVER (ORDER BY ordinal) AS previous_key
      FROM jsonb_array_elements(projection_objects) WITH ORDINALITY AS element(value, ordinal)
    ) AS ordered
    WHERE previous_type > value->>'object_type'
       OR (previous_type = value->>'object_type' AND previous_key >= value->>'canonical_key')
  ) THEN
    RETURN;
  END IF;
  IF EXISTS (
    SELECT 1
    FROM jsonb_array_elements(projection_objects) AS element(value)
    WHERE CASE value->>'object_type'
      WHEN 'pop' THEN
        (SELECT count(*) <> 5 OR bool_or(name NOT IN ('pop_code','iso_country','region','operator_state','version')) FROM jsonb_object_keys(value->'body') AS key(name))
        OR value->>'canonical_key' <> value->'body'->>'pop_code'
        OR value->'body'->>'operator_state' <> 'disabled'
        OR value->'body'->>'version' <> '1'
      WHEN 'failure_domain_definition' THEN
        (SELECT count(*) <> 4 OR bool_or(name NOT IN ('failure_domain_id','domain_type','stable_id','version')) FROM jsonb_object_keys(value->'body') AS key(name))
        OR value->>'canonical_key' <> value->'body'->>'failure_domain_id'
        OR value->'body'->>'version' <> '1'
      WHEN 'capacity_profile_definition' THEN
        (SELECT count(*) <> 14 OR bool_or(name NOT IN ('profile_id','version','adapter','egress_limit_bps','connection_limit','handshake_limit_per_second','cpu_quota_millicores','cpu_limit_basis_points','memory_limit_bytes','task_limit','file_descriptor_limit','queue_limit','packet_loss_limit_basis_points','required_metrics')) FROM jsonb_object_keys(value->'body') AS key(name))
        OR value->>'canonical_key' <> (value->'body'->>'profile_id') || '/' || (value->'body'->>'version')
      WHEN 'node_reconstruction_seed' THEN
        (SELECT count(*) <> 15 OR bool_or(name NOT IN ('node_id','pop_code','operator_state','security_state','identity_state','health_state','identity_epoch','inventory_version','security_version','next_desired_generation','next_recovery_generation','resume_operator_state_or_null','pending_operator_transition_or_null','active_pointer_set','authority_anchor_set')) FROM jsonb_object_keys(value->'body') AS key(name))
        OR value->>'canonical_key' <> value->'body'->>'node_id'
        OR value->'body'->>'operator_state' <> 'disabled'
        OR value->'body'->>'security_state' <> 'quarantined'
        OR value->'body'->>'identity_state' <> 'unauthorized'
        OR value->'body'->>'health_state' <> 'unknown'
        OR value->'body'->>'identity_epoch' <> '0'
        OR value->'body'->>'inventory_version' <> '1'
        OR value->'body'->>'security_version' <> '1'
        OR value->'body'->>'next_desired_generation' <> '1'
        OR value->'body'->>'next_recovery_generation' <> '1'
        OR value->'body'->'resume_operator_state_or_null' <> 'null'::jsonb
        OR value->'body'->'pending_operator_transition_or_null' <> 'null'::jsonb
        OR value->'body'->'active_pointer_set' <> '[]'::jsonb
        OR value->'body'->'authority_anchor_set' <> '[]'::jsonb
      ELSE true
    END
  ) OR NOT EXISTS (
    SELECT 1 FROM jsonb_array_elements(projection_objects) AS element(value)
    WHERE value->>'object_type' = 'node_reconstruction_seed'
  ) THEN
    RETURN;
  END IF;
  projection := convert_from(projection_body_jcs, 'UTF8')::jsonb;
  IF projection_digest IS DISTINCT FROM capability.expected_post_import_inventory_digest
     OR (projection->>'target_activation_id')::uuid <> capability.target_activation_id
     OR (projection->>'target_database_identity_digest') <> encode(capability.target_database_identity_digest, 'hex')
     OR (projection->>'object_count')::bigint <> jsonb_array_length(projection_objects)
     OR projection->'objects' IS DISTINCT FROM (
       SELECT jsonb_agg(jsonb_build_object(
         'object_type', value->>'object_type',
         'canonical_key', value->>'canonical_key',
         'normalized_payload', value->'body'
       ) ORDER BY ordinal)
       FROM jsonb_array_elements(projection_objects) WITH ORDINALITY AS item(value, ordinal)
     ) THEN
    RETURN;
  END IF;
  imported_count := jsonb_array_length(projection_objects);
  complete_node_set_digest := sha256(convert_to('talenro.c12.complete-node-set.v1', 'UTF8') || decode('00','hex') || convert_to((SELECT coalesce(jsonb_agg(value->>'canonical_key' ORDER BY value->>'canonical_key'),'[]'::jsonb)::text FROM jsonb_array_elements(projection_objects) AS item(value) WHERE value->>'object_type' = 'node_reconstruction_seed'), 'UTF8'));
  forbidden_state_zero_digest := sha256(convert_to('talenro.c12.fresh-import-forbidden-state-zero.v1', 'UTF8') || decode('00','hex') || convert_to('{"forbidden_row_count":"0"}', 'UTF8'));
  transaction_snapshot_digest := sha256(convert_to(txid_current_snapshot()::text, 'UTF8'));
  application_body := convert_to(jsonb_build_object(
    'single_use_apply_id', capability.single_use_apply_id,
    'staging_import_capability_digest', encode(capability_digest, 'hex'),
    'staging_import_capability_recovery_intent_digest_or_null', NULL,
    'staging_import_capability_recovery_application_digest_or_null', NULL,
    'manifest_digest', encode(manifest_digest, 'hex'),
    'target_activation_id', capability.target_activation_id,
    'current_database_identity_digest', encode(capability.target_database_identity_digest, 'hex'),
    'database_timeline_lineage_chain_digest', encode(capability.database_timeline_lineage_chain_digest, 'hex'),
    'target_database_incarnation_registration_digest', encode(capability.target_database_incarnation_registration_digest, 'hex'),
    'runtime_rebind_chain_digest', encode(capability.runtime_rebind_chain_digest, 'hex'),
    'runtime_instance_binding_digest', encode(capability.runtime_instance_binding_digest, 'hex'),
    'staging_exclusion_lease_digest', encode(exclusion_lease_digest, 'hex'),
    'acquisition_locked_provider_head_digest', encode(acquisition_head_digest, 'hex'),
    'database_route_closed_digest', encode(route_closed_digest, 'hex'),
    'pre_import_inventory_digest', encode(capability.pre_import_inventory_digest, 'hex'),
    'post_import_inventory_digest', encode(capability.expected_post_import_inventory_digest, 'hex'),
    'imported_object_count', imported_count::text,
    'complete_node_set_digest', encode(complete_node_set_digest, 'hex'),
    'forbidden_state_zero_digest', encode(forbidden_state_zero_digest, 'hex'),
    'database_transaction_id', txid_current()::text,
    'transaction_snapshot_digest', encode(transaction_snapshot_digest, 'hex'),
    'applied_at', transaction_timestamp()
  )::text, 'UTF8');
  application_digest := sha256(convert_to('talenro.c12.fresh-restore-import-application.v1', 'UTF8') || decode('00','hex') || application_body);
  PERFORM set_config('nodecontrol.v7_staging_import', 'verified', true);

  INSERT INTO nodecontrol.node_pops (
    pop_code, iso_country, region, operator_state, version, created_at, updated_at
  )
  SELECT value->'body'->>'pop_code', value->'body'->>'iso_country', value->'body'->>'region',
         'disabled', 1, applied_time, applied_time
  FROM jsonb_array_elements(projection_objects) AS item(value)
  WHERE value->>'object_type' = 'pop';

  INSERT INTO nodecontrol.node_failure_domains (
    failure_domain_id, domain_type, stable_id, version, created_at, updated_at
  )
  SELECT (value->'body'->>'failure_domain_id')::uuid, value->'body'->>'domain_type',
         value->'body'->>'stable_id', 1, applied_time, applied_time
  FROM jsonb_array_elements(projection_objects) AS item(value)
  WHERE value->>'object_type' = 'failure_domain_definition';

  INSERT INTO nodecontrol.node_capacity_profiles (
    profile_id, version, adapter, egress_limit_bps, connection_limit, handshake_limit_per_second,
    cpu_quota_millicores, cpu_limit_basis_points, memory_limit_bytes, task_limit,
    file_descriptor_limit, queue_limit, packet_loss_limit_basis_points, required_metrics, created_at
  )
  SELECT value->'body'->>'profile_id', (value->'body'->>'version')::bigint,
         value->'body'->>'adapter', (value->'body'->>'egress_limit_bps')::bigint,
         (value->'body'->>'connection_limit')::integer,
         (value->'body'->>'handshake_limit_per_second')::integer,
         (value->'body'->>'cpu_quota_millicores')::integer,
         (value->'body'->>'cpu_limit_basis_points')::integer,
         (value->'body'->>'memory_limit_bytes')::bigint,
         (value->'body'->>'task_limit')::integer,
         (value->'body'->>'file_descriptor_limit')::integer,
         (value->'body'->>'queue_limit')::integer,
         (value->'body'->>'packet_loss_limit_basis_points')::integer,
         ARRAY(SELECT jsonb_array_elements_text(value->'body'->'required_metrics')),
         applied_time
  FROM jsonb_array_elements(projection_objects) AS item(value)
  WHERE value->>'object_type' = 'capacity_profile_definition';

  INSERT INTO nodecontrol.node_inventory (
    node_id, pop_code, operator_state, security_state, identity_state, resume_operator_state,
    pending_operator_transition, pending_transition_signing_id, identity_epoch, lineage_id,
    inventory_version, security_version, next_desired_generation, next_recovery_generation,
    created_at, updated_at
  )
  SELECT (value->'body'->>'node_id')::uuid, value->'body'->>'pop_code', 'disabled',
         'quarantined', 'unauthorized', NULL, NULL, NULL, 0, NULL, 1, 1, 1, 1,
         applied_time, applied_time
  FROM jsonb_array_elements(projection_objects) AS item(value)
  WHERE value->>'object_type' = 'node_reconstruction_seed';

  RETURN QUERY
  INSERT INTO nodecontrol.control_plane_authority_fresh_restore_import_applications (
    single_use_apply_id, staging_import_capability_digest,
    staging_import_capability_recovery_intent_digest_or_null, staging_import_capability_recovery_application_digest_or_null,
    manifest_digest, target_activation_id, current_database_identity_digest,
    database_timeline_lineage_chain_digest, target_database_incarnation_registration_digest,
    runtime_rebind_chain_digest, runtime_instance_binding_digest, staging_exclusion_lease_digest,
    acquisition_locked_provider_head_digest, database_route_closed_digest, pre_import_inventory_digest,
    post_import_inventory_digest, imported_object_count, complete_node_set_digest, forbidden_state_zero_digest,
    database_transaction_id, transaction_snapshot_digest, applied_at,
    canonical_evidence_bundle_jcs, canonical_body_jcs, body_digest
  ) VALUES (
    capability.single_use_apply_id, capability_digest, NULL, NULL,
    manifest_digest, capability.target_activation_id, capability.target_database_identity_digest,
    capability.database_timeline_lineage_chain_digest, capability.target_database_incarnation_registration_digest,
    capability.runtime_rebind_chain_digest, capability.runtime_instance_binding_digest, exclusion_lease_digest,
    acquisition_head_digest, route_closed_digest, capability.pre_import_inventory_digest,
    capability.expected_post_import_inventory_digest, imported_count, complete_node_set_digest, forbidden_state_zero_digest,
    txid_current(), transaction_snapshot_digest, applied_time,
    projection_body_jcs, application_body, application_digest
  )
  RETURNING *;
END
$fn$;

-- talenro:statement
REVOKE ALL ON ALL TABLES IN SCHEMA nodecontrol FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_text_array_is_sorted_unique(text[]) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_acquire_source_freeze_for_seal() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_reject_immutable_mutation() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_source_is_frozen() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_assert_source_writable() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_assert_activation_barrier() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_authority_proof_group_valid(bytea,bytea,bytea,bytea,bytea,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_guard_authority_proof_transition() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb) FROM PUBLIC;

-- talenro:statement
ALTER FUNCTION nodecontrol.v7_acquire_source_freeze_for_seal() OWNER TO nodecontrol_upgrade_executor;
ALTER FUNCTION nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea) OWNER TO nodecontrol_migration_downgrader;
ALTER FUNCTION nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea) OWNER TO nodecontrol_migration_downgrader;
ALTER FUNCTION nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb) OWNER TO nodecontrol_staging_importer;

-- talenro:statement
GRANT USAGE ON SCHEMA nodecontrol TO nodecontrol_upgrade_executor;
GRANT EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) TO nodecontrol_upgrade_executor;

-- talenro:statement
GRANT USAGE ON SCHEMA nodecontrol, public TO nodecontrol_migration_downgrader;
GRANT SELECT ON
  nodecontrol.control_plane_authority_protocol_migration_latches,
  nodecontrol.control_plane_authority_protocol_downgrade_authorizations,
  nodecontrol.control_plane_authority_epoch_transition_applications,
  nodecontrol.control_plane_authority_epoch_transition_cancellations,
  nodecontrol.control_plane_authority_epoch_transition_intents,
  nodecontrol.control_plane_authority_epoch_transition_recovery_applications,
  nodecontrol.control_plane_authority_epoch_transition_recovery_intents,
  nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions,
  nodecontrol.control_plane_authority_epoch_transition_resolutions,
  nodecontrol.control_plane_authority_epoch_transition_terminal_applications,
  nodecontrol.control_plane_authority_fresh_restore_import_applications,
  nodecontrol.control_plane_authority_fresh_restore_requirements,
  nodecontrol.control_plane_authority_indeterminate_source_seals,
  nodecontrol.control_plane_authority_legacy_database_source_retirements,
  nodecontrol.control_plane_authority_legacy_source_seals,
  nodecontrol.control_plane_authority_protocol_activation_completions,
  nodecontrol.control_plane_authority_protocol_activation_releases,
  nodecontrol.control_plane_authority_protocol_activations,
  nodecontrol.control_plane_authority_protocol_upgrade_attempts,
  nodecontrol.control_plane_authority_protocol_upgrade_intents,
  nodecontrol.control_plane_authority_runtime_rebind_results,
  nodecontrol.control_plane_authority_runtime_registration_results,
  nodecontrol.control_plane_authority_staging_import_capabilities,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
  nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
  nodecontrol.control_plane_authority_fences,
  nodecontrol.control_plane_trust_bundle_high_waters,
  nodecontrol.node_capacity_profiles,
  nodecontrol.node_certificate_issuances,
  nodecontrol.node_certificates,
  nodecontrol.node_desired_states,
  nodecontrol.node_endpoints,
  nodecontrol.node_enrollment_grants,
  nodecontrol.node_failure_domain_membership,
  nodecontrol.node_failure_domains,
  nodecontrol.node_inventory,
  nodecontrol.node_observed_states,
  nodecontrol.node_operator_audit,
  nodecontrol.node_pops,
  nodecontrol.node_process_slots,
  nodecontrol.node_recovery_sessions,
  nodecontrol.node_recovery_states,
  nodecontrol.node_resource_envelopes,
  nodecontrol.node_restore_reauthorization_approvals,
  nodecontrol.node_root_metadata_publish_intents,
  nodecontrol.node_root_metadata_signature_shares,
  nodecontrol.node_security_fault_receipts,
  nodecontrol.node_security_incidents,
  nodecontrol.node_state_signing_intents,
  nodecontrol.node_state_transitions
TO nodecontrol_migration_downgrader;
GRANT MAINTAIN ON
  nodecontrol.control_plane_authority_protocol_migration_latches,
  nodecontrol.control_plane_authority_protocol_downgrade_authorizations,
  nodecontrol.control_plane_authority_epoch_transition_applications,
  nodecontrol.control_plane_authority_epoch_transition_cancellations,
  nodecontrol.control_plane_authority_epoch_transition_intents,
  nodecontrol.control_plane_authority_epoch_transition_recovery_applications,
  nodecontrol.control_plane_authority_epoch_transition_recovery_intents,
  nodecontrol.control_plane_authority_epoch_transition_recovery_prefix_decisions,
  nodecontrol.control_plane_authority_epoch_transition_resolutions,
  nodecontrol.control_plane_authority_epoch_transition_terminal_applications,
  nodecontrol.control_plane_authority_fresh_restore_import_applications,
  nodecontrol.control_plane_authority_fresh_restore_requirements,
  nodecontrol.control_plane_authority_indeterminate_source_seals,
  nodecontrol.control_plane_authority_legacy_database_source_retirements,
  nodecontrol.control_plane_authority_legacy_source_seals,
  nodecontrol.control_plane_authority_protocol_activation_completions,
  nodecontrol.control_plane_authority_protocol_activation_releases,
  nodecontrol.control_plane_authority_protocol_activations,
  nodecontrol.control_plane_authority_protocol_upgrade_attempts,
  nodecontrol.control_plane_authority_protocol_upgrade_intents,
  nodecontrol.control_plane_authority_runtime_rebind_results,
  nodecontrol.control_plane_authority_runtime_registration_results,
  nodecontrol.control_plane_authority_staging_import_capabilities,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
  nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
  nodecontrol.control_plane_authority_fences,
  nodecontrol.control_plane_trust_bundle_high_waters,
  nodecontrol.node_capacity_profiles,
  nodecontrol.node_certificate_issuances,
  nodecontrol.node_certificates,
  nodecontrol.node_desired_states,
  nodecontrol.node_endpoints,
  nodecontrol.node_enrollment_grants,
  nodecontrol.node_failure_domain_membership,
  nodecontrol.node_failure_domains,
  nodecontrol.node_inventory,
  nodecontrol.node_observed_states,
  nodecontrol.node_operator_audit,
  nodecontrol.node_pops,
  nodecontrol.node_process_slots,
  nodecontrol.node_recovery_sessions,
  nodecontrol.node_recovery_states,
  nodecontrol.node_resource_envelopes,
  nodecontrol.node_restore_reauthorization_approvals,
  nodecontrol.node_root_metadata_publish_intents,
  nodecontrol.node_root_metadata_signature_shares,
  nodecontrol.node_security_fault_receipts,
  nodecontrol.node_security_incidents,
  nodecontrol.node_state_signing_intents,
  nodecontrol.node_state_transitions,
  public.goose_db_version
TO nodecontrol_migration_downgrader;
GRANT INSERT, DELETE ON nodecontrol.control_plane_authority_protocol_downgrade_authorizations TO nodecontrol_migration_downgrader;
GRANT DELETE ON nodecontrol.control_plane_authority_protocol_migration_latches TO nodecontrol_migration_downgrader;
GRANT EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) TO nodecontrol_migration_downgrader;

-- talenro:statement
GRANT USAGE ON SCHEMA nodecontrol TO nodecontrol_staging_importer;
GRANT SELECT ON
  nodecontrol.control_plane_authority_staging_import_capabilities,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_intents,
  nodecontrol.control_plane_authority_staging_import_capability_recovery_applications,
  nodecontrol.control_plane_authority_staging_import_capability_revocation_applications,
  nodecontrol.control_plane_authority_fresh_restore_import_applications
TO nodecontrol_staging_importer;
GRANT UPDATE (body_digest) ON nodecontrol.control_plane_authority_staging_import_capabilities TO nodecontrol_staging_importer;
GRANT INSERT ON
  nodecontrol.node_pops,
  nodecontrol.node_failure_domains,
  nodecontrol.node_capacity_profiles,
  nodecontrol.node_inventory,
  nodecontrol.control_plane_authority_fresh_restore_import_applications
TO nodecontrol_staging_importer;
GRANT EXECUTE ON FUNCTION nodecontrol.v7_require_role(name) TO nodecontrol_staging_importer;

-- talenro:statement
GRANT EXECUTE ON FUNCTION nodecontrol.v7_acquire_source_freeze_for_seal() TO CURRENT_USER;
GRANT EXECUTE ON FUNCTION nodecontrol.v7_insert_downgrade_authorization(uuid,uuid,bytea,bytea,bigint,bytea,bytea,bytea,bytea,numeric,bytea,text,timestamptz,timestamptz,bytea,bytea,bytea,bytea,bytea,bytea) TO CURRENT_USER;
GRANT EXECUTE ON FUNCTION nodecontrol.v7_consume_down_guard(uuid,bytea,bytea,bytea,bytea,bytea) TO CURRENT_USER;
GRANT EXECUTE ON FUNCTION nodecontrol.begin_staging_import(bytea,bytea,bytea,bytea,bytea,bytea,bytea,jsonb) TO CURRENT_USER;

DO $postflight$
DECLARE
  bootstrap_oid oid;
  role_count bigint;
  drift_count bigint;
BEGIN
  IF session_user IS DISTINCT FROM current_user THEN
    RAISE EXCEPTION 'authority v7 postflight requires session_user=current_user' USING ERRCODE = '42501';
  END IF;
  SELECT oid INTO bootstrap_oid
  FROM pg_catalog.pg_roles
  WHERE rolname = current_user AND rolsuper;
  IF bootstrap_oid IS NULL
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_database
       WHERE datname = current_database() AND datdba = bootstrap_oid
     )
     OR NOT EXISTS (
       SELECT 1 FROM pg_catalog.pg_namespace
       WHERE nspname = 'nodecontrol' AND nspowner = bootstrap_oid
     ) THEN
    RAISE EXCEPTION 'authority v7 postflight requires the exact bootstrap owner' USING ERRCODE = '42501';
  END IF;

  SELECT count(*) INTO role_count
  FROM pg_catalog.pg_roles AS role_catalog
  JOIN pg_catalog.pg_authid AS authorization_catalog ON authorization_catalog.oid = role_catalog.oid
  WHERE role_catalog.rolname IN (
      'nodecontrol_upgrade_executor',
      'nodecontrol_migration_downgrader',
      'nodecontrol_staging_importer'
    )
    AND NOT role_catalog.rolcanlogin
    AND NOT role_catalog.rolsuper
    AND NOT role_catalog.rolcreatedb
    AND NOT role_catalog.rolcreaterole
    AND role_catalog.rolinherit
    AND NOT role_catalog.rolreplication
    AND NOT role_catalog.rolbypassrls
    AND role_catalog.rolconnlimit = -1
    AND role_catalog.rolvaliduntil IS NULL
    AND role_catalog.rolconfig IS NULL
    AND authorization_catalog.rolpassword IS NULL;
  SELECT count(*) INTO drift_count
  FROM pg_catalog.pg_db_role_setting AS setting
  WHERE setting.setrole IN (
    'nodecontrol_upgrade_executor'::regrole,
    'nodecontrol_migration_downgrader'::regrole,
    'nodecontrol_staging_importer'::regrole
  );
  drift_count := drift_count + (
    SELECT count(*)
    FROM pg_catalog.pg_auth_members AS membership
    WHERE membership.member IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
       OR membership.roleid IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
  );
  drift_count := drift_count + (
    SELECT count(*)
    FROM pg_catalog.pg_default_acl AS default_acl
    LEFT JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = default_acl.defaclnamespace
    CROSS JOIN LATERAL pg_catalog.aclexplode(default_acl.defaclacl) AS acl
    WHERE default_acl.defaclrole IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
       OR acl.grantee IN (
        'nodecontrol_upgrade_executor'::regrole,
        'nodecontrol_migration_downgrader'::regrole,
        'nodecontrol_staging_importer'::regrole
      )
       OR (
        default_acl.defaclrole = bootstrap_oid
        AND default_acl.defaclobjtype IN ('r', 'S', 'f', 'T', 'n')
        AND (default_acl.defaclnamespace = 0 OR namespace.nspname = 'nodecontrol')
        AND acl.grantee <> bootstrap_oid
      )
  );
  IF role_count <> 3 OR drift_count <> 0 THEN
    RAISE EXCEPTION 'authority v7 bootstrap postflight drift' USING ERRCODE = '55000';
  END IF;
END
$postflight$;

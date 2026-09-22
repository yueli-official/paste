-- Paste authorization catalog v2 adds developer-token capabilities. Catalog
-- changes are explicit so the PostgreSQL adapter continues to fail closed on
-- unknown drift while preserving existing grants and policy revisions.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM authorization_instances
         WHERE instance_key LIKE 'paste:%'
           AND (
               catalog_version <> 1
               OR catalog_digest <> '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb'
           )
    ) THEN
        RAISE EXCEPTION 'Paste authorization catalog is not at the expected v1 digest';
    END IF;
END
$$;

INSERT INTO authorization_policy_bindings (
    instance_key, revision, target_kind, target_key, capability_key
)
SELECT revisions.instance_key,
       revisions.revision,
       'access_layer',
       'authenticated',
       capabilities.capability_key
  FROM authorization_policy_revisions AS revisions
  JOIN authorization_instances AS instances
    ON instances.instance_key = revisions.instance_key
 CROSS JOIN (VALUES
       ('paste.paste.create'),
       ('paste.paste.read'),
       ('paste.paste.update'),
       ('paste.paste.delete')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'paste:%'
   AND instances.catalog_version = 1
   AND instances.catalog_digest = '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb'
ON CONFLICT DO NOTHING;

INSERT INTO authorization_policy_bindings (
    instance_key, revision, target_kind, target_key, capability_key
)
SELECT revisions.instance_key,
       revisions.revision,
       'role',
       'administrator',
       capabilities.capability_key
  FROM authorization_policy_revisions AS revisions
  JOIN authorization_instances AS instances
    ON instances.instance_key = revisions.instance_key
 CROSS JOIN (VALUES
       ('paste.paste.moderate'),
       ('paste.settings.manage')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'paste:%'
   AND instances.catalog_version = 1
   AND instances.catalog_digest = '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb'
ON CONFLICT DO NOTHING;

INSERT INTO authorization_projection_rules (
    instance_key, policy_revision, rule_kind, subject_key, role_key,
    capability_key, scope_id, provenance
)
SELECT instances.instance_key,
       instances.active_policy_revision,
       'access_layer',
       'authenticated',
       'authenticated',
       capabilities.capability_key,
       instances.root_scope_id,
       '{}'::jsonb
  FROM authorization_instances AS instances
 CROSS JOIN (VALUES
       ('paste.paste.create'),
       ('paste.paste.read'),
       ('paste.paste.update'),
       ('paste.paste.delete')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'paste:%'
   AND instances.catalog_version = 1
   AND instances.catalog_digest = '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb'
ON CONFLICT DO NOTHING;

INSERT INTO authorization_projection_rules (
    instance_key, policy_revision, rule_kind, subject_key, role_key,
    capability_key, scope_id, provenance
)
SELECT instances.instance_key,
       instances.active_policy_revision,
       'permission',
       '',
       roles.id,
       capabilities.capability_key,
       instances.root_scope_id,
       jsonb_build_object('role_key', roles.role_key)
  FROM authorization_instances AS instances
  JOIN authorization_role_definitions AS roles
    ON roles.instance_key = instances.instance_key
   AND roles.role_key = 'administrator'
   AND roles.protected = TRUE
 CROSS JOIN (VALUES
       ('paste.paste.moderate'),
       ('paste.settings.manage')
 ) AS capabilities(capability_key)
 WHERE instances.instance_key LIKE 'paste:%'
   AND instances.catalog_version = 1
   AND instances.catalog_digest = '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb'
ON CONFLICT DO NOTHING;

UPDATE authorization_instances
   SET catalog_version = 2,
       catalog_digest = 'c420c9b71152c9cbe1b48c629f3d7a5f2ab47ef299a8391430e82e6042324c67',
       updated_at = NOW()
 WHERE instance_key LIKE 'paste:%'
   AND catalog_version = 1
   AND catalog_digest = '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb';

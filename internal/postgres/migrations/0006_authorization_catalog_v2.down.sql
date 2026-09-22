DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM authorization_instances
         WHERE instance_key LIKE 'paste:%'
           AND (
               catalog_version <> 2
               OR catalog_digest <> 'c420c9b71152c9cbe1b48c629f3d7a5f2ab47ef299a8391430e82e6042324c67'
           )
    ) THEN
        RAISE EXCEPTION 'Paste authorization catalog is not at the expected v2 digest';
    END IF;
END
$$;

DELETE FROM authorization_projection_rules
 WHERE instance_key LIKE 'paste:%'
   AND capability_key IN (
       'paste.paste.create',
       'paste.paste.read',
       'paste.paste.update',
       'paste.paste.delete',
       'paste.paste.moderate',
       'paste.settings.manage'
   );

DELETE FROM authorization_policy_bindings
 WHERE instance_key LIKE 'paste:%'
   AND capability_key IN (
       'paste.paste.create',
       'paste.paste.read',
       'paste.paste.update',
       'paste.paste.delete',
       'paste.paste.moderate',
       'paste.settings.manage'
   );

UPDATE authorization_instances
   SET catalog_version = 1,
       catalog_digest = '9a23c1d75581807104c198d21ca22e4185ae621ea2de44f3179220cd26533aeb',
       updated_at = NOW()
 WHERE instance_key LIKE 'paste:%'
   AND catalog_version = 2
   AND catalog_digest = 'c420c9b71152c9cbe1b48c629f3d7a5f2ab47ef299a8391430e82e6042324c67';

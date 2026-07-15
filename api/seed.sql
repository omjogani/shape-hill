-- Local demo data. Re-runnable: it clears its own user first.
--   docker compose exec -T postgres psql -U postgres -d shapehill < seed.sql

DELETE FROM users WHERE email = 'om@shapehill.dev';

INSERT INTO users (id, email, username, name)
VALUES ('a0000000-0000-4000-8000-000000000001', 'om@shapehill.dev', 'omj', 'Om');

INSERT INTO hills (id, owner_id, slug, title, description, is_public, starts_on, ends_on)
VALUES ('b0000000-0000-4000-8000-000000000001',
        'a0000000-0000-4000-8000-000000000001',
        'demo-billing',
        'Billing v2',
        'Six week cycle, four scopes.',
        true,
        current_date - 28, current_date + 14);

INSERT INTO scopes (id, hill_id, title, color, sort_order) VALUES
  ('c0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000001', 'Card on file',          '#2F4C64', 1),
  ('c0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000001', 'Retry failed charges',  '#55704B', 2),
  ('c0000000-0000-4000-8000-000000000003', 'b0000000-0000-4000-8000-000000000001', 'Invoice emails',        '#937129', 3),
  ('c0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000001', 'Admin refunds',         '#6B4A63', 4);

-- Each scope's trail. The newest row is where the dot sits now.
INSERT INTO scope_positions (scope_id, position, note, moved_by, created_at) VALUES
  ('c0000000-0000-4000-8000-000000000001',  8, 'Spiking Stripe tokenization',   'a0000000-0000-4000-8000-000000000001', now() - interval '26 days'),
  ('c0000000-0000-4000-8000-000000000001', 45, 'Approach proven end to end',    'a0000000-0000-4000-8000-000000000001', now() - interval '14 days'),
  ('c0000000-0000-4000-8000-000000000001', 95, 'Shipped behind a flag',         'a0000000-0000-4000-8000-000000000001', now() - interval '1 day'),

  ('c0000000-0000-4000-8000-000000000002',  5, 'Reading up on dunning',         'a0000000-0000-4000-8000-000000000001', now() - interval '25 days'),
  ('c0000000-0000-4000-8000-000000000002', 31, 'Backoff schedule chosen',       'a0000000-0000-4000-8000-000000000001', now() - interval '11 days'),
  ('c0000000-0000-4000-8000-000000000002', 72, 'Writing the retry worker',      'a0000000-0000-4000-8000-000000000001', now() - interval '2 days'),

  ('c0000000-0000-4000-8000-000000000003',  3, 'Blocked on template system',    'a0000000-0000-4000-8000-000000000001', now() - interval '24 days'),
  ('c0000000-0000-4000-8000-000000000003', 36, 'Dropped HTML tables for MJML',  'a0000000-0000-4000-8000-000000000001', now() - interval '9 days'),
  ('c0000000-0000-4000-8000-000000000003', 61, 'Rendering correctly',           'a0000000-0000-4000-8000-000000000001', now() - interval '3 days'),

  -- Stalled: last moved 12 days ago, still uphill. This is the one the chart exists to surface.
  ('c0000000-0000-4000-8000-000000000004',  4, 'Scoping the admin surface',     'a0000000-0000-4000-8000-000000000001', now() - interval '23 days'),
  ('c0000000-0000-4000-8000-000000000004', 15, 'Permissions model unclear',     'a0000000-0000-4000-8000-000000000001', now() - interval '12 days');

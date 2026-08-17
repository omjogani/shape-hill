-- This supabase cron would ensure API is up and running

-- Enable the scheduler and the async HTTP client
create extension if not exists pg_cron;
create extension if not exists pg_net;

-- Hit API healthz endpoint every 14 mintues (intentional)
select cron.schedule(
  'keep-api-awake',
  '*/14 * * * *',
  $$
  select net.http_get(
    url := 'https://API-DOMAIN/api/healthz',
    timeout_milliseconds := 10000
  );
  $$
);

-- Verify:

-- Fetch the scheduled cron and confirm it is active
select * from cron.job where jobname = 'keep-api-awake';

-- Fetch all the already executed crons, newest first
select status, return_message, start_time from cron.job_run_details
    where jobid = (select jobid from cron.job where jobname = 'keep-api-awake')
    order by start_time desc limit 10;

-- Fetch what the API actually answered, 200 means the ping landed
select status_code, created from net._http_response order by created desc limit 10;

-- Remove:
-- Delete the cron so it stops firing
select cron.unschedule('keep-api-awake');

BEGIN;
DO 28545
DECLARE
  i int := 60;
  ts int := 1750794780;
BEGIN
  WHILE i >= 17 LOOP
    EXECUTE format('UPDATE p_post SET created_on = %s, latest_replied_on = %s WHERE (user_id->>0)::bigint = 45 AND tags = %L;', ts, ts, i::text);
    ts := ts + 86400;
    i := i - 1;
  END LOOP;
END
28545;
COMMIT;

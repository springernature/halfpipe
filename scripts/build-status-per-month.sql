WITH months AS (
    SELECT
        to_char(start_time, 'YYYY-MM') ym,
        status,
        count(id) AS total
    FROM
        builds
    WHERE
            start_time > '2019-01-01'::TIMESTAMP
      AND aborted = FALSE
      AND team_id != 10 /* exclude ee because we run lots of tests */
    GROUP BY ym, status
),
     times AS (
         SELECT
             to_char(start_time, 'YYYY-MM') ym,
             avg(end_time - start_time) mean,
             percentile_cont(0.95) WITHIN group (ORDER BY end_time - start_time) percentile_95
         FROM
             builds
         WHERE
                 start_time > '2019-01-01'::TIMESTAMP
           AND status = 'succeeded'
           AND end_time IS NOT NULL
           AND team_id != 10 /* exclude ee because we run lots of tests */
         GROUP BY ym
     )



SELECT
    m.ym year_month,
    SUM(m.total)::INT AS total_builds,
    (SELECT TRUNC(100 * m1.total / sum(m.total), 1) FROM months m1 WHERE m1.ym = m.ym AND m1.status = 'succeeded') succeeded,
    (SELECT TRUNC(100 * m1.total / sum(m.total), 1) FROM months m1 WHERE m1.ym = m.ym AND m1.status = 'failed') failed,
    (SELECT TRUNC(100 * m1.total / sum(m.total), 1) FROM months m1 WHERE m1.ym = m.ym AND m1.status = 'errored') errored,
    (SELECT to_char(mean, 'MIm SSs') FROM times t WHERE t.ym = m.ym) duration_mean,
    (SELECT to_char(percentile_95, 'MIm SSs') FROM times t WHERE t.ym = m.ym) duration_95_percentile
FROM months m
GROUP BY m.ym
ORDER BY m.ym

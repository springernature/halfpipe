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
    GROUP BY 
        ym, status
)


SELECT
    m.ym year_month,
    SUM(m.total)::INT AS total_builds,
    (SELECT TRUNC(100 * m1.total / sum(m.total), 1) FROM months m1 WHERE m1.ym = m.ym AND m1.status = 'succeeded') succeeded,
    (SELECT TRUNC(100 * m1.total / sum(m.total), 1) FROM months m1 WHERE m1.ym = m.ym AND m1.status = 'failed') failed,
    (SELECT TRUNC(100 * m1.total / sum(m.total), 1) FROM months m1 WHERE m1.ym = m.ym AND m1.status = 'errored') errored
FROM months m
GROUP BY m.ym
ORDER BY m.ym

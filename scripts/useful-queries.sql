-- which pipelines have the most builds?
select b.pipeline_id, p.name, t.name, count(0) n_builds
from builds b
inner join pipelines p on p.id = b.pipeline_id
inner join teams t on t.id = p.team_id
group by b.pipeline_id, p.name, t.name
order by n_builds desc
limit 50;


-- how many resource versions does a pipeline have?
select count(0)
from build_resource_config_version_outputs i
inner join builds b on b.id = i.build_id
inner join pipelines p on p.id = b.pipeline_id
where p.name = 'halfpipe-janitor'
limit 10;


-- which pipelines have the most resource input versions?
select p.name, t.name, count(0) n_versions
from build_resource_config_version_inputs i
inner join builds b on b.id = i.build_id
inner join pipelines p on p.id = b.pipeline_id
inner join teams t on b.team_id = t.id
group by p.name, t.name
order by n_versions desc
limit 20;

-- count a pipeline's resource versions by name
select i.name, count(0) as n
from build_resource_config_version_inputs i
inner join builds b on b.id = i.build_id
inner join pipelines p on p.id = b.pipeline_id
where p.name = 'oscar-sites-nature'
group by i.name
order by n desc
limit 10;


-- table sizes
SELECT
    schema_name,
    relname,
    pg_size_pretty(table_size) AS size,
    table_size,
    reltuples

FROM (
         SELECT
             pg_catalog.pg_namespace.nspname           AS schema_name,
             relname,
             pg_relation_size(pg_catalog.pg_class.oid) AS table_size,
                reltuples

         FROM pg_catalog.pg_class
                  JOIN pg_catalog.pg_namespace ON relnamespace = pg_catalog.pg_namespace.oid
     ) t
WHERE schema_name NOT LIKE 'pg_%'
ORDER BY table_size DESC;


-- running queries
SELECT pid, age(clock_timestamp(), query_start), usename, state, query
FROM pg_stat_activity
WHERE query != '<IDLE>' AND query NOT ILIKE '%pg_stat_activity%' --and usename='cloudsqladmin'
ORDER BY query_start asc;


-- slow queries
SELECT
        total_time / calls AS avg_time,
        calls,
        total_time,
        rows,
        100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent,
        regexp_replace(query, '[\s\t\n]+', ' ', 'g')
FROM pg_stat_statements
WHERE query NOT LIKE '%EXPLAIN%'
  AND query NOT LIKE '%INDEX%'
ORDER BY avg_time DESC
LIMIT 10;

-- kill the 10 longest running queries
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE query != '<IDLE>' AND query NOT ILIKE '%pg_stat_activity%' and usename='concourse'
ORDER BY query_start asc
limit 10;


-- optimised LoadVersionsDB() query
SELECT v.id, v.check_order, r.id, i.build_id, i.name, b.job_id, b.status = 'succeeded'
FROM build_resource_config_version_inputs i
    JOIN builds b ON b.id = i.build_id
    JOIN resource_config_versions v ON v.version_md5 = i.version_md5
    JOIN resources r ON r.id = i.resource_id
WHERE r.resource_config_scope_id = v.resource_config_scope_id
  --AND (r.id, v.version_md5) NOT IN (SELECT resource_id, version_md5 from resource_disabled_versions)
  AND NOT EXISTS (SELECT 1 from resource_disabled_versions where resource_id = r.id and version_md5 = v.version_md5)
  AND v.check_order <> 0
  AND r.pipeline_id = 619;




-- count disabled versions by team
select t.name, count(0) c
from resource_disabled_versions rdv
         join resources r on r.id = rdv.resource_id
         join pipelines p on r.pipeline_id = p.id
         join teams t on p.team_id = t.id
where r.name = 'version'
group by t.name
order by c desc;


-- delete disabled versions - change "select count(0)" to "delete"
--delete
select count(0)
from resource_disabled_versions
where (resource_id, version_md5) in (
    select rdv.*
    from resource_disabled_versions rdv
    join resources r on r.id = rdv.resource_id
    join pipelines p on r.pipeline_id = p.id
    join teams t on p.team_id = t.id
    where r.name = 'version'
);

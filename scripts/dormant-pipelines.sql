with latest as (
    select max(start_time) latest_build, pipeline_id
    from builds
    group by pipeline_id
)


select p.id, p.name, l.latest_build
from pipelines p
inner join latest l on l.pipeline_id = p.id
where l.latest_build  < now() - interval '3 months'
order by l.latest_build asc


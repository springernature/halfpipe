with deploys as (
    select id, team_id, to_char(start_time, 'YYYY-MM-DD') ymd
    from builds
    where public_plan::text like '%promote%'
      and status = 'succeeded'
      and start_time > current_timestamp - interval '1 month'
)


select ymd, count(id) as total
from deploys
group by ymd
order by ymd

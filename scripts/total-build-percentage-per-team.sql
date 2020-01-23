with totalHours as (select round(extract(epoch from sum(end_time - start_time))) as totalHours from builds where start_time > current_timestamp - interval '1 day')
select  x.name as team, round(x.totalDiff/3600) as totalBuildHours, (x.totalDiff/t.totalHours)*100 as percentage
from (
         select round(extract(epoch from sum(end_time - start_time))) as totalDiff, t.name
         from builds
                  join teams t on builds.team_id = t.id
         where start_time > current_timestamp - interval '1 day'
         group by t.name
         order by totalDiff desc
     ) as x, totalHours as t;

select round(extract(epoch from totalDiff) / 3600), name
from (
         select sum(end_time - start_time) as totalDiff, t.name
         from builds
                  join teams t on builds.team_id = t.id
         where start_time > current_timestamp - interval '2 month'
         group by t.name
         order by totalDiff desc
     ) as x;
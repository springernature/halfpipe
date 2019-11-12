select ce, count(name)
from (select name, config::json->>'check_every' ce from resources) x
group by ce

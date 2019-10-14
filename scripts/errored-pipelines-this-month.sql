create or replace function urlencode(in_str text, OUT _result text) returns text as $$
  select
    string_agg(
      case
        when ol>1 or ch !~ '[0-9a-za-z:/@._?#-]+'
          then regexp_replace(upper(substring(ch::bytea::text, 3)), '(..)', E'%\\1', 'g')
        else ch
      end,
      ''
    )
  from (
    select ch, octet_length(ch) as ol
    from regexp_split_to_table($1, '') as ch
  ) as s;
$$ language sql immutable strict;

select
	b.id,
	concat (
		'https://concourse.halfpipe.io/teams/',
		t.name,
		'/pipelines/',
		p.name,
		'/jobs/',
		urlencode(j.name),
		'/builds/',
		b.name
	)
from
	builds b
LEFT JOIN
	pipelines p
	ON b.pipeline_id = p.id
LEFT JOIN
	jobs j
	ON b.job_id = j.id
LEFT JOIN
	teams t
	ON t.id = p.team_id
WHERE
	b.create_time > date_trunc('month', CURRENT_DATE) AND
	b.status = 'errored'
ORDER BY
	b.id;

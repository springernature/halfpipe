select
	concat (
		'https://concourse.halfpipe.io/teams/',
		t.name,
		'/pipelines/',
		p.name,
		'/jobs/',
		j.name,
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
	b.status = 'errored';

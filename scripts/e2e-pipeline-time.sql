WITH vars (PIPELINE_NAME, JOB_NAME) AS (
    VALUES ('ee-docs', 'deploy')
),
end_job AS (
    SELECT
        inputs.*,
        builds.end_time
    FROM
        resource_config_versions versions,
        build_resource_config_version_inputs AS inputs,
        builds,
        resources,
        jobs,
        pipelines,
        vars
    WHERE
        builds.job_id = jobs.id
        AND versions.check_order != 0
        AND versions.metadata != 'null'
        AND inputs.build_id = builds.id
        AND inputs.version_md5 = versions.version_md5
        AND resources.resource_config_scope_id = versions.resource_config_scope_id
        AND resources.id = inputs.resource_id
        AND pipelines.id = jobs.pipeline_id
        AND NOT EXISTS (
            SELECT 1
            FROM build_resource_config_version_outputs outputs
            WHERE
                outputs.version_md5 = versions.version_md5
                AND versions.resource_config_scope_id = resources.resource_config_scope_id
                AND outputs.resource_id = resources.id
                AND outputs.build_id = inputs.build_id
        )
        AND resources.name = 'git'
        AND builds.job_id = jobs.id
        AND status = 'succeeded'
        AND pipelines.name = vars.PIPELINE_NAME
        AND jobs.name = vars.JOB_NAME
 )
, times AS (
    SELECT min(builds.start_time) AS start_time, end_job.end_time
    FROM
        builds,
        build_resource_config_version_inputs AS inputs,
        end_job
    WHERE
        inputs.build_id = builds.id
        AND inputs.version_md5 = end_job.version_md5
        -- AND $__timeFilter(end_job.end_time)
    GROUP BY
        inputs.version_md5, end_job.end_time
)

SELECT end_time AS time, extract(EPOCH FROM (end_time - start_time)) AS duration
FROM times
order by end_time
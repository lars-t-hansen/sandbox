-- psql -f q.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > foo
--
-- what I want is for cluster = fox.educloud.no and a time window t1..t2
--
-- temporary table JOB with max(time) per sample_slurm_job and user_name not blank
--   we want job_id, job_name, account, user_name
--
-- temporary table PROC with max(time) per (process,node) from sample_process and pid not zero and job id not zero
--   we want cpu_time, job
--
-- left join PROC and JOB by job id - there will be many processes per job, all get the
-- same job data (really we care about account, user, job name only)
--
-- now we have a table with slurm and job data for every process

DROP TABLE IF EXISTS job;
SELECT job_id, job_state, account, user_name, time, job_name
INTO job
FROM sample_slurm_job
WHERE ( job_id, time ) IN
( SELECT job_id, max(time)
  FROM sample_slurm_job
  WHERE cluster = 'fox.educloud.no'
  AND time >= '2026-04-01'
  GROUP BY job_id )
AND user_name != ''
-- AND job_state != 'FAILED'
-- AND job_state != 'TIMEOUT'
AND job_state != 'PENDING'
-- AND not (job_state like 'CANCELLED%')
ORDER BY job_id
;

DROP TABLE IF EXISTS proc;
SELECT cpu_time, job, node, pid, cmd, time
INTO proc
FROM sample_process
WHERE (job, node, time) IN
( SELECT job, node, max(time)
  FROM sample_process
  WHERE cluster = 'fox.educloud.no'
  AND time >= '2026-04-01'
  GROUP BY job, node )
AND pid != 0
AND job != 0
ORDER BY job
;

SELECT * from proc left join job on proc.job = job.job_id ;

-- For GPU, we don't have gpu_time like we have cpu_time, but we can probably create it by averaging gpu_util
-- and then multiplying that by end-start time.  In this case we don't want the max of the proc table, we
-- want to just average by process I think and then start and end are max and min across time.  Otherwise the
-- logic is probably the same.

-- psql -f gpu_time.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > foo
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

DROP TABLE IF EXISTS proc_gpu;

SELECT avg(gpu_util), t1.job, t1.node, t1.pid, cmd, uuid
INTO proc_gpu
FROM sample_process_gpu AS t1
LEFT JOIN sample_process AS t2
  ON t1.pid = t2.pid AND t1.node = t2.node AND t1.time = t2.time
WHERE t1.cluster = 'fox.educloud.no'
AND t1.time >= '2026-02-01'
AND t1.pid != 0
GROUP BY t1.job, t1.node, t1.pid, cmd, uuid
ORDER BY t1.job
;

DROP TABLE IF EXISTS job;

SELECT job_id, job_state, account, user_name, time, job_name, start_time, end_time
INTO job
FROM sample_slurm_job
WHERE ( job_id, time ) IN
( SELECT job_id, max(time)
  FROM sample_slurm_job
  WHERE cluster = 'fox.educloud.no'
  AND time >= '2026-02-01'
  GROUP BY job_id )
AND user_name != ''
AND job_state != 'PENDING'
ORDER BY job_id
;

-- DROP TABLE IF EXISTS proc;
-- SELECT cpu_time, job, node, pid, time
-- INTO proc
-- FROM sample_process
-- WHERE (job, node, time) IN
-- ( SELECT job, node, max(time)
--   FROM sample_process
--   WHERE cluster = 'fox.educloud.no'
--   AND time >= '2026-02-01'
--   GROUP BY job, node )
-- AND pid != 0
-- AND job != 0
-- ORDER BY job
-- ;

-- This table has a lot of pid==0, I don't like it.  Mostly for things that don't use GPU but not always.

SELECT * from proc_gpu left join job on proc_gpu.job = job.job_id ;

-- derived from gpu_time.sql and could probably just replace that.

DROP TABLE IF EXISTS proc_gpu;

SELECT avg(gpu_util), t1.job, t1.node, t1.pid, cmd, t1.uuid, manufacturer, model, memory
INTO proc_gpu
FROM sample_process_gpu AS t1
LEFT JOIN sample_process AS t2
  ON t1.pid = t2.pid AND t1.node = t2.node AND t1.time = t2.time
LEFT JOIN sysinfo_gpu_card as t3
  ON t1.uuid = t3.uuid
WHERE t1.cluster = 'fox.educloud.no'
AND t1.time >= '2026-02-01'
AND t1.pid != 0
GROUP BY t1.job, t1.node, t1.pid, cmd, t1.uuid, manufacturer, model, memory
ORDER BY t1.job
;

DROP TABLE IF EXISTS job;

SELECT job_id, time, start_time, end_time
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

-- Once we have this, we want to aggregate gpu time by (manufacturer, model, memory) probably,
-- not by command name or job name, but it amounts to the same.

SELECT * FROM proc_gpu LEFT JOIN job ON proc_gpu.job = job.job_id ;

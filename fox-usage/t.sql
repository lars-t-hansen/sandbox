-- psql -f t.sql postgresql://naicadmin:naicpw@naic-monitor.uio.no:10102/naicmon > foo
-- 
-- "select row with max time attribute from table"

SELECT * FROM cluster_attributes WHERE (cluster, time) IN ( SELECT cluster, MAX(time) from cluster_attributes group by cluster );

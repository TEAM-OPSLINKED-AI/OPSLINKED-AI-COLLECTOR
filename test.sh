$ go run -ldflags="-extldflags=-lssp" main.go
2025/09/03 12:59:49 설정 로드가 완료되었습니다.
2025/09/03 12:59:50 메트릭 수집 및 전달 스케줄러를 시작합니다. 1분마다 실행됩니다.
2025/09/03 12:59:50 작업을 시작합니다...
2025/09/03 12:59:50 INFO: 'MySQL Exporter'의 메트릭 수집을 시작합니다...
2025/09/03 12:59:50 INFO: 'Spring BE Actuator'의 메트릭 수집을 시작합니다...
2025/09/03 12:59:50 INFO: 'Node Exporter'의 메트릭 수집을 시작합니다...
2025/09/03 12:59:50 INFO: 'Node Exporter'의 메트릭 1464개를 'node-exporter-metrics' 토픽으로 성공적으로 전송 요청했습니다.
2025/09/03 12:59:50 INFO: 'MySQL Exporter'의 메트릭 1064개를 'mysql-exporter-metrics' 토픽으로 성공적으로 전송 요청했습니다.
2025/09/03 12:59:50 INFO: 'Spring BE Actuator'의 메트릭 329개를 'spring-actuator-metrics' 토픽으로 성공적으로 전송 요청했습니다.
2025/09/03 12:59:50 모든 작업이 완료되었습니다.
2025/09/03 13:00:50 작업을 시작합니다...
2025/09/03 13:00:50 INFO: 'Node Exporter'의 메트릭 수집을 시작합니다...
2025/09/03 13:00:50 INFO: 'Spring BE Actuator'의 메트릭 수집을 시작합니다...
2025/09/03 13:00:50 INFO: 'MySQL Exporter'의 메트릭 수집을 시작합니다...
2025/09/03 13:00:50 INFO: 'Spring BE Actuator'의 메트릭 329개를 'spring-actuator-metrics' 토픽으로 성공적으로 전송 요청했습니다.
2025/09/03 13:00:50 INFO: 'MySQL Exporter'의 메트릭 1064개를 'mysql-exporter-metrics' 토픽으로 성공적으로 전송 요청했습니다.
2025/09/03 13:00:50 INFO: 'Node Exporter'의 메트릭 1464개를 'node-exporter-metrics' 토픽으로 성공적으로 전송 요청했습니다.
2025/09/03 13:00:50 모든 작업이 완료되었습니다.
...

$ bin/kafka-topics.sh --bootstrap-server _:9095  --list
kafka-deploy-test-topic
mysql-exporter-metrics
node-exporter-metrics
spring-actuator-metrics

$ bin/kafka-console-consumer.sh \
  --bootstrap-server _:9095  \
  --topic spring-actuator-metrics --from-beginning --partition 0
application_ready_time_seconds{main_application_class="org.example.whenwillwemeet.WhenWillWeMeetApplication"} 29.699
application_started_time_seconds{main_application_class="org.example.whenwillwemeet.WhenWillWeMeetApplication"} 29.389
disk_free_bytes{path="/."} 6.3171764224E10
disk_total_bytes{path="/."} 6.7925028864E10
executor_active_threads{name="applicationTaskExecutor"} 0.0
executor_completed_tasks_total{name="applicationTaskExecutor"} 0.0
executor_pool_core_threads{name="applicationTaskExecutor"} 8.0
executor_pool_max_threads{name="applicationTaskExecutor"} 2.147483647E9
executor_pool_size_threads{name="applicationTaskExecutor"} 0.0
executor_queue_remaining_tasks{name="applicationTaskExecutor"} 2.147483647E9
executor_queued_tasks{name="applicationTaskExecutor"} 0.0
hikaricp_connections{pool="HikariPool-1"} 10.0
...
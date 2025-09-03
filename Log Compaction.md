현재 main.go의 데이터 수집 방식은 Kafka의 **Log Compaction** 기능을 활용하여 데이터가 무한정 쌓이는 것을 방지하고 항상 최신 상태만 유지하도록 설계되었습니다.

```go
// main.go
package main
...
		// 메트릭 라인을 공백 기준으로 분리하여 Key 추출
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 1 {
			continue
		}
		key := parts[0]
		value := line

		// Kafka 토픽으로 메시지 전송
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &source.Topic, Partition: 0},
			Key:            []byte(key), // Log Compaction
			Value:          []byte(value),
		}, nil)

		sentCount++
	}
```
Kafka 토픽을 생성할 때 `cleanup.policy=compact`로 설정하면, Kafka는 **동일한 Key**를 가진 메시지 중 **가장 최신 버전의 메시지 하나만 남기고** 나머지는 주기적으로 삭제합니다.

위 Go 코드의 `p.Produce` 부분을 보면, 각 메트릭 라인에서 메트릭 이름과 레이블(`node_cpu_seconds_total{...}` 등)을 추출하여 메시지의 \*\*`Key`\*\*로 사용하고 있습니다.

```go
Key:            []byte(key), // Log Compaction
```
1분마다 새로운 메트릭이 들어와도 **Key가 동일하기 때문에**, Kafka는 이전에 저장된 값을 새로운 값으로 '덮어쓰는(갱신하는)' 효과를 냅니다.

1.  **초기 증가**: 모든 종류의 메트릭 Key가 최소 한 번씩 들어올 때까지 디스크 사용량이 증가합니다.
2.  **안정 상태**: 그 이후로는 새로운 값이 들어와 일시적으로 용량이 늘어나더라도, Kafka의 컴팩션 작업이 주기적으로 오래된 값을 삭제하므로 평균적인 디스크 사용량은 더 이상 늘어나지 않습니다.

```bash
[root@master opslinked-ai]# kubectl exec -it kafka-broker-0 -n kafka -- /bin/bash
Defaulted container "kafka" out of: kafka, kafka-init (init)
1001@kafka-broker-0:/$ /opt/bitnami/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 --entity-type topics --entity-name spring-actuator-metrics --describe --all

synonyms={DEFAULT_CONFIG:log.cleaner.min.compaction.lag.ms=0}
  min.insync.replicas=1 sensitive=false synonyms={DEFAULT_CONFIG:min.insync.replicas=1}
  preallocate=false sensitive=false synonyms={DEFAULT_CONFIG:log.preallocate=false}
  remote.storage.enable=false sensitive=false synonyms={}
  retention.bytes=-1 sensitive=false synonyms={DEFAULT_CONFIG:log.retention.bytes=-1}
  retention.ms=604800000 sensitive=false synonyms={}
  segment.bytes=1073741824 sensitive=false synonyms={DEFAULT_CONFIG:log.segment.bytes=1073741824}


1001@kafka-broker-0:/$ /opt/bitnami/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 --entity-type topics --entity-name spring-actuator-metrics --alter --add-config segment.bytes=1024

[root@master opslinked-ai]# kubectl logs kafka-broker-0 -n kafka | tail -n 10
...
INFO [Admin Manager on Broker 100]: Updating topic spring-actuator-metrics with new configuration : segment.bytes -> 1024
INFO Starting cleaning of log spring-actuator-metrics-0 (kafka.log.LogCleaner)

1001@kafka-broker-0:/$ /opt/bitnami/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 --entity-type topics --entity-name spring-actuator-metrics --alter --add-config segment.bytes=1073741824
```
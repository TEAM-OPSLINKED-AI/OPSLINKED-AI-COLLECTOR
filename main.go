package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	// 사용자 정의 config 패키지를 import 합니다.
	"go-metric-producer/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// MetricSource 구조체는 메트릭 수집 대상 소스를 정의합니다.
// 여러 메트릭 소스를 순회하며 작업하기 위해 사용됩니다.
type MetricSource struct {
	Name  string // 메트릭 소스의 이름
	URL   string // 메트릭을 수집할 URL
	Topic string // 수집한 메트릭을 전송할 Kafka 토픽 이름
}

func main() {
	// 애플리케이션 시작 시점에 설정 파일을 로드합니다.
	config.Load()

	log.Println("메트릭 수집 및 전달 스케줄러를 시작합니다. 1분마다 실행됩니다.")

	// 1분 간격으로 이벤트를 발생시키는 Ticker를 생성합니다.
	ticker := time.NewTicker(1 * time.Minute)
	// main 함수가 종료될 때 티커를 정지시켜 리소스 누수를 방지합니다.
	defer ticker.Stop()

	// 애플리케이션 시작 즉시 작업을 한 번 실행합니다.
	runJobs()

	// 티커의 채널을 통해 신호가 올 때까지 대기하다가, 신호가 오면 runJobs()를 다시 실행합니다.
	// 이 for-range 루프는 프로그램이 종료될 때까지 계속됩니다.
	for range ticker.C {
		runJobs()
	}
}

// runJobs 함수는 모든 메트릭 소스에 대한 수집 작업을 동시에 시작합니다.
func runJobs() {
	// 모든 고루틴(goroutine)이 완료될 때까지 대기하기 위한 WaitGroup을 생성합니다.
	var wg sync.WaitGroup
	log.Println("작업을 시작합니다...")

	// 설정 파일에서 로드된 값을 사용하여 메트릭 소스 목록을 정의합니다.
	metricSources := []MetricSource{
		{Name: "Node Exporter", URL: config.Cfg.NodeExporterURL, Topic: "node-exporter-metrics"},
		{Name: "Spring BE Actuator", URL: config.Cfg.SpringActuatorURL, Topic: "spring-actuator-metrics"},
		{Name: "MySQL Exporter", URL: config.Cfg.MySQLExporterURL, Topic: "mysql-exporter-metrics"},
	}

	// 각 메트릭 소스에 대해 별도의 고루틴을 생성하여 병렬로 처리합니다.
	for _, source := range metricSources {
		// WaitGroup 카운터를 1 증가시킵니다.
		wg.Add(1)
		// fetchAndProduceMetrics 함수를 고루틴으로 실행합니다.
		// 반복문의 변수 'source'를 고루틴에 직접 전달하여, 각 고루틴이 올바른 값을 참조하도록 합니다.
		go func(s MetricSource) {
			// 고루틴이 종료될 때 WaitGroup 카운터를 1 감소시킵니다.
			defer wg.Done()
			fetchAndProduceMetrics(s)
		}(source)
	}

	// WaitGroup 카운터가 0이 될 때까지 여기서 대기합니다. 즉, 모든 고루틴이 종료될 때까지 기다립니다.
	wg.Wait()
	log.Println("모든 작업이 완료되었습니다.")
}

// fetchAndProduceMetrics 함수는 특정 소스에서 메트릭을 가져와 Kafka로 전송합니다.
func fetchAndProduceMetrics(source MetricSource) {
	// 설정 파일의 Kafka 브로커 주소를 사용하여 새로운 Kafka Producer를 생성합니다.
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.Cfg.KafkaBroker, // 접속할 Kafka 클러스터 주소
		"acks":              "all",                   // 모든 ISR(In-Sync Replica)의 확인을 기다려 메시지 손실 방지
		"retries":           5,                       // 메시지 전송 실패 시 재시도 횟수
	})

	if err != nil {
		log.Printf("ERROR: %s의 Kafka Producer 생성 실패: %v\n", source.Name, err)
		return
	}
	// 함수가 종료될 때 Producer를 안전하게 닫습니다.
	defer p.Close()

	// 별도의 고루틴에서 Producer의 이벤트를 비동기적으로 처리합니다.
	// 메시지 전송 성공/실패 여부를 확인하기 위함입니다.
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				// 메시지 전송에 실패한 경우 에러 로그를 출력합니다.
				if ev.TopicPartition.Error != nil {
					log.Printf("ERROR: 메시지 전송 실패: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	log.Printf("INFO: '%s'의 메트릭 수집을 시작합니다... (URL: %s)\n", source.Name, source.URL)
	// HTTP GET 요청을 통해 메트릭 소스 URL에서 데이터를 가져옵니다.
	resp, err := http.Get(source.URL)
	if err != nil {
		log.Printf("ERROR: '%s'의 메트릭 수집 실패: %v\n", source.Name, err)
		return
	}
	// 함수가 종료될 때 응답 본문(Body)을 닫습니다.
	defer resp.Body.Close()

	// HTTP 응답 상태 코드가 200 (OK)가 아닌 경우 에러를 기록하고 함수를 종료합니다.
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: '%s'의 응답 상태 코드가 올바르지 않습니다: %s\n", source.Name, resp.Status)
		return
	}

	// 응답 본문을 한 줄씩 효율적으로 읽기 위해 bufio.Reader를 사용합니다.
	reader := bufio.NewReader(resp.Body)
	sentCount := 0 // 전송 요청한 메트릭의 수를 세는 카운터

	// 무한 루프를 통해 응답 본문의 모든 줄을 읽어 처리합니다.
	for {
		// 개행 문자('\n')를 만날 때까지 한 줄을 읽습니다.
		line, err := reader.ReadString('\n')
		if err != nil {
			// io.EOF(End of File)는 파일의 끝을 의미하는 정상적인 종료 신호이므로 에러로 처리하지 않습니다.
			if err != io.EOF {
				log.Printf("ERROR: '%s'의 메트릭 파싱 중 오류 발생: %v\n", source.Name, err)
			}
			// 파일의 끝에 도달했거나 다른 오류가 발생하면 루프를 종료합니다.
			break
		}

		// 읽어온 줄의 앞뒤 공백을 제거합니다.
		line = strings.TrimSpace(line)
		// Prometheus 메트릭 형식에서 주석(#)으로 시작하거나 빈 줄인 경우는 건너뜁니다.
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// 메트릭 라인을 공백 기준으로 분리하여 메트릭 이름(key)을 추출합니다.
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 1 {
			continue // 유효하지 않은 형식의 라인은 건너뜁니다.
		}
		key := parts[0]
		value := line // 전체 라인을 값으로 사용합니다.

		// 모든 메시지를 동일한 파티션(0)으로 보냅니다.
		// 특정 키에 대한 순서 보장이 필요하다면 키를 해싱하여 파티션을 결정할 수 있습니다.
		partition := int32(0)
		// Kafka 토픽으로 메시지를 비동기적으로 전송(Produce)합니다.
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &source.Topic, Partition: partition},
			Key:            []byte(key),
			Value:          []byte(value),
		}, nil) // nil을 전달하면 별도의 전송 결과 채널을 사용하지 않음을 의미합니다.

		sentCount++
	}

	// Producer 내부에 버퍼링된 모든 메시지가 전송될 때까지 최대 15초간 대기합니다.
	remaining := p.Flush(15 * 1000)
	if remaining > 0 {
		log.Printf("WARN: '%s'의 메시지 %d개가 타임아웃 내에 전송되지 못했습니다.\n", remaining, source.Name)
	} else {
		log.Printf("INFO: '%s'의 메트릭 %d개를 '%s' 토픽으로 성공적으로 전송 요청했습니다.\n", source.Name, sentCount, source.Topic)
	}
}

# Go 모듈 초기화
go mod init go-metric-producer

# 환경변수 설정을 위한 라이브러리 설치치
go get github.com/joho/godotenv

# Confluent Kafka 라이브러리 설치
go get github.com/confluentinc/confluent-kafka-go/v2/kafka

#confluent-kafka-go는 내부적으로 C 라이브러리인 librdkafka를 사용합니다. 
# 시스템에 gcc와 librdkafka-dev (또는 librdkafka-devel) 패키지가 설치되어 있어야 사용할 수 있습니다.
# Ubuntu/Debian: 
sudo apt-get install -y librdkafka-dev
# CentOS/RHEL: 
sudo yum install -y librdkafka-devel
# macOS (Homebrew): 
brew install librdkafka

# 코드 빌드
go build -o metric-producer
# linker 오류시 ldflags 플래그로 빌드
go build -ldflags="-extldflags=-lssp" -o metric-producer

# 실행 파일 실행
./metric-producer

# main.go 실행
go run -ldflags="-extldflags=-lssp" main.go

# Spring Actuator Topic 소비 예시
# --property print.key=true 옵션으로 메시지 Key도 함께 확인
bin/kafka-console-consumer.sh \
  --bootstrap-server _:9095 \
  --topic spring-actuator-metrics \
  --from-beginning \
  --property print.key=true \
  --property key.separator=" | "
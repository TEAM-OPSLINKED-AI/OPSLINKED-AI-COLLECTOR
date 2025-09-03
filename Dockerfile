### Builder Stage
FROM golang:1.25 AS builder

# confluent-kafka-go 라이브러리가 CGo를 사용하므로,
# gcc와 관련 도구를 설치합니다.
RUN apt-get update && apt-get install -y gcc

# 작업 디렉토리를 설정합니다.
WORKDIR /app

# 의존성 관리를 위해 go.mod와 go.sum 파일을 먼저 복사합니다.
COPY go.mod go.sum ./
# go mod download를 실행하여 의존성을 미리 다운로드합니다.
RUN go mod download

# 프로젝트의 모든 소스 코드를 작업 디렉토리로 복사합니다.
COPY . .

# Go 프로그램을 빌드합니다.
# CGO_ENABLED=1은 CGo 사용을 명시적으로 활성화합니다.
# -o /metric-producer는 빌드 결과물인 실행 파일의 경로와 이름을 지정합니다.
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /metric-producer .

### Final Stage
FROM debian:12-slim

# SSL 통신 등에 필요한 ca-certificates를 설치합니다.
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# 작업 디렉토리를 설정합니다.
WORKDIR /

# 빌더 스테이지에서 생성된 실행 파일만 최종 이미지로 복사합니다.
# --from=builder 옵션을 통해 이전 스테이지를 참조할 수 있습니다.
COPY --from=builder /metric-producer .

# 컨테이너가 시작될 때 실행할 명령어를 지정합니다.
CMD ["./metric-producer"]
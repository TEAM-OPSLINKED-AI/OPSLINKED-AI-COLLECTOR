package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig 구조체는 애플리케이션에 필요한 모든 설정 값을 저장합니다.
// 이 값들은 환경 변수로부터 읽어옵니다.
type AppConfig struct {
	KafkaBroker       string // Kafka 브로커의 주소
	NodeExporterURL   string // Node Exporter의 메트릭 수집 URL
	SpringActuatorURL string // Spring Actuator의 메트릭 수집 URL
	MySQLExporterURL  string // MySQL Exporter의 메트릭 수집 URL
}

// Cfg는 로드된 애플리케이션 설정을 담는 전역 변수입니다.
// 다른 패키지에서 이 변수를 통해 설정 값에 접근할 수 있습니다.
var Cfg AppConfig

// Load 함수는 .env 파일이나 운영체제의 환경 변수에서 설정을 읽어와
// 전역 변수인 Cfg를 초기화합니다.
func Load() {
	// godotenv.Load() 함수는 현재 디렉터리에서 .env 파일을 찾아 그 안의 변수들을
	// 시스템의 환경 변수로 로드합니다. 개발 환경에서 유용합니다.
	err := godotenv.Load()
	if err != nil {
		// .env 파일이 없어도 에러는 아니며, 운영체제에 설정된 환경 변수를 사용하게 됩니다.
		// Docker나 Kubernetes 환경에서는 보통 .env 파일 없이 직접 환경 변수를 주입합니다.
		log.Println("경고: .env 파일을 찾을 수 없습니다. 운영체제의 환경 변수를 사용합니다.")
	}

	// 각 환경 변수를 읽어와 Cfg 구조체 필드를 채웁니다.
	Cfg = AppConfig{
		KafkaBroker:       getEnv("KAFKA_BROKER"),
		NodeExporterURL:   getEnv("EXPORTER_URL_NODE"),
		SpringActuatorURL: getEnv("EXPORTER_URL_SPRING"),
		MySQLExporterURL:  getEnv("EXPORTER_URL_MYSQL"),
	}
	log.Println("설정 로드가 완료되었습니다.")
}

// getEnv 함수는 지정된 키(key)에 해당하는 환경 변수 값을 읽어옵니다.
// 만약 환경 변수가 설정되어 있지 않다면, 프로그램을 즉시 종료시켜 설정 누락으로 인한 오작동을 방지합니다.
func getEnv(key string) string {
	value, ok := os.LookupEnv(key)
	// 'ok'는 해당 키의 환경 변수가 존재하는지 여부를 나타내는 boolean 값입니다.
	if !ok {
		// log.Fatalf는 에러 메시지를 출력하고 프로그램을 즉시 종료(exit status 1)시킵니다.
		// 필수 설정이 없는 상태로 프로그램이 실행되는 것을 막는 중요한 부분입니다.
		log.Fatalf("FATAL: 필수 환경 변수 '%s'가 설정되지 않았습니다.", key)
	}
	return value
}

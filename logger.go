package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const BuffSize = 1024

type Config struct {
	port          int
	dataMapSize   int
	sendingPeriod int
	opensearchURL string
}

var config Config
var mutex sync.Mutex

func main() {

	errSetValues := setValues()
	if errSetValues != nil {
		log.Printf("error setting values: %s\n", errSetValues)
		return
	}

	conn, errMakeConn := makeUDPConn("0.0.0.0", config.port)
	if errMakeConn != nil {
		log.Printf("error making connection: %s\n", errMakeConn)
		return
	}
	defer conn.Close()

	// Таймер
	ticker := time.NewTicker(time.Duration(config.sendingPeriod) * time.Second)
	defer ticker.Stop()

	// Массив хранения записей
	dataMap := make([]string, config.dataMapSize)
	recordNumber := 0

	// Буфер для чтения входящих данных
	buf := make([]byte, BuffSize)

	// Отправка по таймеру
	go func() {
		for {
			select {
			case <-ticker.C:
				mutex.Lock()
				if len(dataMap) > 0 {
					send(dataMap)
					dataMap = make([]string, config.dataMapSize)
				}
				mutex.Unlock()
			}
		}
	}()

	fmt.Println("waiting for logs...")
	for {
		// Чтение из UDP-соединения
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("data reading error: %s\n", err)
			continue
		}
		fmt.Println("new log:", string(buf[:n]))

		// Добавление данных в массив
		mutex.Lock()
		dataMap[recordNumber] = string(buf[:n])
		recordNumber++

		// Отправка по количеству записей
		if len(dataMap) >= config.dataMapSize {
			send(dataMap)
			clear(dataMap)
			recordNumber = 0
		}
		mutex.Unlock()
	}
}

func setValues() error {
	var err error

	// Получение порта
	config.port, err = getIntEnv("GOLOGGER_PORT")
	if err != nil {
		return fmt.Errorf("failed to set port: %w", err)
	}

	// Получение размера буфера логов
	config.dataMapSize, err = getIntEnv("GOLOGGER_DATAMAP_SIZE")
	if err != nil {
		return fmt.Errorf("failed to set data map size: %w", err)
	}

	// Получение промежутка времени отправки
	config.sendingPeriod, err = getIntEnv("GOLOGGER_SENDING_PERIOD_SEC")
	if err != nil {
		return fmt.Errorf("failed to set sending period: %w", err)
	}

	config.opensearchURL = "http://localhost:9200/php-logs/_doc"

	return nil
}

func getIntEnv(key string) (int, error) {
	//Получить int сразу из env
	value, exists := os.LookupEnv(key)
	if !exists {
		return 0, fmt.Errorf("%s not set", key)
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("failed to convert %s: %w", key, err)
	}
	return val, nil
}

func makeUDPConn(ip string, p int) (*net.UDPConn, error) {
	// Создать адрес
	addr := net.UDPAddr{
		Port: p,
		IP:   net.ParseIP(ip),
	}

	// Создать listener
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil, fmt.Errorf("error creating UDP-listener: %w", err)
	}

	return conn, nil
}

func send(dataMap []string) error {
	// Преобразование json
	jsonLogs, err := json.Marshal(dataMap)
	if err != nil {
		return fmt.Errorf("json conversion error: %w", err)
	}

	// Запрос
	req, err := http.NewRequest("POST", config.opensearchURL, bytes.NewBuffer(jsonLogs))
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}

	// Заголовки
	req.Header.Set("Content-Type", "application/json")

	// Отправка запроса.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

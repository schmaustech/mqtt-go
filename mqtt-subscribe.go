package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabriel-vasile/mimetype"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	server   = "tcp://192.168.0.206:30001"
	topic    = "transfer"
	savePath = "/root/inbound"
)

func main() {
	opts := MQTT.NewClientOptions().AddBroker(server)
	opts.SetClientID("file_subscriber")

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Error connecting to MQTT broker:", token.Error())
	}
	defer client.Disconnect(250)

	if token := client.Subscribe(topic, 0, onMessageReceived); token.Wait() && token.Error() != nil {
		log.Fatal("Error subscribing to topic:", token.Error())
	}
	fmt.Printf("Subscribed to topic: %s\n", topic)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Terminating...")
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	
	mimeType := mimetype.Detect(message.Payload())
	
	fileName := fmt.Sprintf("file_%d%s", time.Now().UnixNano(), mimeType.Extension())

	filePath := fmt.Sprintf("%s/%s", savePath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(message.Payload())
	if err != nil {
		log.Println("Error writing to file:", err)
		return
	}

	fmt.Printf("File saved successfully: %s\n", filePath)
}

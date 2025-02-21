package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// broker: mosquitto -v -p 1883
	// consumer: mosquitto_sub -p 1883 -t my-topic

	// simple producer
	opt := mqtt.NewClientOptions()
	opt.AddBroker("tcp://localhost:1883")
	client := mqtt.NewClient(opt)
	tok := client.Connect()
	println(tok.WaitTimeout(10*time.Second))
	if err := tok.Error() ; err != nil {
		fmt.Printf("%v\n", tok.Error())
		return
	}
	tok = client.Publish("my-topic", 0, false, "Hi there")
	println(tok.WaitTimeout(10*time.Second))
	if err := tok.Error() ; err != nil {
		fmt.Printf("%v\n", tok.Error())
		return
	}
}

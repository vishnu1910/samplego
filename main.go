package main

import (
	"fmt"
	"log"

	pb "github.com/vishnu1910/samplego"
	"google.golang.org/protobuf/proto"
)

func main() {
	person := &personpb.Person{
		Name: "John Doe",
		Age: 30,
	}

	data, err := proto.Marshall(person)
	if err != nil {
		log.Fatalf("Failed to encode person: %v", err)
	}

	fmt.Println("Serialized Data:", data)

	var newPerson personpb.Person
	err = proto.Unmarshal(data, &newPerson)
	if err != nil {
		log.Fatalf("Failed to decode person: %v", err)
	}

	fmt.Println("Deserialized Person:", newPerson)
}

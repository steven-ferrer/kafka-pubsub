package usermanagementsvc

import (
	"context"
	"fmt"

	"github.com/basvanbeek/pubsub"
)

//Service describes user management serviice
type Service interface {
	CreateUser(ctx context.Context, email, password string) error
}

//NewService returns an instance of user management service
//Note: don't forget to close asyncProd
func NewService(
	pubbers Publishers,
) Service {
	return service{
		pubbers: pubbers,
	}
}

type service struct {
	pubbers Publishers
}

//CreateUser invokes CreateUserPublisher to publish to UserCreate topic on kafka
func (s service) CreateUser(ctx context.Context, email, password string) error {
	//create the user
	msg := fmt.Sprintf("User Created %s: password:%s", email, password)

	//send a message to UserCreated topic
	return s.pubbers.CreateUserPublisher.PublishRaw("", []byte(msg))
}

//Publishers contains publishers/producers
type Publishers struct {
	CreateUserPublisher pubsub.Publisher
}

//Subscribers contains subscribers/consumers
type Subscribers struct {
	CreateUserSubscriber pubsub.Subscriber
}

//Start starts all subscribers/consumers
func (s Subscribers) Start() chan error {
	errc := make(chan error)

	//Listen to UserCreated topic
	go func(subber pubsub.Subscriber) {
		sm := subber.Start()
		for m := range sm {
			fmt.Println(string(m.Message()))
			if err := m.Done(); err != nil {
				errc <- err
				return
			}
		}

		errc <- subber.Stop()
	}(s.CreateUserSubscriber)

	return errc
}

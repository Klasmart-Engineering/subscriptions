package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
	"subscriptions.demo/models"
)

type Config struct {
	BrokerAddrs []string
	Context     context.Context
	Logger      log.Logger
}

func StartConsumers(ctx context.Context) {
	go ConsumeCreateSubscription(ctx)
	go ConsumeSubscriptionUpdated(ctx)
	go ConsumeDeleteSubscription(ctx)
	go ConsumeAccountAction(ctx)
}

func ConsumeDeleteSubscription(ctx context.Context) {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Kafka.Brokers,
		Topic:   "subscriptions.cmd.DeleteSubscription",
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			//TODO sort how to make this transactional? Need to rollback if anything fails.
			msg, err := kafkaReader.ReadMessage(ctx)
			if err != nil {
				panic("could not read message " + err.Error())
			}
			subscriptionId := string(msg.Value) //TODO sort schema properly

			valid, err := dbInstance.IsValidSubscriptionId(subscriptionId)

			if !valid || err != nil {
				PublishSubscriptionDeleteError(err.Error(), ctx)
			}

			// Get a subscription evaluation (all products)
			evaluation, err := dbInstance.GetSubscriptionEvaluation(subscriptionId)
			if err != nil {
				PublishSubscriptionDeleteError(err.Error(), ctx)
			}

			err = dbInstance.UnsubscribeFromProducts(subscriptionId)

			if err != nil {
				PublishSubscriptionDeleteError(err.Error(), ctx)
			}

			// Create a summary
			EvaluateSubscription(evaluation)

			subscriptionUUID, err := uuid.Parse(subscriptionId)
			if err != nil {
				PublishSubscriptionDeleteError(err.Error(), ctx)
			}

			publishSubscriptionDeleted(ctx, subscriptionUUID)
		}
	}
}

func ConsumeAccountAction(ctx context.Context) {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Kafka.Brokers,
		Topic:   "subscriptions.cmd.AccountAction", //TODO correct name?
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := kafkaReader.ReadMessage(ctx)
			if err != nil {
				panic("could not read message " + err.Error())
			}
			accountActionString := string(msg.Value)
			fmt.Println("received: ", accountActionString)

			var accountAction models.SubscriptionAccountAction
			json.Unmarshal([]byte(accountActionString), &accountAction)

			action := LogAction(accountAction)
			publishSubscriptionAccountLogged(ctx, action)
		}
	}
}

func ConsumeSubscriptionUpdated(ctx context.Context) {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Kafka.Brokers,
		Topic:   "subscriptions.cmd.UpdateSubscription",
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := kafkaReader.ReadMessage(ctx)
			if err != nil {
				panic("could not read message " + err.Error())
			}
			fmt.Println("received: ", string(msg.Value))

			/* TODO

			- Validation/Checks
			- Update the subscription (validate and check which type of update it is as we're merging different functions here)
			- Publish to subscriptions.cmd.SubscriptionUpdated

			*/

		}
	}
}

func ConsumeCreateSubscription(ctx context.Context) {

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Kafka.Brokers,
		Topic:   "subscriptions.cmd.CreateSubscription",
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := kafkaReader.ReadMessage(ctx)
			if err != nil {
				panic("could not read message " + err.Error())
			}
			fmt.Println("received: ", string(msg.Value)) //TODO what metadata is passed here? Tracking ID etc. should we persist/validate

			/* TODO
			- Validation/Checks (Subscription already created - if we get any info?)

			*/

			subscriptionId, err := dbInstance.CreateSubscription()
			if err != nil {
				panic("Unable to create subscription")
			}

			publishSubscriptionCreated(ctx, subscriptionId)
		}
	}
}

func publishSubscriptionCreated(ctx context.Context, uuid uuid.UUID) {
	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(config.Kafka.Brokers...),
		Topic: "subscriptions.cmd.SubscriptionCreated",
		//Logger: &config.Logger,
	}

	// Put the row on the topic
	var err = kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("SubscriptionId"),
			Value: []byte(uuid.String()),
		},
	)
	if err != nil {
		panic("could not write message " + err.Error())
	}
}

func publishSubscriptionDeleted(ctx context.Context, uuid uuid.UUID) {
	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(config.Kafka.Brokers...),
		Topic: "subscriptions.cmd.SubscriptionDeleted",
		//Logger: &config.Logger,
	}

	// Put the row on the topic
	var err = kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("SubscriptionId"),
			Value: []byte(uuid.String()),
		},
	)
	if err != nil {
		panic("could not write message " + err.Error())
	}
}

func publishSubscriptionAccountLogged(ctx context.Context, action models.LogResponse) {
	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(config.Kafka.Brokers...),
		Topic: "subscriptions.cmd.SubscriptionAccountLogged",
		//Logger: &config.Logger,
	}

	// Put the row on the topic
	marshal, err := json.Marshal(action)
	if err != nil {
		panic("could not marshall LogResponse " + err.Error())
	}

	err = kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("AccountLogged"), //TODO sort
			Value: marshal,
		},
	)
	if err != nil {
		panic("could not write message " + err.Error())
	}
}

func PublishSubscriptionUsage(subscriptionEvaluation models.EvaluatedSubscription, ctx context.Context) {

	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(config.Kafka.Brokers...),
		Topic: "subscriptions.cdc.SubscriptionUsage",
		//Logger: &config.Logger,
	}

	marshal, err := json.Marshal(subscriptionEvaluation)
	if err != nil {
		panic("could not marshall LogResponse " + err.Error())
	}

	err = kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("SubscriptionEvaluation"), //TODO sort
			Value: marshal,
		},
	)
	if err != nil {
		panic("could not write message " + err.Error())
	}
}

func PublishSubscriptionDeleteError(errorMessage string, ctx context.Context) {

	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(config.Kafka.Brokers...),
		Topic: "subscriptions.fct.SubscriptionDeleteError",
		//Logger: &config.Logger,
	}

	error := kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("SubscriptionDeleteError"), //TODO sort
			Value: []byte(errorMessage),
		},
	)
	if error != nil {
		panic("could not write message " + error.Error())
	}
}

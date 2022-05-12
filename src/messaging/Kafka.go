package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
	"subscriptions.demo/config"
	db "subscriptions.demo/database"
	"subscriptions.demo/handler"
	"subscriptions.demo/models"
)

type Config struct {
	BrokerAddrs []string
	Context     context.Context
	Logger      log.Logger
}

var dbInstance db.Database

func StartConsumers(ctx context.Context, config *config.Config, db db.Database) {
	dbInstance = db
	go ConsumeCreateSubscription(ctx, config)
	go ConsumeSubscriptionUpdated(ctx, config)
	go ConsumeDeleteSubscription(ctx, config)
	go ConsumeAccountAction(ctx, config)
}

func ConsumeDeleteSubscription(ctx context.Context, config *config.Config) {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Kafka.Brokers,
		Topic:   "subscriptions.cmd.DeleteSubscription",
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
			- Unsubscribe from all the product
			- Create a subscription summary based on last time and now (need to 'pay' for what they've used)
			- Cancel/Remove that subscription (update its state)
			- Publish to subscriptions.cmd.SubscriptionDeleted

			*/

		}
	}
}

func ConsumeAccountAction(ctx context.Context, config *config.Config) {
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

			action := handler.LogAction(accountAction)
			publishSubscriptionAccountLogged(ctx, config, action)
		}
	}
}

func ConsumeSubscriptionUpdated(ctx context.Context, config *config.Config) {
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

func ConsumeCreateSubscription(ctx context.Context, config *config.Config) {

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

			publishSubscriptionCreated(ctx, config, subscriptionId)
		}
	}
}

func publishSubscriptionCreated(ctx context.Context, config *config.Config, uuid uuid.UUID) {
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

func publishSubscriptionAccountLogged(ctx context.Context, config *config.Config, action models.LogResponse) {
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

func PublishSubscriptionUsage(subscriptionEvaluation models.EvaluatedSubscription, config config.Kafka, ctx context.Context) {
	kafkaWriter := kafka.Writer{
		Addr:  kafka.TCP(config.Brokers...),
		Topic: "subscriptions.cdc.SubscriptionUsage",
		//Logger: &config.Logger,
	}

	// Put the row on the topic
	marshal, err := json.Marshal(subscriptionEvaluation)
	if err != nil {
		panic("could not marshall LogResponse " + err.Error())
	}

	err = kafkaWriter.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte("SubscrtiptionEvaluation"), //TODO sort
			Value: marshal,
		},
	)
	if err != nil {
		panic("could not write message " + err.Error())
	}
}

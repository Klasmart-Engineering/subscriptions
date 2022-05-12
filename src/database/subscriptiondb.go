package db

import (
	"database/sql"
	"fmt"
	uuid2 "github.com/google/uuid"
	"log"
	"subscriptions.demo/models"
	"time"
)

func (db Database) Healthcheck() (bool, error) {
	var up int
	if err := db.Conn.QueryRow(`
			SELECT 1 AS up 
			FROM subscription_account`).Scan(&up); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("unable to get connection to the database: %s", err)
		}
	}

	return up == 1, nil
}

func (db Database) IsSubscriptionActive(subscriptionId string) (bool, error) {
	var state string
	if err := db.Conn.QueryRow(`
			SELECT ss.name 
			FROM subscription_account sa 
			JOIN subscription_state ss
			  ON sa.state = ss.id
			WHERE sa.id = $1`, subscriptionId).Scan(&state); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("unable to check if subscription is active: %s", err)
		}
	}

	return state == "Active", nil
}

func (db Database) UpdateLastProcessed(subscription *models.SubscriptionEvaluation) {

	sqlStatement := `
						UPDATE subscription_account
						 SET last_processed = NOW()
						WHERE id = $1;`

	_, err := db.Conn.Exec(sqlStatement, subscription.ID)
	if err != nil {
		log.Panic(fmt.Printf("Unable to update the last processed time of subscription id %d\n", subscription.ID))
	}

}

func (db Database) CreateSubscription() (uuid uuid2.UUID, err error) {
	var minutes = 43200
	var state = 1 // Active

	var subscriptionId uuid2.UUID
	sqlStatement := `
						INSERT INTO subscription_account (run_frequency_minutes, state)
						VALUES($1, $2) RETURNING id;`

	err = db.Conn.QueryRow(sqlStatement, minutes, state).Scan(&subscriptionId)
	if err != nil {
		log.Panic("Unable to create subscription", err)
	}

	return subscriptionId, err
}

func (db Database) UpdateSubscriptionStatus(subscriptionId int, active int) {

	sqlStatement := `
UPDATE subscription_account
 SET state = $1
WHERE id = $1;`

	db.Conn.Exec(sqlStatement, &active, &subscriptionId)

}

func (db Database) UsageOfSubscription(subscriptionEvaluation models.SubscriptionEvaluation) (int, error) {

	var subscriptionUsage int

	var countInteractionsSql = `
			SELECT COUNT(1) AS subscription_usage 
			FROM subscription_account_log 
			WHERE subscription_id = $1 AND product = $2 and interaction_at > $3`

	if err := db.Conn.QueryRow(countInteractionsSql,
		subscriptionEvaluation.ID, subscriptionEvaluation.Product, subscriptionEvaluation.LastProcessedTime).Scan(&subscriptionUsage); err != nil {
		if err == sql.ErrNoRows {
			return subscriptionUsage, fmt.Errorf("unknown usage on subscription: %s", subscriptionEvaluation.ID)
		}
	}
	return subscriptionUsage, nil
}

func (db Database) SubscriptionsToProcess() (*models.SubscriptionEvaluations, error) {

	list := &models.SubscriptionEvaluations{}
	rows, err := db.Conn.Query(`
		SELECT subAccount.id, subProduct.product, subProduct.threshold, subProduct.product, subAccount.last_processed
		FROM subscription_account subAccount
		JOIN subscription_account_product subProduct
		  ON subAccount.id = subProduct.subscription_id
		WHERE subAccount.last_processed IS NULL OR (now() < last_processed + ((SELECT run_frequency_minutes from subscription_account)||' minutes')::interval)`)
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var subscriptionEvaluation models.SubscriptionEvaluation
		var lastProcessed sql.NullString
		err := rows.Scan(&subscriptionEvaluation.ID, &subscriptionEvaluation.Product, &subscriptionEvaluation.Threshold, &subscriptionEvaluation.Name, &lastProcessed)

		if lastProcessed.Valid {
			subscriptionEvaluation.LastProcessedTime = lastProcessed.String
		} else {
			subscriptionEvaluation.LastProcessedTime = ""
		}

		if err != nil {
			return list, err
		}

		list.SubscriptionEvaluations = append(list.SubscriptionEvaluations, subscriptionEvaluation)
	}
	return list, nil
}

func (db Database) GetSubscriptionTypes() (*models.SubscriptionTypeList, error) {
	list := &models.SubscriptionTypeList{}
	sqlQuery := "SELECT id, name FROM subscription_type ORDER BY id DESC"
	rows, err := db.Conn.Query(sqlQuery)
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var subscription models.SubscriptionType
		err := rows.Scan(&subscription.ID, &subscription.Name)
		if err != nil {
			return list, err
		}
		list.Subscriptions = append(list.Subscriptions, subscription)
	}
	return list, nil
}

func (db Database) GetAllSubscriptionActions() (*models.SubscriptionActionList, error) {
	list := &models.SubscriptionActionList{}
	rows, err := db.Conn.Query("SELECT name, description, unit FROM subscription_action")
	if err != nil {
		return list, err
	}
	for rows.Next() {
		var action models.SubscriptionAction
		err := rows.Scan(&action.Name, &action.Description, &action.Unit)
		if err != nil {
			return list, err
		}
		list.Actions = append(list.Actions, action)
	}
	return list, nil
}

func (db Database) LogUserAction(accountAction models.SubscriptionAccountAction) {

	stmt, es := db.Conn.Prepare(`
			INSERT INTO subscription_account_log (subscription_id, action_type, usage, product_name, interaction_at)
			VALUES ($1, $2, $3, $4, $5, NOW())`)
	if es != nil {
		panic(es.Error())
	}

	_, er := stmt.Exec(accountAction.SubscriptionId, accountAction.ActionType, accountAction.UsageAmount, accountAction.Product)
	if er != nil {
		panic(er.Error())
	}

}

func (db Database) CountInteractionsForSubscription(userAction models.SubscriptionAccountAction) (int, error) {

	var lastProcessedTime time.Time
	if err := db.Conn.QueryRow(`
			SELECT last_processed
			FROM subscription_account
			WHERE id = $1 `,
		userAction.SubscriptionId).Scan(&lastProcessedTime); err != nil {
		if err == sql.ErrNoRows {
			panic(err)
		}
	}

	var countInteractionsSql = `
			SELECT COUNT(1) AS user_interactions 
			FROM subscription_account_log 
			WHERE subscription_id = $1 AND product = $2 `
	var countUserInteractions int
	if !lastProcessedTime.IsZero() {
		countInteractionsSql = countInteractionsSql + "AND interaction_at > $3"
		if err := db.Conn.QueryRow(countInteractionsSql,
			userAction.SubscriptionId, userAction.Product, lastProcessedTime).Scan(&countUserInteractions); err != nil {
			if err == sql.ErrNoRows {
				return countUserInteractions, fmt.Errorf("unknown count on user: %s", userAction.SubscriptionId)
			}
		}
	} else {
		if err := db.Conn.QueryRow(countInteractionsSql,
			userAction.SubscriptionId, userAction.Product).Scan(&countUserInteractions); err != nil {
			if err == sql.ErrNoRows {
				return countUserInteractions, fmt.Errorf("unknown count on user: %s", userAction.SubscriptionId)
			}
		}
	}
	return countUserInteractions, nil
}

func (db Database) GetThresholdForSubscriptionProduct(userAction models.SubscriptionAccountAction) (int, error) {

	var subscriptionThreshold int
	if err := db.Conn.QueryRow(`
			SELECT sap.threshold 
			FROM subscription_account_product sap 
			JOIN subscription_account sa
			  ON sap.subscription_id = sa.id
			WHERE sa.id = $1 AND sap.product = $2 `,
		userAction.SubscriptionId, userAction.Product).Scan(&subscriptionThreshold); err != nil {
		if err == sql.ErrNoRows {
			return subscriptionThreshold, fmt.Errorf("unknown threshold on user: %s", userAction.SubscriptionId)
		}
	}
	return subscriptionThreshold, nil
}

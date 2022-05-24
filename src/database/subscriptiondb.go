package db

import (
	"database/sql"
	"fmt"
	uuid2 "github.com/google/uuid"
	"go.uber.org/zap"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
	"time"
)

const (
	UserLimit = "UserLogin"
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

func (db Database) GetSubscriptionEvaluation(subscriptionId string) (models.SubscriptionEvaluation, error) {

	subscriptionEvaluation := models.SubscriptionEvaluation{}
	rows, err := db.Conn.Query(`
		SELECT subAccount.id, subProduct.threshold, subProduct.product, subAccount.last_processed
		FROM subscription_account subAccount
		JOIN subscription_account_product subProduct
		  ON subAccount.id = subProduct.subscription_id
		WHERE subAccount.last_processed IS NULL OR (now() < last_processed + ((SELECT run_frequency_minutes from subscription_account)||' minutes')::interval)
		AND subAccount.id = $1`, subscriptionId)
	if err != nil {
		return subscriptionEvaluation, err
	}

	for rows.Next() {

		var subId string
		var threshold int
		var name string
		var lastProcessed sql.NullString

		err := rows.Scan(&subId, &name, &threshold, &lastProcessed)

		if err != nil {
			return subscriptionEvaluation, err
		}

		if len(subscriptionEvaluation.Products) == 0 {
			subscriptionEvaluation.ID = subId
			products := append(subscriptionEvaluation.Products, models.SubscriptionEvaluationProduct{Threshold: threshold, Name: name})
			subscriptionEvaluation.Products = products

			if lastProcessed.Valid {
				subscriptionEvaluation.LastProcessedTime = lastProcessed.String
			} else {
				subscriptionEvaluation.LastProcessedTime = ""
			}
		} else {
			products := append(subscriptionEvaluation.Products, models.SubscriptionEvaluationProduct{Threshold: threshold, Name: name})
			subscriptionEvaluation.Products = products
		}
	}
	return subscriptionEvaluation, nil
}

func (db Database) IsValidSubscriptionId(subscriptionId string) (bool, error) {
	var valid int
	if err := db.Conn.QueryRow(`
			SELECT 1 AS up 
			FROM subscription_account
			WHERE id = $1`, subscriptionId).Scan(&valid); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("no rows returned.Unable to check if subscription is valid %s", err)
		}
	}

	return valid == 1, nil
}

func (db Database) UnsubscribeFromProducts(subscriptionId string) error {
	sqlStatement := `
			DELETE FROM subscription_account_product
			WHERE subscription_id = $1`
	_, err := db.Conn.Exec(sqlStatement, subscriptionId)
	if err != nil {
		return fmt.Errorf("unable to unsubscribe products from subscription", subscriptionId)
	}

	return nil
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

func (db Database) UpdateLastProcessed(monitoringContext *monitoring.Context, subscription *models.SubscriptionEvaluation) {

	sqlStatement := `
						UPDATE subscription_account
						 SET last_processed = NOW()
						WHERE id = $1;`

	_, err := db.Conn.Exec(sqlStatement, subscription.ID)
	if err != nil {
		monitoringContext.Error("Unable to update the last processed time of subscription id",
			zap.String("subscription", subscription.ID))
	}

}

func (db Database) SubscriptionExists(monitoringContext *monitoring.Context, accountId string) (subscriptionId uuid2.UUID, state int, err error) {

	var subId uuid2.UUID
	var subscriptionState int
	sqlStatement := `SELECT id, state FROM subscription_account
						WHERE account_id = $1;`

	err = db.Conn.QueryRow(sqlStatement, accountId).Scan(&subId, &subscriptionState)

	if err != nil {
		if err == sql.ErrNoRows {
			return subId, 0, fmt.Errorf("no rows returned. The subscription does not exist for account %s, %s", accountId, err)
		} else {
			monitoringContext.Panic("Unable to verify if subscription exists", zap.Error(err))
		}
	}

	return subId, subscriptionState, err
}

func (db Database) CreateSubscription(monitoringContext *monitoring.Context, accountId string) (uuid uuid2.UUID, err error) {
	var minutes = 43200 //30 days by default for now
	var state = 1       // Active by default

	var subscriptionId uuid2.UUID
	sqlStatement := `
						INSERT INTO subscription_account (account_id, run_frequency_minutes, state)
						VALUES($1, $2, $3) RETURNING id;`

	err = db.Conn.QueryRow(sqlStatement, accountId, minutes, state).Scan(&subscriptionId)
	if err != nil {
		monitoringContext.Panic("Unable to create subscription", zap.Error(err))
	}

	return subscriptionId, err
}

func (db Database) UpdateSubscriptionStatus(subscriptionId string, active int) error {

	sqlStatement := `
			UPDATE subscription_account
			 SET state = $1
			WHERE id = $2;`

	_, err := db.Conn.Exec(sqlStatement, &active, &subscriptionId)
	if err != nil {
		return err
	}

	return nil
}

func (db Database) UsageOfSubscription(subscriptionEvaluation models.SubscriptionEvaluation) (map[models.SubscriptionEvaluationProduct]int, error) {

	var countInteractionsSql = `
			SELECT COUNT(1) AS subscription_usage, sap.product As productName, sap.threshold, sap.type 
			FROM subscription_account_product sap
			LEFT JOIN subscription_account_log sal
			 ON sal.subscription_id = sap.subscription_id AND sal.product_name = sap.product
			WHERE sal.subscription_id = $1 AND sal.product_name = $2 AND sal.valid_usage = TRUE `

	var countInteractionWithTimestamp = " and sal.interaction_at > $3"
	var groupBySql = " GROUP BY sap.product, sap.threshold, sap.type"
	productToProductUsage := make(map[models.SubscriptionEvaluationProduct]int)

	subIdUUID, es := uuid2.Parse(subscriptionEvaluation.ID)

	if es != nil {
		panic(es.Error())
	}

	for _, product := range subscriptionEvaluation.Products {
		var productUsage int
		var productName string
		var productThreshold sql.NullInt16
		var productType string

		if subscriptionEvaluation.LastProcessedTime == "" {
			if err := db.Conn.QueryRow(countInteractionsSql+groupBySql,
				subIdUUID, product.Name).Scan(&productUsage, &productName, &productThreshold, &productType); err != nil {
				if err == sql.ErrNoRows {
					continue
				}
			}

		} else {
			if err := db.Conn.QueryRow(countInteractionsSql+countInteractionWithTimestamp+groupBySql,
				subIdUUID, product.Name, subscriptionEvaluation.LastProcessedTime).Scan(&productUsage, &productName, &productThreshold, &productType); err != nil {
				if err == sql.ErrNoRows {
					continue
				}
			}
		}

		var threshold int
		if productThreshold.Valid {
			threshold = int(productThreshold.Int16)
		} else {
			threshold = 0
		}

		productToProductUsage[models.SubscriptionEvaluationProduct{Name: productName, Threshold: threshold, Type: productType}] = productUsage
	}

	return productToProductUsage, nil
}

func (db Database) SubscriptionsToProcess() (*models.SubscriptionEvaluations, error) {

	evaluations := &models.SubscriptionEvaluations{}
	rows, err := db.Conn.Query(`
		SELECT subAccount.id, subProduct.threshold, subProduct.product, subProduct.type, subAccount.last_processed
		FROM subscription_account subAccount
		JOIN subscription_account_product subProduct
		  ON subAccount.id = subProduct.subscription_id
		WHERE subAccount.last_processed IS NULL OR (now() < last_processed + ((SELECT run_frequency_minutes from subscription_account)||' minutes')::interval) 
		ORDER BY subAccount.id DESC `)
	if err != nil {
		return evaluations, err
	}
	for rows.Next() {
		var subscriptionEvaluation models.SubscriptionEvaluation
		var subId string
		var productThreshold sql.NullInt16
		var name string
		var productType string
		var lastProcessed sql.NullString

		err := rows.Scan(&subId, &productThreshold, &name, &productType, &lastProcessed)

		if err != nil {
			return evaluations, err
		}

		if len(evaluations.SubscriptionEvaluations) == 0 {

			subscriptionEvaluation.ID = subId

			var threshold int
			if productThreshold.Valid {
				threshold = int(productThreshold.Int16)
			} else {
				threshold = 0
			}

			products := append(subscriptionEvaluation.Products, models.SubscriptionEvaluationProduct{Threshold: threshold, Name: name, Type: productType})
			subscriptionEvaluation.Products = products

			if lastProcessed.Valid {
				subscriptionEvaluation.LastProcessedTime = lastProcessed.String
			} else {
				subscriptionEvaluation.LastProcessedTime = ""
			}

			evaluations.SubscriptionEvaluations = append(evaluations.SubscriptionEvaluations, subscriptionEvaluation)
		} else {
			var found = false

			for i := 0; i < len(evaluations.SubscriptionEvaluations); i++ {
				if evaluations.SubscriptionEvaluations[i].ID == subId {

					var threshold int
					if productThreshold.Valid {
						threshold = int(productThreshold.Int16)
					} else {
						threshold = 0
					}

					products := append(evaluations.SubscriptionEvaluations[i].Products, models.SubscriptionEvaluationProduct{Threshold: threshold, Name: name, Type: productType})
					evaluations.SubscriptionEvaluations[i].Products = products
					found = true
				}
			}

			if found == false {
				subscriptionEvaluation.ID = subId

				var threshold int
				if productThreshold.Valid {
					threshold = int(productThreshold.Int16)
				} else {
					threshold = 0
				}

				products := append(subscriptionEvaluation.Products, models.SubscriptionEvaluationProduct{Threshold: threshold, Name: name, Type: productType})
				subscriptionEvaluation.Products = products

				if lastProcessed.Valid {
					subscriptionEvaluation.LastProcessedTime = lastProcessed.String
				} else {
					subscriptionEvaluation.LastProcessedTime = ""
				}
			}
		}

		if err != nil {
			return evaluations, err
		}

	}
	return evaluations, nil
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
			VALUES ($1, $2, $3, $4, to_timestamp($5))`)
	if es != nil {
		panic(es.Error())
	}
	subIdUUID, es := uuid2.Parse(accountAction.SubscriptionId)

	if es != nil {
		panic(es.Error())
	}
	_, er := stmt.Exec(subIdUUID, accountAction.ActionType, accountAction.UsageAmount, accountAction.Product, accountAction.InteractionTimeEpoch)
	if er != nil {
		panic(er.Error())
	}

}

func (db Database) UpdateChargeableLog(accountAction models.SubscriptionAccountAction) {

	stmt, es := db.Conn.Prepare(`
			UPDATE subscription_account_log SET valid_usage = FALSE
			WHERE subscription_id = $1 AND action_type = $2 AND usage = $3 AND product_name = $4 AND interaction_at = to_timestamp($5)`)
	if es != nil {
		panic(es.Error())
	}
	subIdUUID, es := uuid2.Parse(accountAction.SubscriptionId)

	if es != nil {
		panic(es.Error())
	}
	_, er := stmt.Exec(subIdUUID, accountAction.ActionType, accountAction.UsageAmount, accountAction.Product, accountAction.InteractionTimeEpoch)
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
			SELECT SUM(usage) AS user_interactions 
			FROM subscription_account_log 
			WHERE subscription_id = $1 AND product_name = $2 AND valid_usage = TRUE `

	var countUserInteractions int
	if !lastProcessedTime.IsZero() {

		interactionTimeSql := "AND interaction_at > $3"
		var query = countInteractionsSql + interactionTimeSql

		if err := db.Conn.QueryRow(query,
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

func (db Database) AddProductToSubscription(addProduct models.AddProduct) error {

	stmt, err := db.Conn.Prepare(`
			INSERT INTO subscription_account_product (subscription_id, product, type, threshold, action)
			VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return err
	}

	subIdUUID, es := uuid2.Parse(addProduct.SubscriptionId)

	if es != nil {
		return es
	}

	_, er := stmt.Exec(subIdUUID, addProduct.Product, addProduct.Type, addProduct.Threshold, addProduct.Action)
	if er != nil {
		return er
	}

	return nil
}

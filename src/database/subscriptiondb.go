package db

import (
	"database/sql"
	"fmt"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
)

func Healthcheck() (bool, error) {
	var up int
	if err := dbConnection.QueryRow(`SELECT 1 AS up`).Scan(&up); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("unable to get connection to the database: %s", err)
		}
	}

	return up == 1, nil
}

func GetSubscriptionByAccountId(monitoringContext *monitoring.Context, accountId string) (exists bool, subscription models.Subscription, err error) {
	err = dbConnection.GetContext(monitoringContext, &subscription, `
		SELECT * FROM subscription WHERE account_id = $1`, accountId)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, subscription, nil
		}

		return false, subscription, err
	}

	return true, subscription, nil
}

func GetSubscriptionById(monitoringContext *monitoring.Context, id string) (exists bool, subscription models.Subscription, err error) {
	err = dbConnection.GetContext(monitoringContext, &subscription, `
		SELECT * FROM subscription WHERE id = $1`, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, subscription, nil
		}

		return false, subscription, err
	}

	return true, subscription, nil
}

func CreateSubscription(monitoringContext *monitoring.Context, subscription models.Subscription) error {
	sqlStatement := `INSERT INTO subscription (id, account_id, state, created_at)
						VALUES($1, $2, 1, $3);`

	_, err := dbConnection.ExecContext(monitoringContext, sqlStatement, subscription.Id, subscription.AccountId, subscription.CreatedAt)

	return err
}

func UpdateSubscriptionStatus(monitoringContext *monitoring.Context, subscriptionId string, state models.SubscriptionState) error {

	sqlStatement := `
			UPDATE subscription 
			 SET state = $1
			WHERE id = $2;`

	_, err := dbConnection.ExecContext(monitoringContext, sqlStatement, &state, &subscriptionId)
	if err != nil {
		return err
	}

	return nil
}

func GetSubscriptionTypes(monitoringContext *monitoring.Context) (*models.SubscriptionTypeList, error) {
	list := &models.SubscriptionTypeList{}
	sqlQuery := "SELECT id, name FROM subscription_type ORDER BY id DESC"
	rows, err := dbConnection.QueryContext(monitoringContext, sqlQuery)
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

func GetAllSubscriptionActions(monitoringContext *monitoring.Context) (*models.SubscriptionActionList, error) {
	list := &models.SubscriptionActionList{}

	rows, err := dbConnection.QueryContext(monitoringContext, "SELECT name, description, unit FROM subscription_action")
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

func GetSubscriptionsPage(monitoringContext *monitoring.Context, pageSize int, offset int) ([]models.Subscription, error) {
	var result []models.Subscription

	err := dbConnection.SelectContext(monitoringContext, &result,
		fmt.Sprintf("SELECT * FROM subscription LIMIT %d OFFSET %d", pageSize, offset))

	return result, err
}

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

func GetSubscription(monitoringContext *monitoring.Context, accountId string) (exists bool, subscription models.Subscription, err error) {
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

func CreateSubscription(monitoringContext *monitoring.Context, subscription models.Subscription) error {
	sqlStatement := `INSERT INTO subscription (id, account_id, state)
						VALUES($1, $2, 1);`

	_, err := dbConnection.ExecContext(monitoringContext, sqlStatement, subscription.Id, subscription.AccountId)

	return err
}

func UpdateSubscriptionStatus(monitoringContext *monitoring.Context, subscriptionId string, active int) error {

	sqlStatement := `
			UPDATE subscription 
			 SET state = $1
			WHERE id = $2;`

	_, err := dbConnection.ExecContext(monitoringContext, sqlStatement, &active, &subscriptionId)
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

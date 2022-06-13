package db

import (
	"database/sql"
	"fmt"
	uuid2 "github.com/google/uuid"
	"go.uber.org/zap"
	"subscriptions/src/models"
	"subscriptions/src/monitoring"
)

func (db Database) Healthcheck() (bool, error) {
	var up int
	if err := db.Conn.QueryRow(`SELECT 1 AS up`).Scan(&up); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("unable to get connection to the database: %s", err)
		}
	}

	return up == 1, nil
}

func (db Database) IsValidSubscriptionId(monitoringContext *monitoring.Context, subscriptionId string) (bool, error) {
	var valid int
	if err := db.Conn.QueryRowContext(monitoringContext, `
			SELECT 1 AS up 
			FROM subscription_account
			WHERE id = $1`, subscriptionId).Scan(&valid); err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("no rows returned.Unable to check if subscription is valid %s", err)
		}
	}

	return valid == 1, nil
}

func (db Database) IsSubscriptionActive(monitoringContext *monitoring.Context, subscriptionId string) (bool, error) {
	var state string
	if err := db.Conn.QueryRowContext(monitoringContext, `
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

func (db Database) SubscriptionExists(monitoringContext *monitoring.Context, accountId string) (subscriptionId uuid2.UUID, state int, err error) {

	var subId uuid2.UUID
	var subscriptionState int
	sqlStatement := `SELECT id, state FROM subscription_account
						WHERE account_id = $1;`

	err = db.Conn.QueryRowContext(monitoringContext, sqlStatement, accountId).Scan(&subId, &subscriptionState)

	if err != nil {
		if err == sql.ErrNoRows {
			return subId, 0, fmt.Errorf("no rows returned. The subscription does not exist for account %s, %s", accountId, err)
		} else {
			monitoringContext.Panic("Unable to verify if subscription exists", zap.Error(err))
		}
	}

	return subId, subscriptionState, nil
}

func (db Database) CreateSubscription(monitoringContext *monitoring.Context, accountId string) (uuid uuid2.UUID, err error) {
	var minutes = 43200 //30 days by default for now
	var state = 1       // Active by default

	var subscriptionId uuid2.UUID
	sqlStatement := `INSERT INTO subscription_account (account_id, run_frequency_minutes, state)
						VALUES($1, $2, $3) RETURNING id;`

	err = db.Conn.QueryRowContext(monitoringContext, sqlStatement, accountId, minutes, state).Scan(&subscriptionId)
	if err != nil {
		monitoringContext.Panic("Unable to create subscription", zap.Error(err))
	}

	return subscriptionId, err
}

func (db Database) UpdateSubscriptionStatus(monitoringContext *monitoring.Context, subscriptionId string, active int) error {

	sqlStatement := `
			UPDATE subscription_account
			 SET state = $1
			WHERE id = $2;`

	_, err := db.Conn.ExecContext(monitoringContext, sqlStatement, &active, &subscriptionId)
	if err != nil {
		return err
	}

	return nil
}

func (db Database) GetSubscriptionTypes(monitoringContext *monitoring.Context) (*models.SubscriptionTypeList, error) {
	list := &models.SubscriptionTypeList{}
	sqlQuery := "SELECT id, name FROM subscription_type ORDER BY id DESC"
	rows, err := db.Conn.QueryContext(monitoringContext, sqlQuery)
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

func (db Database) GetAllSubscriptionActions(monitoringContext *monitoring.Context) (*models.SubscriptionActionList, error) {
	list := &models.SubscriptionActionList{}

	rows, err := db.Conn.QueryContext(monitoringContext, "SELECT name, description, unit FROM subscription_action")
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

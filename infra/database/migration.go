package database

import (
	"org-forms-config-management/models"
)

//Add list of model add for migrations
//var migrationModels = []interface{}{&ex_models.Example{}, &model.Example{}, &model.Address{})}
var migrationModels = []interface{}{&models.Organization{}, &models.User{}, &models.Project{}, &models.Environment{}, &models.Variant{}, &models.Access{}, &models.SSOConfig{}}


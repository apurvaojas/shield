package services

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}

type userContext struct {
	Username string
}

var singleInstance *userContext

func (userContextService *userContext) SetUsername(username string) {
	userContextService.Username = username
}

func GetUserContextInstance() *userContext {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			fmt.Println("Creating single instance now.")
			singleInstance = &userContext{}
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}

	return singleInstance
}

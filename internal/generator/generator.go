package generator

import (
	"time"
	"math/rand"
	"github.com/google/uuid"
	"github.com/florita1/ingestion-service/internal/model"
)

var actions = [] string {"login", "click", "purchase", "logout"}

func GenerateEvent() model.Event {
    return model.Event {
        Timestamp:  time.Now(),
        UserID:     getUserId(),
        Action:     actions[rand.Intn(len(actions))],
        Payload:    "example-payload",
    }
}

func getUserId() string {
    return "user-" + uuid.NewString()
}
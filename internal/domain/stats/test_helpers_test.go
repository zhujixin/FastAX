package stats

import (
	"fmt"
	"time"
)

var idCounter uint64

func uniqueID() string {
	idCounter++
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), idCounter)
}

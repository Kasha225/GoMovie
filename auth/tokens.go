package auth

import "time"

func AccessTTL() time.Duration {
	return accessTTL
}

func RefreshTTL() time.Duration {
	return refreshTTL
}

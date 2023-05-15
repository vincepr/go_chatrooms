/*
*	Basic One Time Password - OTP solution for Authentification
*	happens before Websocket connection is established
 */

package main

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// holds all active/valid tokens
type RetentionMap map[string] OTP

func NewRetentionMap(ctx context.Context, duration time.Duration) RetentionMap {
	rm := make(RetentionMap)
	go rm.Retention(ctx, duration)	// runs in background and removes expired OTPs
	return rm
}


// a single (valid) login token
type OTP struct {
	Key		string
	Created	time.Time
}

func (rm RetentionMap) NewOTP() OTP {
	otp := OTP{
		Key: 		uuid.NewString(),
		Created:	time.Now(),
	}
	rm[otp.Key] = otp
	return otp
}

func (rm RetentionMap) VertifyOTP(otp string) bool {
	if _, ok := rm[otp]; !ok {
		return false
	}
	delete(rm, otp)
	return true
}

// Runs as a Goroutine (is blocking so async save)
// makes sure old OTP tokens are removed when no longer valid
func (rm RetentionMap) Retention(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, otp := range rm{
				if otp.Created.Add(duration).Before(time.Now()) {
					delete(rm, otp.Key)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

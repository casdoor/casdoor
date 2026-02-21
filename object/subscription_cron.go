// Copyright 2025 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"fmt"
	"time"
)

func runSubscriptionRenewal() {
	subscriptions := []*Subscription{}
	err := ormer.Engine.Find(&subscriptions, &Subscription{State: SubStateExpired, IsAutoRenew: true})
	if err != nil {
		fmt.Printf("runSubscriptionRenewal() error: %s\n", err.Error())
		return
	}

	for _, sub := range subscriptions {
		// Skip if the user already has an active or upcoming subscription for the same plan
		hasActive, err := HasActiveSubscriptionForPlan(sub.Owner, sub.User, sub.Plan)
		if err != nil {
			fmt.Printf("runSubscriptionRenewal() HasActiveSubscriptionForPlan error: %s\n", err.Error())
			continue
		}
		if hasActive {
			continue
		}

		newSub, err := renewSubscription(sub)
		if err != nil {
			fmt.Printf("runSubscriptionRenewal() renewSubscription error: %s\n", err.Error())
			continue
		}

		affected, err := AddSubscription(newSub)
		if err != nil {
			fmt.Printf("runSubscriptionRenewal() AddSubscription error: %s\n", err.Error())
			continue
		}
		if !affected {
			fmt.Printf("runSubscriptionRenewal() failed to add subscription: %s\n", newSub.Name)
			continue
		}

		fmt.Printf("runSubscriptionRenewal() renewed subscription: %s -> %s for user: %s\n", sub.Name, newSub.Name, sub.User)
	}
}

func RunSubscriptionRenewalJob() {
	// Run once at startup
	runSubscriptionRenewal()

	// Schedule to run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		runSubscriptionRenewal()
	}
}

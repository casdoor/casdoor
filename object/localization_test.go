// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
	notify "github.com/casdoor/notify2"
)

func TestLocalizationMfaFailureBoundaries(t *testing.T) {
	// A configured push provider still cannot turn a sent notification into approval.
	push := &PushMfa{MfaProps: &MfaProps{Secret: "synthetic-receiver"}, provider: &Provider{}}
	for _, verify := range []func(string, string) error{push.Verify, push.SetupVerify} {
		err := verify("synthetic-code", "zh-TW")
		if !errors.Is(err, errPushMfaUnverifiable) {
			t.Fatalf("unsupported push outcome changed: %v", err)
		}
		if got := i18n.LocalizeError("zh-TW", err); !strings.Contains(got, "尚無法確認裝置上的核准結果") {
			t.Fatalf("push guidance was not localized: %q", got)
		}
		joined := i18n.JoinErrors(", ", err, i18n.Errorf("check:password or code is incorrect"))
		if !errors.Is(joined, errPushMfaUnverifiable) || !strings.Contains(i18n.LocalizeError("zh-TW", joined), "尚無法確認") {
			t.Fatal("composed MFA failure lost its cause or localization")
		}
	}
	totp := &TotpMfa{MfaProps: &MfaProps{}}
	if err := totp.Verify("synthetic-code", "zh-TW"); err == nil || i18n.LocalizeError("zh-TW", err) != "尚未啟用 TOTP" {
		t.Fatalf("missing TOTP secret changed: %v", err)
	}
	if err := MfaRecover(&User{}, "synthetic-recovery"); err == nil || i18n.LocalizeError("zh-TW", err) != "尚未設定復原碼" {
		t.Fatalf("missing recovery codes changed: %v", err)
	}
	if err := MfaRecover(&User{RecoveryCodes: []string{"synthetic-existing"}}, "synthetic-other"); err == nil || i18n.LocalizeError("zh-TW", err) != "找不到此復原碼" {
		t.Fatalf("rejected recovery code changed: %v", err)
	}
	radius := &RadiusMfa{MfaProps: &MfaProps{Secret: "synthetic-user"}}
	if err := radius.Verify("synthetic-code", "zh-TW"); err == nil || i18n.LocalizeError("zh-TW", err) != "尚未設定 RADIUS 服務供應商" {
		t.Fatalf("missing RADIUS provider changed: %v", err)
	}
}

func TestLocalizationMagicLinkDefaultsAndCustomContent(t *testing.T) {
	link := "https://example.invalid/synthetic-magic-link"
	for _, language := range []string{"zh-TW", "en"} {
		content := getMagicLinkEmailContent(&Provider{}, link, nil, language)
		if !strings.Contains(content, link) || !strings.Contains(content, fmt.Sprint(getMagicLinkTimeout())) || strings.Contains(content, "%link") || strings.Contains(content, "%expireTime") {
			t.Fatalf("default substitutions changed: %q", content)
		}
		if language == "zh-TW" && (!strings.Contains(content, "申請登入連結") || !strings.Contains(content, "瀏覽器中使用一次")) {
			t.Fatalf("Traditional guidance missing: %q", content)
		}
		if language == "en" && !strings.Contains(content, "You have requested a sign-in link") {
			t.Fatalf("English fallback changed: %q", content)
		}
		provider := &Provider{Content: `<p>Customer-owned <a href="%link">go</a> %expireTime %{user.friendlyName}</p>`, Title: "Customer-owned subject"}
		user := &User{DisplayName: "Synthetic Name"}
		custom := getMagicLinkEmailContent(provider, link, user, language)
		want := fmt.Sprintf(`<p>Customer-owned <a href="%s">go</a> %d Synthetic Name</p>`, link, getMagicLinkTimeout())
		if custom != want || provider.Title != "Customer-owned subject" {
			t.Fatalf("custom content was translated or rewritten: %q", custom)
		}
	}
}

func TestLocalizationNewPaymentMessageProvenance(t *testing.T) {
	owned := util.NewMessageError("payment:Payment failed")
	result := &pp.NotifyResult{PaymentStatus: pp.PaymentStateError, NotifyMessage: owned.Error(), MessageError: owned, OrderId: "synthetic-order", Price: 1.25, Currency: "USD"}
	if got := getPaymentNotificationMessage(result, "zh-TW"); got != "付款失敗" {
		t.Fatalf("owned outcome not localized: %q", got)
	}
	if got := getPaymentNotificationMessage(result, "en"); got != "Payment failed" {
		t.Fatalf("English outcome changed: %q", got)
	}
	if result.NotifyMessage != "Payment failed" || result.PaymentStatus != pp.PaymentStateError || result.Price != 1.25 || result.Currency != "USD" {
		t.Fatal("presentation changed provider outcome or amounts")
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "MessageError") {
		t.Fatal("presentation provenance leaked into wire data")
	}
	opaque := &pp.NotifyResult{NotifyMessage: "payment:Payment failed (customer-owned)"}
	if getPaymentNotificationMessage(opaque, "zh-TW") != opaque.NotifyMessage {
		t.Fatal("opaque provider or stored prose was translated")
	}
	opaque.MessageError = errors.New("Payment failed")
	if getPaymentNotificationMessage(opaque, "zh-TW") != opaque.NotifyMessage {
		t.Fatal("an arbitrary error was treated as provenance")
	}
}

func TestLocalizationKnownNotificationJoin(t *testing.T) {
	child := util.MessageErrorf("service:WeCom webhook returned HTTP status code: %d", 429)
	joined := errors.Join(notify.ErrSendNotification, child)
	value := notificationPresentationError(joined)
	if value.Error() != joined.Error() || !errors.Is(value, notify.ErrSendNotification) || !errors.Is(value, child) {
		t.Fatal("known notification wrapper lost diagnostics or identity")
	}
	if got := i18n.LocalizeError("zh-TW", value); got != "通知傳送失敗：WeCom Webhook 傳回 HTTP 狀態碼：429" {
		t.Fatalf("owned nested notification was not localized: %q", got)
	}
	opaque := errors.New("provider:opaque <message> 100%")
	value = notificationPresentationError(errors.Join(notify.ErrSendNotification, opaque))
	if got := i18n.LocalizeError("zh-TW", value); got != "通知傳送失敗："+opaque.Error() {
		t.Fatalf("provider diagnostic was interpreted: %q", got)
	}
	unknown := errors.Join(errors.New("different outer context"), child)
	if notificationPresentationError(unknown) != unknown || notificationPresentationError(nil) != nil {
		t.Fatal("unknown aggregate was reconstructed")
	}
}

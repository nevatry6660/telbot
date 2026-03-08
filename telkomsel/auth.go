package telkomsel

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"telkomsel-bot/config"
	"telkomsel-bot/model"
	"telkomsel-bot/util"
)

type OTPCallback func() (otp string, err error)

type Auth struct {
	mu sync.Mutex
}

func NewAuth() *Auth {
	return &Auth{}
}

func saveDebug(ctx context.Context, step string) {
	if !config.Verbose {
		return
	}
	dir := "debug"
	os.MkdirAll(dir, 0o755)

	var buf []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, err = page.CaptureScreenshot().
			WithFormat(page.CaptureScreenshotFormatPng).
			WithQuality(80).
			Do(ctx)
		return err
	})); err == nil && len(buf) > 0 {
		path := filepath.Join(dir, step+".png")
		os.WriteFile(path, buf, 0o644)
		log.Printf("[Debug] Screenshot saved: %s (%d bytes)", path, len(buf))
	} else if err != nil {
		log.Printf("[Debug] Screenshot failed for %s: %v", step, err)
	}

	var html string
	if err := chromedp.Run(ctx, chromedp.OuterHTML("html", &html)); err == nil {
		path := filepath.Join(dir, step+".html")
		os.WriteFile(path, []byte(html), 0o644)
		log.Printf("[Debug] HTML saved: %s (%d bytes)", path, len(html))
	} else {
		log.Printf("[Debug] HTML capture failed for %s: %v", step, err)
	}

	var url string
	if err := chromedp.Run(ctx, chromedp.Location(&url)); err == nil {
		log.Printf("[Debug] Current URL at %s: %s", step, url)
	}
}

func (a *Auth) Login(ctx context.Context, localPhone string, otpCallback OTPCallback) (*model.Session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	fullPhone := "62" + localPhone

	session := &model.Session{
		Phone:     localPhone,
		FullPhone: fullPhone,
		State:     model.StateLoggingIn,
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-zygote", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	var tokenMu sync.Mutex

	chromedp.ListenTarget(taskCtx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			reqURL := ev.Request.URL

			if strings.Contains(reqURL, "telkomsel.com") && config.Verbose {
				log.Printf("[Network] → %s %s", ev.Request.Method, reqURL)
			}

			if !strings.Contains(reqURL, "tdw.telkomsel.com") && !strings.Contains(reqURL, "api.telkomsel.com") {
				return
			}

			tokenMu.Lock()
			defer tokenMu.Unlock()

			for key, val := range ev.Request.Headers {
				keyLower := strings.ToLower(key)
				strVal := fmt.Sprintf("%v", val)

				if strings.Contains(keyLower, "accessauthorization") {
					token := strings.TrimPrefix(strVal, "Bearer ")
					token = strings.TrimSpace(token)
					if len(token) > 100 {
						session.AccessAuth = token
						log.Printf("✓ Captured accessauthorization: %s...", token[:50])
					}
				} else if strings.Contains(keyLower, "authorization") && !strings.Contains(keyLower, "access") {
					token := strings.TrimPrefix(strVal, "Bearer ")
					token = strings.TrimSpace(token)
					if len(token) > 100 {
						session.Authorization = token
						log.Printf("✓ Captured authorization: %s...", token[:50])
					}
				} else if strings.Contains(keyLower, "x-device") {
					session.XDevice = strVal
				} else if strings.Contains(keyLower, "hash") && len(strVal) == 56 {
					session.Hash = strVal
				} else if strings.Contains(keyLower, "mytelkomsel-web-app-version") {
					session.WebAppVersion = strVal
				}
			}
		}
	})

	log.Println("[Login] Navigating to my.telkomsel.com...")
	err := chromedp.Run(taskCtx,
		network.Enable(),
		chromedp.Navigate("https://my.telkomsel.com"),
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	saveDebug(taskCtx, "01_after_navigate")

	log.Println("[Login] Looking for phone input and filling...")
	phoneSelectors := []string{
		"input[type=\"tel\"]",
		"input[placeholder*=\"nomor\"]",
		"input[placeholder*=\"phone\"]",
		"input[placeholder*=\"Nomor\"]",
		"input[placeholder*=\"Phone\"]",
		"input[type=\"text\"]",
		"input[type=\"number\"]",
	}

	phoneFilled := false
	for _, sel := range phoneSelectors {
		var nodes []cdp.NodeID
		_ = chromedp.Run(taskCtx, chromedp.NodeIDs(sel, &nodes, chromedp.AtLeast(0)))
		log.Printf("[Login] Selector %q → %d nodes", sel, len(nodes))
		if len(nodes) > 0 && !phoneFilled {
			if err := chromedp.Run(taskCtx, chromedp.SendKeys(sel, localPhone)); err == nil {
				log.Printf("[Login] ✓ Filled phone using selector: %s", sel)
				phoneFilled = true
			} else {
				log.Printf("[Login] ✗ SendKeys failed for %s: %v", sel, err)
			}
		}
	}

	if !phoneFilled {
		saveDebug(taskCtx, "02_phone_not_found")

		log.Println("[Login] No input found with selectors, trying JS injection...")
		var jsResult string
		jsInject := fmt.Sprintf(`
			(function() {
				var inputs = document.querySelectorAll('input');
				var filled = false;
				for (var i = 0; i < inputs.length; i++) {
					var inp = inputs[i];
					var info = inp.type + '|' + inp.placeholder + '|' + inp.name + '|' + inp.id;
					console.log('Found input: ' + info);
					if (!filled) {
						var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
						nativeInputValueSetter.call(inp, '%s');
						inp.dispatchEvent(new Event('input', { bubbles: true }));
						inp.dispatchEvent(new Event('change', { bubbles: true }));
						filled = true;
					}
				}
				return 'inputs found: ' + inputs.length + ', filled: ' + filled;
			})()
		`, localPhone)
		if err := chromedp.Run(taskCtx, chromedp.Evaluate(jsInject, &jsResult)); err == nil {
			log.Printf("[Login] JS injection result: %s", jsResult)
		} else {
			log.Printf("[Login] JS injection failed: %v", err)
		}
	}

	saveDebug(taskCtx, "02_after_phone_fill")

	log.Println("[Login] Clicking submit button...")
	submitButtonScripts := []string{
		`Array.from(document.querySelectorAll('button')).find(el => el.textContent.match(/Lanjut|Masuk|Continue|Login/i))?.click()`,
		`document.querySelector('button[type="submit"]')?.click()`,
	}

	for _, script := range submitButtonScripts {
		var res interface{}
		_ = chromedp.Run(taskCtx, chromedp.EvaluateAsDevTools(script, &res))
		log.Printf("[Login] Submit script result: %v", res)
	}
	time.Sleep(3 * time.Second)

	saveDebug(taskCtx, "03_after_submit")

	log.Println("[Login] Waiting for OTP from user...")
	otp, err := otpCallback()
	if err != nil {
		return nil, fmt.Errorf("OTP callback: %w", err)
	}

	log.Println("[Login] Filling OTP via JS (6 individual digit inputs)...")

	otpJS := fmt.Sprintf(`
		(function() {
			var otp = '%s';

			var allInputs = Array.from(document.querySelectorAll('input'));

			var otpInputs = allInputs.filter(function(inp) {
				var rect = inp.getBoundingClientRect();
				var ml = inp.getAttribute('maxlength');
				return rect.width > 0 && rect.height > 0 && (ml === '1' || ml === null || rect.width < 80);
			});


			if (otpInputs.length < 6) {

				otpInputs = allInputs.filter(function(inp) {
					var rect = inp.getBoundingClientRect();
					return rect.width > 0 && rect.height > 0;
				});
			}


			if (otpInputs.length > 6) {
				otpInputs = otpInputs.slice(otpInputs.length - 6);
			}

			if (otpInputs.length < 6) {
				return 'error: found only ' + otpInputs.length + ' inputs, need 6';
			}

			var nativeSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
			for (var i = 0; i < 6; i++) {
				nativeSetter.call(otpInputs[i], otp[i]);
				otpInputs[i].dispatchEvent(new Event('input', { bubbles: true }));
				otpInputs[i].dispatchEvent(new Event('change', { bubbles: true }));
			}


			var submitBtn = Array.from(document.querySelectorAll('button')).find(function(el) {
				return el.textContent.match(/Submit|Verifikasi|Kirim|Lanjut/i);
			});
			if (submitBtn) {
				setTimeout(function() { submitBtn.click(); }, 500);
			}

			return 'ok: filled ' + otpInputs.length + ' inputs, submit=' + (submitBtn ? 'clicked' : 'not found');
		})()
	`, otp)

	var otpResult string
	otpFilled := false
	if err := chromedp.Run(taskCtx, chromedp.Evaluate(otpJS, &otpResult)); err != nil {
		log.Printf("[Login] ✗ OTP JS fill failed: %v", err)
	} else {
		log.Printf("[Login] OTP JS result: %s", otpResult)
		otpFilled = strings.HasPrefix(otpResult, "ok:")
	}

	time.Sleep(2 * time.Second)

	saveDebug(taskCtx, "04_after_otp")

	log.Println("[Login] Waiting for token capture (timeout 30s)...")
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		tokenMu.Lock()
		hasTokens := session.AccessAuth != "" && session.Authorization != ""
		tokenMu.Unlock()

		if hasTokens {
			break
		}
		time.Sleep(1 * time.Second)
	}

	saveDebug(taskCtx, "05_final")

	if session.XDevice == "" {
		session.XDevice = fmt.Sprintf("%s-%s-%s-%s-%s",
			util.RandomHex(4), util.RandomHex(2), util.RandomHex(2), util.RandomHex(2), util.RandomHex(6))
	}
	if session.Hash == "" {
		session.Hash = util.RandomHex(28)
	}

	if session.AccessAuth != "" && session.Authorization != "" {
		session.State = model.StateLoggedIn
		session.LastLoginAt = time.Now()
		log.Println("[Login] ✓ Login successful, tokens captured!")
		return session, nil
	}

	return nil, fmt.Errorf("failed to capture tokens during login (phoneFilled=%v, otpFilled=%v)", phoneFilled, otpFilled)
}

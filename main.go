package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/getlantern/systray"
	"github.com/getsentry/raven-go"
	"github.com/robfig/cron"
)

const (
	btc = "BTC"
	eth = "ETH"
	xrp = "XRP"
	ltc = "LTC"
)

var SentryDSN string

type state struct {
	Cron             *cron.Cron
	SelectedCurrency string
	CurrencyNames    map[string]string
	MenuItems        map[string]*systray.MenuItem
}

func main() {
	s := &state{
		SelectedCurrency: btc,
		CurrencyNames: map[string]string{
			btc: "bitcoin",
			eth: "ethereum",
			xrp: "ripple",
			ltc: "litecoin",
		},
		MenuItems: map[string]*systray.MenuItem{},
	}
	systray.Run(s.onReady, s.onExit)
}

func (s *state) onReady() {
	raven.SetDSN(SentryDSN)

	s.updatePrice()

	s.Cron = cron.New()
	s.Cron.AddFunc("@every 30s", s.updatePrice)
	s.Cron.Start()

	for currency := range s.CurrencyNames {
		s.MenuItems[currency] = systray.AddMenuItem(currency, "")
	}
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "")

	for {
		select {
		case <-s.MenuItems[btc].ClickedCh:
			s.SelectedCurrency = btc
			s.updatePrice()
		case <-s.MenuItems[eth].ClickedCh:
			s.SelectedCurrency = eth
			s.updatePrice()
		case <-s.MenuItems[xrp].ClickedCh:
			s.SelectedCurrency = xrp
			s.updatePrice()
		case <-s.MenuItems[ltc].ClickedCh:
			s.SelectedCurrency = ltc
			s.updatePrice()
		case <-quit.ClickedCh:
			systray.Quit()
		}
	}
}

func (s *state) onExit() {
	s.Cron.Stop()
}

func (s *state) updatePrice() {
	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://coinmarketcap.com/currencies/" + s.CurrencyNames[s.SelectedCurrency]
	res, err := client.Get(url)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		raven.CaptureErrorAndWait(errors.New("Status code is not OK"), nil)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		return
	}

	price := doc.Find(".details-panel-item--price__value").Text()
	systray.SetTitle(s.SelectedCurrency + " $" + price)
}

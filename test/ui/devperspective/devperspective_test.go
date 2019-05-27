package devperspective

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/redhat-developer/devconsole-operator/test/support"
	"github.com/redhat-developer/devconsole-operator/test/ui"
	"github.com/stretchr/testify/require"

	"github.com/tebeka/selenium"
)

var tag string

func TestDevPerspective(t *testing.T) {
	tag = support.Getenv(t, "TAG", fmt.Sprintf("%d", time.Now().Unix()))
	userIsAdmin := support.Getenv(t, "USER_IS_ADMIN", "true")
	chBin := support.Getenv(t, "CHROMEDRIVER_BINARY", "/usr/bin/chromedriver")
	chPort, err := strconv.Atoi(support.Getenv(t, "CHROMEDRIVER_PORT", "9515"))
	require.NoError(t, err, "Chromedriver port")

	devconsoleUsername := support.Getenv(t, "DEVCONSOLE_USERNAME", "consoledeveloper")
	devconsolePassword := support.Getenv(t, "DEVCONSOLE_PASSWORD", "developer")
	openshiftConsoleURL := support.Getenv(t, "OS_CONSOLE_URL", "http://localhost")

	wd, svc := ui.InitSelenium(
		t,
		chBin,
		chPort,
	)

	defer tearDown(t, wd, svc)

	defaultWait := 10 * time.Second

	t.Logf("Open URL: %s", openshiftConsoleURL)
	err = wd.Get(openshiftConsoleURL)
	require.NoErrorf(t, err, "Open URL: %s", openshiftConsoleURL)
	consoleIsUp := false
	for attempt := 0; attempt < 10; attempt++ {
		err = wd.Refresh()
		require.NoErrorf(t, err, "Refresh URL: %s", openshiftConsoleURL)
		el, _ := wd.FindElement(selenium.ByXPATH, "//*[contains(text(),'Application is not available')]")
		if el != nil {
			t.Logf("Openshift Console is not available, try again after 2s.")
			time.Sleep(2 * time.Second)
		} else {
			t.Logf("Openshift Console is up.")
			consoleIsUp = true
			break
		}
	}
	if !consoleIsUp {
		require.FailNow(t, "Openshift Console is not available.")
	}

	require.NoError(t, err, fmt.Sprintf("Open console starting URL: %s", openshiftConsoleURL))
	ui.WaitForURLToContain(t, wd, "oauth", defaultWait)

	var elem selenium.WebElement

	if userIsAdmin == "true" {
		elem = ui.FindElementBy(t, wd, selenium.ByLinkText, "kube:admin")
	} else {
		elem = ui.FindElementBy(t, wd, selenium.ByLinkText, devconsoleUsername)
	}

	ui.WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	ui.ClickToElement(t, elem)

	elem = ui.FindElementBy(t, wd, selenium.ByID, "inputUsername")
	ui.WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	ui.SendKeysToElement(t, elem, devconsoleUsername)

	elem = ui.FindElementBy(t, wd, selenium.ByID, "inputPassword")
	ui.WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	ui.SendKeysToElement(t, elem, devconsolePassword)

	elem = ui.FindElementBy(t, wd, selenium.ByXPATH, "//*/button[contains(text(),'Log In')]")
	ui.ClickToElement(t, elem)

	elem = ui.FindElementBy(t, wd, selenium.ByID, "nav-toggle")
	ui.WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	ui.ClickToElement(t, elem)

	elem = ui.FindElementBy(t, wd, selenium.ByXPATH, "//*/a[@target][contains(text(),'Developer')]")
	ui.WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
}

func tearDown(t *testing.T, wd selenium.WebDriver, svc *selenium.Service) {
	err := wd.Quit()
	if err != nil {
		t.Log(err)
	}
	err = svc.Stop()
	if err != nil {
		t.Log(err)
	}
}

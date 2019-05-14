package devperspective

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tebeka/selenium"
)

func TestDevPerspective(t *testing.T) {
	chBin := Getenv("CHROMEDRIVER_BINARY", "chromedriver")
	chPort, err := strconv.Atoi(Getenv("CHROMEDRIVER_PORT", "9515"))
	require.NoError(t, err, "Chromedriver port")
	wd, svc := InitSelenium(
		t,
		chBin,
		chPort,
	)

	defer svc.Stop()
	defer wd.Quit()

	defaultWait := 10 * time.Second

	startURL := Getenv("OS_CONSOLE_URL", "https://console-openshift-console.apps.pmacik.devcluster.openshift.com")
	err = wd.Get(startURL)
	require.NoError(t, err, fmt.Sprintf("Open console starting URL: %s", startURL))
	WaitForURLToContain(t, wd, "oauth", defaultWait)

	elem := FindElementBy(t, wd, selenium.ByLinkText, "consoledeveloper")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	elem.Click()

	elem = FindElementBy(t, wd, selenium.ByID, "inputUsername")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	SendKeysToElement(t, elem, Getenv("DEVCONSOLE_USERNAME", "consoledeveloper"))

	elem = FindElementBy(t, wd, selenium.ByID, "inputPassword")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	SendKeysToElement(t, elem, Getenv("DEVONSOLE_PASSWORD", "developer"))

	elem = FindElementBy(t, wd, selenium.ByXPATH, "//*/button[contains(text(),'Log In')]")
	elem.Click()

	elem = FindElementBy(t, wd, selenium.ByID, "nav-toggle")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	elem.Click()

	elem = FindElementBy(t, wd, selenium.ByXPATH, "//*/a[@target][contains(text(),'Developer')]")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
}

func FindElementBy(t *testing.T, wd selenium.WebDriver, by string, selector string) selenium.WebElement {
	elem, err := wd.FindElement(by, selector)
	require.NoError(t, err, fmt.Sprintf("Find element by %s=%s", by, selector))
	return elem
}

func WaitForElementToBeDisplayed(t *testing.T, wd selenium.WebDriver, element selenium.WebElement, duration time.Duration) {
	err := wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		return element.IsDisplayed()
	}, duration)
	require.NoError(t, err, fmt.Sprintf("Wait for element %s to be displayed", element))
}

func WaitForURLToContain(t *testing.T, wd selenium.WebDriver, text string, duration time.Duration) {
	err := wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		currentURL, err2 := wd.CurrentURL()
		return strings.Contains(currentURL, text), err2
	}, duration)
	currentURL, err2 := wd.CurrentURL()
	require.NoError(t, err2, fmt.Sprintf("Get current URL"))
	require.NoError(t, err, fmt.Sprintf("Wait for URL to contain '%s'. The current URL is '%s'.", text, currentURL))
}

func SendKeysToElement(t *testing.T, element selenium.WebElement, keys string) {
	err := element.SendKeys(keys)
	require.NoError(t, err, fmt.Sprintf("Send keys to element '%s'", element))
}

func InitSelenium(t *testing.T, chromedriverPath string, chromedriverPort int) (selenium.WebDriver, *selenium.Service) {

	service, err := selenium.NewChromeDriverService(chromedriverPath, chromedriverPort)
	require.NoError(t, err)

	chromeOptions := map[string]interface{}{
		"args": []string{
			"--no-cache",
			"--no-sandbox",
			"--headless",
			"--window-size=1920,1080",
			"--window-position=0,0",
			"--enable-features=NetworkService",
		},
	}

	caps := selenium.Capabilities{
		"browserName":   "chrome",
		"chromeOptions": chromeOptions,
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", chromedriverPort))
	require.NoError(t, err)
	return wd, service
}

// Getenv returns a value of environment variable, if it exists.
// Returns the default value otherwise.
func Getenv(key string, defaultValue string) string {
	value, found := os.LookupEnv(key)
	if found {
		return value
	}
	return defaultValue
}

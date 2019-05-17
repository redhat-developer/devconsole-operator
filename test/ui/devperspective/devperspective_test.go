package devperspective

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tebeka/selenium"
)

var tag string

func TestDevPerspective(t *testing.T) {
	tag = Getenv(t, "TAG", fmt.Sprintf("%d", time.Now().Unix()))
	userIsAdmin := Getenv(t, "USER_IS_ADMIN", "true")
	chBin := Getenv(t, "CHROMEDRIVER_BINARY", "/usr/bin/chromedriver")
	chPort, err := strconv.Atoi(Getenv(t, "CHROMEDRIVER_PORT", "9515"))
	require.NoError(t, err, "Chromedriver port")

	devconsoleUsername := Getenv(t, "DEVCONSOLE_USERNAME", "consoledeveloper")
	devconsolePassword := Getenv(t, "DEVCONSOLE_PASSWORD", "developer")
	openshiftConsoleURL := Getenv(t, "OS_CONSOLE_URL", "http://localhost")

	wd, svc := InitSelenium(
		t,
		chBin,
		chPort,
	)

	defer tearDown(t, wd, svc)

	defaultWait := 10 * time.Second

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
	WaitForURLToContain(t, wd, "oauth", defaultWait)

	var elem selenium.WebElement

	if userIsAdmin == "true" {
		elem = FindElementBy(t, wd, selenium.ByLinkText, "kube:admin")
	} else {
		elem = FindElementBy(t, wd, selenium.ByLinkText, devconsoleUsername)
	}

	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	ClickToElement(t, elem)

	elem = FindElementBy(t, wd, selenium.ByID, "inputUsername")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	SendKeysToElement(t, elem, devconsoleUsername)

	elem = FindElementBy(t, wd, selenium.ByID, "inputPassword")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	SendKeysToElement(t, elem, devconsolePassword)

	elem = FindElementBy(t, wd, selenium.ByXPATH, "//*/button[contains(text(),'Log In')]")
	ClickToElement(t, elem)

	elem = FindElementBy(t, wd, selenium.ByID, "nav-toggle")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
	ClickToElement(t, elem)

	elem = FindElementBy(t, wd, selenium.ByXPATH, "//*/a[@target][contains(text(),'Developer')]")
	WaitForElementToBeDisplayed(t, wd, elem, defaultWait)
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
	counter := 1
	err := wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		currentURL, err2 := wd.CurrentURL()
		//t.Logf("current url = %s", currentURL)
		//SaveScreenShotToPNG(t, wd, fmt.Sprintf("%s/%s/%d.png", tag, text, counter))
		counter++
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

func ClickToElement(t *testing.T, element selenium.WebElement) {
	err := element.Click()
	require.NoErrorf(t, err, "Click to element %s", element)
}

func InitSelenium(t *testing.T, chromedriverPath string, chromedriverPort int) (selenium.WebDriver, *selenium.Service) {

	service, err := selenium.NewChromeDriverService(chromedriverPath, chromedriverPort)
	require.NoError(t, err)

	chromeOptions := map[string]interface{}{
		"args": []string{
			"--verbose",
			"--no-cache",
			"--no-sandbox",
			"--headless",
			"--window-size=1920,1080",
			"--window-position=0,0",
			"--enable-features=NetworkService", // to ignore invalid HTTPS certificates
			//"--whitelisted-ips=''",             // to support running in a container
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
func Getenv(t *testing.T, key string, defaultValue string) string {
	value, found := os.LookupEnv(key)
	var retVal string
	if found {
		retVal = value
	} else {
		retVal = defaultValue
	}
	t.Logf("Using env variable: %s=%s", key, retVal)
	return retVal
}

func SaveScreenShotToPNG(t *testing.T, wd selenium.WebDriver, filename string) {
	err := os.MkdirAll(path.Dir(filename), 0775)
	require.NoError(t, err, "Create screenshot directory")
	// convert []byte to image for saving to file
	imgByte, err := wd.Screenshot()
	require.NoError(t, err, "Taking screenshot")
	img, _, _ := image.Decode(bytes.NewReader(imgByte))

	//save the imgByte to file
	out, err := os.Create(filename)
	defer out.Close()
	require.NoError(t, err, "Creating a file for the screenshot")

	err = png.Encode(out, img)
}

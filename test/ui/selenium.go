package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

//FindElementBy look for a web element by a given selector and returs it back when found.
func FindElementBy(t *testing.T, wd selenium.WebDriver, by string, selector string) selenium.WebElement {
	elem, err := wd.FindElement(by, selector)
	require.NoError(t, err, fmt.Sprintf("Find element by %s=%s", by, selector))
	return elem
}

//WaitForElementToBeDisplayed for a given web element to be displayed/visible for a given time duration.
func WaitForElementToBeDisplayed(t *testing.T, wd selenium.WebDriver, element selenium.WebElement, duration time.Duration) {
	err := wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		return element.IsDisplayed()
	}, duration)
	require.NoError(t, err, fmt.Sprintf("Wait for element %s to be displayed", element))
}

//WaitForURLToContain waits for a current URL to contain the given text. It waits for a given time duration.
func WaitForURLToContain(t *testing.T, wd selenium.WebDriver, text string, duration time.Duration) {
	counter := 1
	err := wd.WaitWithTimeout(func(wd selenium.WebDriver) (bool, error) {
		currentURL, err2 := wd.CurrentURL()
		counter++
		return strings.Contains(currentURL, text), err2
	}, duration)
	currentURL, err2 := wd.CurrentURL()
	require.NoError(t, err2, fmt.Sprintf("Get current URL"))
	require.NoError(t, err, fmt.Sprintf("Wait for URL to contain '%s'. The current URL is '%s'.", text, currentURL))
}

//SendKeysToElement sends keys to a given web element
func SendKeysToElement(t *testing.T, element selenium.WebElement, keys string) {
	err := element.SendKeys(keys)
	require.NoError(t, err, fmt.Sprintf("Send keys to element '%s'", element))
}

//ClickToElement performs a click on a given web element.
func ClickToElement(t *testing.T, element selenium.WebElement) {
	err := element.Click()
	require.NoErrorf(t, err, "Click to element %s", element)
}

//InitSelenium creates and initializes a new ChromeDriver service.
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

//SaveScreenShotToPNG saves current screen to a given PNG file.
func SaveScreenShotToPNG(t *testing.T, wd selenium.WebDriver, filename string) {
	err := os.MkdirAll(path.Dir(filename), 0775)
	require.NoError(t, err, "Create screenshot directory")
	// convert []byte to image for saving to file
	imgByte, err := wd.Screenshot()
	require.NoError(t, err, "Take screenshot")
	img, _, _ := image.Decode(bytes.NewReader(imgByte))

	//save the imgByte to file
	out, err := os.Create(filename)
	defer close(t, out)

	require.NoError(t, err, "Create a file for the screenshot")

	err = png.Encode(out, img)
	require.NoError(t, err, "Write screanshot to PNG file")
}

func close(t *testing.T, f *os.File) {
	err := f.Close()
	require.NoError(t, err, "Close output file")
}

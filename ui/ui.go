package ui

import "C"
import (
	"elektromagnetskiAkcelerator/dataProcessor"
	"elektromagnetskiAkcelerator/serialComms"
	"fmt"
	"github.com/gonutz/wui/v2"
	"github.com/labstack/gommon/log"
	"strconv"
	"strings"
	"time"
)

var serialHandler = serialComms.NewSerialHandler()
var errorLabel *wui.Label

func getCOMPortDropdown() *wui.ComboBox {
	c := wui.NewComboBox()
	c.SetWidth(150)
	c.SetHeight(30)
	c.SetSelectedIndex(0)
	c.SetX(6)
	c.SetY(6)

	for _, port := range serialComms.GetPorts() {
		c.AddItem(port)
	}

	return c
}

func getConnectButton(dropdown *wui.ComboBox, label *wui.Label) *wui.Button {
	b := wui.NewButton()
	b.SetText("Connect")
	b.SetBounds(170, 5, 150, 26)

	b.SetOnClick(func() {
		serialHandler.Close()

		label.SetText("...")

		portName := dropdown.Items()[dropdown.SelectedIndex()]
		log.Infof(portName)

		isPort, err := serialComms.IsInterfaceAtPort(portName)
		if err != nil {
			SetErrorText(err)
			log.Error(err)
			label.SetText("failed")
			return
		}

		if isPort {
			label.SetText("connected")
		} else {
			label.SetText("failed")
			return
		}

		err = serialHandler.InitialiseHandler(portName)
		if err != nil {
			SetErrorText(err)
			log.Error(err)
			label.SetText("failed")
			return
		}
		ClearErrorText()

	})

	return b
}

func getConnectText() *wui.Label {
	t := wui.NewLabel()
	t.SetBounds(340, 5, 150, 26)
	return t
}

func getBasesCheckbox(num int) *wui.CheckBox {
	b := wui.NewCheckBox()
	b.SetText(fmt.Sprintf("Accelerator %d", num))
	b.SetChecked(true)
	b.SetBounds(6, 45+45*num, 150, 30)
	b.SetOnChange(func(checked bool) {
		err := serialHandler.SetBase(num-1, !checked)
		if err != nil {
			SetErrorText(err)
			log.Error(err)
		}
		ClearErrorText()
	})

	return b
}

func getLogDurationInput() *wui.EditLine {
	e := wui.NewEditLine()
	e.SetBounds(350, 90, 150, 26)
	return e
}

func getLogDurationLabel() *wui.Label {
	l := wui.NewLabel()
	l.SetBounds(200, 90, 150, 26)
	l.SetText("Recording duration(s):")

	return l
}

func handleStartRecording(durationInput *wui.EditLine, progressBar *wui.ProgressBar) {
	duration, err := strconv.Atoi(durationInput.Text())
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	err = serialHandler.StartRecording(duration)
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	go handleRecord(progressBar, int64(duration))

	ClearErrorText()
}

func handleStopRecording() {
	err := serialHandler.StopRecording()
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	ClearErrorText()
}
func getLogStartButton(durationInput *wui.EditLine, progressBar *wui.ProgressBar) *wui.Button {
	b := wui.NewButton()
	b.SetBounds(510, 90, 150, 26)
	b.SetText("Start recording")
	b.SetOnClick(func() {
		if b.Text() == "Start recording" {
			handleStartRecording(durationInput, progressBar)
		} else {
			handleStopRecording()
		}
	})

	go handleStartStopButton(b)

	return b
}

func getProgressBar() *wui.ProgressBar {
	p := wui.NewProgressBar()
	p.SetBounds(200, 135, 460, 26)
	return p
}

func handleRecord(progressBar *wui.ProgressBar, duration int64) {
	startTime := time.Now().UnixMilli()

	for time.Now().UnixMilli()-startTime <= duration*1000 {
		if !serialHandler.IsRecording {
			progressBar.SetValue(0)
			return
		}

		progressBar.SetValue(float64(time.Now().UnixMilli()-startTime) / float64(duration*1000))
		time.Sleep(10 * time.Millisecond)
	}

	progressBar.SetValue(1)
	time.Sleep(1 * time.Second)
	progressBar.SetValue(0)

	buffer, err := serialHandler.GetWholeBuffer()
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	dataPoints, err := dataProcessor.GetDataPointsFromRawSerial(strings.ReplaceAll(string(buffer), "\000", ""))
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	err = dataProcessor.WriteToXSLSFle(dataPoints)
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	err = serialHandler.StopRecording()
	if err != nil {
		SetErrorText(err)
		log.Error(err)
		return
	}

	ClearErrorText()
}

func handleStartStopButton(button *wui.Button) {
	for {
		if serialHandler.IsRecording {
			button.SetText("Stop recording")
		} else {
			button.SetText("Start recording")
		}
	}
}

func getErrorText() *wui.Label {
	l := wui.NewLabel()
	l.SetBounds(200, 350, 460, 52)
	return l
}

func SetErrorText(err error) {
	errorLabel.SetText("Error: " + err.Error())
}

func ClearErrorText() {
	errorLabel.SetText("")
}

func StartUI() {
	window := wui.NewWindow()
	window.SetWidth(800)

	dropdown := getCOMPortDropdown()
	connectStatusLabel := getConnectText()
	progressBar := getProgressBar()
	durationInput := getLogDurationInput()
	errorLabel = getErrorText()

	window.Add(dropdown)
	window.Add(connectStatusLabel)
	window.Add(getConnectButton(dropdown, connectStatusLabel))
	window.Add(getLogDurationLabel())
	window.Add(durationInput)
	window.Add(progressBar)
	window.Add(getLogStartButton(durationInput, progressBar))
	window.Add(errorLabel)

	for i := 1; i <= 6; i++ {
		window.Add(getBasesCheckbox(i))
	}

	window.Show()

}

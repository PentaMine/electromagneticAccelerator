package dataProcessor

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DataPoint struct {
	timeMicros int64
	states     [6]int
}

func GetDataPointsFromRawSerial(serialExchange string) ([]DataPoint, error) {
	if strings.HasSuffix(serialExchange, "slf\n") {
		return nil, errors.New("received malformed serialExchange")
	}
	reportedPoints := strings.Split(serialExchange, "\n")

	var (
		startTime  int64
		dataPoints []DataPoint
		prevStates [6]int
	)

	for i, data := range reportedPoints[:len(reportedPoints)-2] {

		if data == "slf" {
			break
		}

		parameters := strings.Split(data[2:len(data)-1], ",")
		statesParm := parameters[0]
		timeMicros, err := strconv.ParseInt(parameters[1], 10, 64)

		if err != nil {
			return nil, err
		}

		if i == 0 {
			startTime = timeMicros
		}

		var states [6]int

		for i, state := range statesParm {
			if state == '0' {
				states[i] = 0
			} else {
				states[i] = 1
			}
		}

		if prevStates == states {
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			timeMicros: timeMicros - startTime,
			states:     states,
		})

		prevStates = states
	}

	sort.Slice(dataPoints[:], func(i, j int) bool {
		return dataPoints[i].timeMicros < dataPoints[j].timeMicros
	})

	return dataPoints, nil
}

func WriteToXlsxFle(dataPoints []DataPoint) error {
	f := excelize.NewFile()
	defer f.Close()

	index, err := f.NewSheet("results")
	if err != nil {
		return err
	}

	f.SetCellValue("results", "A1", "time / µs")
	f.SetCellValue("results", "B1", "accelerator 1")
	f.SetCellValue("results", "C1", "accelerator 2")
	f.SetCellValue("results", "D1", "accelerator 3")
	f.SetCellValue("results", "E1", "accelerator 4")
	f.SetCellValue("results", "F1", "accelerator 5")
	f.SetCellValue("results", "G1", "accelerator 6")

	for i, dataPoint := range dataPoints {
		f.SetCellValue("results", fmt.Sprintf("A%d", i+2), dataPoint.timeMicros)
		f.SetCellValue("results", fmt.Sprintf("B%d", i+2), dataPoint.states[0])
		f.SetCellValue("results", fmt.Sprintf("C%d", i+2), dataPoint.states[1])
		f.SetCellValue("results", fmt.Sprintf("D%d", i+2), dataPoint.states[2])
		f.SetCellValue("results", fmt.Sprintf("E%d", i+2), dataPoint.states[3])
		f.SetCellValue("results", fmt.Sprintf("F%d", i+2), dataPoint.states[4])
		f.SetCellValue("results", fmt.Sprintf("G%d", i+2), dataPoint.states[5])
	}

	f.SetActiveSheet(index)

	if err := f.SaveAs(fmt.Sprintf("results%d.xlsx", time.Now().Unix())); err != nil {
		return err
	}

	return nil
}

package serialComms

import (
	"errors"
	"fmt"
	"go.bug.st/serial"
	"log"
	"strings"
	"time"
)

type SerialHandler struct {
	port        serial.Port
	bases       []bool
	IsRecording bool
}

func (s *SerialHandler) InitialiseHandler(portName string) error {

	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return err
	}

	s.port = port

	s.StopRecording()

	for i := 0; i < 6; i++ {
		s.SetBase(0, false)
	}

	return nil
}

func (s *SerialHandler) GetWholeBuffer() ([]byte, error) {
	if s.port == nil {
		return nil, errors.New("not connected")
	}

	n := 1024
	var (
		err    error
		buffer []byte
	)

	for n == 1024 {
		tempBuffer := make([]byte, 1024)

		n, err = s.port.Read(tempBuffer)

		if err != nil {
			return nil, err
		}

		buffer = append(buffer, tempBuffer...)
	}

	return buffer, nil
}

func (s *SerialHandler) Close() error {
	if s.port == nil {
		return errors.New("not connected")
	}

	return s.port.Close()
}
func (s *SerialHandler) StartRecording(duration int) error {
	if s.port == nil {
		return errors.New("not connected")
	}

	_, err := s.port.Write([]byte(fmt.Sprintf("l%d\n", duration)))
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	buffer := make([]byte, 50)
	_, err = s.port.Read(buffer)
	if err != nil {
		return err
	}

	msg := strings.Split(string(buffer), "\r\n")[0]

	if msg != "sls" {
		return errors.New(msg)
	}

	s.IsRecording = true

	return nil
}

func (s *SerialHandler) StopRecording() error {
	if s.port == nil {
		return errors.New("not connected")
	}

	_, err := s.port.Write([]byte("l-1\n"))
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Millisecond)

	buffer := make([]byte, 50)
	_, err = s.port.Read(buffer)
	if err != nil {
		return err
	}

	msg := strings.Split(string(buffer), "\r\n")[0]

	if msg != "slf" {
		return errors.New(msg)
	}

	s.IsRecording = false

	return nil
}

func (s *SerialHandler) UpdateBases() error {
	msg := "b"

	for _, state := range s.bases {
		if state {
			msg += "1"
		} else {
			msg += "0"
		}
	}

	msg += "\n"

	_, err := s.port.Write([]byte(msg))

	if err != nil {
		return err
	}

	buffer := make([]byte, 50)

	time.Sleep(10 * time.Millisecond)

	_, err = s.port.Read(buffer)
	if err != nil {
		return err
	}

	msg = strings.Split(string(buffer), "\r\n")[0]

	if msg != "sb" {
		return errors.New(msg)
	}

	return nil
}

func (s *SerialHandler) SetBase(index int, state bool) error {
	s.bases[index] = state

	return s.UpdateBases()
}

func NewSerialHandler() *SerialHandler {
	sh := new(SerialHandler)
	sh.bases = []bool{false, false, false, false, false, false}
	return sh
}

func IsInterfaceAtPort(portName string) (bool, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return false, err
	}

	_, err = port.Write([]byte("pTest\n"))
	if err != nil {
		return false, nil
	}

	time.Sleep(20 * time.Millisecond)

	buffer := make([]byte, 10)

	_, err = port.Read(buffer)
	if err != nil {
		return false, err
	}

	err = port.Close()
	if err != nil {
		return false, err
	}

	msg := strings.Split(string(buffer), "\r\n")[0]

	if msg == "sPingTest" {
		return false, errors.New(msg)
	}

	return true, nil
}

func GetPorts() []string {
	ports, err := serial.GetPortsList()

	if err != nil {
		log.Fatal(err)
	}

	return ports
}

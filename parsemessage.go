package main

import (
	"errors"
	"strconv"
	"strings"
)

// Message is a parsed incoming message
type Message struct {
	accountName string
	class       string
	capacity    int
	inc         int
}

// parseMessage takes an incoming UDP message string and parses it looking for
// <accountName>,<class>,<capacity>,<inc>\n
// where accountName that uniquely identifies each client, class is l/w/q,
// capacity is the bucket capacity for that class/accountName and inc is
// the amount that is being asked to be removed from the bucket value.
func parseMessage(str string) (*Message, error) {
	// parse the incoming string - account,class,max_per_second,inc_by
	bits := strings.Split(str, ",")
	if len(bits) != 4 {
		return nil, errors.New("message string must contain 4 strings separated by commas")
	}

	// sanity checks
	accountName := bits[0]
	class := bits[1]
	capacityStr := bits[2]
	incrementStr := bits[3]
	if len(accountName) == 0 || len(class) == 0 || len(capacityStr) == 0 || len(incrementStr) == 0 {
		return nil, errors.New("missing account/class/capacity/inc strings")
	}
	if class != "l" && class != "w" && class != "q" {
		return nil, errors.New("class must be l/w/q")
	}
	capacity, err := strconv.Atoi(capacityStr)
	if err != nil {
		return nil, errors.New("cannot convert capacity from string to integer")
	}
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	inc, err := strconv.Atoi(incrementStr)
	if err != nil {
		return nil, errors.New("cannot convert increment from string to integer")
	}
	if inc <= 0 {
		return nil, errors.New("inc must be positive")
	}
	message := Message{
		accountName: accountName,
		class:       class,
		capacity:    capacity,
		inc:         inc,
	}
	return &message, nil

}

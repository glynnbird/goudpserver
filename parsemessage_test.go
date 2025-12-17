package main

import "testing"

func Test_parsemessage_nocommas(t *testing.T) {
	_, err := parseMessage("gibberish")
	if err == nil {
		t.Error("Expected error for supplying gibberish message, got nil")
	}
}

func Test_parsemessage_twocommas(t *testing.T) {
	_, err := parseMessage("gibb,er,ish")
	if err == nil {
		t.Error("Expected error for supplying insufficient components in message, got nil")
	}
}

func Test_parsemessage_threecommas_missing_data(t *testing.T) {
	var err error
	_, err = parseMessage(",l,10,1")
	if err == nil {
		t.Error("Expected error for missing name component in message, got nil")
	}
	_, err = parseMessage("gb,,10,1")
	if err == nil {
		t.Error("Expected error for missing class component in message, got nil")
	}
	_, err = parseMessage("gb,l,,1")
	if err == nil {
		t.Error("Expected error for missing capacity component in message, got nil")
	}
	_, err = parseMessage("gb,l,10,")
	if err == nil {
		t.Error("Expected error for missing inc component in message, got nil")
	}
}

func Test_parsemessage_invalid_class(t *testing.T) {
	var err error
	_, err = parseMessage("gb,x,10,1")
	if err == nil {
		t.Error("Expected error for invalid class component in message, got nil")
	}
}

func Test_parsemessage_invalid_capacity(t *testing.T) {
	var err error
	_, err = parseMessage("gb,w,ten,1")
	if err == nil {
		t.Error("Expected error for invalid capacity component in message, got nil")
	}
}

func Test_parsemessage_negative_capacity(t *testing.T) {
	var err error
	_, err = parseMessage("gb,w,-10,1")
	if err == nil {
		t.Error("Expected error for invalid capacity component in message, got nil")
	}
}

func Test_parsemessage_invalid_inc(t *testing.T) {
	var err error
	_, err = parseMessage("gb,w,10,one")
	if err == nil {
		t.Error("Expected error for invalid inc component in message, got nil")
	}
}

func Test_parsemessage_negative_inc(t *testing.T) {
	var err error
	_, err = parseMessage("gb,w,10,-1")
	if err == nil {
		t.Error("Expected error for negative inc component in message, got nil")
	}
}

func Test_parsemessage_success_lookup(t *testing.T) {
	message, err := parseMessage("gb,l,10,1")
	if err != nil {
		t.Errorf("Expected no error for valid message, got %v", err)
	}
	if message.accountName != "gb" {
		t.Errorf("Expected accountName to be %v, got %v", "gb", message.accountName)
	}
	if message.class != "l" {
		t.Errorf("Expected class to be %v, got %v", "l", message.class)
	}
	if message.capacity != 10 {
		t.Errorf("Expected capacity to be %v, got %v", 10, message.capacity)
	}
	if message.inc != 1 {
		t.Errorf("Expected inc to be %v, got %v", 1, message.inc)
	}
}

func Test_parsemessage_success_write(t *testing.T) {
	message, err := parseMessage("gb,w,10,1")
	if err != nil {
		t.Errorf("Expected no error for valid message, got %v", err)
	}
	if message.accountName != "gb" {
		t.Errorf("Expected accountName to be %v, got %v", "gb", message.accountName)
	}
	if message.class != "w" {
		t.Errorf("Expected class to be %v, got %v", "w", message.class)
	}
	if message.capacity != 10 {
		t.Errorf("Expected capacity to be %v, got %v", 10, message.capacity)
	}
	if message.inc != 1 {
		t.Errorf("Expected inc to be %v, got %v", 1, message.inc)
	}
}

func Test_parsemessage_success_query(t *testing.T) {
	message, err := parseMessage("gb,q,10,1")
	if err != nil {
		t.Errorf("Expected no error for valid message, got %v", err)
	}
	if message.accountName != "gb" {
		t.Errorf("Expected accountName to be %v, got %v", "gb", message.accountName)
	}
	if message.class != "q" {
		t.Errorf("Expected class to be %v, got %v", "q", message.class)
	}
	if message.capacity != 10 {
		t.Errorf("Expected capacity to be %v, got %v", 10, message.capacity)
	}
	if message.inc != 1 {
		t.Errorf("Expected inc to be %v, got %v", 1, message.inc)
	}
}

package text

import (
	"strconv"
	"time"
)

const (
	esc = "\033"
	csi = esc + "["

	noColor    = csi + "m"
	green      = csi + "32;40m"
	lightGreen = csi + "1;32;40m"
	yellow     = csi + "1;33;40m"
	brown      = csi + "33;40m"

	Clear = csi + "2J"

	moveUp = csi + "1A" //ESC[#A
)

type Message struct {
	Msg     string
	Sender  string
	MsgType string
}

func Brown(t string) string {
	return brown + t + noColor
}

func Yellow(t string) string {
	return yellow + t + noColor
}

func LightGreen(t string) string {
	return lightGreen + t + noColor
}

func Move(x int, y int) string {
	return csi + strconv.Itoa(x) + ";" + strconv.Itoa(y) + "f"
}

func FormatChatMsg(msg Message, sender string) string {
	var result string

	t := time.Now().Format("15:04:05")

	if msg.MsgType == "chat" {
		if msg.Sender == sender {
			sender = Yellow(msg.Sender)
		} else {
			sender = Brown(msg.Sender)
		}
		result = moveUp + t + " " + sender + ": " + msg.Msg
	} else {
		result = "* " + msg.Msg
	}

	return result
}

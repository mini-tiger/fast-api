package dLogger

import (
	"testing"
	"time"

	"github.com/mini-tiger/fast-api/core"
)

func TestWrite(t *testing.T) {
	core.Start()
	Write(LeverWaning, "dfdsf", "fdf")
	time.Sleep(time.Second)
}

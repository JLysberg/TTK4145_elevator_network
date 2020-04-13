package config

import (
	"time"
)

const (
	UpdateRate = 500 * time.Millisecond

	ClearTimeout = 1 * time.Second
	DoorTimeout  = 3 * time.Second

	MFloors = 4
	NElevs  = 2
)

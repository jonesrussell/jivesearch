package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestSetup(t *testing.T) {
	v := viper.New()
	setup(v)
}

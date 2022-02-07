package main

import (
	"github.com/alexflint/go-arg"
	mpeg_dash_tools "github.com/denysvitali/mpeg-dash-tools/pkg"
	"github.com/sirupsen/logrus"
	"net/url"
)

var args struct {
	Url string `arg:"positional,required"`
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	arg.MustParse(&args)

	_, err := url.Parse(args.Url)
	if err != nil {
		logger.Fatalf("invalid URL provided: %v", err)
	}

	dc := mpeg_dash_tools.NewDumpClient()
	dc.SetLogger(logger)
	dc.Process(args.Url)
}
